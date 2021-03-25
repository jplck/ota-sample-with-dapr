package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var twinTopic string = "$iothub/twin/PATCH/properties/desired/"

type Manifest struct {
	Definitions map[string]Definition `json:"devicesoftwaredefinition"`
}

type Definition struct {
	ImageName string `json:"imageName"`
	Version   string `json:"version"`
}

func main() {
	mqttHost, mqttPort, mqttPassword, mqttClientId, isOk := setupContext()

	if !isOk {
		os.Exit(0)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ssl://%s:%s", mqttHost, mqttPort))
	opts.SetClientID(mqttClientId)
	opts.SetUsername(fmt.Sprintf("%s/%s/?api-version=2018-06-30", mqttHost, mqttClientId))
	//az iot hub generate-sas-token --device-id [DEVICE_ID] --hub-name [HUB_NAME]
	opts.SetPassword(mqttPassword)

	opts.SetDefaultPublishHandler(messagePubHandler)

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	opts.SetKeepAlive(60 * 2 * time.Second)

	tlsConfig := NewTlsConfig()
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	<-c
}

func NewTlsConfig() *tls.Config {
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("IoTHubRootCA_Baltimore.pem")
	if err != nil {
		log.Fatalln(err.Error())
	}

	certPool.AppendCertsFromPEM(ca)
	return &tls.Config{
		RootCAs: certPool,
	}
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	if strings.HasPrefix(msg.Topic(), twinTopic) {
		fmt.Printf("Received device twin update: %s from\n", msg.Payload())
		ExecuteDefinitions(msg.Payload())
	}
}

func ExecuteDefinitions(payload []byte) {
	var manifest Manifest

	err := json.Unmarshal(payload, &manifest)

	if err == nil {

		for key, definition := range manifest.Definitions {
			fmt.Printf("Running preparations for definition with key %s\n", key)
			fileUrl, err := NewDefinitionFileUrl(definition.ImageName, definition.Version)

			if err != nil {
				fmt.Println(err.Error())
			} else {
				ExecKubectlWithManifest(fileUrl)
			}
		}

	} else {
		fmt.Println(err)
	}
}

func ExecKubectlWithManifest(fileUrl string) {
	//exec
	fmt.Printf("Executing kubectl with definition file '%s'\n", fileUrl)
	cmd := exec.Command("kubectl", "apply", "-f", fmt.Sprintf("'%s'", fileUrl))

	if err := cmd.Run(); err != nil {
		fmt.Println(cmd.String())
		fmt.Println(err.Error())
	}
}

func NewDefinitionFileUrl(definitionName string, definitionVersion string) (string, error) {
	fmt.Printf("Start downloading definition file %s with version %s\n", definitionName, definitionVersion)

	accName, accNameOk := os.LookupEnv("STORAGE_ACC_NAME")
	accKey, accKeyOk := os.LookupEnv("STORAGE_ACC_KEY")
	accContainerName, containerNameOk := os.LookupEnv("CONTAINER_NAME")
	if !accNameOk || !accKeyOk || !containerNameOk {
		fmt.Println("Missing ENV VAR for storage connection")
		os.Exit(1)
	}

	credentials, err := azblob.NewSharedKeyCredential(accName, accKey)

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	} else {

		sasQueryParams, err := azblob.AccountSASSignatureValues{
			Protocol:      azblob.SASProtocolHTTPS,
			ExpiryTime:    time.Now().UTC().Add(15 * time.Minute), //2 Min
			Permissions:   azblob.AccountSASPermissions{Read: true, List: true}.String(),
			Services:      azblob.AccountSASServices{Blob: true}.String(),
			ResourceTypes: azblob.AccountSASResourceTypes{Container: true, Object: true}.String(),
		}.NewSASQueryParameters(credentials)
		if err != nil {
			log.Fatal(err)
		}

		encodedSASParams := sasQueryParams.Encode()

		pipeline := azblob.NewPipeline(credentials, azblob.PipelineOptions{})
		u, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net?%s", accName, encodedSASParams))
		serviceUrl := azblob.NewServiceURL(*u, pipeline)
		containerUrl := serviceUrl.NewContainerURL(accContainerName)

		blobUrl := containerUrl.NewBlockBlobURL(fmt.Sprintf("%s.yaml", definitionName))

		return blobUrl.String(), nil
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")

	if token := client.Subscribe(fmt.Sprintf("%s%s", twinTopic, "#"), 0, nil); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func setupContext() (mqttHost string,
	mqttPort string,
	mqttPassword string,
	mqttClientId string,
	isOk bool) {

	var shouldExit = false
	mqttClientId, clientIdExists := os.LookupEnv("CLIENT_ID")
	if !clientIdExists {
		fmt.Println("Missing ENV VAR CLIENT_ID")
		shouldExit = true
	}

	mqttPort, portExists := os.LookupEnv("IOT_HUB_PORT")
	if !portExists {
		fmt.Println("Missing ENV VAR IOT_HUB_PORT")
		shouldExit = true
	}

	mqttHost, hostExists := os.LookupEnv("IOT_HUB_HOST")
	if !hostExists {
		fmt.Println("Missing ENV VAR IOT_HUB_HOST")
		shouldExit = true
	}

	mqttPassword, passwordExists := os.LookupEnv("PASSWORD")
	if !passwordExists {
		fmt.Println("Missing ENV VAR PASSWORD")
		shouldExit = true
	}
	isOk = !shouldExit
	return
}

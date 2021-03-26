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
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/*
	Definiton of IoT Hub topics for device twin and d2c topics.
*/
var twinTopic string = "$iothub/twin/PATCH/properties/desired/#"
var directMethodTopic string = "$iothub/methods/POST/#"
var directMethodResponseTopic = "$iothub/methods/res/%d/?$rid=%s"
var d2cPublishTopic = "devices/%s/messages/events/"

var userNameTemplate = "%s/%s/?api-version=2018-06-30"
var brokerHostTemplate = "ssl://%s:%s"

var brokerCACert = "certs/IoTHubRootCA_Baltimore.pem"

type Manifest struct {
	Definitions map[string]Definition `json:"devicesoftwaredefinition"`
}

type Definition struct {
	ImageName string `json:"imageName"`
	Version   string `json:"version"`
}

type SecurePackageDownloadTokenRequest struct {
	PackageName string `json:"packageName"`
	DeviceID    string `json:"deviceId"`
}

type SecurePackageDownloadTokenResponse struct {
	PackageName string `json:"packageName"`
	DeviceID    string `json:"deviceId"`
	DlToken     string `json:"dltoken"`
}

type Context struct {
	Host     string
	Port     string
	ClientID string
}

func main() {
	context, contextValid := setupContext()

	if !contextValid {
		panic("Invalid context due to missing env variables.")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf(brokerHostTemplate, context.Host, context.Port))
	opts.SetClientID(context.ClientID)
	opts.SetUsername(fmt.Sprintf(userNameTemplate, context.Host, context.ClientID))

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	opts.SetKeepAlive(240 * time.Minute)

	tlsConfig := newTlsConfig()
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	<-c
}

func newTlsConfig() *tls.Config {
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(brokerCACert)
	if err != nil {
		log.Fatalln(err.Error())
	}

	certPool.AppendCertsFromPEM(ca)

	context, _ := setupContext()

	certFile := fmt.Sprintf("certs/%s-public.pem", context.ClientID)
	keyFile := fmt.Sprintf("certs/%s-private.pem", context.ClientID)

	clientKeyPair, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		log.Fatalln(err.Error())
	}

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientKeyPair},
	}
}

func execKubectlWithManifest(fileUrl string) {
	//exec
	log.Printf("Executing kubectl with definition file '%s'\n", fileUrl)
	cmd := exec.Command("kubectl", "apply", "-f", fmt.Sprintf("'%s'", fileUrl))

	if err := cmd.Run(); err != nil {
		log.Println(err.Error())
	}
}

var directMethodHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received direct method call with %s\n", msg.Payload())

	parsedUrl, err := url.Parse(msg.Topic())

	if err != nil {
		log.Fatal(err.Error())
	}

	queryParts, err := url.ParseQuery(parsedUrl.RawQuery)

	if err != nil {
		log.Fatal(err.Error())
	}

	var dlCreds SecurePackageDownloadTokenResponse
	unmarshalErr := json.Unmarshal(msg.Payload(), &dlCreds)

	status := 1
	if unmarshalErr != nil {
		log.Println(unmarshalErr.Error())
		status = 0
	}

	reqID := queryParts["$rid"]

	/*
		Reply to direct method request with empty response
		Response topic needs the request ID from the incoming message and a status as int (0,1?)
		"$iothub/methods/res/%d/?$rid=%s"
	*/

	respTopic := fmt.Sprintf(directMethodResponseTopic, 1, reqID[0])

	if token := client.Publish(respTopic, byte(status), false, ""); token.Wait() && token.Error() != nil {
		log.Printf("Unable to publish to reply to direct method call on topic %s and error: %s", respTopic, token.Error())
	}

}

var deviceTwinUpdateHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received device twin update: %s with\n", msg.Payload())

	var manifest Manifest
	err := json.Unmarshal(msg.Payload(), &manifest)

	if err != nil {
		log.Println(err.Error())
		return
	}

	context, _ := setupContext()

	d2cTopic := fmt.Sprintf(d2cPublishTopic, context.ClientID)

	for key, definition := range manifest.Definitions {

		fmt.Printf("Received definition: %s", definition)

		payload := SecurePackageDownloadTokenRequest{
			PackageName: key,
			DeviceID:    context.ClientID,
		}

		json, _ := json.Marshal(payload)

		if token := client.Publish(d2cTopic, 1, false, json); token.Wait() && token.Error() != nil {
			log.Printf("Unable to publish token request on topic %s and error: %s", d2cTopic, token.Error())
		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")

	//Subscribe to changes in device twin
	if token := client.Subscribe(twinTopic, 0, deviceTwinUpdateHandler); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	//Subscribe to direct method calls
	if token := client.Subscribe(directMethodTopic, 0, directMethodHandler); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func setupContext() (Context, bool) {
	clientID, clientIdExists := os.LookupEnv("CLIENT_ID")
	port, portExists := os.LookupEnv("IOT_HUB_PORT")
	host, hostExists := os.LookupEnv("IOT_HUB_HOST")

	context := Context{
		Host:     host,
		Port:     port,
		ClientID: clientID,
	}
	contextValid := clientIdExists && portExists && hostExists
	return context, contextValid
}

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var userNameTemplate = "%s/%s/?api-version=2018-06-30"
var brokerHostTemplate = "ssl://%s:%s"

var ctx Context

func main() {
	ctx = Context{}.Init()

	if !ctx.Valid() {
		panic("Invalid context due to missing env variables.")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf(brokerHostTemplate, ctx.Host, ctx.Port))
	opts.SetClientID(ctx.ClientID)
	opts.SetUsername(fmt.Sprintf(userNameTemplate, ctx.Host, ctx.ClientID))

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

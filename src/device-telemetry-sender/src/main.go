package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/google/uuid"
)

var sub = &common.Subscription{
	PubsubName: "credentialrequest-pubsub",
	Topic:      "credentialrequests",
	Route:      "/command",
}

const (
	pubTopic            = "telemetry"
	telemetryPubSubName = "telemetry-pubsub"
	port                = 8000
)

func main() {

	signalChan := make(chan bool)
	simStatus := true
	e := make(chan os.Signal, 1)
	signal.Notify(e, os.Interrupt, syscall.SIGTERM)

	/*
		Run simulation always on startup. The signal channel is used to
		stop the simulation if the associated command comes in from
		remote
	*/
	go runSim(signalChan)

	/*
		Run dapr service in a goroutine. This is required to be able to
		synchronize the channels between the two components (sim, dapr)
	*/
	go func(signalChan chan bool) {
		s := daprd.NewService(fmt.Sprintf(":%d", port))

		//Receive commands and signal the currently running goroutine if stop is required
		commandHandler := func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
			command := parseCommand(e, signalChan)
			if !command {
				//STOP
				signalChan <- false
				simStatus = false
			} else {
				//START NEW SIM
				if !simStatus {
					go runSim(signalChan)
				}
			}
			return false, nil
		}

		if err := s.AddTopicEventHandler(sub, commandHandler); err != nil {
			log.Fatalf("error adding topic subscription: %v", err)
		}

		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error listenning: %v", err)
		}
	}(signalChan)

	<-e
}

func runSim(stopSignal chan bool) {
	log.Println("Start publishing telemetry.")

	daprClient, err := dapr.NewClient()
	if err != nil {
		log.Fatal("Unable to setup dapr client.")
	}

	defer daprClient.Close()

	ctx := context.Background()

	for {
		select {
		case <-stopSignal:
			break
		default:
			log.Println("Publishing telemetry")
			pubTelemetry(daprClient, ctx)
			time.Sleep(time.Second)
		}
	}
}

func pubTelemetry(daprClient dapr.Client, ctx context.Context) {
	err := daprClient.PublishEvent(
		ctx,
		telemetryPubSubName,
		pubTopic,
		[]byte(fmt.Sprintf("telemetry payload: %s", uuid.New().String())))

	if err != nil {
		log.Printf("Unable to publish: %v", err.Error())
	}
}

func parseCommand(e *common.TopicEvent, signal chan bool) bool {
	log.Printf("event - PubsubName: %s, Topic: %s, ID: %s, Data: %s", e.PubsubName, e.Topic, e.ID, e.Data)

	mode := false

	return mode
}

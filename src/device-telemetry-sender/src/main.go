package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
)

var (
	sub = &common.Subscription{
		PubsubName: "credentialrequest-pubsub",
		Topic:      "credentialrequests",
		Route:      "/command",
	}

	simActive = true
	modeChan  chan bool
)

const (
	pubTopic            = "telemetry"
	telemetryPubSubName = "telemetry-pubsub"
	port                = 8000
)

func init() {
	modeChan = make(chan bool)
}

func main() {

	defer close(modeChan)
	go publishTelemetry()

	s := daprd.NewService(fmt.Sprintf(":%d", port))

	if err := s.AddTopicEventHandler(sub, commandEventHandler); err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}

func publishTelemetry() {
	log.Println("Start publishing telemetry.")

	daprClient, err := dapr.NewClient()
	if err != nil {
		log.Fatal("Unable to setup dapr client.")
	}

	defer daprClient.Close()

	go func() {
		ctx := context.Background()
		i := 0
		for {

			if !simActive {
				log.Println("Exiting sim loop.")
				break
			}

			err := daprClient.PublishEvent(
				ctx,
				telemetryPubSubName,
				pubTopic,
				[]byte(fmt.Sprintf("telemetry payload: %d", i)))

			if err != nil {
				log.Printf("Unable to publish: %v", err.Error())
			}

			time.Sleep(time.Second)
			i++
		}
	}()

	simActive = <-modeChan
}

func commandEventHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	log.Printf("event - PubsubName: %s, Topic: %s, ID: %s, Data: %s", e.PubsubName, e.Topic, e.ID, e.Data)

	mode := false

	if mode && !simActive {
		publishTelemetry()
	} else if !mode && simActive {
		simActive = false
	}

	return false, nil
}

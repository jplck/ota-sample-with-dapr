package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/google/uuid"
)

var (
	sub = &common.Subscription{
		PubsubName: "credentialrequest-pubsub",
		Topic:      "credentialrequests",
		Route:      "/command",
	}
	simStatus  = false
	sig        chan bool
	daprClient dapr.Client
)

const (
	pubTopic            = "telemetry"
	telemetryPubSubName = "telemetry-pubsub"
	port                = 8000
)

func init() {
	c, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("Unable to create new dapr client due to err: %v", err.Error())
	}
	daprClient = c
}

func main() {

	defer daprClient.Close()

	go func() {

		time.Sleep(time.Second * 10)

		triggerCmd("start")

		time.Sleep(time.Second * 5)

		triggerCmd("stop")

		time.Sleep(time.Second * 5)

		triggerCmd("start")
	}()

	s := daprd.NewService(fmt.Sprintf(":%d", port))

	if err := s.AddTopicEventHandler(sub, commandHandler); err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	if err := s.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error listenning: %v", err)
	}
}

func commandHandler(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	var cmdReq CommandRequest
	if err := json.Unmarshal([]byte(fmt.Sprint(e.Data)), &cmdReq); err != nil {
		log.Fatalf("Unable to parse incoming command due to error: %v", err.Error())
	}

	if cmdReq.Command == "stop" {
		//STOP
		close(sig)
		simStatus = false
	} else {
		//START NEW SIM
		if !simStatus {
			sig = make(chan bool)
			go runSim(sig)
		}
	}
	return false, nil
}

func triggerCmd(cmd string) {
	cmdReq := CommandRequest{
		Command: cmd,
	}

	json, _ := json.Marshal(cmdReq)

	ctx := context.Background()

	if err := daprClient.PublishEvent(ctx, sub.PubsubName, sub.Topic, json); err != nil {
		log.Printf("Unable to trigger sim from self due to error: %v", err.Error())
	}
}

func runSim(sig chan bool) {
	log.Println("Start publishing telemetry.")

	for {
		select {
		case <-sig:
			log.Println("Stopping simulation.")
			return
		default:
			log.Println("Publishing telemetry")
			pubTelemetry(daprClient)
			time.Sleep(time.Second)
		}
	}
}

func pubTelemetry(daprClient dapr.Client) {
	ctx := context.Background()
	err := daprClient.PublishEvent(
		ctx,
		telemetryPubSubName,
		pubTopic,
		[]byte(fmt.Sprintf("telemetry payload: %s", uuid.New().String())))

	if err != nil {
		log.Printf("Unable to publish: %v", err.Error())
	}
}

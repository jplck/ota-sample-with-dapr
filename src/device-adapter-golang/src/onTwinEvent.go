package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var deviceTwinUpdateHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received device twin update: %s with\n", msg.Payload())

	var manifest Manifest
	err := json.Unmarshal(msg.Payload(), &manifest)

	if err != nil {
		log.Println(err.Error())
		return
	}

	d2cTopic := fmt.Sprintf(d2cPublishTopic, ctx.ClientID)

	for key, definition := range manifest.Definitions {

		fmt.Printf("Received definition: %s", definition)

		payload := SecurePackageDownloadTokenRequest{
			PackageName: key,
			DeviceID:    ctx.ClientID,
		}

		cloudEvent := CloudEvent{
			ID:          uuid.New().String(),
			Source:      fmt.Sprintf("/device/%s/credentials/request", ctx.ClientID),
			SpecVersion: "1.0",
			Data:        payload,
			Type:        "credentialrequest",
			Time:        time.Now(),
		}

		json, _ := json.Marshal(cloudEvent)

		if token := client.Publish(d2cTopic, 1, false, json); token.Wait() && token.Error() != nil {
			log.Printf("Unable to publish token request on topic %s and error: %s", d2cTopic, token.Error())
		}
	}
}

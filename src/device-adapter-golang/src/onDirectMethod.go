package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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
	} else {

		/*
			This is currently a simplification. The idea is to use the received DLToken and the URL to download the package
			with updated files from for example a storage.
			The currently implementation uses the Url directly in the kubectl command. The url is a SAS URL for a file on a
			blob storage.

			{
				"url": "[url to blob file on blob storage]",
				"packageName": "[PackageName]"
				"deviceId": "[ClientID]",
				"dlToken": "[Token]"
			}
		*/
		status = execKubectlWithManifest(dlCreds.Url)
	}

	reqID := queryParts["$rid"]

	/*
		Reply to direct method request with empty response
		Response topic needs the request ID from the incoming message and a status as int (0,1?)
		"$iothub/methods/res/%d/?$rid=%s"
	*/
	respTopic := fmt.Sprintf(directMethodResponseTopic, status, reqID[0])

	if token := client.Publish(respTopic, 1, false, ""); token.Wait() && token.Error() != nil {
		log.Printf("Unable to publish to reply to direct method call on topic %s and error: %s", respTopic, token.Error())
	}

}

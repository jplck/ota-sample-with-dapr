package main

/*
	Definiton of IoT Hub topics for device twin and d2c topics.
*/
var twinTopic = "$iothub/twin/PATCH/properties/desired/#"
var directMethodTopic = "$iothub/methods/POST/#"
var directMethodResponseTopic = "$iothub/methods/res/%d/?$rid=%s"
var d2cPublishTopic = "devices/%s/messages/events/"

var mqtt = require('mqtt')
var fs = require('fs')
var path = require('path')
require("dotenv").config();

var _caList = fs.readFileSync(path.join(__dirname, '/IoTHubRootCA_Baltimore.pem'))

var _port = parseInt(process.env.IOT_HUB_PORT)
var _host = process.env.IOT_HUB_HOST
var _clientId = process.env.CLIENT_ID
var _username = `${_host}/${_clientId}/?api-version=2018-06-30`

var _desiredPropsTopic = "$iothub/twin/PATCH/properties/desired/#"

//az iot hub generate-sas-token --device-id device1 --hub-name daprhub1
var _password = process.env.PASSWORD

var options = {
    port: _port,
    host: _host,
    rejectUnauthorized: true,
    ca: _caList,
    protocol: 'mqtts',
    username: _username,
    password: _password,
    clientId: _clientId
}

var client = mqtt.connect(options)

client.on("connect", () => {
    console.log("connection done!")
    client.subscribe(_desiredPropsTopic, (err) => {
        if (!err) {
            console.log("subscribed to changes on device twin properties.")
        }
    })
    client.on("message", (topic, message) => {
        if (topic.startsWith(_desiredPropsTopic.replace('#', ''))) {

            const manifest = JSON.parse(message.toString())
            const definitions = manifest.devicesoftwaredefinition

            console.log(definitions)

            for (defIdx in definitions) {
                const def = definitions[defIdx]
                const defUri = def.imageName

                //exec
                console.log(`kubectl apply -f ${defUri}`)
            }
        }
    })
})
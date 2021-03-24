var mqtt = require('mqtt')
var fs = require('fs')
var path = require('path')
var _caList = fs.readFileSync(path.join(__dirname, '/IoTHubRootCA_Baltimore.pem'))

var _port = 8883
var _host = "daprhub1.azure-devices.net"
var _clientId = "device1"
var _username = `${_host}/${_clientId}/?api-version=2018-06-30`

var _desiredPropsTopic = "$iothub/twin/PATCH/properties/desired/#"

//az iot hub generate-sas-token --device-id device1 --hub-name daprhub1
var _password = "SharedAccessSignature sr=daprhub1.azure-devices.net%2Fdevices%2Fdevice1&sig=VSCyFJbInVh3O%2BkjRY3s9FvRNnWSdsLWeITRUY%2BgYkA%3D&se=1616583940"

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
            console.log(JSON.parse(message.toString()))
        }
    })
})
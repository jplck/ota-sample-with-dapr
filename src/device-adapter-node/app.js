var mqtt = require('mqtt')
var fs = require('fs')
var path = require('path')
var { exec } = require('child_process');
require("dotenv").config();

const { StorageSharedKeyCredential, 
        BlobServiceClient, 
        generateBlobSASQueryParameters, 
        BlobSASPermissions } = require("@azure/storage-blob");

var _caList = fs.readFileSync(path.join(__dirname, '/IoTHubRootCA_Baltimore.pem'))

var _port = parseInt(process.env.IOT_HUB_PORT)
var _host = process.env.IOT_HUB_HOST
var _clientId = process.env.CLIENT_ID
var _username = `${_host}/${_clientId}/?api-version=2018-06-30`

var _desiredPropsTopic = "$iothub/twin/PATCH/properties/desired/#"

//az iot hub generate-sas-token --device-id [DEVICE_ID] --hub-name [HUB_NAME]
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
                
                var defUrl = GetDeploymentManifestUrl(def.imageName)

                ExecuteKubectlApply(defUrl)
            }
        }
    })
})

function GetDeploymentManifestUrl(manifestName) {
    console.log(manifestName)

    const sharedKeyCredential = new StorageSharedKeyCredential(
        process.env.MANIFEST_STORAGE_ACCOUNT, 
        process.env.MANIFEST_STORAGE_ACCESS_KEY
    );
    
    const blobServiceClient = new BlobServiceClient(
        `https://${process.env.MANIFEST_STORAGE_ACCOUNT}.blob.core.windows.net`,
        sharedKeyCredential
    );
    
    const containerName = process.env.MANIFEST_STORAGE_CONTAINER_NAME

    const containerClient = blobServiceClient.getContainerClient(containerName);
    const blockBlobClient = containerClient.getBlockBlobClient(manifestName);

    const sasToken = generateBlobSASQueryParameters({
        containerName: containerName,
        blobName: manifestName,
        expiresOn: new Date(new Date().valueOf() + 86400),
        permissions: BlobSASPermissions.parse("racwd")
    }, sharedKeyCredential);
      
    const sasUrl = `${blockBlobClient.url}?${sasToken}`;

    return sasUrl
}

function ExecuteKubectlApply(defUrl) {
    //exec
    exec(`sudo kubectl apply -f ${defUrl}`, (error, stdout, stderr) => {
        console.log(`executed: kubectl apply -f ${defUrl}`)
        if (error) {
            console.log(error)
        }
        else if (stdout) {
            console.log(stdout)
        }
        else {
            console.log(stderr)
        }
    })
}
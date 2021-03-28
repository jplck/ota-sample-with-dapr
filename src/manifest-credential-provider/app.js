const express = require('express')
const SASBlobUrl = require('./blob')
const DirectMethod = require('./directMethod')
require("dotenv").config();
require('isomorphic-fetch');

const app = express()
app.use(express.json({ type: 'application/*+json' }));

const daprPort = process.env.DAPR_HTTP_PORT || 3500;
const port = 8000

const secretsUrl = `http://localhost:${daprPort}/v1.0/secrets`;
const secretStoreName = 'secretstore'

async function getSecret(secretName) {
    return await fetch(`${secretsUrl}/${secretStoreName}/${secretName}`)
        .then(async (response) => {
            if (!response.ok) {
                throw "Could not get secret";
            }
            return (await response.json())[secretName];
        })
}

app.get('/dapr/subscribe', (req, res) => {
    res.json(
        [
            {
                "pubsubname": "credentialrequest-pubsub",
                "topic": "credentialrequests",
                "route": "credentialrequest"
            }
        ]
    )
})

/*
    {
        "packageName": "[packageName]",
        "deviceId": "[deviceId]"
    }
*/

app.post("/credentialrequest", async (req, res) => {
    console.log(req.body)

    var packageName = req.body.data.packageName
    var deviceId = req.body.data.deviceId

    try {
        const accountName = await getSecret("PACKAGESTORAGEACCOUNT")
        const accountKey = await getSecret("PACKAGESTORAGEACCOUNTKEY")
        const containerName = await getSecret("PACKAGESTORAGECONTAINERNAME")
        const iotHubConnectionString = await getSecret("IotHubConnectionString")

        console.log(accountKey)

        const sasURL = SASBlobUrl(`${packageName}.yaml`, accountName, accountKey, containerName)

        DirectMethod(iotHubConnectionString, deviceId, "sendcredentials", {
            url: sasURL,
            deviceId: deviceId,
            dlToken: "123456789",
            packageName: packageName
        })

        res.sendStatus(200)

    } catch (ex) {
        console.log(ex)
        res.sendStatus(500)
    }
})

app.listen(port, () => console.log(`Node App listening on port ${port}!`));
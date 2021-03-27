const express = require('express')
const SASBlobUrl = require('./blob')
const DirectMethod = require('./directMethod')
require("dotenv").config();

const app = express()
app.use(express.json({ type: 'application/*+json' }));

const port = 8000

const containerName = process.env.MANIFEST_STORAGE_CONTAINER_NAME
const accoutName = process.env.MANIFEST_STORAGE_ACCOUNT
const accountKey = process.env.MANIFEST_STORAGE_ACCESS_KEY


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

app.post("/credentialrequest", (req, res) => {
    console.log(req.body)

    var packageName = req.body.data.packageName
    var deviceId = req.body.data.deviceId

    const sasURL = SASBlobUrl(`${packageName}.yaml`, accoutName, accountKey, containerName)

    DirectMethod(deviceId, "sendcredentials", {
        url: sasURL,
        deviceId: deviceId,
        dlToken: "123456789",
        packageName: packageName
    })

    res.sendStatus(200)
})

app.listen(port, () => console.log(`Node App listening on port ${port}!`));
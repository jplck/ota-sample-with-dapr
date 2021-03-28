const express = require('express')
const { generateSASUrl } = require('./blob')
const { callDirectMethod } = require('./directMethod')
const { getSecrets } = require('./helpers')
require("dotenv").config();
require('isomorphic-fetch');

const app = express()
app.use(express.json({ type: 'application/*+json' }));

const port = 8000

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
        ++ Required CloudEvent fields
        "packageName": "[packageName]",
        "deviceId": "[deviceId]"
    }
*/

app.post("/credentialrequest", async (req, res) => {
    console.log(req.body)

    var packageName = req.body.data.packageName
    var deviceId = req.body.data.deviceId

    try {
        const secrets = await getSecrets()

        const sasURL = generateSASUrl(
            `${packageName}.yaml`, 
            secrets["PACKAGESTORAGEACCOUNT"], 
            secrets["PACKAGESTORAGEACCOUNTKEY"],
            secrets["PACKAGESTORAGECONTAINERNAME"]
        )

        callDirectMethod(
            secrets["IotHubConnectionString"], 
            deviceId, 
            "sendcredentials", 
            {
                url: sasURL,
                deviceId: deviceId,
                dlToken: "123456789",
                packageName: packageName
            }
        )

        res.sendStatus(200)

    } catch (ex) {
        console.log(ex)
        res.sendStatus(500)
    }
})

app.listen(port, () => console.log(`Package credentrial provider listening on port ${port}!`));
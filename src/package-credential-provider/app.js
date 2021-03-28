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

    console.log("Starting credential provider...")

    try {

        var packageName = req.body.data.packageName
        var deviceId = req.body.data.deviceId

        if (!packageName || !deviceId) { throw "PackageName or DeviceId cannot be empty."}

        /*
            Fetching all secrets is not a good approach but 
            sufficient for this demo.
        */
        const secrets = await getSecrets()

        /*
            Generate a SAS Url to a "package" file on an Azure Blob Storage.
            The file URL can than be used for a limited time on the device to
            either downloads its contents or apply it to a deployment.
        */
        const sasURL = generateSASUrl(
            `${packageName}.yaml`, 
            secrets["PACKAGESTORAGEACCOUNT"], 
            secrets["PACKAGESTORAGEACCOUNTKEY"],
            secrets["PACKAGESTORAGECONTAINERNAME"],
            new Date(new Date().valueOf() + 2 * 60) //2 Min
        )

        /*
            Calls a direct method via an Azure Iot Hub on the device 
            that triggered this request. The direct method call contains
            the SAS Url and additional information the device can use to 
            download and use the update package
        */
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
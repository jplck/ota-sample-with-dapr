var Client = require('azure-iothub').Client;

//https://github.com/Azure/azure-iot-sdk-node/blob/master/service/samples/javascript/device_method.js

module.exports = function(connectionString, deviceId, methodName, payload, timeout = 15) {
    var methodParams = {
        methodName: methodName,
        payload: payload,
        responseTimeoutInSeconds: timeout
    };

    var client = Client.fromConnectionString(connectionString);

    client.invokeDeviceMethod(deviceId, methodParams, function (err, result) {
        if (err) {
            console.error('Failed to invoke method \'' + methodParams.methodName + '\': ' + err.message);
        } else {
            console.log(methodParams.methodName + ' on ' + deviceId + ':');
            console.log(JSON.stringify(result, null, 2));
        }
    });
}
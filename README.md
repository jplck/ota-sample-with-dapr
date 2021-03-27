# Device update flow with DAPR

Recently I was challenged quite frequently with the question on how to handle updates of functionalty on memory and/or compute constrained IoT devices at scale. Usually this does not mean embedded but more the "raspberry pi" style of device.

To make matters even more interesting, the challenge often includes requirements of having edge capabilities like container orchestrations, hight availability and so on.

## Criterias

The criterias I have used as baseline requirements for this project were the following:
1. Leverage DAPR in the Cloud and on the device
2. Provide mean of container orchestration on the device
3. Provide a light way mechanism for cloud to device and device to cloud communication
4. Provide a secure mechanism for downloading firmware updates
5. Do everything at scale for potentially millions of devices
6. Consider being cloud agnostic

## Architecture
The architecture is split into a device and a cloud part. The cloud portion is build upon Azure components like an IoT Hub and an Azure Kubernetes Service. The device part is build on top of a K3S and native components, running direcly on the OS without any orechstrations or containerizations.

### Tech Stack
- Azure IoT Hub (https://azure.microsoft.com/en-us/services/iot-hub/)
- Azure Kubernetes Service (https://azure.microsoft.com/en-us/services/kubernetes-service/)
- Azure Blob Storage (https://azure.microsoft.com/en-us/services/storage/blobs/)
- K3S (https://k3s.io/)


![Architecture](/docs/images/architecture.png)

### Beakdown - Step by step

Some initial things to know:
- DeviceTwin: The Device Twin is part of the device definition in the Azure Iot Hub. It can be updated and synched with the corresponding device. Each device as one device twin.

1. Initial part of the flow is the update of something called IoT Device Configuration. The device config is a manifest, that defines specific properties for a group of devices. The mechanism for that is something proprietary to the IoT Hub but can potentially be build by using custom code as well. What it does, is to apply the properties defined in the manifest to all devices (device twins) that comply with that configration.
The device twin update triggers a MQTT message from the IoT Hub to the device.


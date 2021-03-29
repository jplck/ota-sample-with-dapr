const { StorageSharedKeyCredential, 
        BlobServiceClient, 
        generateBlobSASQueryParameters, 
        BlobSASPermissions } = require("@azure/storage-blob");

const generateSASUrl = (manifestName, accountName, accountKey, containerName, expiresIn) => {
    const sharedKeyCredential = new StorageSharedKeyCredential(
        accountName, 
        accountKey
    );
    
    const blobServiceClient = new BlobServiceClient(
        `https://${accountName}.blob.core.windows.net`,
        sharedKeyCredential
    );

    const containerClient = blobServiceClient.getContainerClient(containerName);
    const blockBlobClient = containerClient.getBlockBlobClient(manifestName);

    var expiryDate = new Date()
    expiryDate.setSeconds(expiryDate.getSeconds() + expiresIn)

    const sasToken = generateBlobSASQueryParameters({
        containerName: containerName,
        blobName: manifestName,
        startsOn: new Date(),
        expiresOn: expiryDate,
        permissions: BlobSASPermissions.parse("racwd")
    }, sharedKeyCredential);
      
    const sasUrl = `${blockBlobClient.url}?${sasToken}`;

    return sasUrl
}

module.exports = generateSASUrl
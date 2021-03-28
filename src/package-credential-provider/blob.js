const { StorageSharedKeyCredential, 
        BlobServiceClient, 
        generateBlobSASQueryParameters, 
        BlobSASPermissions } = require("@azure/storage-blob");

function generateSASUrl (manifestName, accountName, accountKey, containerName, expiresOn) {
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

    const sasToken = generateBlobSASQueryParameters({
        containerName: containerName,
        blobName: manifestName,
        expiresOn: expiresOn,
        permissions: BlobSASPermissions.parse("racwd")
    }, sharedKeyCredential);
      
    const sasUrl = `${blockBlobClient.url}?${sasToken}`;

    return sasUrl
}

module.exports = generateSASUrl
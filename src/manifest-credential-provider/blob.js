const { StorageSharedKeyCredential, 
    BlobServiceClient, 
    generateBlobSASQueryParameters, 
    BlobSASPermissions } = require("@azure/storage-blob");

module.exports = function (manifestName, accountName, accountKey, containerName) {
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
        expiresOn: new Date(new Date().valueOf() + 86400),
        permissions: BlobSASPermissions.parse("racwd")
    }, sharedKeyCredential);
      
    const sasUrl = `${blockBlobClient.url}?${sasToken}`;

    return sasUrl
}
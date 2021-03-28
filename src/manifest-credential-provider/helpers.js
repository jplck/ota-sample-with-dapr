const daprPort = process.env.DAPR_HTTP_PORT || 3500;
const secretsUrl = `http://localhost:${daprPort}/v1.0/secrets`;
const secretStoreName = 'secretstore'

async function getSecrets() {
    return await fetch(`${secretsUrl}/${secretStoreName}/bulk`)
        .then(async (response) => {
            if (!response.ok) {
                throw "Could not get secret";
            }
            return (await response.json());
        })
}

module.exports = getSecrets
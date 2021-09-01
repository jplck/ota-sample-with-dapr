# Note: You can use any Debian/Ubuntu based image you want. 
FROM mcr.microsoft.com/vscode/devcontainers/base:0-buster

# Options
ARG INSTALL_ZSH="true"
ARG UPGRADE_PACKAGES="false"
ARG USE_MOBY="true"
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
# Install needed packages and setup non-root user. Use a separate RUN statement to add your own dependencies.
COPY setup-scripts/*.sh /tmp/setup-scripts/
RUN apt-get update && /bin/bash /tmp/setup-scripts/common-debian.sh "${INSTALL_ZSH}" "${USERNAME}" "${USER_UID}" "${USER_GID}" "${UPGRADE_PACKAGES}" "true" "true" \
    && /bin/bash /tmp/setup-scripts/docker-in-docker-debian.sh "true" "${USERNAME}" "${USE_MOBY}" \ 
    && /bin/bash /tmp/setup-scripts/kubectl-helm-debian.sh "latest" "latest" "latest" \
    && apt-get autoremove -y && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/setup-scripts/

COPY setup-scripts/k3d-setup.sh /home

RUN wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | TAG=v4.4.8 bash \
    && wget -q https://raw.githubusercontent.com/dapr/cli/master/install/install.sh -O - | /bin/bash \
    && wget -q https://golang.org/dl/go1.17.linux-amd64.tar.gz \
    && tar -C /usr/local/ -xzf go1.17.linux-amd64.tar.gz \
    && apt-get install -y apt-transport-https && \
       apt-get update && \
       apt-get install -y dotnet-sdk-5.0 \
    && apt-get install -y npm

ENV PATH "${PATH}:/usr/local/go/bin"

RUN go get -v golang.org/x/tools/gopls && go get -v github.com/rogpeppe/godef

VOLUME [ "/var/lib/docker" ]

# Setting the ENTRYPOINT to docker-init.sh will start up the Docker Engine 
# inside the container "overrideCommand": false is set in devcontainer.json. 
# The script will also execute CMD if you need to alter startup behaviors.
ENTRYPOINT [ "/usr/local/share/docker-init.sh" ]

USER $USER_UID
CMD ["/bin/bash", "/home/k3d-setup.sh"]


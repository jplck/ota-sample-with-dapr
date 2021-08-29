#!/bin/bash
s=1
while [ $s -ne 0 ]
do
      docker version > /dev/null 2>&1
      s=$?
      echo "."
      sleep 1
done
k3d cluster create devicecluster --api-port 127.0.0.1:8080 -p 80:80@loadbalancer -p 443:443@loadbalancer --k3s-server-arg "--no-deploy=traefik" --volume /etc/resolv.conf:/etc/resolv.conf

echo "Switching context..."

k3d kubeconfig merge devicecluster --kubeconfig-merge-default --kubeconfig-switch-context

kubectl cluster-info

dapr init --kubernetes --wait

sleep infinity
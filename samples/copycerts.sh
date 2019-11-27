#!/bin/bash

set -ex

cp /var/lib/minikube/certs/etcd/ca.crt /tmp
cp /var/lib/minikube/certs/apiserver-etcd-client.crt /tmp/client.crt
cp /var/lib/minikube/certs/apiserver-etcd-client.key /tmp/client.key


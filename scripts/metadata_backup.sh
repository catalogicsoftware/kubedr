#!/bin/bash

set -ex

SCRIPT_PATH=`which $0`
export SCRIPT_DIR=`dirname $SCRIPT_PATH`

### Input Data #####

CERTS_DIR=/var/lib/minikube/certs
ETCD_ENDPOINT=https://127.0.0.1:2379

export AWS_ACCESS_KEY_ID=minio
export AWS_SECRET_ACCESS_KEY=minio123
export RESTIC_PASSWORD=testpass
AWS_URL=http://10.107.63.254:9000
AWS_BUCKET=testbucket

### Input Data - end #####

TARGET_ROOTDIR=`pwd`/data
TARGET_CERTS_DIR=$TARGET_ROOTDIR/certs
TARGET_ETCD_DIR=$TARGET_ROOTDIR/etcd

mkdir -p $TARGET_ROOTDIR
mkdir -p $TARGET_CERTS_DIR
mkdir -p $TARGET_ETCD_DIR

cp -R $CERTS_DIR/* $TARGET_CERTS_DIR

docker run --rm -it --env ETCDCTL_API=3 -v $CERTS_DIR:/certs_dir:ro -v $TARGET_ETCD_DIR:/backup \
    --network host k8s.gcr.io/etcd:3.3.10 etcdctl --endpoints=$ETCD_ENDPOINT \
    --cacert=/certs_dir/etcd/ca.crt --cert /certs_dir/apiserver-etcd-client.crt \
    --key /certs_dir/apiserver-etcd-client.key snapshot save /backup/etcd-snapshot.db 

restic -r s3:$AWS_URL/$AWS_BUCKET --verbose backup $TARGET_ROOTDIR







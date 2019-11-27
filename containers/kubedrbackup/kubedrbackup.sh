#!/bin/bash

set -e

if [ -z "$AWS_ACCESS_KEY" ]
then
    echo "AWS_ACCESS_KEY is required."
    exit 1
fi

if [ -z "$AWS_SECRET_KEY" ]
then
    echo "AWS_SECRET_KEY is required."
    exit 1
fi

if [ -z "$RESTIC_PASSWORD" ]
then
    echo "RESTIC_PASSWORD is required."
    exit 1
fi

# Setting this after checking all the env variables containing sensitive
# information.
set -x

if [ -z "$ETCD_ENDPOINT" ]
then
    echo "ETCD_ENDPOINT is required."
    exit 1
fi

if [ -z "$ETCD_CREDS_DIR" ]
then
    echo "ETCD_CREDS_DIR is required."
    exit 1
fi

if [ -z "$ETCD_SNAP_PATH" ]
then
    echo "ETCD_SNAP_PATH is required."
    exit 1
fi

if [ -z "$RESTIC_REPO" ]
then
    echo "RESTIC_REPO is required."
    exit 1
fi

if [ -z "$BACKUP_SRC" ]
then
    echo "BACKUP_SRC is required."
    exit 1
fi

if [ -z "$KDR_POLICY_NAME" ]
then
    echo "KDR_POLICY_NAME is required."
    exit 1
fi

export ETCDCTL_API=3
etcdctl --endpoints="$ETCD_ENDPOINT" --cacert=$ETCD_CREDS_DIR/ca.crt \
    --cert=$ETCD_CREDS_DIR/client.crt --key=$ETCD_CREDS_DIR/client.key \
	--debug snapshot save $ETCD_SNAP_PATH

if [ -n "$CERTS_SRC_DIR" -a -n "$CERTS_DEST_DIR" ]
then
    mkdir -p $CERTS_DEST_DIR && cp -R $CERTS_SRC_DIR/* $CERTS_DEST_DIR
fi

export RESTIC_OUTPUT_FILE=/tmp/restic.txt
restic --json -r $RESTIC_REPO --verbose backup $BACKUP_SRC | tee $RESTIC_OUTPUT_FILE

snapshot_id=$(grep snapshot_id $RESTIC_OUTPUT_FILE | tail -1 | jq -r '.snapshot_id')

cat <<EOF | kubectl apply -f -
apiVersion: kubedr.catalogicsoftware.com/v1alpha1
kind: MetadataBackupRecord
metadata:
  name: mbr-$snapshot_id
  finalizers:
  - mbr.finalizers.kubedr.catalogicsoftware.com
spec:
  snapshotId: "$snapshot_id"
  policy: $KDR_POLICY_NAME
EOF



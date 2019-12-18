#!/bin/bash

set -x

shopt -s expand_aliases
alias k="kubectl -n kubedr-system"

# Delete all CRs
k delete metadatabackuppolicy --all
k delete metadatabackuprecord --all
k delete backuplocation --all

# Delete the namespace
kubectl delete namespace kubedr-system

# Delete CRDs
kubectl delete crd metadatabackuppolicies.kubedr.catalogicsoftware.com
kubectl delete crd metadatabackuprecords.kubedr.catalogicsoftware.com
kubectl delete crd backuplocations.kubedr.catalogicsoftware.com

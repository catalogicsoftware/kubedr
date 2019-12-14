# Kubernetes Metadata Backup

[![pipeline status](https://gitlab.ad.catalogic.us/kubernetes/kcx/badges/master/pipeline.svg)](https://gitlab.ad.catalogic.us/kubernetes/kcx/commits/master)

## Introduction

This project, code named *kubedr*, is about protecting metadata of
Kubernetes stored in *etcd*. In addition, certificates can be backed up
as well but that is optional.

**This is a very high level document that only provides broad overview
  of the project at this point. More detailed documentation will be
  provided soon. Familiarity with Kubernetes is assumed throughout
  the document.**

We plan to release an alpha version some time in *December 2019*
time frame but details are yet to be worked out.

## Rationale

Kubernetes stores all its objects in *etcd* so backing up data in
*etcd* is crucial for DR purposes. However, there are no active
projects or products in the wild that provide a seamless backup of
this important piece of data. That gives us an opportunity to enter
Kubernetes world with a simple and yet a very useful tool.

Working on a narrowly defined tool has another advantage in that we
can focus more on the mechanics of building applications for Kubernetes
(called *operators*). This knowledge will come very handy
when we introduce more involved products such as backup of
applications.

## Overview

This project implements a tool that backs up *etcd* data and
certificates to any S3 bucket. It follows the *operator* pattern that
is popular in Kubernetes world. 

An operator is basically a combination of *custom resources (CRs)*
coupled with *controllers* that manage the CRs. There would be one
controller for each CR. In addition to controllers, operators can also
contain *webhooks* that can be used to validate the data in resources
as well as to set defaults when some fields are not provided in the
input. Our operator uses webhooks for both these purposes.

For data transfer to S3, we currently use a tool called
[restic](https://restic.net) but we will be able to change specific
backup tool in a backwards compatible manner.

## Limitations

At present, the tool works only for clusters where *etcd* is
accessible. This mostly rules out Kubernetes clusters in the
cloud. We will investigate how to extend the tool to cloud providers
in a future release.

## Installation

For now, you can only install directly from the repo. A more "customer
centric" installation instructions will be provided soon. 

- Make sure you have a test cluster and you can access it using
  `kubectl`. This means that a valid *kubeconfig* is set up providing
  access to the cluster.

- Install
  [cert-manager](https://github.com/jetstack/cert-manager). This is
  required for webhooks to work. A simple way to install *cert-manager*:

```bash
$ kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml
```
- clone this repo.

- Do

```bash
$ cd <GIT_CLONE_DIR>

# Build the image that contains backup script.
$ cd kubedr/containers/kubedrbackup
$ docker build -t kubedrbackup:0.42 .

# Build the image that contains controllers and webhooks.
$ cd ../../kubedr
$ make docker-build IMG=kubedr:0.42

$ make deploy IMG=kubedr:0.42
```
- At this point, you should see a new namespace ``kubedr-system``
created and the following command should show a pod, replica set,
and a deployment:

```bash
$ kubectl -n kubedr-system get all
```

- Before backups are run, make sure that one or more master nodes has
  a label identifying them (if not already present). Some clusters are
  set up with *node-role.kubernetes.io/master* label on the master
  nodes. If this is the case, nothing more needs to be done. If this
  label is not present, create it. If it is not possible to use the
  above name for label, choose any name and pass that name in policy
  options. 

## Custom Resources

There are 3 custom resources currently supported. Every resource in
Kubernetes is uniquely identified by a *Group*, *Version*, and
*Kind*. This allows introducing new versions of resources in a
seamless manner. 

We will use the following values for *Group* and *Version*.  

- Group: *kubedr.catalogicsoftware.com*
- Version: *v1alpha1*

The following sections will list each resource *Kind* and briefly
explain their schema. Note that all CRs should be created in the name
space *kubedr-system*. 

### BackupLocation

This resource represents an S3 bucket that will be used as destination
for backup.

A sample resource:

```yaml
apiVersion: kubedr.catalogicsoftware.com/v1alpha1
kind: BackupLocation
metadata:
  name: local-minio
spec:
  url: http://10.96.57.3:9000
  bucketName: testbucket
  credentials: minio-creds
```

<dl>
  <dt>name</dt>
  <dd>Logical name of the resource.</dd>

  <dt>url</dt>
  <dd>S3 end point</dd>

  <dt>bucketName</dt>
  <dd>Name of the S3 bucket. It will be created if it doesn't
      exist.</dd> 

  <dt>credentials</dt>
  <dd><p>Name of the Kubernetes "secret" resource containing
  credentials.</p>

  The secret should contain three pieces of information - S3 access
  key, S3 secret key, and a password to be used to encrypt
  backups. Here is one way to create such a secret:

```bash
$ echo -n 'sample_access_key' > access_key
$ echo -n 'sample_secret_key' > secret_key
$ echo -n 'sample_restic_repo_password' > restic_repo_password

$ kubectl -n kubedr-system create secret generic minio-creds \
        --from-file=access_key --from-file=secret_key \
        --from-file restic_repo_password 
```
  </dd>

</dl>

### MetadataBackupPolicy

This resource defines a backup policy. Here is a sample policy
resource: 

```yaml
apiVersion: kubedr.catalogicsoftware.com/v1alpha1
kind: MetadataBackupPolicy
metadata:
  name: test-backup
spec:
  # Name of the "BackupLocation" resource where backup will go.
  destination: local-minio

  certsDir: /var/lib/minikube/certs

  etcdEndpoint: https://127.0.0.1:2379
  etcdCreds: etcd-creds # secret

  schedule: "*/10 * * * *"

  retainNumBackups: 1
```

<dl>
  <dt>name</dt>
  <dd>Logical name of the resource.</dd>

  <dt>destination</dt>
  <dd>Name of *BackupLocation* resource where backups should be
        stored.</dd> 

  <dt>certsDir</dt>
  <dd>Directory containing Kubernetes certificates. Optional.
        If given, contents of entire directory will be backed up.</dd>

  <dt>etcdEndpoint</dt>
  <dd>Optional. In most cases, the default value of
        "https://127.0.0.1:2379" would work.</dd>

  <dt>etcdCreds</dt>
  <dd><p>Optional. Name of the Kubernetes "secret" resource containing
  etcd credentials. If the name "etcd-creds" is used for the secret,
  there is no need to include this field.</p>

  The secret should contain three pieces of information that are
  required to connect to *etcd*:

  - ca.crt
  - client.crt
  - client.key

  One can find this information by looking at the command line
  arguments passed to *kube-apiserver* process. Once these three files
  are found (note that files must be named exactly as shown above,
  make copies if required), the secret can be created as follows: 

```bash
$ kubectl -n kubedr-system create secret generic etcd-creds
        --from-file=ca.crt --from-file=client.crt --from-file=client.key
```
  </dd>

  <dt>schedule</dt>
  <dd><p>A string in the format of "cronjob" schedule.</p>
  
  For example, "*/10 * * * *" results in backups every 10 minutes. 
  </dd>

  <dt>retainNumBackups</dt>
  <dd>Optional. An integer specifying how many successful backups
        should be stored on the target. Default value is 120. </dd>

</dl>

### MetadataBackupRecord

This resource represents a successful backup and is automatically
created at the end of each successful backup.

Here is one example:

```yaml
apiVersion: kubedr.catalogicsoftware.com/v1alpha1
kind: MetadataBackupRecord
metadata:
  name: mbr-9a9808ce
  namespace: kubedr-system
spec:
  policy: test-backup
  snapshotId: 9a9808ce
```

<dl>

  <dt>policy</dt>
  <dd>Name of the policy that created this resource.</dd>

  <dt>snapshotId</dt>
  <dd>Snapshot ID of the backup. </dd>

</dl>

## Implementation

- The tool is built using
  [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder). Note
  that another way of building operators is by using
  [operator-sdk](https://github.com/operator-framework/operator-sdk)
  but *kubebuilder* is chosen because it is more active, supports
  webhooks, and is better documented.

  More over, there are efforts underway to merge these two projects.

- When a *BackupLocation* resource is created, a *restic* repo is
  initialized. This is done by running a ```restic init``` command in
  a pod. We use ```restic/restic``` docker image for this purpose.

- When a *MetadataBackupPolicy* resource is created, a
  [CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
  is created with the given schedule. 

  The cronjob runs a backup pod at the scheduled times. The pod uses
  our own docker image *kubedrbackup*. The image contains "restic",
  "etcdctl", and "kubectl" binaries in addition to a shell script
  "kubedrbackup.sh". 

  The script creates etcd snapshot, copies certificates if required,
  and then performs a restic backup. It then creates the resource
  "MetadataBackupRecord" corresponding to just finished backup.

  The script will be converted to a full fledged Python script in a
  future release.

## Restore

The main restore use case is in a DR scenario when the master nodes
are lost and you are setting up a new cluster. In this case, first
browse backups on the target and then pick a snapshot to restore
from. 

To browse backups:
```bash
$ docker run --rm -it -e AWS_ACCESS_KEY_ID=minio \
        -e AWS_SECRET_ACCESS_KEY=minio123 \
        -e RESTIC_PASSWORD=testpass \
        restic/restic 
        -r s3:<S3-END-POINT>/<BUCKET-NAME> snapshots
```

To restore from a backup snapshot:

```bash
$ docker run --rm -it -e AWS_ACCESS_KEY_ID=minio \
        -e AWS_SECRET_ACCESS_KEY=minio123 \
        -e RESTIC_PASSWORD=testpass \
        restic/restic \
        -r s3:<S3-END-POINT>/<BUCKET-NAME> restore <SNAPSHOT-ID> \
        --target <TARGET_DIR> 
```

Once restore is done, etcd snapshot file and (optionally) certificates
will be available in <TARGET_DIR>. One can then configure etcd server
to recover data from the snapshot. For more details, see
[Restoring etcd
cluster](https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/recovery.md#restoring-a-cluster)
and docs for your cluster distro.

In future, we will provide an option to browse Kubernetes resources
from a backup without having to restore entire etcd snapshot.

## Tasks before Alpha release

- Figure out a name for the project. The current code name *kubedr* is
  used at several places and it needs to be replaced if and when we
  come up with a different name.

- Set up automatic builds. We need to figure out how container images
  will be made available. The tool requires two custom images at
  present in addition to external ones like *restic*.

- Figure out installation procedure. Clearly document the
  requirements (especially certificate related ones). 

- Implement (and document) referential integrity semantics between
  resources.

- Implement metrics through [Prometheus](https://prometheus.io/). 

- Implement reporting through [Grafana](https://grafana.com/).

- Provide an easier way of restoring in case the restore is being done
  from the same cluster. 

- Write detailed documentation.

- Write automated tests.

## Tests

- Usage of S3 using http and https.




















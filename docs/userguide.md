
- [Kubernetes Metadata Backup](#kubernetes-metadata-backup)
  * [Installation](#installation)
  * [Configuration](#configuration)
    + [Master nodes](#master-nodes)
    + [S3 end point](#s3-end-point)
  * [Backup](#backup)
  * [Pausing backups](#pausing-backups)
  * [Restore](#restore)
  * [Monitoring](#monitoring)
  * [Troubleshooting](#troubleshooting)
  * [Uninstall](#uninstall)
  
# Kubernetes Metadata Backup

## Introduction

This project, code named *Kubedr*, is about protecting metadata of
Kubernetes stored in *etcd*. In addition, certificates can be backed up
as well but that is optional.

**The project is currently in Alpha phase. Changes can be made that
may not be backwards compatible. There may be many corner cases that
may not work.**

## Overview

Kubernetes stores all its objects in *etcd* so backing up data in
*etcd* is crucial for DR purposes. This project implements a tool that
backs up *etcd* data and certificates to any S3 bucket. It follows the
*operator* pattern that is popular in Kubernetes world.

An operator is basically a combination of *custom resources (CRs)*
coupled with *controllers* that manage the CRs. There would be one
controller for each CR. In addition to controllers, operators can also
contain *webhooks* that can be used to validate the data in resources
as well as to set defaults when some fields are not provided in the
input. Our operator uses webhooks for both these purposes.

For data transfer to S3, we currently use a tool called
[restic](https://restic.net) but we will be able to change specific
backup tool in a backwards compatible manner.

## Requirements

- Since we need direct access to etcd, *Kubedr* currently works only
  for on-prem clusters or where ever etcd can be accessed. We are
  investigating how to extend the same functionality to clusters where
  etcd snapshot cannot be taken.

- The project is tested using 1.16 but we are currently verifying how
  many older versions can be supported.

## Installation

- Install [cert-manager](https://cert-manager.io/). This is required
  for webhooks to work.

- Make sure that `kubectl` is set up to access your cluster and
  install `kubedr/kubedr.yaml` (available at the root of this repo) by
  running the following command:


  ```bash
  $ kubectl apply -f kubedr.yaml
  ```

  For now, use the file *kubedr.yaml* directly from the master
  branch. Soon, we will create a "release" version of it.

  Running this command will create a new namespace called
  *kubedr-system* and starts all the necessary pods, services,
  webhooks, and deployments in that namespace. It also installs the
  following *Custom Resource Definitions (CRDs)*:

  * BackupLocation
  * MetadataBackupPolicy
  * MetadataBackupRecord

- Note that the following two images are required for *kubedr*  to
  work. They are:

  * docker-registry.devad.catalogic.us:5000/kubedr:0.42
  * docker-registry.devad.catalogic.us:5000/kubedrutil:0.42

  Currently, these images are available only from internal Catalogic
  repo. But they will be moved to an external repo soon (private or
  otherwise).

- To verify that installation is successful, run the following command
  and ensure that all the resources are in running state.

```bash
$ kubectl -n kubedr-system get all

# Sample output, your output may vary.
NAME                                             READY   STATUS    RESTARTS   AGE
pod/kubedr-controller-manager-859bc794bb-p9pf2   1/1     Running   126        3d23h

NAME                                                TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/kubedr-controller-manager-metrics-service   ClusterIP   10.109.226.24   <none>        8443/TCP,8080/TCP   6d3h
service/kubedr-webhook-service                      ClusterIP   10.97.71.44     <none>        443/TCP             6d3h

NAME                                        READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/kubedr-controller-manager   1/1     1            1           6d3h

NAME                                                   DESIRED   CURRENT   READY   AGE
replicaset.apps/kubedr-controller-manager-6868c85d7    0         0         0       4d20h
replicaset.apps/kubedr-controller-manager-79f5684d8    0         0         0       6d3h
replicaset.apps/kubedr-controller-manager-859bc794bb   1         1         1       3d23h
replicaset.apps/kubedr-controller-manager-8cc9cb9fb    0         0         0       4d21h
```

## Configuration

### Master nodes

Before backups are run, make sure that one or more master nodes has
a label identifying them (if not already present). Some clusters are
set up with *node-role.kubernetes.io/master* label on the master
nodes. If this is the case, nothing more needs to be done. If this
label is not present, create it. 

If it is not possible to use the above name for label, choose any name
and pass that name in policy options (as described below).

### S3 end point

Before defining and running backups, you need to configure a S3 end
point that acts as target for the backups. 

To define the S3 target, you need to create a custom resource called
`BackupLocation`.

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

  The secret should contain three pieces of information. Here is the
  description of each item in the secret and the key with which they
  should be created. 

  * S3 access key ("access_key")
  * S3 secret key ("secret_key")
  * Password to be used to encrypt backups ("restic_repo_password").

  Here is one way to create such a secret:

  ```bash
  $ echo -n 'sample_access_key' > access_key
  $ echo -n 'sample_secret_key' > secret_key
  $ echo -n 'sample_restic_repo_password' > restic_repo_password
  
  $ kubectl -n kubedr-system create secret generic minio-creds \
          --from-file=access_key --from-file=secret_key \
          --from-file restic_repo_password 
  ```

  Note that the secret should be created in the namespace
  *kubedr-system*.
  </dd>
</dl>

Assuming you defined the `BackupLocation` resource in a file called
`backuplocation.yaml`, create the resource by running the command:

```bash
$ kubectl -n kubedr-system apply -f backuplocation.yaml
```

At this time, *Kubedr* will initialize a backup repository at the
configured bucket (creating the bucket if necessary). To verify the
initialization process, run the following command and ensure that
status is "Completed".

```bash
$ kubectl -n kubedr-system get pod/<BACKUP_LOCATION_NAME>-init-pod
```


## Backup

- Create a `BackupLocation` resource if not already done. 

- Create a "secret" containing three pieces of information that are
  required to connect to *etcd*:

  Since *kube-apiserver* also connects to *etcd*, one can find this
  information by looking at the command line arguments passed to
  *kube-apiserver* process. The options to look at are "etcd-cafile",
  "etcd-certfile", and "etcd-keyfile". 

  Once these three files are found, copy the files with the following
  names:

  - etcd-cafile => ca.crt
  - etcd-certfile => client.crt
  - etcd-keyfile => client.key

  and create the secret as follows:

  ```bash
  $ kubectl -n kubedr-system create secret generic etcd-creds
          --from-file=ca.crt --from-file=client.crt --from-file=client.key
  ```

  Note that once the secret is created, copies of the files can be
  deleted. They are not used any more.

At this point, you are ready to create a `MetadataBackupPolicy`
resource that defines a backup policy. Here is a sample policy
resource and the description of each field:

```yaml
apiVersion: kubedr.catalogicsoftware.com/v1alpha1
kind: MetadataBackupPolicy
metadata:
  name: test-backup
spec:
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
  <dd>Describes the endpoint where etcd server is
    available. Optional. In most cases, the default value of
    "https://127.0.0.1:2379" would work. You can check the end point
    by looking at "kube-apiserver" command line option
    "etcd-servers". 
  </dd>

  <dt>etcdCreds</dt>
  <dd><p>Optional. Name of the Kubernetes "secret" resource containing
  etcd credentials. If the name "etcd-creds" is used for the secret,
  there is no need to include this field.</p>
  </dd>

  <dt>schedule</dt>
  <dd><p>A string in the format of Kubernetes
    <a
  href="https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/">cronjob</a>
    resource.</p>
  
   <p>For example, "*/10 * * * *` results in backups every 10 minutes.</p>
  </dd>

  <dt>retainNumBackups</dt>
  <dd>Optional. An integer specifying how many successful backups
        should be stored on the target. Default value is 120. </dd>

</dl>

In addition to above fields, the `MetadataBackupPolicy` resource also
supports a field called *options* which is a map of string keys and
string values. Currently, only one option is supported. 

<dl>
  <dt>master-node-label-name</dt>
  <dd><p>Describes the label that is used to designate master
  nodes.<p>

   <p>Note that if the label "node-role.kubernetes.io/master" is
  present, there is no need to specify it in the options here. If some
  other name (say "ismasternode") is used, it can be set as follows:

  ```json
  options:
    "master-node-label-name": ismasternode
  ```
  </dd>
</dl>

Assuming you defined the `MetadataBackupPolicy` resource in a file
called `policy.yaml`, create the resource by running the command:

```bash
$ kubectl -n kubedr-system apply -f policy.yaml
```

At this time, *Kubedr* will create a 
[cronjob](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/#writing-a-cron-job-spec)
resource. It can be seen as follows:

```bash
$ kubectl -n kubedr-system get all

...
...
```

## Pausing backups

It is possible to pause backups if there is a need to do so (and
resume them later).

To pause a backup, you need to patch the `MetadataBackupPolicy`
resource by following standard Kubernetes way of 
[making partial changes](https://kubernetes.io/docs/tasks/run-application/update-api-object-kubectl-patch/)
to a resource. 

First, create a file called `suspend.yaml` (you can choose any name
you want) with the following contents:

```yaml
spec:
  suspend: true
```

Replace "<NAME>" in the following command with the name of the policy
resource and then run it:

```bash
$ kubectl -n kubedr-system patch \
  metadatabackuppolicy.kubedr.catalogicsoftware.com/<NAME> \
  --patch "$(cat suspend.yaml)" --type merge 
```

You can verify that the backups are indeed suspended by checking the
cronjob resource as follows:

*TBD* Copy cronjob output here.

To resume backups, follow the same procedure as above but this time,
use the following snippet:

```yaml
spec:
  suspend: false
```

## Restore

The main restore use case is in a DR scenario when the master nodes
are lost and you are setting up a new cluster. In this case, first
browse backups on the target and then pick a snapshot to restore
from. 

To browse backups (replace access key, secret key, and restic password
values with the ones you used while creating `BackupLocation`
resource): 

```bash
$ docker run --rm -it -e AWS_ACCESS_KEY_ID=<ACCESS_KEY> \
        -e AWS_SECRET_ACCESS_KEY=<SECRET_KEY> \
        -e RESTIC_PASSWORD=<REPO_PASSWORD> \
        restic/restic 
        -r s3:<S3-END-POINT>/<BUCKET-NAME> snapshots
```

To restore from a backup snapshot:

```bash
$ docker run --rm -it -e AWS_ACCESS_KEY_ID=<ACCESS_KEY> \
        -e AWS_SECRET_ACCESS_KEY=<SECRET_KEY> \
        -e RESTIC_PASSWORD=<REPO_PASSWORD> \
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

**In future, we will provide an option to browse Kubernetes resources
from a backup without having to restore entire etcd snapshot.**

## Monitoring

TBD.

## Troubleshooting

In case of any problems, please get output of the following commands. 

```bash

$ alias k="kubectl -n kubedr-system"

$ k get all

$ k describe all

# For pods that show errors
$ k logs <PODNAME> --all-containers
```

## Uninstall

To uninstall, delete the namespace *kubedr-system*

```bash
$ kubectl delete namespace kubedr-system
```

If you don't need the backups already done, go ahead and delete the
bucket on S3.

TBD: Need to delete CRDs as well?

## Roadmap

The following list includes feature enhancements as well as
robustness improvements to the project.

- Perform *Referential Integrity* checks and prevent deletion of
  resources that are currently in use.

- Make it easy to switch backup tool. Currently, we use
  [restic](https://restic.net) but the design should support easily
  switching to any other tool.

- The current restore support assumes a DR use case where entire etcd
  snapshot needs to be restored. But we also want to support granular
  restore where one can restore individual resources.

- Auditing of changes. Provide a way to check what changed between
  backups and also to see how a particular resource has evolved over
  time.

- Support clusters in the cloud. Currently, *Kubedr* requires direct
  access to etcd so that a snapshot can be created. This may not be
  possible in the cloud so we may need to iterate over all the
  objects and back them up.

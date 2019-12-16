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
  $ kubectl -n kubedr-system create secret generic etcd-creds \ 
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


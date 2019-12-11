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


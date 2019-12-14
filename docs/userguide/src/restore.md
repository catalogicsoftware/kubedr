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
        restic/restic \
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


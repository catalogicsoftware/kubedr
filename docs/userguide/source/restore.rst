=========
 Restore
=========

The main restore use case is a DR scenario when the master nodes
are lost and you are setting up a new cluster. In this case, first
browse backups on the target and then pick a snapshot to restore
from.

To browse backups (replace "<ACCESS_KEY>", <SECRET_KEY>, and
"<REPO_PASSWORD>" with values you used while creating
``BackupLocation`` resource):

.. code-block:: bash

  $ docker run --rm -it -e AWS_ACCESS_KEY_ID=<ACCESS_KEY> \
          -e AWS_SECRET_ACCESS_KEY=<SECRET_KEY> \
          -e RESTIC_PASSWORD=<REPO_PASSWORD> \
          restic/restic \
          -r s3:<S3-END-POINT>/<BUCKET-NAME> snapshots

To restore data from a snapshot into the directory ``/tmp/restore``:

.. code-block:: bash

  $ docker run --rm -it -e AWS_ACCESS_KEY_ID=<ACCESS_KEY> \
          -e AWS_SECRET_ACCESS_KEY=<SECRET_KEY> \
          -e RESTIC_PASSWORD=<REPO_PASSWORD> \
          -v /tmp/restore:/tmp/restore \
          restic/restic \
          --target <TARGET_DIR> \
          -r s3:<S3-END-POINT>/<BUCKET-NAME> restore <SNAPSHOT-ID>

.. note::

   You must mount the target dir using -v option.

Once restore is done, etcd snapshot file and (optionally) certificates
will be available in ``/tmp/restore/data``. One can then configure etcd
server to recover data from the snapshot. For more details, see
`Restoring etcd cluster`_ and docs for your cluster distro.

.. note::

  The restore support in *Kubedr* is very limited at present, only
  covering DR scenario. In the future, we will provide an option to
  browse and restore individual or group of Kubernetes resources from
  a backup without having to restore entire etcd snapshot.** 

.. _Restoring etcd cluster: https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/recovery.md#restoring-a-cluster



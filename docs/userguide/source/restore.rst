=========
 Restore
=========

There are two types of restores supported by *KubeDR*. 

DR restore
    Used when master nodes are lost and you are setting up a new
    cluster.

Regular restore
    Used when cluster is up and running but you need access to the
    certificates or etcd snapshot.

DR Restore
==========

In this case, first browse backups on the target and then pick a
snapshot to restore from.

To help simplify restore operations, a Python library called
`kubedrctl`_ can be used. It can be installed by running:

.. code-block:: bash

  $ pip install kubedrctl

Please note that you need to use Python 3 for ``kubedrctl``. At this
point, this library is a thin wrapper over the corresponding docker
commands but its functionality will be enhanced in the future.

In rest of the document, we will provide sample commands using
``kubedrctl`` as well as using ``docker`` directly. 

Browsing Backups
----------------

To browse backups, please run the following command (replace
"<ACCESS_KEY>", <SECRET_KEY>, and "<REPO_PASSWORD>" with values you
used while creating ``BackupLocation`` resource):

.. code-block:: bash

  $ kubedrctl list backups --accesskey <ACCESSKEY> \
        --secretkey <SECRETKEY> --repopwd <REPO_PASSWORD> \
        --endpoint <S3-ENDPOINT> --bucket <BUCKET>

Alternatively, 

.. code-block:: bash

  $ docker run --rm -it -e AWS_ACCESS_KEY_ID=<ACCESSKEY> \
          -e AWS_SECRET_ACCESS_KEY=<SECRETKEY> \
          -e RESTIC_PASSWORD=<REPO_PASSWORD> \
          restic/restic \
          -r s3:<S3-END-POINT>/<BUCKET-NAME> snapshots

These commands will print list of backup snapshots available at the
given backup location. Here is a sample output::

    ID        Time                 Host        Tags        Paths
    ------------------------------------------------------------
    abe28f0f  2020-01-15 01:28:10  beast                   /data
    a0f7dbf7  2020-01-15 01:29:10  beast                   /data
    734af8c7  2020-01-15 01:30:10  beast                   /data
    
You need snapshot ID printed in the first column for restore.


Restoring a Backup
------------------

To restore data from a snapshot into the directory ``/tmp/restore``:

.. code-block:: bash

  $ kubedrctl restore --accesskey <ACCESSKEY> \
        --secretkey <SECRETKEY> --repopwd <REPO_PASSWORD> \
        --endpoint <S3-ENDPOINT> --bucket <BUCKET> \
        --targetdir /tmp/restore <SNAPSHOT-ID>

Alternatively,

.. code-block:: bash

  $ docker run --rm -it -e AWS_ACCESS_KEY_ID=<ACCESSKEY> \
          -e AWS_SECRET_ACCESS_KEY=<SECRETKEY> \
          -e RESTIC_PASSWORD=<REPO_PASSWORD> \
          -v /tmp/restore:/tmp/restore \
          restic/restic \
          -r s3:<S3-END-POINT>/<BUCKET-NAME> \
          --target /tmp/restore \
          restore <SNAPSHOT-ID>

Once restore is done, etcd snapshot file and (optionally) certificates
will be available in ``/tmp/restore/data``. One can then configure etcd
server to recover data from the snapshot. For more details, see
`Restoring etcd cluster`_ and docs for your cluster distro.

Regular Restore
===============

This type allows you to restore certificates and etcd snapshot by
simply creating a custom resource. The assumption is that the Cluster
is up and running but you need access to this data for one reason or
another. 

Browsing Backups
----------------

As has already be seen, *KubeDR* creates a resource of the type
``MetadataBackupRecord`` after every successful backup. To list all
the backups in the chronological order, run the following command:

.. code-block:: bash

  $ kubectl -n kubedr-system get  metadatabackuprecords \
       --sort-by=.metadata.creationTimestamp \
       -o custom-columns=NAME:.metadata.name,CTIME:.metadata.creationTimestamp

  NAME           CTIME
  mbr-00f2bb92   2020-02-21T18:35:10Z
  mbr-30efb3f4   2020-02-21T18:36:11Z
  mbr-a27e5153   2020-02-21T18:36:11Z
  mbr-9353053f   2020-02-21T18:45:11Z

Based on the timestamp, select the backup you want to restore from and
note the name.

Restoring a Backup
------------------

In the previous step, you selected the source for the restore and now
you need to tell *KubeDR* where the files need to be restored. This is
done by creating a `PersistentVolumeClaim`_.

`PersistentVolume`_ (PV) and `PersistentVolumeClaim`_ (PVC) resources
are the primary mechanism by which storage is provided to pods and
containers. In this case, the user needs to create a
PV of the type "FileSystem" and then create a PVC that binds to
it. There are many types of PVs supported by Kubernetes. One such type
is "HostPath" which allows a local directory on a node to be used.

Here is a sample "HostPath" PV that points to the local directory
``/tmp/restoredir``.

.. code-block:: bash

  $ cat pv.yaml

  kind: PersistentVolume
  metadata:
    name: mrtest
  spec:
    accessModes:
    - ReadWriteOnce
    capacity:
      storage: 8Gi
    hostPath:
      path: /tmp/restoredir
    persistentVolumeReclaimPolicy: Delete
    storageClassName: standard

  $ kubectl apply -f pv.yaml

The following PVC will bind to the above PV.

.. code-block:: bash

  $ cat pvc.yaml

  apiVersion: v1
  kind: PersistentVolumeClaim
  metadata:
    name: mrtest-claim
  spec:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 8Gi
    volumeMode: Filesystem
    volumeName: mrtest

  # Note that PVC needs to be created in the KubeDR namespace.
  $ kubectl -n kubedr-system apply -f pvc.yaml

At this point, PVC ``mrtest-claim`` should be bound to the PV
``mrtest`` and should be ready to be used. You can verify it like so: 

.. code-block:: bash

  $ kubectl -n kubedr-system get pvc mrtest-claim

  NAME           STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
  mrtest-claim   Bound    mrtest   8Gi        RWO            standard       1d

Now, we are ready to create the resource that would trigger the restore.

.. code-block:: bash

  $ cat restore.yaml

  apiVersion: kubedr.catalogicsoftware.com/v1alpha1
  kind: MetadataRestore
  metadata:
    name: mrtest
  spec:
    mbrName: mbr-e5014782
    pvcName: mrtest-claim

  $ kubectl -n kubedr-system apply -f restore.yaml

When restore is complete, the status of the ``MetadataRestore``
resource created above would be updated. Example of a successful
restore:: 

    apiVersion: kubedr.catalogicsoftware.com/v1alpha1
    kind: MetadataRestore
    metadata:
      ...
      name: mrtest
      namespace: kubedr-system
    spec:
      mbrName: mbr-c41edb29
      pvcName: mrtest-claim
    status:
      observedGeneration: 1
      restoreErrorMessage: ""
      restoreStatus: Completed
      restoreTime: "2020-02-21T21:14:05Z"

Once restore is complete, the restored files (``etcd-snapshot.db`` and
certificate files) can be found in the directory pointed to by the
persistent volume. At this point, you can safely delete the
``MetadataRestore`` resource.

At the end of restore, *KubeDR* generates an event. Please check
"Monitoring" section for more details.

.. _Restoring etcd cluster: https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/recovery.md#restoring-a-cluster
.. _kubedrctl: https://pypi.org/project/kubedrctl/
.. _PersistentVolumeClaim: https://kubernetes.io/docs/concepts/storage/persistent-volumes/
.. _PersistentVolume: https://kubernetes.io/docs/concepts/storage/persistent-volumes/

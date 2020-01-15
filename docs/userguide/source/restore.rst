=========
 Restore
=========

The main restore use case is a DR scenario when the master nodes
are lost and you are setting up a new cluster. In this case, first
browse backups on the target and then pick a snapshot to restore
from.

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
================

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
==================

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

.. note::

  The restore support in *Kubedr* is very limited at present, only
  covering DR scenario. In the future, we will provide an option to
  browse and restore individual or group of Kubernetes resources from
  a backup without having to restore entire etcd snapshot.** 

.. _Restoring etcd cluster: https://github.com/etcd-io/etcd/blob/master/Documentation/op-guide/recovery.md#restoring-a-cluster
.. _kubedrctl: https://pypi.org/project/kubedrctl/


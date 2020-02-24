========
 Backup
========

- Create a ``BackupLocation`` resource if not already done.

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

  .. code-block:: bash

    $ kubectl -n kubedr-system create secret generic etcd-creds \
          --from-file=ca.crt --from-file=client.crt --from-file=client.key

  Note that once the secret is created, copies of the files can be
  deleted. They are not used any more.

At this point, you are ready to create a ``MetadataBackupPolicy``
resource that defines a backup policy. Here is a sample policy
resource and the description of each field:

.. code-block:: yaml

  apiVersion: kubedr.catalogicsoftware.com/v1alpha1
  kind: MetadataBackupPolicy
  metadata:
    name: test-backup
  spec:
    destination: remote-minio

    certsDir: /etc/kubernetes/pki

    etcdEndpoint: https://127.0.0.1:2379
    etcdCreds: etcd-creds # secret

    schedule: "*/10 * * * *"

    retainNumBackups: 1

name
    Name of the policy.

destination
    Name of *BackupLocation* resource where backups should be stored.

certsDir
    Directory containing Kubernetes certificates. Optional. If given,
    contents of entire directory will be backed up.

etcdEndpoint
    Describes the endpoint where etcd server is
    available. Optional. In most cases, the default value of
    "https://127.0.0.1:2379" would work. You can check the end point
    by looking at "kube-apiserver" command line option "etcd-servers".

etcdCreds
    Optional. Name of the Kubernetes "secret" resource containing etcd
    credentials. If the name "etcd-creds" is used for the secret,
    there is no need to include this field.

schedule
    A string in the format of Kubernetes `cronjob`_ resources's
    "schedule" field.

    For example, "\*/10 \* \* \* \*` results in backups every 10
    minutes.

retainNumBackups
    Optional. An integer specifying how many successful backups should
    be stored on the target. Default value is 120.

In addition to above fields, the ``MetadataBackupPolicy`` resource also
supports a field called *options* which is a map of string keys and
string values. Currently, only one option is supported.

master-node-label-name
    Describes the label that is used to designate master nodes.

    Note that if the label "node-role.kubernetes.io/master" is
    present, there is no need to specify it in the options here. If
    some other name (say "ismasternode") is used, it can be set as
    follows:

    .. code-block:: yaml

       options:
         "master-node-label-name": ismasternode

Assuming you defined the ``MetadataBackupPolicy`` resource in a file
called ``policy.yaml``, create the resource by running the command:

.. code-block:: bash

  $ kubectl -n kubedr-system apply -f policy.yaml

At this time, *Kubedr* will create a `cronjob`_ resource.

After every successful backup, *KubeDR* creates a resource of the type
``MetadataBackupRecord`` which contains the snapshot ID of the
backup. This resource acts as a "catalog" for the backups. Here is one
such sample resource::

    apiVersion: kubedr.catalogicsoftware.com/v1alpha1
    kind: MetadataBackupRecord
    metadata:
      creationTimestamp: "2020-02-21T18:35:10Z"
      finalizers:
      - mbr.finalizers.kubedr.catalogicsoftware.com
      generation: 2
      name: mbr-00f2bb92
      namespace: kubedr-system
      resourceVersion: "1739627"
      selfLink: /apis/kubedr.catalogicsoftware.com/v1alpha1/namespaces/kubedr-system/metadatabackuprecords/mbr-00f2bb92
      uid: 50cf3088-7763-4d8a-bb8b-0c308b1fbdac
    spec:
      backuploc: tests3-1582310048
      policy: backup-1582310055
      snapshotId: 00f2bb92

As can be seen, the spec of ``MetadataBackupRecord`` has three pieces
of information.

backuploc
    Points to the ``BackupLocation`` resource used for the backup.

policy
    Name of the ``MetadataBackupPolicy`` resource.

snapshotId
    Snapshot ID of the backup. This value is used in restores.

In addition to creating the above resource, *KubeDR* also generates an
event both in case of success as well as in case of any
failures. Please check :ref:`Backup Events<Backup events>` for more
details.
    

.. _cronjob: https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs


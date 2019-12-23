=================
 Troubleshooting
=================

Collecting Info
===============

In case of any problems, please get output of the following commands. 

.. code-block:: bash

  $ alias k="kubectl -n kubedr-system"

  $ k get all

  $ k describe all

  # For pods that show errors
  $ k logs <PODNAME> --all-containers

Artifacts
=========

This section is aimed at cluster admins who want to or need to know
all the artifacts that comprise of *Kubedr*.

Custom Resources
----------------

BackupLocation
    Represents a S3 backup target.

MetadataBackupPolicy
    Describes the backup policy.

MetadataBackupRecord
    Represents a successful backup. It is not being used in restores
    at present but once granular restore feature is added, this
    resource will be used.

Kubernetes Resources
--------------------

Controller Manager Pod
    This has controllers for all the custom resources. In addition, it
    also serves metrics and implements webhook end points (used for
    validation and initialization of unset fields). 

    Corresponding to this pod, there is a Replica Set, Deployment, and
    two services.

Cronjobs
    There will be one `cronjob`_ for each backup policy.

Job
    One job created for each backup instance (managed by "Cronjob").

Repo initialization pod
    When a new ``BackupLocation`` is added, a pod is created that
    initializes the repo. It is named as "<NAME>-init-pod" where
    "<NAME>" is the name of ``BackupLocation`` resource.

    This pod is not deleted currently but in the future, it will be
    cleaned up.

Snapshot deletion pods
    In order to support retention setting and clean up expired
    snapshots, a pod is created that deletes the backup snapshot. Such
    pods are named "mbr-<SNAPSHOTID>-del".

    At most three such deletion pods are kept and others are cleaned
    up.

.. _cronjob: https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs

    











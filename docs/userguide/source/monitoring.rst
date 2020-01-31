============
 Monitoring
============

*Kubedr* exposes several metrics that can be scraped with
`Prometheus`_ and visualised using `Grafana`_. Most of the metrics at
present deal with the internal implementation and hence, are not being
documented here. 

In the future, several metrics that are useful to the users will be
added. Some examples:

- Number of successful/failed backups
- Backup size

Status of Resources
===================

All Kubernetes resources have two sections - *spec* and *status*.

*spec* describes the intent of the user and the the cluster constantly
drives towards matching it. On the other hand, *status* is for the
cluster components to set and it typically contains useful information
about the current state of the resource.

*KubeDR* makes use of the *status* field to set the results of backup
and other operations. The following sections describe the *status*
details for each resource.

BackupLocation
--------------

This resource defines a backup target (which is an S3 bucket) and
when it is created, *KubeDR* initializes a backup repo at the given
bucket. The *status* field of the resource indicates success or
failure of such operation.

Here is an example of an error condition::

    status:
      initErrorMessage: |+
        Fatal: create key in repository at s3:http://10.106.189.174:9000/testbucket50 failed: repository master key and config already initialized

      initStatus: Failed
      initTime: Thu Jan 30 16:02:53 2020

When initialization succeeds::

    status:
      initErrorMessage: ""
      initStatus: Completed
      initTime: Thu Jan 30 16:05:56 2020

MetadataBackupPolicy
--------------------

This resource defines the backup policy and its *status* field
indicates details about the most recent backup.

An example::

    status:
      backupErrorMessage: ""
      backupStatus: Completed
      backupTime: Thu Jan 30 16:04:05 2020
      dataAdded: 1573023
      filesChanged: 1
      filesNew: 0
      snapshotId: b0f347ef
      totalBytesProcessed: 15736864
      totalDurationSecs: "0.318463127"

Events
======

*KubeDR* generates events after some operations that can be monitored
by admins. The following sections provide more details about each such
event. Note that events are generated in the namespace
*kubedr-system*. 

Backup repo initialization
--------------------------

When a ``BackupLocation`` resource is created first time, a backup
repo is initialized at the given S3 bucket. An event is generated at
the end of such init process. 

Here is an example of the event generated after successful
initialization.::

    $ kubectl -n kubedr-system get event

    ...
    25s  Normal  InitSucceeded    backuplocation/local-minio   Repo at s3:http://10.106.189.174:9000/testbucket62 is successfully initialized

In case of error::

    $ kubectl -n kubedr-system get event

    ...
    5s   Error  InitFailed        backuplocation/local-minio   Fatal: create key in repository at s3:http://10.106.189.174:9000/testbucket62 failed: repository master key and config already initialized

.. _Prometheus: https://prometheus.io
.. _Grafana: https://grafanalabs.io



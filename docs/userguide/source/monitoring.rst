============
 Monitoring
============

There are multiple ways in which KubeDR's operations can be
monitored. They include:

- Prometheus metrics

- "Status" section of individual resources.

- Kubernetes events

The following sections will elaborate on each of these mechanisms. 

Prometheus Metrics
==================

*KubeDR* exposes several metrics that can be scraped with
`Prometheus`_ and visualized using `Grafana`_. Most of the metrics
deal with the internal implementation but the following ones provide
very useful information to the user. They are widely known as
`RED`_ metrics.

kubedr_backup_size_bytes (Gauge)
    Size of the backup in bytes.

kubedr_num_backups (Counter)
    Total number of backups.

kubedr_num_successful_backups (Counter)
    Total number of successful backups.

kubedr_num_failed_backups (Counter)
    Total number of successful backups.

kubedr_backup_duration_seconds (Histogram)
    Time (seconds) taken for the backup.

    This metric is a histogram with the following buckets::

        15s, 30s, 1m, 5m, 10m, 15m, 30m, 1h, ...., 10h

All the metrics will have a label called ``policyName`` set to the
name of the ``MetadataBackupPolicy`` resource.

.. note::

   More details on how exactly Prometheus can be configured to scrape
   KubeDR's metrics will be provided soon. If you are interested,
   please check out `issue 26`_.

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
      mbrName': mbr-4c1223d6
      snapshotId: b0f347ef
      totalBytesProcessed: 15736864
      totalDurationSecs: "0.318463127"

Apart from the stats regarding the backup, the status also contains
the name of the ``MetadataBackupRecord`` resource that is required to
perform restores.

MetadataRestore
---------------

This resource defines a restore and its *status* field indicates
success or failure of the operation.

Success::

    restoreErrorMessage: ""
    restoreStatus: Completed

Error::

    restoreErrorMessage: Error in creating restore pod
    restoreStatus: Failed

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

.. _Backup events:


Backup
------

After every backup, an event is generated containing details about
success or failure and in the case of latter, the event will
contain relevant error message. Here are couple of sample events.

Success::

    Normal  BackupSucceeded  metadatabackuppolicy/test-backup  Backup completed, snapshot ID: 34abbf1b

Error::

    Error  BackupFailed  metadatabackuppolicy/test-backup  subprocess.CalledProcessError: 
        Command '['restic', '--json', '-r', 's3:http://10.106.189.174:9000/testbucket63', 
            '--verbose', 'backup', '/data']' returned non-zero exit status 1. 
            (Fatal: unable to open config file: Stat: The access key ID you provided does not exist 
            in our records. Is there a repository at the following location?
            s3:http://10.106.189.174:9000/testbucket63

Restore
-------

After every restore, an event is generated containing details about
success or failure and in the case of latter, the event will
contain relevant error message. Here are couple of sample events.

Success::

    Normal  RestoreSucceeded metadatarestore/mrtest  Restore from snapshot 5bbc8b1a completed

Error::

    Error RestoreFailed  metadatarestore/mrtest subprocess.CalledProcessError: 
        Command '['restic', '-r', 's3:http://10.106.189.175:9000/testbucket110', 
        '--verbose', 'restore', '--target', '/restore', '5bbc8b1a']' returned non-zero exit 
        status 1. (Fatal: unable to open config file: Stat: 
        Get http://10.106.189.175:9000/testbucket110/?location=: 
        dial tcp 10.106.189.175:9000: i/o timeout
        Is there a repository at the following location?
        s3:http://10.106.189.175:9000/testbucket110)

.. _Prometheus: https://prometheus.io
.. _Grafana: https://grafana.com
.. _RED: https://www.scalyr.com/blog/red-and-monitoring-three-key-metrics-and-why-they-matter/
.. _issue 26: https://github.com/catalogicsoftware/kubedr/issues/26


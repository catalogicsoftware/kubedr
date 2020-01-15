==========
 Overview
==========

Kubernetes stores all its objects in *etcd* so backing up data in
*etcd* is crucial for DR purposes. This project implements a tool that
backs up *etcd* data and certificates to any S3 bucket. It follows the
*operator* pattern that is popular in Kubernetes world.

An operator is basically a combination of *custom resources (CRs)*
coupled with *controllers* that manage the CRs. There would be one
controller for each CR. In addition to controllers, operators can also
contain *webhooks* that can be used to validate the data in resources
as well as to set defaults when some fields are not set in the
resource specs. Our operator uses webhooks for both these purposes.

For data transfer to S3, we currently use a tool called `restic`_. In
the future, it will be possible to change the specific backup tool in
a backwards compatible manner.

High level features of KubeDR
=============================

- Backup of *etcd* data and certificates to S3.
- Backups are encrypted and deduplicated.
- Can pause and resume backups.
- Can configure "retention" that controls how many backups are kept.

Requirements
============

- Since direct access to etcd is needed, *Kubedr* currently works
  only for clusters where *etcd* is accessible and a snapshot can be
  taken. 

  This includes on-prem clusters as well as those in the cloud that
  are explicitly set up on the compute instances.

- Supported Versions: 1.13 - 1.17.

.. _restic: https://restic.net

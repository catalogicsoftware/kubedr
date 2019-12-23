=========
 Roadmap
=========

The *Kubedr* project is currently in *Alpha* phase with several
enhancements currently being worked on. 

The following list includes such enhancements as well as robustness
improvements that are being considered for the next release. 

- Support *Helm* installation.

- Perform *Referential Integrity* checks and prevent deletion of
  resources that are currently in use.

- Add more metrics.

- Make it easy to switch backup tool. Currently, we use
  `restic`_ but the design should support easily switching to any
  other tool. 

- Support a file system target in addition to S3 (or any
  `PersistentVolume`).

- The current restore support assumes a DR use case where entire etcd
  snapshot needs to be restored. But we also want to support granular
  restore where one can restore individual resources.

- Auditing of changes. Provide a way to check what changed between
  backups and also to see how a particular resource has evolved over
  time.

- Support clusters in the cloud. Currently, *Kubedr* requires direct
  access to etcd so that a snapshot can be created. This may not be
  possible in the cloud so we may need to iterate over all the
  objects and back them up.

.. _restic: https://restic.net

=========
 Roadmap
=========

The *Kubedr* project is currently in *Alpha* phase with several
enhancements currently being worked on. 

The following list includes such enhancements as well as robustness
improvements that are being considered for the next release. 

- Support *Helm* installation.

- Make it easy to switch backup tool. Currently, we use
  `restic`_ but the design should support easily switching to any
  other tool. 

- Support a file system target in addition to S3 (or any
  `PersistentVolume`).

- Support more restore use cases.

.. _restic: https://restic.net

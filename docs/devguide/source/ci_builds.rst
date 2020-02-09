===========
 CI Builds
===========

.. note::

   All KubeDR `Catalogic Software`_ CI/CD Builds are now handled 
   by `Concourse CI`_.

All artifacts are created using proper `Semantic Versioning`_ (`semver`) schemes
and use S3 storage backends to store version history.

Nothing runs on the base shell level on any host, even docker builds
utilize docker-in-docker (DND) to build / push docker images from
within a container.

Pipeline Basics
===============

All pipeline configuration is written in YAML in accordance with the 
Concourse CI pipeline specification.

Currently the following automatically triggers a pipeline run:

  - Pushing to 'master'

  - Opening a new Pull Request

  - Committing to an open Pull Request

Artifacts
=========

In total, there are four artifacts being produced:

  1. ``kubedr.yaml`` 
  2. ``kubedr`` Docker Image
  3. userguide 
  4. devguide 

Of chief importance is the ``kubedr.yaml`` file, which holds the bundled
operator resource definition for *KubeDR* and is applied against
Kubernetes masters during the `kubedr-apply` job, and tested against in the
`smoke-tests` job. 

Releases
========

Provided all smoke-tests are passing, if the release job is started 
from Concourse, the pipeline will continue on to package the appropriate 
semver-formatted release assets and trigger a GitHub release.

.. _Semantic Versioning: https://semver.org 
.. _Concourse CI: https://concourse-ci.org
.. _Catalogic Software: https://catalogicsoftware.com

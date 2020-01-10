===========
 CI Builds
===========

.. note::

   For now, CI/CD is done by internal gitlab infrastructure at
   `Catalogic Software`_. This will continue until we figure out how
   best to port the logic to github based tooling.

The CI build, or 'pipeline' in GitLab terms, runs using the GitLab
CI/CD toolchain, every 'job' is run in its own pre-determined
container. 

Nothing runs on the base shell level on any host, even docker builds
utilize docker-in-docker (DND) to build / push docker images from
within a container.

Pipeline Basics
===============

All pipeline configuration is written in YAML, all currently
self-contained within the ``gitlab-ci.yml`` file at the root of the
repository.

Currently the following triggers a pipeline run:

- Pushing to master

- Pushing to a branch with an open merge request

- Pushing a tag

The pipeline is visible from the GitLab projects screen > CI/CD. Each
'circle' at the top level represents a 'stage', if you click into it
you will see the individual jobs running in parallel within that same
stage. 

Artifacts
=========

The three artifacts being produced are:

  1. `kubedr.yaml` 
  2. userguide 
  3. devguide 

Of chief importance is the `kubedr.yaml` file, which holds the bundled
operator resource definition for *KubeDR* and is applied against
Kubernetes masters during the `apply` stage, and tested against in the
`test` stage. 

Releases
========

Releases are controlled by *git tags*, a push with a tag will gear the
pipeline for a 'release', causing it to change Docker push targets to
DockerHub with the specific tagged version, and build the operator
with the DockerHub images in mind. 


References
==========

Some important documentation pages relating to GitLab's CI/CD pipeline
configuration: 

- For default environment variables accessible in every job, see:
  `Predefined environment variables reference`_

- For available key/values to define the pipeline in
  ``gitlab-ci.yml``, see: `GitLab CI/CD Pipeline Configuration
  Reference`_ 

.. _Predefined environment variables reference: https://docs.gitlab.com/ee/ci/variables/predefined_variables.html
.. _GitLab CI/CD Pipeline Configuration Reference: https://docs.gitlab.com/ee/ci/yaml/README.html
.. _Catalogic Software: https://catalogicsoftware.com

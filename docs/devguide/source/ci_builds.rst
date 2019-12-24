===========
 CI Builds
===========

The CI build, or 'pipeline' in GitLab terms, runs using the GitLab CI/CD toolchain, every 'job' is run in its own pre-determined container.
Nothing runs on a simple shell level on any host, even docker builds utilize docker-in-docker (DND) to 
build / push docker images from within a container.


Architecture
=================

Pipeline Basics
-----
All pipeline configuration is written in YAML, all currently self-contained within the ``gitlab-ci.yml`` file
at the root of the repository. 

Currently the following triggers a pipeline run:

  - Pushing to master

  - Pushing to a branch with an open merge request

  - Pushing a tag

References
=================

Some important documentation pages relating to the CI/CD Pipeline:

  - For default environment variables accessible in every job - `Predefined environment variables reference`

  - For available key/values to define the pipeline in ``gitlab-ci.yml`` - `GitLab CI/CD Pipeline Configuration Reference`



.. _Predefined environment variables reference https://docs.gitlab.com/ee/ci/variables/predefined_variables.html
.. _GitLab CI/CD Pipeline Configuration Reference https://docs.gitlab.com/ee/ci/yaml/README.html


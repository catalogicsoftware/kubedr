===========
 CI Builds
===========

The CI build, or 'pipeline' in GitLab terms, runs using the GitLab CI/CD toolchain, every 'job' is run in its own pre-determined container.
Nothing runs on a simple shell level on any host, even docker builds utilize docker-in-docker (DND) to 
build / push docker images from within a container.


Pipeline Structure
=================

All pipeline configuration is written in YAML, all currently self-contained within the `gitlab-ci.yml` file
at the root of the repository. 

Currently the following triggers a pipeline run:

  - Pushing to master

  - Pushing to a branch with an open merge request

  - Pushing a tag
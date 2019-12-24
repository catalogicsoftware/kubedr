===========
 CI Builds
===========

The CI build or 'pipeline' runs using the GitLab CI/CD toolchain, every 'job' is run in its own pre-determined container.
Nothing runs on a simple shell level on any host, even docker builds utilize docker-in-docker (DND) to 
build / push docker images from within a container.

Currently the following triggers a pipeline run:

  - Pushing to master
  - Pushing to a branch with an open merge request
  - Pushing a tag
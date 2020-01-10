==============
 Local Builds
==============

Operator
=================
The *KubeDR* operator resource can be built locally through utilizing the `Makefile` at the root of the repository, simply run `make build`, and *KubeDR*
will start both its local docker build and operator customization.

The resulting output artifact will be the `kubedr.yaml` file under `kubedr/`.
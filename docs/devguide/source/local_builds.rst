==============
 Local Builds
==============

To build locally:

.. code-block:: bash

   $ make build

This builds two artifacts:

- ``kubedr.yaml``
- Docker image ``kubedr:latest``

The image tag can be changed by using env variable
``DOCKER_KUBEDR_IMAGE_TAG``. 

Before applying ``kubedr.yaml``, make sure that the image is accessible
in your test environment. For example, if you are using `minikube`_,
you may need to add the image to its cache, like so:

.. code-block:: bash

   $ minikube cache add kubedr:latest

.. _minikube: https://github.com/kubernetes/minikube

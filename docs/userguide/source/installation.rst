==============
 Installation
==============

- Install `cert-manager`_.

- Make sure that ``kubectl`` is set up to access your cluster.

- Download `kubedr.yaml` from the 
  `Releases page <https://github.com/catalogicsoftware/kubedr/releases>`_.

  .. note::

    We are also working on supporting *Helm* installs in the future.

- Apply the downloaded ``kubedr.yaml``, like so:

  .. code-block:: bash

    $ kubectl apply -f kubedr.yaml

  Note that the following two images are required for *Kubedr*  to
  work.

  * catalogicsoftware/kubedrutil:0.1.0
  * catalogicsoftware/kubedr:0.1.0

- Applying ``kubedr.yaml`` will create a new namespace called
  *kubedr-system* and starts all the necessary pods, services,
  webhooks, and deployments in that namespace. It also installs the
  following *Custom Resource Definitions (CRDs)*: 

  * BackupLocation
  * MetadataBackupPolicy
  * MetadataBackupRecord

- To verify that installation is successful, run the following command
  and ensure that all the resources are in running state.

.. code-block:: bash

  $ kubectl -n kubedr-system get all

  NAME                                             READY   STATUS    RESTARTS   AGE
  pod/kubedr-controller-manager-7bc7dc96f6-h8v28   2/2     Running   0          4s

  NAME                                                TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
  service/kubedr-controller-manager-metrics-service   ClusterIP   10.104.87.59    <none>        8443/TCP   4s
  service/kubedr-webhook-service                      ClusterIP   10.109.153.83   <none>        443/TCP    4s

  NAME                                        READY   UP-TO-DATE   AVAILABLE   AGE
  deployment.apps/kubedr-controller-manager   1/1     1            1           4s

  NAME                                                   DESIRED   CURRENT   READY   AGE
  replicaset.apps/kubedr-controller-manager-7bc7dc96f6   1         1         1       4s
  
.. _cert-manager: https://cert-manager.io/

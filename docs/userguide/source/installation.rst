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

  * catalogicsoftware/kubedrutil:0.0.2
  * catalogicsoftware/kubedr:0.42

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

  # Sample output, your output may vary.
  NAME                                             READY   STATUS    RESTARTS   AGE
  pod/kubedr-controller-manager-859bc794bb-p9pf2   1/1     Running   126        3d23h
  
  NAME                                                TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
  service/kubedr-controller-manager-metrics-service   ClusterIP   10.109.226.24   <none>        8443/TCP,8080/TCP   6d3h
  service/kubedr-webhook-service                      ClusterIP   10.97.71.44     <none>        443/TCP             6d3h
  
  NAME                                        READY   UP-TO-DATE   AVAILABLE   AGE
  deployment.apps/kubedr-controller-manager   1/1     1            1           6d3h
  
  NAME                                                   DESIRED   CURRENT   READY   AGE
  replicaset.apps/kubedr-controller-manager-6868c85d7    0         0         0       4d20h
  replicaset.apps/kubedr-controller-manager-79f5684d8    0         0         0       6d3h
  replicaset.apps/kubedr-controller-manager-859bc794bb   1         1         1       3d23h
  replicaset.apps/kubedr-controller-manager-8cc9cb9fb    0         0         0       4d21h
  
.. _cert-manager: https://cert-manager.io/

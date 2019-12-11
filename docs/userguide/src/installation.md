## Installation

- Install [cert-manager](https://cert-manager.io/). This is required
  for webhooks to work.

- Make sure that `kubectl` is set up to access your cluster and
  install `kubedr/kubedr.yaml` (available at the root of this repo) by
  running the following command:


  ```bash
  $ kubectl apply -f kubedr.yaml
  ```

  For now, use the file *kubedr.yaml* directly from the master
  branch. Soon, we will create a "release" version of it.

  Running this command will create a new namespace called
  *kubedr-system* and starts all the necessary pods, services,
  webhooks, and deployments in that namespace. It also installs the
  following *Custom Resource Definitions (CRDs)*:

  * BackupLocation
  * MetadataBackupPolicy
  * MetadataBackupRecord

- Note that the following two images are required for *kubedr*  to
  work. They are:

  * docker-registry.devad.catalogic.us:5000/kubedr:0.42
  * docker-registry.devad.catalogic.us:5000/kubedrutil:0.42

  Currently, these images are available only from internal Catalogic
  repo. But they will be moved to an external repo soon (private or
  otherwise).

- To verify that installation is successful, run the following command
  and ensure that all the resources are in running state.

```bash
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
```

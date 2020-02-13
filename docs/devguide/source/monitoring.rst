============
 Monitoring
============

Since *KubeDR* is built with `kubebuilder`_ framework, it
automatically comes with a way to export `Prometheus`_ metrics. In
fact, `controller runtime`_ exports several metrics dealing with
internal implementation details but they are only relevant for
developers. 

*KubeDR* adds several metrics that are of interest to the
users. For details about these metrics, check the `user guide`_. 

Here are some high level details regarding the implementation and
testing of metrics sub-system. 

- All metrics are defined in the file ``kubedr/metrics/metrics.go``. 

The metrics end point is protected by RBAC so until we figure out how
to configure Prometheus in this setting, the feature was tested in the
following way:

- Remove RBAC by commenting out the line::

     - manager_auth_proxy_patch.yaml 

  and uncommenting the line::

     - manager_prometheus_metrics_patch.yaml

  in the file::

      kubedr/config/default/kustomization.yaml. 

You need to build *KubeDR* after this change. Once *KubeDR* is
deployed after this change, run the following command to make metrics
end point accessible on local host:

.. code-block:: bash

  $ kubectl -n kubedr-system port-forward <KUBEDR-CONTROLLER-POD> 8080:8080

Here is an example:

.. code-block:: bash

  $ kubectl -n kubedr-system port-forward kubedr-controller-manager-bd9f4467c-ljblq 8080:8080

Now, the following command will show all the relevant metrics:

.. code-block:: bash

  $ curl -s http://localhost:8080/metrics | grep kubedr_

.. _kubebuilder: https://book.kubebuilder.io/
.. _Prometheus: https://prometheus.io
.. _controller runtime: https://github.com/kubernetes-sigs/controller-runtime
.. _user guide: https://catalogicsoftware.com/clab-docs/kubedr/userguide




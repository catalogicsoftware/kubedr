================
 Implementation
================

- The project uses `kubebuilder`_ tool to scaffold new controllers and
  types. 

- Before deciding on `kubebuilder`_, `opsdk`_ was considered but
  `kubebuilder`_ is more active and has better documentation. Moreover,
  it has better support for webhooks. Also, there is an effort
  underway to integrate these two projects as far as Go operators are
  concerned. 

More details about the implementation will be added soon.

.. _kubebuilder: https://book.kubebuilder.io/
.. _opsdk: https://github.com/operator-framework/operator-sdk

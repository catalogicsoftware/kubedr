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

Status updates
==============

By default, any updates to status of a resource results in
reconciling. This is a problem if status is being updated in the
controller and might result in infinite loop of update followed by a
reconcile. In many cases, you just want to update the status and don't
want to have to process that update again.

One way to achieve that is by checking to see if the "generation"
number of a resource changed. This number is bumped up only if the
"spec" of a resource changes. So if we skip reconcile in case the
"generation" number hasn't changed, we will be avoiding reconciles
triggered by status updates. But keep in mind that you may miss other
changes to the metadata (such as changes to annotations) as well.

To implement this technique, do the following (using ``backupLoc`` as
the resource in the examples below):

- Add the following field to the status struct:

  .. code-block::

   // +kubebuilder:validation:Optional
   ObservedGeneration int64 `json:"observedGeneration"`

- Set the generation number in status (do this whether or not the
  current operation succeeds).

  .. code-block:: python

   backupLoc.Status.ObservedGeneration = backupLoc.ObjectMeta.Generation

- Do the following check in the relevant controller:

  .. code-block:: go

   if backupLoc.Status.ObservedGeneration == backupLoc.ObjectMeta.Generation {
       return ctrl.Result{}, nil
   }


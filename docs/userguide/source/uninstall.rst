
===========
 Uninstall
===========

To uninstall:

- Delete all the CRs
- Delete the namespace *kubedr-system*
- Delete CRDs

It is important to follow the order of deletions as otherwise,
deletion of namespace may hang. Here are the commands to uninstall:


.. code-block:: bash

  $ alias k="kubectl -n kubedr-system"
  
  # Delete all CRs
  $ k delete metadatabackuppolicy --all
  $ k delete metadatabackuprecord --all
  $ k delete backuplocation --all
  
  # Delete the namespace
  $ kubectl delete namespace kubedr-system
  
  # Delete CRDs
  $ kubectl delete crd metadatabackuppolicies.kubedr.catalogicsoftware.com
  $ kubectl delete crd metadatabackuprecords.kubedr.catalogicsoftware.com
  $ kubectl delete crd backuplocations.kubedr.catalogicsoftware.com

If you don't need the backups any more, go ahead and delete the
bucket on S3.

  .. note::

     In the future, you will be able to install and uninstall using
     *Helm*. 



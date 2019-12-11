
## Uninstall

To uninstall, delete the namespace *kubedr-system*

```bash
$ kubectl delete namespace kubedr-system
```

If you don't need the backups already done, go ahead and delete the
bucket on S3.

TBD: Need to delete CRDs as well?

## Pausing backups

It is possible to pause backups if there is a need to do so (and
resume them later).

To pause a backup, you need to patch the `MetadataBackupPolicy`
resource by following standard Kubernetes way of
[making partial changes](https://kubernetes.io/docs/tasks/run-application/update-api-object-kubectl-patch/)
to a resource.

First, create a file called `suspend.yaml` (you can choose any name
you want) with the following contents:

```yaml
spec:
  suspend: true
```

Replace "<NAME>" in the following command with the name of the policy
resource and then run it:

```bash
$ kubectl -n kubedr-system patch \
  metadatabackuppolicy.kubedr.catalogicsoftware.com/<NAME> \
  --patch "$(cat suspend.yaml)" --type merge
```

You can verify that the backups are indeed suspended by checking the
cronjob resource as follows:

```bash
$ kubectl -n kubedr-system get cronjobs
NAME                             SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
test-backup-new-backup-cronjob   */2 * * * *   False     0        5m59s           13m
```

To resume backups, follow the same procedure as above but this time,
use the following snippet:

```yaml
spec:
  suspend: false
```


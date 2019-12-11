## Troubleshooting

In case of any problems, please get output of the following commands. 

```bash

$ alias k="kubectl -n kubedr-system"

$ k get all

$ k describe all

# For pods that show errors
$ k logs <PODNAME> --all-containers
```

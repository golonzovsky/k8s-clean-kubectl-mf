# cleanup kubectl field owners
You deployed thing with helm and edited some object with kubectl?   
Now `helm rollback` doesent work and neither new helm deployments?

Well, for some reason kubectl is considered a valid resource owner in k8s despite not being active controller in cluster.   
Yes, you can do `helm rollback --force`, but it will mess up active HPAs and other stuff.  

Here we clean up this nonsense.  

```sh
k8s-clean-kubectl-mf --dry-run=false --log-level=debug
```


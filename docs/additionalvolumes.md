# Additional Volumes
Additional volumes may be needed for things like adding spill directories for the Presto cluster or for maintaining the data directory on a long term storage. For such a use case, additional volumes can be specified for Presto Cluster just like how they are specified with pods. The lifecycle of volumes would be tied with the lifecycle of the Presto Cluster. The specified volumes are mounted on the Presto server and Presto workers. 

For specifying the volume, the mount point and the volumes can be specified together. For e.g.

```bash
# An empty dir (Kubernetes construct) is specified 
# as the volume and is mounted on /prestospill
spec:
  volumes:
  - name: spillvol
    emptyDir: {}
    mountPath: /prestospill
# /prestospill folder is used as the spill folder
  coordinator:
    additionalProps:
      spill-enabled: "true"
      spiller-spill-path: "/prestospill"
  coordinator:
    additionalProps:
      spill-enabled: "true"
      spiller-spill-path: "/prestospill"
```

# Additional Volumes
Additional volumes may be needed for things like adding spill directories for the Presto cluster or for maintaining the data directory on a long term storage. For such a use case, additional volumes can be specified for Presto Cluster just like how they are specified with pods. For non-persistent volumes, the lifecycle of volumes would be tied with the lifecycle of the Presto Cluster. The specified volumes are mounted on the Presto server and Presto workers. 

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
## Spill Volumes

In the case of memory intensive operations, Presto allows offloading intermediate operation results to disk. The goal of this mechanism is to enable execution of queries that require amounts of memory exceeding per query or per node limits.

You can add additional volumes to enable disk spilling. EmptyDir and Hostpath are two good candidates for disk spilling.
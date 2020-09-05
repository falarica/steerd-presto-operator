# Managing Presto Cluster

Kubernetes API and `kubectl` command can be used to list the presto servers. 

```bash
$ kubectl get prestos -A 
NAMESPACE   NAME        COORDINATOR              CLUSTERSTATE   COORDINATORCPU   WORKERSCPU   DESIREDWORKERS   CURRENTWORKERS
default     mycluster   10.8.10.201:8100/30002   Ready          42%              42%          1                1

```
Kubernetes API and `kubectl` command can also be used to describe a Presto cluster. The `describe` command outputs the Status of a presto cluster along with the Spec. The status of the presto cluster looks something like this.  

```bash
$ kubectl describe prestos mycluster
 
Name:         mycluster
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"falarica.io/v1alpha1","kind":"Presto","metadata":{"annotations":{},"name":"mycluster","namespace":"default"},"spec":{"addit...
API Version:  falarica.io/v1alpha1
Kind:         Presto
Metadata:
  Creation Timestamp:  2020-06-16T10:05:32Z
  Generation:          1
  Resource Version:    62569
  Self Link:           /apis/falarica.io/v1alpha1/namespaces/default/prestos/mycluster
  UID:                 ea4e17e6-afb8-11ea-af85-42010a80010e
Spec:
  Additional Presto Prop Files:
    access-control.properties:  access-control.name=read-only
  Catalogs:
    Catalog Spec:
      Content:
        connector.name:  tpch
      Name:              newtpch
      Content:
        connector.name:  tpcds
      Name:              newtpcds
  Coordinator:
    Cpu Limit:                   0.5
    Https Enabled:               true
    Https Key Pair Password:     hemant
    Https Key Pair Secret Key:   prestoserverkeystore.jks
    Https Key Pair Secret Name:  prestokeystore
    Memory Limit:                1Gi
  Service:
    Node Port:  30002
    Port:       8100
    Type:       NodePort
  Worker:
    Additional Props:
      shutdown.grace-period:  10s
    Autoscaling:
      Enabled:                            false
      Max Replicas:                       3
      Min Replicas:                       2
      Target CPU Utilization Percentage:  20
    Count:                                1
    Cpu Limit:                            0.5
    Memory Limit:                         1Gi
Status:
  Catalog Config:          catalogconfig-03f118d2
  Cluster State:           Ready
  Coordinator Address:     10.8.10.201:8100/30002
  Coordinator CPU:         18%
  Coordinator Config:      coordinatorconfig-03f118d2
  Coordinator Replicaset:  coordinatorreplicaset-03f118d2
  Current Workers:         1
  Desired Workers:         1
  Error Reason:            
  Headless Service:        pod-discovery-03f118d2
  Hpa Name:                
  Modification Time:       2020-06-16T10:08:29Z
  Service:                 external-presto-svc-03f118d2
  Uuid:                    03f118d2-d399-4e66-8417-b5caa96988ec
  Worker CPU:              13%
  Worker Config:           workerconfig-03f118d2
  Worker Replicaset:       workerreplicaset-03f118d2zk768
Events:
  Type    Reason   Age    From               Message
  ----    ------   ----   ----               -------
  Normal  Created  2m52s  presto-controller  Created Headless Service. pod-discovery-03f118d2
  Normal  Created  2m51s  presto-controller  Created Service. external-presto-svc-03f118d2
  Normal  Created  2m50s  presto-controller  Created Coordinator Config. coordinatorconfig-03f118d2
  Normal  Created  2m49s  presto-controller  Created Worker Config. workerconfig-03f118d2
  Normal  Created  2m48s  presto-controller  Created Catalog Config. catalogconfig-03f118d2
  Normal  Created  2m48s  presto-controller  Created Coordinator Replicaset. coordinatorreplicaset-03f118d2
  Normal  Created  2m47s  presto-controller  Created Worker Replicaset. workerreplicaset-03f118d2zk768
```

Kubernetes API and `kubectl` command can be used to delete a Presto cluster

```bash
$ kubectl delete prestos mycluster
```
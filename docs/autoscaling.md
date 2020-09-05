# Autoscaling

Workers in the presto cluster can be scaled up and down based on the CPU utilization in the Presto cluster. Operator uses K8S's Horizontal Pod Autoscaling for scaling.

With autoscaling, initially the `spec.worker.count` number of workers are created. Based on `spec.worker.autoscaling.targetCPUUtilizationPercentage`, the workers are scaled up and down. The minium number of workers to which it scales down is `spec.worker.autoscaling.minReplicas`. The maximum number of workers to which it scales up is `spec.worker.autoscaling.maxReplicas`. 

A sample snippet of the Presto YAML. 
```bash
apiVersion: falarica.io/v1alpha1
kind: Presto
metadata:
  name: mycluster
spec:
 worker:
    memoryLimit: "1Gi"
    count: 1
    autoscaling:
      enabled: false
      minReplicas: 2
      maxReplicas: 3
      targetCPUUtilizationPercentage: 20
``` 


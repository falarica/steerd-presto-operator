# Autoscaling

Kubernetes' [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) (HPA) automatically scales the number of pods in a replication controller, deployment, replica set or stateful set based on observed CPU utilization. 

Presto Operator uses HPA to scale up and down Presto workers in the presto cluster based on the CPU utilization.

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


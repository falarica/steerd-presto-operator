# Services

There are two services created inside Kubernetes for the communication between and with Presto processes

- Headless Coordinator Discovery - This service is created so that workers can discover the coordinator using a hostname. This is internal to the operator and it cannot be configured by the user. 
- External service - This service is created to enable Presto Coordinator port from outside the cluster so that JDBC and REST clients can access the end point. 

External service can be configured to be either a ClusterIP, NodePort or LoadBalancer. By default, the external service is of type ClusterIP.

External service can be configured to be of type NodePort by defining it like this. 
```bash
spec:
  service:
    type: "NodePort"
    port: 8100
    nodePort: 30002
```
External service can be configured to be of type LoadBalancer by defining it like this.
```bash
spec:
  service:
    type: "LoadBalancer"
    port: 8100
```

All Kubernetes Service properties that one specifies  while defining a Kubernetes service can also be defined for external service. For e.g. `externalIPs`, `loadBalancerIP`, `loadBalancerSourceRanges`  


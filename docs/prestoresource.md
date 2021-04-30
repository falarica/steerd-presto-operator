# Presto Resource YAML

Presto cluster can be created by specifying the cluster properties as YAML and then using that YAML to create the Presto resource.

Following is an example of Presto resource YAML.

```bash
apiVersion: falarica.io/v1alpha1
kind: Presto
metadata:
  name: mycluster
spec:
  service:
    type: "NodePort"
    port: 8100
    nodePort: 30002
  catalogs:
    catalogSpec:
      - name: newtpch
        content:
          connector.name: tpch
      - name: newtpcds
        content:
          connector.name: tpcds
  coordinator:
    memoryLimit: "1Gi"
    cpuLimit: "0.5"
    httpsEnabled: false
    httpsKeyPairSecretName: "prestokeystore"
    httpsKeyPairSecretKey: "prestoserverkeystore.jks"
    httpsKeyPairPassword: "hemant"
  worker:
    memoryLimit: "1Gi"
    cpuLimit: "0.5"
    count: 1
    autoscaling:
      enabled: false
      minReplicas: 2
      maxReplicas: 3
      targetCPUUtilizationPercentage: 20
    additionalProps:
      shutdown.grace-period: 10s
  additionalPrestoPropFiles:
    access-control.properties:
      access-control.name=read-only
```
## Creating Presto Cluster

Presto cluster can be created using this YAML.

```bash
kubectl apply   -f  deploy/crds/falarica_prestodb.yaml
```

The main components of the YAML are service, catalogs, coordinator and worker. Coordinator and worker are used to specify the properties of coordinator and worker. All the workers in the cluster would have the same properties.

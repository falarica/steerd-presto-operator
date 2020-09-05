# SteerD Presto Kubernetes Operator

[Kubernetes](https://kubernetes.io/docs/home/) is an open source container orchestration engine for automating deployment, scaling, and management of containerized applications. [Operators](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) are software extensions to Kubernetes that make use of custom resources to manage applications and their components. Operators follow Kubernetes principles, notably the control loop.

SteerD Presto Operator is a Kubernetes Operator for Presto to manage Presto clusters which are deployed as custom resources. In short, the task of configuring, creating, managing, automatically scaling up and scaling-in of Presto cluster(s) in a Kubernetes environment has been made simple, easy and quick.

SteerD Presto Operator supports both **PrestoDB** and **PrestoSQL** forks. 

## Deploying Operator

### Deploying Operator - Locally
*Step 1:* Enable metrics server for k8s, if not already enabled. [See this](https://github.com/kubernetes-sigs/metrics-server). This is needed for horizontal pod autoscaling.

*Step 2:* Build the operator 
```bash
$ go build  -o steerd-presto-operator cmd/manager/main.go 
```

*Step 3:* Deploy the CRD
```bash
$  kubectl apply -f    deploy/crds/falarica.io_prestos_crd.yaml
```

*Step 4:* Start the controller with the right credentials
```bash
$ ./steerd-presto-operator -kubeconfig /home/hemant/.kube/config
```

### Deploying Operator - GKE

*Step 1:* Enable metrics server for GKE, if not already enabled. [See this](https://github.com/kubernetes-sigs/metrics-server). This is needed for horizontal pod autoscaling.

*Step 2:* Create Operator Image Using Google CloudBuild
```bash
$ docker/gcloudDockerBuild.sh
```
*Step 3:* Deploy the CRD
```bash
$  kubectl apply -f    deploy/crds/falarica.io_prestos_crd.yaml
```
*Step 4:* Update the Operator yaml with image name 
```bash
# Here gcr.io/fluid-tangent-249303/steerd-presto-operator is the name of image. 
# Replace it with your image name
$ sed -i 's/REPLACE_IMAGE/gcr.io\/fluid-tangent-249303\/steerd-presto-operator/g' deploy/operator.yaml
```

*Step 5:* Launch the operator
```bash
$ kubectl apply -f deploy/operator.yaml
```  
 
## Deploy Presto Cluster

Deploy the presto cluster
```bash
$ ## Deploy Presto SQL. For PrestoDB, refer deploy/crds/falarica_prestodb.yaml 
$ kubectl apply -f    deploy/crds/falarica.io_v1alpha1_presto_cr.yaml
```

## Further Details 

- [Creating Presto Cluster](docs/prestoresource.md)
- [Managing Presto Cluster](docs/status.md)
- [Autoscaling](docs/autoscaling.md)
- [Catalogs](docs/catalog.md)
- [Services](docs/service.md)
- [Additional Volumes](docs/additionalvolumes.md)
- [HTTPS Support](docs/https.md)
- [Caveats/Future Work](docs/caveats.md)

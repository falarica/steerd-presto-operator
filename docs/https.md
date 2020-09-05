# Presto Cluster - Enabling HTTPS 

Currently, when HTTPS is enabled for Presto server, operator will keep the internal communication still on HTTP. This ensures that the performance of the distributed joins and other processing where data is exchanged between nodes is not impacted. And since, Presto server runs inside kubernetes environment, the HTTP ports won't be visible to an external network

Presto cluster needs a server side keystore for enabling HTTPS. For accessing a HTTPS enabled Presto server, client needs to possess the certificate. For non-production environment, keytool can be used to generate the keyvalue pair. If you already have a key and certificate, the following is not needed. 

Generate a key and certificate for the prestoserver:
```bash 
# Replace PATH_FOR_KEYSTORE, PRESTO_KEYSTORE_PASSWORD, PRESTO_HOSTNAME in the command below
# PRESTO_HOSTNAME is needed while verifying the key by the client. 
# For test environments, specify any name and then update your /etc/hosts with this name pointing to presto server
# Server key would be stored in PATH_FOR_KEYSTORE as prestoserverkeystore.jks

keytool -genkeypair -alias prestoserver  -keyalg RSA -keystore PATH_FOR_KEYSTORE/prestoserverkeystore.jks  -storepass PRESTO_KEYSTORE_PASSWORD  -dname "CN=PRESTO_HOSTNAME"

# use the following command to generate the keypair if you want to use IP 
keytool -genkeypair -alias prestoserver    -ext SAN=IP:PRESTOSERVER_IP  -keyalg RSA -keystore PATH_FOR_KEYSTORE/prestoserverkeystore.jks   -storepass PRESTO_KEYSTORE_PASSWORD 


# Replace PATH_FOR_KEYSTORE, PRESTO_KEYSTORE_PASSWORD, PATH_FOR_CERTIFICATE in the command below
# Client certificate would be stored in PATH_FOR_CERTIFICATE as prestoserver_clientcertificate.jks

keytool -export -alias prestoserver -keystore PATH_FOR_KEYSTORE/prestoserverkeystore.jks  -storepass PRESTO_KEYSTORE_PASSWORD -rfc -file PATH_FOR_CERTIFICATE/prestoserver_clientcertificate.jks
```

## K8s Secret for HTTPS 

Firstly, server keystore file has to created as the K8s secret so that the Presto server can read it from there. 

```bash
kubectl create secret  generic  prestokeystore  --from-file=/tmp/etc/prestoserverkeystore.jks
```

## Presto Resource YAML - HTTPS properties 

Following properties now needs to be specified to create a HTTPS enabled presto server. 

```bash
spec.coordinator.httpsEnabled: Whether HTTPS should be enabled or not
spec.coordinator.httpsKeyPairSecretName: Name of the secret of the keystore file
spec.coordinator.httpsKeyPairSecretKey: Name of the file that has keystore in the secret. 
spec.coordinator.httpsKeyPairPassword: Password of the Keystore. This must have beeen specified while creating keystore 
```

For e.g.
```bash 
apiVersion: falarica.io/v1alpha1
kind: Presto
metadata:
  name: mycluster
spec:
coordinator:
    memoryLimit: "1Gi"
    cpuLimit: "0.5"
    httpsEnabled: false
    httpsKeyPairSecretName: "prestokeystore"
    httpsKeyPairSecretKey: "prestoserverkeystore.jks"
    httpsKeyPairPassword: "hemant"
....
```
Other HTTPS Server related properties can be specified as `spec.coordinator.additionalProps`
#Caveats/Future work

- Currently validations for Presto Resource yaml are not done. Validating admission webhooks are something that can be added for validating presto resource YAML
- Adding custom plugins to Presto is not supported/tested
- Currently we do not allow the internal communication between the workers and coordinator to be encrypted. This is a conscious decision because Kubernetes network is not exposed. 

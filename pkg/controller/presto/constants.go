package presto

const (
	nodePropertiesKey       = "node.properties"
	configPropertiesKey     = "config.properties"
	jvmConfigKey            = "jvm.config"
	prestoPort              = 8080
	mountPath               = "/etc/presto"
	httpsVolPath            = "/etc/httpssecret"
	prestoShutdownScript    = "presto_shutdown.sh"
	DefaultTerminationGracePeriodSeconds = 7200
	catalogFileSuffix       = ".properties"
	catalogMountPath        = "/catalog/"
	// script called during shutdown. Picked from OneOneStar repo https://gist.github.com/oneonestar/ea75a608d58aa7e40cc952ad20e5a31a
	// Have made it a string so that a separate file is not needed at the run time.
	// the string has to be formatted to pass the mountpath of config.properties
	shutdownScriptContent = `
#!/bin/bash
# This works for worker only. Coordinator doesn't support graceful shutdown.
# This script will block until the server has actually shutdown.
set -x
http_port="$(cat {MOUNT_PATH}/config.properties | grep 'http-server.http.port' | sed 's/^.*=\(.*\)$/\1/')"
https_port="$(cat {MOUNT_PATH}/config.properties | grep 'http-server.https.port' | sed 's/^.*=\(.*\)$/\1/')"

if [ -n "$http_port" ] ; then
    res=$(curl -s -o /dev/null -w "%{http_code}"  -XPUT --data '"SHUTTING_DOWN"' -H "Content-type: application/json" http://localhost:${http_port}/v1/info/state)
fi

if [ -z "$res" -o "$res" != "200" ] && [ -n "$https_port" ]; then
    res=$(curl -k -s -o /dev/null -w "%{http_code}"  -XPUT --data '"SHUTTING_DOWN"' -H "Content-type: application/json" https://localhost:${https_port}/v1/info/state)
fi

if [ -z "$res" -o "$res" != "200" ] ; then
  # Failed to send the shutdown request.
  exit -1
else
  # Server is shutting down. Block until the server is actually down.
  while curl http://localhost:${http_port}/v1/info/state; do
    sleep 3
  done
fi
`
)

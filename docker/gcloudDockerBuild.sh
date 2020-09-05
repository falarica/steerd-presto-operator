#!/usr/bin/env bash
SCRIPTPATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
gcloud builds submit --config $SCRIPTPATH/steerdpresto.yaml $SCRIPTPATH/..

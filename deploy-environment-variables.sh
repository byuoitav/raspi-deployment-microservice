#!/bin/bash

# Get all the environment variables
printenv > /tmp/environment-variables-all

# Find only the custom variables
grep -v -F -x -f /tmp/environment-variables-circle /tmp/environment-variables-all > /tmp/environment-variables

# Remove remnant variables from Circle
sed '/GOPATH/d' environment-variables > tmpfile; mv tmpfile environment-variables
sed '/SHLVL/d' environment-variables > tmpfile; mv tmpfile environment-variables
sed '/PWD/d' environment-variables > tmpfile; mv tmpfile environment-variables

aws s3 cp /tmp/environment-variables $AWS_BUCKET_ADDRESS

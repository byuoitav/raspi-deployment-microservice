#!/bin/bash

# Get all the environment variables
printenv > /tmp/environment-variables-all

# Find only the custom variables
grep -v -F -x -f /tmp/environment-variables-circle /tmp/environment-variables-all > /tmp/environment-variables

# Remove remnant variables from Circle
sed '/GOPATH/d' environment-variables > environment-variables
sed '/SHLVL/d' environment-variables > environment-variables
sed '/PWD/d' environment-variables > environment-variables

aws s3 cp /tmp/environment-variables-circle $AWS_BUCKET_ADDRESS
aws s3 cp /tmp/environment-variables-all $AWS_BUCKET_ADDRESS
aws s3 cp /tmp/environment-variables $AWS_BUCKET_ADDRESS

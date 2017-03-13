#!/bin/bash

# Get all the environment variables
printenv > /tmp/environment-variables-all

# Find only the custom variables
grep -v -F -x -f /tmp/environment-variables-circle /tmp/environment-variables-all > /tmp/environment-variables

# Remove remnant variables from Circle
sed '/GOPATH/d' /tmp/environment-variables > /tmp/tmpfile; mv /tmp/tmpfile /tmp/environment-variables
sed '/SHLVL/d' /tmp/environment-variables > /tmp/tmpfile; mv /tmp/tmpfile /tmp/environment-variables
sed '/PWD/d' /tmp/environment-variables > /tmp/tmpfile; mv /tmp/tmpfile /tmp/environment-variables
sed '/PWD/d' /tmp/environment-variables > /tmp/tmpfile; mv /tmp/tmpfile /tmp/environment-variables

# Remove environment variables we don't need on the Pi's that are needed in Circle to deploy
sed '/RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS/d' /tmp/environment-variables > /tmp/tmpfile; mv /tmp/tmpfile /tmp/environment-variables
sed '/RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER/d' /tmp/environment-variables > /tmp/tmpfile; mv /tmp/tmpfile /tmp/environment-variables

aws s3 cp /tmp/environment-variables $AWS_BUCKET_ADDRESS

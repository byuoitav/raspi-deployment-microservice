#!/bin/bash

# Get all the environment variables
printenv > environment-variables-all

grep -v -F -x -f environment-variables-circle environment-variables-all > environment-variables

aws s3 cp environment-variables $AWS_BUCKET_ADDRESS

#!/bin/bash

printenv > environment-variables

aws s3 cp environment-variables $AWS_BUCKET_ADDRESS

#!/bin/bash 

#get environment variables
curl https://sandbag.byu.edu:2000/deploy/$(hostname)

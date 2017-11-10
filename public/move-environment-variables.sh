#!/bin/bash

echo PI_HOSTNAME=$(cat /etc/hostname) >> ~/.environment-variables
sudo mv ~/.environment-variables /etc/environment


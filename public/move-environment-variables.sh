#!/bin/bash

sudo cat ~/.environment-variables > /etc/environment

echo PI_HOSTNAME=$(cat /etc/hostname) >> /etc/environment

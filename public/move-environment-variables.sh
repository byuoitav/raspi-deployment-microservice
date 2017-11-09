#!/bin/bash

sudo cat ~/.environment-variables > /etc/environment

sudo echo PI_HOSTNAME=$(cat /etc/hostname) >> /etc/environment

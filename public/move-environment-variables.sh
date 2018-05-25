#!/bin/bash

echo "export PI_HOSTNAME=\"$(cat /etc/hostname)\"" >> ~/.environment-variables
sudo mv ~/.environment-variables /etc/environment


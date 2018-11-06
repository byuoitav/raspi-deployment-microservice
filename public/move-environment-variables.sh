#!/bin/bash

echo "export PI_HOSTNAME=\"$(cat /etc/hostname)\"" >> ~/.environment-variables
echo "export ROOM_SYSTEM=\"$(cat /etc/hostname)\"" >> ~/.environment-variables
echo "export SYSTEM_ID=\"$(cat /etc/hostname)\"" >> ~/.environment-variables
sudo mv ~/.environment-variables /etc/environment



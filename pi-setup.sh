#!/usr/bin/env bash

# This script is used to install and set up dependencies on a newly wiped/installed Raspberry Pi
# For clean execution, run this script inside of the /tmp directory with `./pi-setup.sh`
# The script assumes the username of the autologin user is "pi"

# Run the `sudo.sh` code block to install necessary packages and commands
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/sudo.sh > /tmp/sudo.sh
chmod +x /tmp/sudo.sh
sudo sh -c "bash /tmp/sudo.sh"

# Make `startx` result in starting the i3 window manager
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/xinitrc > /home/pi/.xinitrc
chmod +x /home/pi/.xinitrc

# Configure i3
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/i3_config > /home/pi/.i3/config

# Install an ARM-specific Docker version
curl -sSL http://downloads.hypriot.com/docker-hypriot_1.10.3-1_armhf.deb > /tmp/docker-hypriot_1.10.3-1_armhf.deb
sudo sh -c "dpkg -i /tmp/docker-hypriot_1.10.3-1_armhf.deb; usermod -aG docker pi; systemctl enable docker.service"

# Make X start on login
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/bash_profile > /home/pi/.bash_profile

sudo sh -c "reboot"

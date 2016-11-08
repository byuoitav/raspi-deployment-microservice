#!/usr/bin/env bash

# This script is used to install and set up dependencies on a newly wiped/installed Raspberry Pi
# For clean execution, run this script inside of the /tmp directory with `./pi-setup.sh`
# The script assumes the username of the autologin user is "pi"

# Run the `sudo.sh` code block to install necessary packages and commands
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/scripts/sudo.sh > sudo.sh
chmod +x sudo.sh
sudo sh -c "bash sudo.sh"

# Install docker-compose
curl -L https://github.com/docker/compose/releases/download/1.9.0-rc3/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Download and run Docker containers
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/docker-compose.yml > docker-compose.yml
docker-compose up -d

# Make `startx` result in starting the Awesome window manager
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/xinitrc > /home/pi/.xinitrc
chmod +x /home/pi/.xinitrc

# Copy the default Awesome config
rm -rf /home/pi/.config
mkdir /home/pi/.config
cp -r /etc/xdg/awesome/ /home/pi/.config/awesome/

# Make Awesome start Chromium on boot
echo "awful.util.spawn_with_shell('chromium-browser --kiosk http://localhost:8888')" >> /home/pi/.config/awesome/rc.lua

# Install an ARM-specific Docker version
curl -sSL http://downloads.hypriot.com/docker-hypriot_1.10.3-1_armhf.deb > /tmp/docker-hypriot_1.10.3-1_armhf.deb
sudo sh -c "dpkg -i /tmp/docker-hypriot_1.10.3-1_armhf.deb; usermod -aG docker pi; systemctl enable docker.service"

# Make X start on login
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/bash_profile > /home/pi/.bash_profile

sudo sh -c "reboot"

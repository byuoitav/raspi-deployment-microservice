#!/usr/bin/env bash

# This script is used to install and set up dependencies on a newly wiped/installed Raspberry Pi
# Run with `./pi-setup.sh`
# The script assumes the username of the autologin user is "pi"

# Set the proper keyboard layout, update everything, enable autologin, and install our GUI dependencies
sudo sh -c "curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/keyboard > /etc/default/keyboard; apt-get update; apt-get -y upgrade; apt-get -y dist-upgrade; apt-get -y autoremove; apt-get -y autoclean; apt-get -y install xorg awesome chromium-browser; mkdir -pv /etc/systemd/system/getty@tty1.service.d/; curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/autologin.conf > /etc/systemd/system/getty@tty1.service.d/autologin.conf; systemctl enable getty@tty1.service"

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
dpkg -i /tmp/docker-hypriot_1.10.3-1_armhf.deb
rm -f /tmp/docker-hypriot_1.10.3-1_armhf.deb
sh -c 'usermod -aG docker $SUDO_USER'
systemctl enable docker.service

# Make X start on login
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/bash_profile > /home/pi/.bash_profile

reboot

#!/usr/bin/env bash

# Set the proper keyboard layout
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/keyboard > /etc/default/keyboard

# Update everything
apt-get update
apt-get -y upgrade
apt-get -y dist-upgrade
apt-get -y autoremove
apt-get -y autoclean

# Install frontend pieces
apt-get install -y xorg
apt-get install -y awesome
apt-get install -y chromium-browser

# Make `startx` result in starting the Awesome window manager
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/xinitrc > /home/pi/.xinitrc
chmod +x /home/pi/.xinitrc
chown pi /home/pi/.xinitrc
chgrp pi /home/pi/.xinitrc

# Copy the default Awesome config
rm -rf /home/pi/.config
mkdir /home/pi/.config
cp -r /etc/xdg/awesome/ /home/pi/.config/awesome/
chown -R pi /home/pi/.config
chgrp -R pi /home/pi/.config

# Make Awesome start Chromium on boot
echo "awful.util.spawn_with_shell('chromium-browser --kiosk http://localhost:8888')" >> /home/pi/.config/awesome/rc.lua

# Install an ARM-specific Docker version
curl -sSL http://downloads.hypriot.com/docker-hypriot_1.10.3-1_armhf.deb > /tmp/docker-hypriot_1.10.3-1_armhf.deb
dpkg -i /tmp/docker-hypriot_1.10.3-1_armhf.deb
rm -f /tmp/docker-hypriot_1.10.3-1_armhf.deb
sh -c 'usermod -aG docker $SUDO_USER'
systemctl enable docker.service

# Enable autologin
mkdir -pv /etc/systemd/system/getty@tty1.service.d/
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/autologin.conf > /etc/systemd/system/getty@tty1.service.d/autologin.conf
systemctl enable getty@tty1.service

# Make X start on login
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/bash_profile > /home/pi/.bash_profile
chown pi /home/pi/.bash_profile
chgrp pi /home/pi/.bash_profile

reboot

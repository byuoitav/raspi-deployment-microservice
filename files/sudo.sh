#!/usr/bin/env bash

# This script is called automatically by `pi-setup.sh` to run a batch of Pi setup commands that require sudo permissions

echo "Type the desired hostname of this device (E.g. ITB-1006-CP2), followed by [ENTER]:"

read desired_hostname

echo $desired_hostname > /etc/hostname

# Fix the keyboard layout
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/keyboard > /etc/default/keyboard

# Perform general updating
apt-get update
apt-get -y upgrade
apt-get -y dist-upgrade
apt-get -y autoremove
apt-get -y autoclean

# Patch the Dirty COW kernel vulnerability
apt-get -y install raspberrypi-kernel 

# Install UI dependencies
apt-get -y install xorg i3 suckless-tools chromium-browser

# Install an ARM build of docker-compose
apt-get install docker-compose

# Configure automatic login for the `pi` user
mkdir -pv /etc/systemd/system/getty@tty1.service.d/
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/autologin.conf > /etc/systemd/system/getty@tty1.service.d/autologin.conf
systemctl enable getty@tty1.service

# Rotate the Pi's screen 180 degrees
echo "lcd_rotate=2" >> /boot/config.txt

# Enable SSH connections
touch /boot/ssh

# Set the timezone
cp /usr/share/zoneinfo/America/Denver /etc/localtime

# Add the `pi` user to the sudoers group
usermod -aG sudo pi


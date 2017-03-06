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

# Add the Hyperiot repository to our sources
apt-get -y install apt-transport-https 
echo "deb https://packagecloud.io/Hypriot/Schatzkiste/debian/ jessie main" | sudo tee /etc/apt/sources.list.d/hypriot.list
apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 37BBEE3F7AD95B3F
apt-get update

# Add the i3 repository to our sources
echo "deb http://debian.sur5r.net/i3/ $(lsb_release -c -s) universe" >> /etc/apt/sources.list
apt-get update
apt-get --allow-unauthenticated install sur5r-keyring
apt-get update

# Patch the Dirty COW kernel vulnerability
apt-get -y install raspberrypi-kernel 

# Install UI dependencies
apt-get -y install xorg i3 chromium-browser

# Install an ARM build of docker-compose
apt-get install docker-compose

# Configure automatic login for the `pi` user
mkdir -pv /etc/systemd/system/getty@tty1.service.d/
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/autologin.conf > /etc/systemd/system/getty@tty1.service.d/autologin.conf
systemctl enable getty@tty1.service

# Rotate the Pi's screen 180 degrees
echo "lcd_rotate=2" >> /boot/config.txt

# Add the `pi` user to the sudoers group
usermod -aG sudo pi


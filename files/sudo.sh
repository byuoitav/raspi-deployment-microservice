#!/usr/bin/env bash

# This script is called automatically by `pi-setup.sh` to run a batch of Pi setup commands that require sudo permissions

echo "Type the desired hostname of this device (E.g. ITB-1006-CP2), followed by [ENTER]:"

read desired_hostname

echo $desired_hostname > /etc/hostname

# get static ip
echo "Type the desired static ip-address of this device (E.g. 10.5.99.18), followed by [ENTER]:"

read desired_ip

echo "interface eth0" >> /etc/dhcpcd.conf
echo "static ip_address=$desired_ip/24" >> /etc/dhcpcd.conf
routers=$(echo "static routers=$desired_ip" | cut -d "." -f -3)
echo "$routers.1" >> /etc/dhcpcd.conf
echo "static domain_name_servers=10.8.0.19, 10.8.0.26" >> /etc/dhcpcd.conf

# Fix the keyboard layout
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/keyboard > /etc/default/keyboard

# Perform general updating
apt update
apt -y upgrade
apt -y dist-upgrade
apt -y autoremove
apt -y autoclean

# Patch the Dirty COW kernel vulnerability
apt -y install raspberrypi-kernel 

# Install UI dependencies
apt -y install xorg i3 suckless-tools chromium-browser

# Install an ARM build of docker-compose
apt -y install python-pip
easy_install --upgrade pip
pip install docker-compose

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


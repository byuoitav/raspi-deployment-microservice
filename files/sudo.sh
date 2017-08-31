#!/usr/bin/env bash

# This script is called automatically by `pi-setup.sh` to run a batch of Pi setup commands that require sudo permissions
started="/usr/local/games/setup-started"

ip=`ip addr | grep 'state UP' -A2 | tail -n1 | awk '{print $2}' | cut -f1  -d'/'`
echo "\n\nmy ip address: $ip\n\n"

# Update the time (from google, to ensure https works)
date -s "$(wget -qSO- --max-redirect=0 google.com 2>&1 | grep Date: | cut -d' ' -f5-8)Z"

# Fix the keyboard layout
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/keyboard > /etc/default/keyboard

echo "Type the desired hostname of this device (E.g. ITB-1006-CP2), followed by [ENTER]:"

read -e desired_hostname

# get static ip
echo "Type the desired static ip-address of this device (e.g. $ip), followed by [ENTER]:"

read -e desired_ip

# check if script has already been started
if [ -f "$started" ]; then
	touch $started
	echo "setup has been started remotely"
	echo "please wait for me to reboot (about 30 minutes)"
	sleep infinity
	exit
fi

# copy original dhcp file
cp /etc/dhcpcd.conf /etc/dhcpcd.conf.other

# setup hostname
echo $desired_hostname > /etc/hostname
echo "127.0.1.1    $desired_hostname" >> /etc/hosts

# setup static ip
echo "interface eth0" >> /etc/dhcpcd.conf
echo "static ip_address=$desired_ip/24" >> /etc/dhcpcd.conf
routers=$(echo "static routers=$desired_ip" | cut -d "." -f -3)
echo "$routers.1" >> /etc/dhcpcd.conf
echo "static domain_name_servers=10.8.0.19 10.8.0.26" >> /etc/dhcpcd.conf

# set contact points up
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/image/contacts.service > /usr/lib/systemd/system/contacts.service
chmod 664 /usr/lib/systemd/system/contacts.service
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/contacts.py > /usr/bin/contacts.py
chmod 775 /usr/bin/contacts.py
systemctl daemon-reload

# set up screenshutoff
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/screenshutoff-setup.sh > /tmp/sss-setup.sh
chmod +x /tmp/sss-setup.sh
sh -c "bash /tmp/sss-setup.sh"

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

# Enable SSH connections
touch /boot/ssh

# Set the timezone
cp /usr/share/zoneinfo/America/Denver /etc/localtime

# Add the `pi` user to the sudoers group
usermod -aG sudo pi

# set to update from byu servers
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/ntp.conf > /etc/ntp.conf
apt -y install ntpdate
systemctl stop ntp
ntpdate-debian
systemctl start ntp
ntpq -p

#!/bin/bash

if [ "$EUID" -ne 0 ]; then
	echo "Must be run as root/sudo."
	exit 1
fi

# get binary for xssstart
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/rpi-xssstart > /usr/local/bin/xssstart 
chmod +x /usr/local/bin/xssstart

# get start script
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/screenshutoff.sh > /usr/local/bin/screenshutoff
chmod +x /usr/local/bin/screenshutoff

# make script run when x server starts
#curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/xinitrc > /home/pi/.xinitrc

echo "You're all good to go!"

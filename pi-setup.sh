#!/usr/bin/env bash

apt-get update
apt-get -y upgrade
apt-get -y dist-upgrade
apt-get -y autoremove
apt-get -y autoclean

apt-get install -y xorg
apt-get install -y openbox
apt-get install -y chromium-browser

curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/xinitrc > ~/.xinitrc
chmod +x ~/.xinitrc

mkdir ~/.config/
cp -r /etc/xdg/awesome/ ~/.config/awesome/

curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/rc.lua > ~/.config/awesome/rc.lua
chmod +x ~/.config/awesome/rc.lua

startx

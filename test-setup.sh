#!/usr/bin/env bash

apt-get update
apt-get -y upgrade
apt-get -y dist-upgrade
apt-get autoclean

apt-get install -y awesome
apt-get install -y chromium-browser

touch ~/.xinitrc
echo #!/bin/sh >> ~/.xinitrc

cp -r /etc/xdg/awesome/ ~/.config/awesome/
echo awful.util.spawn("chromium-browser --kiosk `http://www.jessemillar.com/`") >> ~/.config/awesome/rc.lua
startx

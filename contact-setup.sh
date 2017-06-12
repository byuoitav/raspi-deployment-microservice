#!/bin/bash

read -r -p "Will this device be monitoring contact points? [y/N]: " response

case "$response" in 
	[yY][eE][sS]|[yY])
		;;
	*)
	echo "contact points not enabled"
		exit 0
	;;
esac
	
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/image/contacts.service > /usr/lib/systemd/system/contacts.service
chmod 664 /usr/lib/systemd/system/contacts.service

curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/contacts.py > /usr/bin/contacts.py
chmod 775 /usr/bin/contacts.py

systemctl daemon-reload
systemctl enable contacts

echo "Contact points enabled"



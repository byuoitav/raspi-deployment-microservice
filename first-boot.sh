#!/bin/bash
# This script should live in /usr/bin/ on the rasbian img. It will run once on the first boot of the pi, and then disable the service.

sleep 15

echo ""
echo ""
echo "Hi from Danny"
echo ""
echo ""

chvt 2

echo "Starting $0"

bootfile="/usr/local/games/firstboot"

if [ -f "$bootfile" ]; then
	echo "First boot."
	sudo rm $bootfile

	# download pi-setup
	curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/pi-setup.sh > /tmp/pi-setup.sh
	chmod +x /tmp/pi-setup.sh
	/tmp/pi-setup.sh
else
	echo "Second boot."

	curl -sSL https://get.docker.com -k | sh
	wait
	sudo usermod -aG docker pi

	printf "Please trigger a build to get the necessary environment variables.\n"
	printf "Waiting...\n"

	# wait for env. variables
	modtime=$(stat -c %Y /etc/environment)
	printf "original mod time to /etc/environment: $modtime"
	newtime=$(stat -c %Y /etc/environment)
	until [ "$modtime" != "$newtime" ]; do
		printf "\tnew mod time to /etc/environment: $newtime"
		sleep 5
	done

	printf "recieved env. variables\n"

	curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/mariadb-setup.sh > /tmp/mariadb-setup.sh
	chmod +x /tmp/mariadb-setup.sh
	/tmp/mariadb-setup.sh

	echo "Removing symlink to startup script."
	sudo systemctl disable first-boot.service
fi

exit 0

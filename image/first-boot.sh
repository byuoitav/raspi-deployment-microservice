#!/bin/bash
# This script should live in /usr/bin/ on the rasbian img. It will run once on the first boot of the pi, and then disable the service.

sleep 15

printf "\n\nHi From Danny\n\n"

sudo chvt 2

bootfile="/usr/local/games/firstboot"

if [ -f "$bootfile" ]; then
	echo "First boot."

	# download pi-setup
	until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/pi-setup.sh > /tmp/pi-setup.sh); do
		echo "Trying again."
	done
	chmod +x /tmp/pi-setup.sh

	echo "Removing first boot file."
	sudo rm $bootfile

	/tmp/pi-setup.sh
else
	echo "Second boot."

	wait 45 
	sudo chvt 2

	until $(sudo usermod -aG docker pi); do
		curl -sSL https://get.docker.com -k | sh
		wait
	done
	echo "Added user pi to the docker group"

	printf "Please trigger a build to get the necessary environment variables.\n"
	printf "Waiting...\n"

	# wait for env. variables
	modtime=$(stat -c %Y /etc/environment)
	printf "original mod time to /etc/environment: $modtime"
	newtime=$(stat -c %Y /etc/environment)
	until [ "$modtime" != "$newtime" ]; do
		newtime=$(stat -c %Y /etc/environment)
		printf "\tnew mod time to /etc/environment: $newtime"
		sleep 10
	done

	printf "\nrecieved env. variables\n"
	source /etc/environment

	until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/mariadb-setup.sh > /tmp/mariadb-setup.sh); do
		echo "Trying again."
	done
	chmod +x /tmp/mariadb-setup.sh
	/tmp/mariadb-setup.sh

	until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/salt-setup.sh > /tmp/salt-setup.sh); do
		echo "Trying again."
	done
	chmod +x /tmp/salt-setup.sh

	until [ -d "/etc/salt/" ]; do
		/tmp/salt-setup.sh
	done

	echo "Removing symlink to startup script."
	sudo rm /usr/lib/systemd/system/default.target.wants/first-boot.service
fi

printf "\n\n\n\n\n"
echo "Setup complete! I'll never see you again. Bye lol"
sleep 30

sudo sh -c "reboot"


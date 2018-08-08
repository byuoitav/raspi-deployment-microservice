#!/bin/bash
# This script should live in /usr/bin/ on the rasbian img. It will run once on the first boot of the pi, and then disable the service.

printf "\n\nHi From Danny\n\n"

sudo chvt 2

bootfile="/usr/local/games/firstboot"

if [ -f "$bootfile" ]; then
	echo "First boot."

	# download pi-setup script
	until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/pi-setup.sh > /tmp/pi-setup.sh); do
		echo "Trying again."
	done
	chmod +x /tmp/pi-setup.sh

	/tmp/pi-setup.sh
	
else
	sleep 30
	sudo chvt 2

	echo "Second boot."

	# download second-boot script
	until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/second-boot.sh > /tmp/second-boot.sh); do
		echo "Trying again."
	done
	chmod +x /tmp/second-boot.sh

	/tmp/second-boot.sh

	echo "Removing symlink to startup script."
	sudo rm /usr/lib/systemd/system/default.target.wants/first-boot.service
    sleep 3
fi

clear
printf "\n\n\n\n\n\n"
printf "Setup complete! I'll never see you again."
printf "\n\tPlease wait for me to reboot.\n"
sleep 10
printf "\n\nBye lol"
sleep 3

sudo sh -c "reboot"

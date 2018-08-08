#!/bin/bash
# This script should live in /usr/bin/ on the rasbian img. It will run once on the first boot of the pi, and then disable the service.

sleep 15

printf "\n\nHi From Danny\n\n"

sudo chvt 2

bootfile="/usr/local/games/firstboot"
resizefile="/usr/local/games/resize"

if [ -f "$resizefile" ]; then
    echo "0th boot. resizing /var partition"
    sleep 3

	until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/image/resizevar > /tmp/resizevar); do
		echo "Downloading resize script"
	done
	chmod +x /tmp/resizevar

    sudo /tmp/resizevar

    # make sure it doesn't get past here
    sudo reboot
fi

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
	sudo systemctl disable first-boot.service
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

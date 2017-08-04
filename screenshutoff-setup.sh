if [ "$EUID" -ne 0 ]; then
	echo "Must be run as root/sudo."
	exit 1
fi

# get binary for xssstart
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/rpi-xssstart > /usr/local/bin/xssstart 
chmod +x /usr/local/bin/xssstart

# get start script
curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/screenshutoff.sh > /home/pi/screenshutoff.sh 
chmod +x /home/pi/screenshutoff.sh

echo "Please reboot device for changes to take effect."

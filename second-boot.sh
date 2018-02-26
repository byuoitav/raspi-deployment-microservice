#!/bin/bash 

echo "Second boot."

until $(sudo usermod -aG docker pi); do
	curl -sSL https://get.docker.com -k | sh
	wait
done
echo "Added user pi to the docker group"

# get environment variables
echo "getting environment variables..."
until curl http://sandbag.byu.edu:2001/deploy/$(hostname); do 
	echo "trying again..."
done

until [ $PI_HOSTNAME ]; do
	echo "PI_HOSTNAME not set"
	source /etc/environment
	sleep 5 
done

printf "\nrecieved env. variables\n"

# maria db setup
until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/mariadb-setup.sh > /tmp/mariadb-setup.sh); do
	echo "Trying again."
done
chmod +x /tmp/mariadb-setup.sh

/tmp/mariadb-setup.sh

# salt setup
until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/salt-setup.sh > /tmp/salt-setup.sh); do
	echo "Trying again."
done
chmod +x /tmp/salt-setup.sh

until [ -f "/etc/salt/setup" ]; do
	/tmp/salt-setup.sh
	wait
done

# docker 
until [ $(docker ps -q | wc -l) -gt 1 ]; do
	echo "Waiting for docker containers to download"
	sleep 10
done

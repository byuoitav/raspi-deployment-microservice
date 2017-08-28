#!/bin/bash 

echo "Second boot."

until $(sudo usermod -aG docker pi); do
	curl -sSL https://get.docker.com -k | sh
	wait
done
echo "Added user pi to the docker group"

# get environment variables
echo "Getting environment variables"
until $(curl http://sandgrains.byu.edu/$(hostname)); do 
	echo "Trying again"
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

# salt setup
until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/salt-setup.sh > /tmp/salt-setup.sh); do
	echo "Trying again."
done
chmod +x /tmp/salt-setup.sh

until [ -d "/etc/salt/" ]; do
	/tmp/salt-setup.sh
done

# docker 
until [ $(docker ps -q | wc -l) -gt 1 ]; do
	echo "Waiting for docker containers to download"
	sleep 10
done

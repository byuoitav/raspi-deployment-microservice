#!/bin/bash

printf "\n\nstarting run...\n\n"

until $(curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/pi-setup.sh > /tmp/first-boot.sh); do
    echo "Cannot download first-boot"
done

chmod +x /tmp/first-boot.sh

/tmp/firt-boot.sh | tee -a /etc/first-boot/logs.log

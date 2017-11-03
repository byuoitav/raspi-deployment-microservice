#!/usr/bin/env bash

if [ "$EUID" -ne 0 ]; then
	echo "Please run as root."
	exit 1
fi

old_hostname=$HOSTNAME
read -p "This pi is currently set up for room $HOSTNAME. Are you sure you want to change it? (Y/n) " -n 1 response
echo ""
case "$response" in
	[yY])
		;;
	*)
		echo "Ok. Bye then."
		exit 1
		;;
esac

printf "Type the new hostname of this device (e.g ITB-1106-CP2), followed by [ENTER]:\n"
read -e new_hostname

printf "\nType the new static ip-address of this device (e.g. 10.5.99.18), followed by [ENTER]:\n"
read -e new_ip

# update hostname
echo $new_hostname > /etc/hostname 
head -n -1 /etc/hosts > temp.txt && mv temp.txt /etc/hosts
echo "127.0.1.1      $new_hostname" >> /etc/hosts

# update the salt minion
sed -i -e "s/$old_hostname/$new_hostname/g" /etc/salt/minion

# update ip
head -n -4 /etc/dhcpcd.conf > temp.txt && mv temp.txt /etc/dhcpcd.conf
echo "interface eth0" >> /etc/dhcpcd.conf
echo "static ip_address=$new_ip/24" >> /etc/dhcpcd.conf
routers=$(echo "static routers=$new_ip" | cut -d "." -f -3)
echo "$routers.1" >> /etc/dhcpcd.conf
echo "static domain_name_servers=10.8.0.19 10.8.0.26" >> /etc/dhcpcd.conf

reboot

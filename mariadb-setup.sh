#!/bin/bash

#Ensure that our environment Variables are set
source /etc/environment
if [ -z "$CONFIGURATION_DATABASE_PASSWORD" ]; then
    echo "Environment Variables not available in script"
    return 1
fi
if [ -z "$CONFIGURATION_DATABASE_USERNAME" ]; then
    echo "Environment Variables not available in script"
    return 1
fi

#------
#Figure out how to set password in automated way.
#-----
sudo export DEBIAN_FRONTEND=noninteractive
sudo debconf-set-selections <<< "maria-db mysql-server/root_password password root
sudo debconf-set-selections <<< "maria-db mysql-server/root_password_again password install

sudo apt-get install mariadb-server mariadb-client -y

#-----
#Set Server ID
#-----

command=$(python -c "a = '$HOSTNAME'; a = a.split('-'); command = 'CALL getIDByHostName(\'' + a[0] + '\',\'' + a[1] + '\',\'' + a[2]+ '\');'; print command")

server_id=$(mysql -f -N --user=$CONFIGURATION_DATABASE_USERNAME --password=$CONFIGURATION_DATABASE_PASSWORD --host=$CONFIGURATION_DATABASE_REPLICATION_HOST configuration -e "$command")

#create my.cnf
echo "[mariadb]" | sudo tee /etc/my.cnf
echo "server_id=$server_id" | sudo tee --append /etc/my.cnf

#restart service
sudo service mysql restart

mysqldump --dump-slave --master-data --gtid --password=$CONFIGURATION_DATABASE_PASSWORD --user=root --host=$CONFIGURATION_DATABASE_REPLICATION_SETUP_HOST --all-databases > /tmp/dump.sql

mysql -f --user=root --password=install < /tmp/dump.sql

mysql -f --user=root --password=install -e "FLUSH PRIVILEGES"

#Set Master
mysql -f --user=$CONFIGURATION_DATABASE_USERNAME --password=$CONFIGURATION_DATABASE_PASSWORD -e "CHANGE MASTER TO master_host='$CONFIGURATION_DATABASE_REPLICATION_HOST', master_port=3306, master_user='$CONFIGURATION_DATABASE_USERNAME', master_password='$CONFIGURATION_DATABASE_PASSWORD', master_use_gtid=slave_pos;"

#START SLAVE;
mysql --user=$CONFIGURATION_DATABASE_USERNAME --password=$CONFIGURATION_DATABASE_PASSWORD -e 'START SLAVE';

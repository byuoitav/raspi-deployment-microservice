#!/bin/bash

#------
#Figure out how to set password in automated way. 
#-----

sudo debconf-set-selections <<< "maria-db mysql-server/root_password password $CONFIGURATION_DATABASE_PASSWORD"
sudo debconf-set-selections <<< "maria-db mysql-server/root_password_again password $CONFIGURATION_DATABASE_PASSWORD"

sudo apt-get install mariadb-server mariadb-client -y

#-----
#Set Server ID
#-----

mysqldump --dump-slave --master-data --gtid --password=$CONFIGURATION_DATABASE_PASSWORD --user=root --host=$CONFIGURATION_DATABASE_REPLICATION_HOST configuration > /tmp/dump.sql

mysql -f --user=$CONFIGURATION_DATABASE_USERNAME --password=$CONFIGURATION_DATABASE_PASSWORD < /tmp/dump.sql

mysql -f --user=$CONFIGURATION_DATABASE_USERNAME --password=$CONFIGURATION_DATABASE_PASSWORD -e "FLUSH PRIVILEGES"

#Set Master
#mysql -f --user=$CONFIGURATION_DATABASE_USERNAME --password=$CONFIGURATION_DATABASE_PASSWORD -e "CHANGE MASTER TO master_host='$CONFIGURATION_DATABASE_REPLICATION_HOST', master_port=3306, master_user='$CONFIGURATION_DATABASE_USERNAME', master_password='$CONFIGURATION_DATABASE_PASSWORD', master_use_gtid=slave_pos;"

#START SLAVE;
#mysql --user=$CONFIGURATION_DATABASE_USERNAME --password= -e 'START SLAVE';

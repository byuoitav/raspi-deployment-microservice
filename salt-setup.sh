source /etc/environment

wget -O - http://repo.saltstack.com/apt/debian/8/armhf/2016.11/SALTSTACK-GPG-KEY.pub | sudo apt-key add -
echo "deb http://repo.saltstack.com/apt/debian/8/armhf/2016.11 jessie main" | sudo tee --append /etc/apt/sources.list.d/saltstack.list
sudo apt update
sudo apt -y install salt-minion 

#Get the Minion Addr
wget https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/minion

sed -i 's/\$SALT_MASTER_HOST/'$SALT_MASTER_HOST'/' minion
sed -i 's/\$SALT_MASTER_FINGER/'$SALT_MASTER_FINGER'/' minion
sed -i 's/\$PI_HOSTNAME/'$PI_HOSTNAME'/' minion
sed -i 's/\$ROOM_SYSTEM/'$ROOM_SYSTEM'/' minion

sudo mv minion /etc/salt/minion 

sudo setfacl -m u:pi:rwx /etc/salt/pki/minion/
sudo setfacl -m u:pi:rwx /etc/salt/pki/minion/*

sudo wget -O /usr/bin/salt-event-proxy https://github.com/byuoitav/salt-event-proxy/releases/download/v0.8.1/salt-event-proxy
sudo chmod +x /usr/bin/salt-event-proxy

sudo wget -O /usr/lib/systemd/system/salt-event-proxy.service https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/image/salt-event-proxy.service
sudo systemctl enable salt-event-proxy
sudo systemctl start salt-event-proxy
sudo iptables -A INPUT ! -s 127.0.0.1 -p tcp -m tcp --dport 7010 -j DROP

sudo systemctl restart salt-minion

sudo touch /etc/salt/setup

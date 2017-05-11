wget -O - https://repo.saltstack.com/apt/debian/8/armhf/2016.11/SALTSTACK-GPG-KEY.pub | sudo apt-key add -
echo "deb http://repo.saltstack.com/apt/debian/8/armhf/2016.11 jessie main" | sudo tee --append /etc/apt/sources.list.d/saltstack.list
sudo apt update
sudo apt -Y install salt-minion

#Get the Minion Addr
wget https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/minion

sed -i 's/\$SALT_MASTER_HOST/'$SALT_MASTER_HOST'/' minion
sed -i 's/\$SALT_MASTER_FINGER/'$SALT_MASTER_FINGER'/' minion

sudo mkdir /etc/salt
sudo mv minion /etc/salt/minion



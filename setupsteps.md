#Setup Steps to Prepare a pi

1. Image using the raspbian-lite image `https://www.raspberrypi.org/downloads/raspbian/`
1. Install docker from hyperiot
  ```
  curl -sSL http://downloads.hypriot.com/docker-hypriot_1.10.3-1_armhf.deb >/tmp/docker-hypriot_1.10.3-1_armhf.deb
  sudo dpkg -i /tmp/docker-hypriot_1.10.3-1_armhf.deb
  rm -f /tmp/docker-hypriot_1.10.3-1_armhf.deb
  sudo sh -c 'usermod -aG docker $SUDO_USER'
  sudo systemctl enable docker.service
  ```
1. Add Device into DB
1. git pu

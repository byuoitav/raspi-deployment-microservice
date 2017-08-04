curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/files/rpi-xssstart > /home/pi/xssstart 
chmod +x /home/pi/xssstart

export DISPLAY=:0

./home/pi/xssstart curl -X PUT http://localhost:8888 &

echo "Waiting for screenoff commands."

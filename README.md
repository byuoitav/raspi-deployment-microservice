# raspi-deployment-microservice
[![CircleCI](https://img.shields.io/circleci/project/byuoitav/raspi-deployment-microservice.svg)](https://circleci.com/gh/byuoitav/raspi-deployment-microservice) [![Apache 2 License](https://img.shields.io/hexpm/l/plug.svg)](https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/LICENSE)

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](http://byuoitav.github.io/swagger-ui/?url=https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/swagger.json)

## Setup
### Environment Variables
The following environment variables need to be set in Circle so the deployment functionality can SSH into the Raspberry Pi's: `PI_SSH_USERNAME`, `PI_SSH_PASSWORD`, `CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS`, `RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS`, and `ELK_ADDRESS`

Additionally, any environment variables the Pi's will need to function need to be set in the Circle web interface.

### Installation
The installation process is mostly automated and very easy
1. Make sure the JSON object representing the UI associated with the PI has been updated in the `ui-configuration` Github Repository
1. When prompted, enter the desired hostname of the Pi, e.g.
	```
	ITB-1101-CP1
	```
1. When prompted, enter the desired IP address of the Pi, e.g.

	```
	10.66.9.13
	```

1. Wait for the Pi to reboot twice, then it's good to go

### Contact Points
The Raspberry Pi's monitor contact closures on their GPIO pins with a python-based systemd service, which is automatically installed during the setup sequence. 
The Pi monitors on pin 7 and grounds pin 9


# raspi-deployment-microservice
[![CircleCI](https://img.shields.io/circleci/project/byuoitav/raspi-deployment-microservice.svg)](https://circleci.com/gh/byuoitav/raspi-deployment-microservice) [![Apache 2 License](https://img.shields.io/hexpm/l/plug.svg)](https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/LICENSE)

[![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](http://byuoitav.github.io/swagger-ui/?url=https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/master/swagger.json)

## Setup
### Environment Variables
The following environment variables need to be set in Circle so the deployment functionality can SSH into the Raspberry Pi's: `PI_SSH_USERNAME`, `PI_SSH_PASSWORD`, `CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS`, `RASPI_DEPLOYMENT_MICROSERVICE_ADDRESS`, and `ELK_ADDRESS`

Additionally, any environment variables the Pi's will need to function need to be set in the Circle web interface.

### Installation
1. Run the following command to install Docker on the Pi in question:

	```
	curl -sSL https://get.docker.com | sh
	```

1. Get `pi-setup.sh` and run it with the following commands:

	```
	curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/pi-setup.sh
	chmod +x pi-setup.sh
	./pi-setup.sh
	```

1. Trigger a deployment from Circle ("Rebuild" the `raspi-deployment-microservice`) to get the necessary environment variables onto the new Pi
1. When the Circle deployment finishes, get `mariadb-setup.sh` and run it with the following commands:

	```
	curl https://raw.githubusercontent.com/byuoitav/raspi-deployment-microservice/mariadb-setup.sh
	chmod +x mariadb-setup.sh
	./mariadb-setup.sh
	```

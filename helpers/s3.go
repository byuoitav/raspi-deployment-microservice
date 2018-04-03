package helpers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/fatih/color"
)

func GetClassAndDesignationID(role, designation string) (int64, int64, error) {

	if (len(role) == 0) || (len(designation) == 0) {
		return 0, 0, errors.New("invalid role or designation")
	}

	//get role ID
	roleId, err := GetRoleId(role)
	if err != nil {
		msg := fmt.Sprintf("role ID not found: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, 0, errors.New(msg)
	}

	//get designation ID
	desigId, err := GetDesignationId(designation)
	if err != nil {
		msg := fmt.Sprintf("designation ID not found: %s", err.Error())
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return 0, 0, errors.New(msg)
	}

	return roleId, desigId, nil
}

func SetToken(request *http.Request) error {

	if len(os.Getenv("LOCAL_ENVIRONMENT")) == 0 {

		log.Printf("[helpers] setting bearer token...")

		token, err := bearertoken.GetToken()
		if err != nil {
			msg := fmt.Sprintf("cannot get bearer token: %s", err.Error())
			log.Printf("%s", color.HiRedString("[helpers] %s", msg))
			return errors.New(msg)
		}

		request.Header.Set("Authorization", "Bearer "+token.Token)
	}

	return nil
}

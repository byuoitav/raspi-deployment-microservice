package helpers

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// UpdateContactState executes a divider sensor to enable/disable the contacts service
func UpdateContactState(hostname string, active bool, output io.Writer) (DeployReport, *nerr.E) {
	report := DeployReport{
		Address:   hostname,
		Timestamp: time.Now().Format(time.RFC3339),
		Success:   false,
	}

	// get device from database
	device, err := db.GetDB().GetDevice(hostname)
	if err != nil {
		return report, nerr.Translate(err).Addf("failed to get %v from the database", hostname)
	}

	// validate it has the divider sensor role
	valid := false
	for i := range device.Roles {
		if strings.EqualFold(device.Roles[i].ID, "DividerSensor") {
			valid = true
			break
		}
	}

	if !valid {
		return report, nerr.Create(fmt.Sprintf("device %v isn't a divider sensor.", device.ID), reflect.TypeOf(errors.New("")).String())
	}

	command := fmt.Sprintf("sudo systemctl enable contacts && sudo systemctl start contacts")
	if !active {
		command = fmt.Sprintf("sudo systemctl stop contacts && sudo systemctl disable contacts")
	}

	log.L.Infof(command)

	/*
		er := SSHAndRunCommand(device.Address, command, os.Stdout)
		if er != nil {
			return report, er.Addf("failed to update contact state on %v", device.Address)
		}
	*/

	report.Success = true
	return report, nil
}

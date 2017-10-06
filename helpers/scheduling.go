package helpers

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

const STAGE_DEPLOYMENT_HOUR = 0
const PROD_DEPLOYMENT_HOUR = 0
const ACCURACY = time.Minute

var scheduledDeployments = make(map[string]bool)

func HoldDeployment(branch string, status bool) {
	if status {
		log.Printf("Disabling %s deployments", branch)
	} else {
		log.Printf("Enabling %s deployments", branch)
	}

	scheduledDeployments[branch] = status
}

// deploys/schedules deployments based on deploymentType
func ScheduleDeployment(deploymentType string) (string, error) {

	if scheduledDeployments[deploymentType] {

		msg := fmt.Sprintf("%s deployments are currently being held")
		log.Printf("%s", color.HiRedString("[helpers] %s", msg))
		return "", errors.New(msg)
	}

	switch deploymentType {
	/*
			case "stage":
				t := GetTimeTomorrowByHour(STAGE_DEPLOYMENT_HOUR)
				schedule, err := Schedule(t, ACCURACY)
				if err != nil {
					return "", err
				}

				go DeployOnSchedule(schedule, deploymentType)
				return fmt.Sprintf("%s deployment scheduled for %s", deploymentType, t), nil
		case "production":
			t := GetTimeTomorrowByHour(PROD_DEPLOYMENT_HOUR)
			schedule, err := Schedule(t, ACCURACY)
			if err != nil {
				return "", err
			}

			go DeployOnSchedule(schedule, deploymentType)
			return fmt.Sprintf("%s deployment scheduled for %s", deploymentType, t), nil
	*/

	default:
		err := Deploy(deploymentType)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s deployment started", deploymentType), nil
	}
}

// waits for s to return, then starts a deployment
func DeployOnSchedule(s chan time.Time, deploymentType string) {
	scheduledDeployments[deploymentType] = true

	color.Set(color.FgBlue)
	log.Printf("Waiting to deploy to %s...", deploymentType)
	color.Unset()

	<-s
	close(s)

	Deploy(deploymentType)
}

// schedules a timer that returns when the specified time is reached
// accuracy specifies how to round the time
func Schedule(t time.Time, accuracy time.Duration) (chan time.Time, error) {
	// verify t is in the future
	if !t.After(time.Now()) {
		log.Printf("Can't create a schedule for a time in the past. (%s)", t)
		return nil, errors.New(fmt.Sprintf("Can't create a schedule for a time in the past. (%s)", t))
	}
	color.Set(color.FgMagenta)
	log.Printf("Creating schedule for time: %s", t)
	color.Unset()

	ret := make(chan time.Time)

	rounded := t.UTC().Round(accuracy)
	ticker := time.NewTicker(time.Nanosecond)
	go func() {
		for tick := range ticker.C {
			rtick := tick.UTC().Round(accuracy)
			//			fmt.Printf("Rounded tick: %s; compared to scheduled time: %s\n", rtick, rounded)
			if rounded.Equal(rtick) {
				color.Set(color.FgHiGreen)
				log.Printf("Scheduled time %s reached\n", rounded)
				color.Unset()
				ret <- rtick
				break
			}
		}

		ticker.Stop()
	}()
	return ret, nil
}

// returns a time tomorrow with the specified hour
func GetTimeTomorrowByHour(hour int) time.Time {
	t := time.Now()
	t = t.AddDate(0, 0, 1)
	t = time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	return t
}

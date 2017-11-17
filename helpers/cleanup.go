package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

var fileTimers map[string]*time.Timer

func MakeMap() {
	fileTimers = make(map[string]*time.Timer)
}

func AddEntry(fileName string, timer *time.Timer) {
	fileTimers[fileName] = timer
}

func GenerateRandomString(numBytes int) (string, error) {

	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.New(fmt.Sprintf("error generating file name: %s", err.Error()))
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

func TrackFile(fileName, fileLocation string) {
	removeFile := func() {
		log.Printf("[helpers] removing old file: %s...", fileName)
		err := os.Remove(fileLocation + fileName)
		if err != nil {
			log.Printf("%s", color.HiCyanString("[helpers] error removing old file: %s", err.Error()))
		}
	}

	AddEntry(fileName, time.AfterFunc(TIMER_DURATION, removeFile))
}

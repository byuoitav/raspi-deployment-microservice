package helpers

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"
)

var fileTimers map[string]*time.Timer

func MakeMap() {
	fileTimers = make(map[string]*time.Timer)
}

func AddEntry(fileName string, timer *time.Timer) {
	fileTimers[fileName] = timer
}

func HashName(name string) string {

	log.Printf("[helpers] hashing file name %s...")

	hasher := md5.New()
	hasher.Write([]byte(name))

	result := hasher.Sum(nil)

	fixed := url.QueryEscape(string(result))

	log.Printf("[helpers] name: %s", fixed)

	return fixed
}

func GenerateRandomString(numBytes int) (string, error) {

	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.New(fmt.Sprintf("error generating file name: %s", err.Error()))
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

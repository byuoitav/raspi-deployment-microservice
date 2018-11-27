package deploy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/db/couch"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/raspi-deployment-microservice/deployment"
	"github.com/byuoitav/raspi-deployment-microservice/docker"
)

var (
	couchdb *couch.CouchDB
	dbName  = "deployment-information"
)

type configQueryResponse struct {
	Docs     []deployment.ConfigInfoWrapper `json:"docs"`
	Bookmark string                         `json:"bookmark"`
	Warning  string                         `json:"warning"`
}

func init() {
	// make sure that database is a couch database
	var ok bool

	database := db.GetDB()
	couchdb, ok = database.(*couch.CouchDB)
	if !ok {
		log.L.Fatalf("must use couch database, not %s", reflect.TypeOf(database).String())
	}
}

// GenerateDockerCompose .
func GenerateDockerCompose(stage string) ([]byte, *nerr.E) {
	var query couch.IDPrefixQuery
	query.Selector.ID.GT = "\x00"
	query.Limit = 1000

	b, err := json.Marshal(query)
	if err != nil {
		return nil, nerr.Translate(err).Addf("failed to marshal query to get docker-compose file")
	}

	var resp configQueryResponse
	err = couchdb.MakeRequest("POST", fmt.Sprintf("%v/_find", dbName), "application/json", b, &resp)
	if err != nil {
		return nil, nerr.Translate(err).Addf("failed to get deployment information")
	}

	var wrappers []deployment.ConfigInfoWrapper
	for i := range resp.Docs {
		wrappers = append(wrappers, resp.Docs[i])
	}

	// build docker-compose
	compose := docker.NewCompose()

	for _, wrapper := range wrappers {
		if config, ok := wrapper.Stages[stage]; ok {
			// build an individual service
			image := fmt.Sprintf("byuoitav/rpi-%s:%s", wrapper.ID, strings.ToLower(stage))

			env := []string{}
			for key, val := range config.EnvironmentValues {
				env = append(env, fmt.Sprintf(`%s=%s`, key, val))
			}

			ports := []string{
				fmt.Sprintf(`%s:%s`, config.Port, config.Port),
			}

			compose.Services[wrapper.ID] = docker.NewServiceWithDefaultOptions(image, ports, env)
		}
	}

	bytes, er := docker.GetFileBytes(compose)
	if er != nil {
		return nil, er.Addf("failed to generate docker compose file")
	}

	return bytes, nil
}

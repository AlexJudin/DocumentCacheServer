package temporal

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
)

const SaveDocument = "save-saga"

func NewTemporalClient(connStr string) (client.Client, error) {
	log.Info("Start connection to temporal")

	c, err := client.Dial(client.Options{
		HostPort: connStr,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to connect to temporal: %+v", err)
	}

	log.Info("Successfully connected to temporal")

	return c, nil
}

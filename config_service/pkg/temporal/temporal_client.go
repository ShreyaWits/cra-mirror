package workflows

import (
	"log"

	temporalClient "go.temporal.io/sdk/client"
)

func InitTemporal(endpoint string) (*temporalClient.Client, error) {

	client, err := temporalClient.Dial(temporalClient.Options{
		HostPort:  "localhost:7233",
		Namespace: "default",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
		return nil, err
	}
	return &client, nil
}

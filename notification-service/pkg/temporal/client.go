package temporal

import "go.temporal.io/sdk/client"

func NewTemporalClient(TemporalUrl string) (client.Client, error) {
	// Create a Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort: TemporalUrl,
	})

	if err != nil {
		return nil, err
	}

	return temporalClient, nil
}

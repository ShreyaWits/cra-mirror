package temporal

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type TemporalService struct {
	client client.Client
	worker worker.Worker
}

func InitTemporalWorker(TemporalUrl string, options worker.Options) (*TemporalService, error) {
	queueName := "notification-service-queue"

	// Create a Temporal client
	temporalClient, err := NewTemporalClient(TemporalUrl)

	if err != nil {
		return nil, err
	}

	log.Println("Temporal client initialized")

	// Create a Temporal worker
	temporalWorker := worker.New(temporalClient, queueName, worker.Options{})

	log.Println("Temporal worker initialized")

	return &TemporalService{
		client: temporalClient,
		worker: temporalWorker,
	}, nil

}

func (ts *TemporalService) RegisterWorkflows(workflows []interface{}) {
	//loop and register
	for i := 0; i < len(workflows); i++ {
		handler := workflows[i]
		ts.worker.RegisterWorkflow(handler)
	}
}

func (ts *TemporalService) RegisterActivities(activities []interface{}) {
	//loop and register
	for i := 0; i < len(activities); i++ {
		handler := activities[i]
		ts.worker.RegisterActivity(handler)
	}
}

func (ts *TemporalService) Run() {
	// Start the worker
	err := ts.worker.Run(worker.InterruptCh())
	if err != nil {
		log.Println("Error in temporal client", err.Error())
		panic(err)
	}
}

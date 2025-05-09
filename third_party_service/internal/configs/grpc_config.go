package configs

import (
	"third_party_service/internal/grpc/client"
)

type App struct {
	ThirdPartyClient *client.ThirdPartyClient
}

func NewApp() *App {
	thirdPartyClient := client.NewThirdPartyClient("localhost:50051") // use env/config
	return &App{
		ThirdPartyClient: thirdPartyClient,
	}
}

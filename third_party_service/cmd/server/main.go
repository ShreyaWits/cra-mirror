package main

import (
	"context"
	"fmt"
	"third_party_service/internal/configs"
	pb "third_party_service/thirdparty-client-service/proto" // Import the generated proto package
)

func main() {

	application := configs.NewApp()
	// Example usage
	response, err := application.ThirdPartyClient.Client.SendEmail(
		context.Background(),
		&pb.EmailRequest{
			To:      "test@example.com",
			Subject: "Hello",
			Body:    "This is a test",
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Email sent:", response.Success)

}

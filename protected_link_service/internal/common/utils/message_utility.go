package utils

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

var (
	messages map[string]string
	once     sync.Once
)

// Load messages from JSON
func LoadMessages() {
	once.Do(func() {
		env := os.Getenv("APP_ENV")
		var path string

		switch env {
		case "docker":
			path = "/app/internal/configs/message_helper_config.json"
		default: // local by default
			dir, _ := os.Getwd()
			log.Println("🗂 Current working directory:", dir)
			path = dir + "/internal/configs/message_helper_config.json"
		}

		file, err := os.Open(path)
		if err != nil {
			log.Fatalf("❌ Failed to open helper_message_config: %v", err)
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&messages)
		if err != nil {
			log.Fatalf("❌ Failed to decode helper_message_config: %v", err)
		}
	})
}

// GetMessage fetches a message by key (e.g., "USER_NOT_FOUND_ERROR")
func GetMessage(key string) string {

	if msg, ok := messages[key]; ok {
		return msg
	}
	return "Message not found for key: " + key
}

package notify_sms

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	// setup
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	m.Run()
	// teardown
	fmt.Println("Teardown test")
}

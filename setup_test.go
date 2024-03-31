package notify_sms

import (
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	// setup
	fmt.Println("Setup test")

	m.Run()
	// teardown
	fmt.Println("Teardown test")
}

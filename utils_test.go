package notify_sms

import (
	"testing"
)

func Test_auth(t *testing.T) {

}

func Test_makeRequest(t *testing.T) {

}

func Test_validateUsername(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		expected bool
	}{
		{
			name:     "Should return true when username is valid",
			username: "260979000000",
			expected: true,
		},
		{
			name:     "Should return false when username is invalid",
			username: "hello",
			expected: false,
		},
		{
			name:     "Should return false when username is empty",
			username: "",
			expected: false,
		},
		{
			name:     "Should return false when username is less than 12 characters",
			username: "26097900000",
			expected: false,
		},
		{
			name:     "Should return false when username is more than 12 characters",
			username: "2609790000000",
			expected: false,
		},
		{
			name:     "Should return false when username has a + sign",
			username: "+26097900000",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := validateUsername(tc.username)
			if actual != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

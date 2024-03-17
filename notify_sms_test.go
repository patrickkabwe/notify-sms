package notify_sms

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	var authCalled bool
	testCases := []struct {
		name                 string
		params               NewClientParams
		expectedErr          error
		expectAuthToBeCalled bool
	}{
		{
			name: "Should return error when username is missing",
			params: NewClientParams{
				Username: "",
				Password: "hello123",
				authFunc: func(n *notify) error {
					authCalled = false
					return nil
				},
			},
			expectedErr:          MissingCredErr,
			expectAuthToBeCalled: false,
		},
		{
			name: "Should return error when password is missing",
			params: NewClientParams{
				Username: "hello",
				Password: "",
				authFunc: func(n *notify) error {
					authCalled = false
					return nil
				},
			},
			expectedErr:          MissingCredErr,
			expectAuthToBeCalled: false,
		},
		{
			name: "Should return error when username is invalid",
			params: NewClientParams{
				Username: "hello",
				Password: "hello123",
				authFunc: func(n *notify) error {
					authCalled = false
					return nil
				},
			},
			expectedErr:          InvalidUsernameErr,
			expectAuthToBeCalled: false,
		},
		{
			name: "Should return nil when username and password are provided",
			params: NewClientParams{
				Username: "260979000000",
				Password: "hello123",
				authFunc: func(n *notify) error {
					authCalled = false
					return nil
				},
			},
			expectedErr:          nil,
			expectAuthToBeCalled: false,
		},
		{
			name: "Should return nil when username and password are provided and authFunc is provided",
			params: NewClientParams{
				Username: "260979000000",
				Password: "hello123",
				authFunc: func(n *notify) error {
					authCalled = true
					return nil
				},
			},
			expectedErr:          nil,
			expectAuthToBeCalled: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient(tc.params)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}

			if tc.expectAuthToBeCalled != authCalled {
				t.Errorf("Expected authFunc to be called: %v, got %v", tc.expectAuthToBeCalled, authCalled)
			}
		})
	}
}

func TestSendToContacts(t *testing.T) {
	testCases := []struct {
		name         string
		params       SendSmsToCustomContactsParams
		expectedErr  error
		expectedBool bool
	}{
		//{
		//	name: "Should send sms to custom contacts",
		//	params: SendSmsToCustomContactsParams{
		//		Contacts: []string{os.Getenv("NOTIFY_SMS_TEST_CONTACT")},
		//		Message:  "Hello Patrick from Go SDK",
		//		SenderID: os.Getenv("NOTIFY_SMS_SENDER_ID"),
		//	},
		//	expectedErr:  nil,
		//	expectedBool: true,
		//},
		{
			name: "Should return error when contacts are missing",
			params: SendSmsToCustomContactsParams{
				Contacts: []string{},
				Message:  "Hello Patrick from Go SDK",
				SenderID: os.Getenv("NOTIFY_SMS_SENDER_ID"),
			},
			expectedErr:  MissingContactsErr,
			expectedBool: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &notify{
				baseURL:  "https://production.olympusmedia.co.zm/api/v1",
				username: os.Getenv("NOTIFY_SMS_USERNAME"),
				password: os.Getenv("NOTIFY_SMS_PASSWORD"),
			}

			client.authFunc = func() error {
				return authFunc(client)
			}

			err := client.authFunc()

			if err != nil {
				t.Errorf("Expected nil, got %v", err)
			}

			_, err = client.SendToContacts(tc.params)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func Test_sendSMS(t *testing.T) {
	str := SendSmsToCustomContactsParams{
		RecipientType: "NOTIFY_RECIEPIENT_TYPE_CUSTOM",
		Contacts:      []string{os.Getenv("NOTIFY_SMS_TEST_CONTACT")},
		Message:       "Hello Patrick from Go SDK",
		SenderID:      os.Getenv("NOTIFY_SMS_SENDER_ID"),
	}

	validBody, _ := json.Marshal(str)

	testCases := []struct {
		name         string
		payload      []byte
		expectedBool bool
		hasToken     bool
	}{
		{
			name:         "Should return error when payload is invalid",
			payload:      []byte(""),
			expectedBool: false,
			hasToken:     true,
		},
		{
			name:         "Should return true when payload is valid",
			payload:      validBody,
			expectedBool: true,
			hasToken:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &notify{
				baseURL:  "https://production.olympusmedia.co.zm/api/v1",
				username: os.Getenv("NOTIFY_SMS_USERNAME"),
				password: os.Getenv("NOTIFY_SMS_PASSWORD"),
			}

			client.authFunc = func() error {
				return authFunc(client)
			}

			if tc.hasToken {
				err := client.authFunc()

				if err != nil {
					t.Errorf("Expected nil, got %v", err)
				}
			}

			result, _ := client.sendSMS(tc.payload)

			if result != tc.expectedBool {
				t.Errorf("Expected %v, got %v", tc.expectedBool, result)
			}
		})
	}
}

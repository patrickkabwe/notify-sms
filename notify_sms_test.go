package notify_sms

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name        string
		params      NewClientParams
		expectedErr error
		client      Client
	}{
		{
			name: "Should return error when username is missing",
			params: NewClientParams{
				Username: "",
				Password: "hello123",
			},
			expectedErr: ErrMissingCred,
			client:      nil,
		},
		{
			name: "Should return error when password is missing",
			params: NewClientParams{
				Username: "hello",
				Password: "",
			},
			expectedErr: ErrMissingCred,
			client:      nil,
		},
		{
			name: "Should return error when username is invalid",
			params: NewClientParams{
				Username: "hello",
				Password: "hello123",
			},
			expectedErr: ErrInvalidUsername,
			client:      nil,
		},
		{
			name: "Should return nil when username and password are provided",
			params: NewClientParams{
				Username:    "260979000000",
				Password:    "hello123",
				makeRequest: mockRequest,
			},
			expectedErr: nil,
			client: &notify{
				token:    "test_token",
				baseURL:  "https://production.olympusmedia.co.zm/api/v1",
				username: "260979000000",
				password: "hello123",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.params)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}

			if client != nil {
				mc := client.(*notify)
				if mc.username != tc.params.Username {
					t.Errorf("Expected %v, got %v", tc.params.Username, mc.username)
				}

				if mc.token == "" {
					t.Errorf("Expected %v, got %v", tc.client.(*notify).token, mc.token)
				}
			}
		})
	}
}

func TestNotify_GetSenders(t *testing.T) {
	testCases := []struct {
		name        string
		params      NewClientParams
		expectedErr error
	}{
		{
			name: "Should return senders when token is valid",
			params: NewClientParams{
				Username:    "260979000000",
				Password:    "hello123",
				makeRequest: mockRequest,
			},
			expectedErr: nil,
		},
		{
			name: "Should not return senders when token is invalid or missing",
			params: NewClientParams{
				Username:    "260979000000",
				Password:    "hello123",
				makeRequest: mockRequest,
			},
			expectedErr: ErrMissingAuth,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.params)
			if err != nil && !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}

			if client != nil {
				if tc.expectedErr == nil {
					_, err = client.GetSenders()
					if err != nil {
						t.Errorf("Expected nil, got %v", err)
					}
				} else {
					client.(*notify).token = ""
					_, err = client.GetSenders()
					if !errors.Is(err, tc.expectedErr) {
						t.Errorf("Expected %v, got %v", tc.expectedErr, err)
					}
				}

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
		{
			name: "Should send sms to custom contacts",
			params: SendSmsToCustomContactsParams{
				Contacts: []string{"+260979600000"},
				Message:  "Hello Patrick from Go SDK",
				SenderID: "test_sender_id",
			},
			expectedErr:  nil,
			expectedBool: true,
		},
		{
			name: "Should return error when contacts are missing",
			params: SendSmsToCustomContactsParams{
				Contacts: []string{},
				Message:  "Hello Patrick from Go SDK",
				SenderID: "test_sender_id",
			},
			expectedErr:  ErrMissingContacts,
			expectedBool: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientParams := NewClientParams{
				Username:    "260979000000",
				Password:    "hello123",
				makeRequest: mockRequest,
			}
			client, err := NewClient(clientParams)

			if err != nil {
				t.Errorf("Expected nil, got %v", err)
			}

			if err != nil {
				t.Errorf("Expected nil, got %v", err)
			}

			err = client.SendToContacts(tc.params)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func Test_SendToChannel(t *testing.T) {

	testCases := []struct {
		name    string
		params  SendSmsToChannelParams
		success bool
	}{
		{
			name: "Should return true when payload is valid",
			params: SendSmsToChannelParams{
				SenderID: "test_sender_id",
				Channel:  "sms_channel",
				Message:  "test message",
			},
			success: true,
		},
		{
			name:    "Should return error when payload is invalid",
			params:  SendSmsToChannelParams{},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientParams := NewClientParams{
				Username:    "260979000000",
				Password:    "hello123",
				makeRequest: mockRequest,
			}
			client, err := NewClient(clientParams)

			if err != nil {
				t.Errorf("Expected nil, got %v", err)
			}

			err = client.SendToChannel(tc.params)

			if err != nil && tc.success {
				t.Errorf("Expected nil, got %v", err)
			}
		})
	}
}

func Test_SendToContactGroup(t *testing.T) {

	testCases := []struct {
		name    string
		params  SendSmsToContactGroup
		success bool
	}{
		{
			name: "Should return error when payload is invalid",
			params: SendSmsToContactGroup{
				SenderID:     "test_sender_id",
				ContactGroup: "sms_channel",
				Message:      "test message",
			},
			success: true,
		},
		{
			name:    "Should return true when payload is valid",
			params:  SendSmsToContactGroup{},
			success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientParams := NewClientParams{
				Username:    "260979000000",
				Password:    "hello123",
				makeRequest: mockRequest,
			}
			client, err := NewClient(clientParams)

			if err != nil {
				t.Errorf("Expected nil, got %v", err)
			}

			err = client.SendToContactGroup(tc.params)

			if err != nil && tc.success {
				t.Errorf("Expected nil, got %v", err)
			}
		})
	}
}

func mockRequest(method, endpoint string, body io.Reader, opt MakeRequestOptions) ([]byte, error) {
	if method == http.MethodPost && strings.Contains(endpoint, "authentication") {
		var res NewClientParams
		err := json.NewDecoder(body).Decode(&res)
		if err != nil {
			return nil, err
		}
		if res.Username != "260979000000" && res.Password != "hello123" {
			return nil, ErrInvalidCred
		}
		return []byte(`{"success":true,"payload":{"token": "test_token"}}`), nil
	}
	authHeader := strings.Replace(opt.Headers["Authorization"], "Bearer ", "", 1)
	if authHeader == "" {
		return nil, ErrMissingAuth
	}

	if method == http.MethodPost && strings.Contains(endpoint, "channels") {
		var res map[string]interface{}
		err := json.NewDecoder(body).Decode(&res)
		if err != nil {
			return nil, err
		}

		if res["senderId"] == "" || res["message"] == "" {
			return nil, ErrInvalidPayload
		}

		if res["reciepientType"] == NOTIFY_RECIPIENT_TYPE_CHANNEL {
			if res["channel"] == "" {
				return nil, ErrInvalidPayload
			}
		}
		if res["reciepientType"] == NOTIFY_RECIPIENT_TYPE_CONTACT_GROUP {
			if res["contactGroup"] == "" {
				return nil, ErrInvalidPayload
			}
		}

		return []byte(`{"success":true,"message":"","payload":{}}`), nil
	}

	if method == http.MethodGet && strings.Contains(endpoint, "sender-ids") {
		return []byte(`{"success":true,"message":"","payload":{"data":[]}}`), nil
	}

	return nil, nil
}

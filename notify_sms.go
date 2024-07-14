package notify_sms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const ErrorPrefix = "NOTIFY_SMS: "

var (
	ErrMissingCred     = errors.New(ErrorPrefix + "username and password is missing")
	ErrInvalidUsername = errors.New(ErrorPrefix + "username is invalid")
	ErrInvalidCred     = errors.New(ErrorPrefix + "failed to authenticate please check your username or password")
	ErrMissingContacts = errors.New(ErrorPrefix + "contacts are missing")
	ErrInvalidPayload  = errors.New(ErrorPrefix + "invalid payload")
	ErrMissingAuth     = errors.New(ErrorPrefix + "authorization header is missing")
)

type Client interface {
	SendToContacts(params Message) (err error)
	SendToChannel(params Message) (err error)
	SendToContactGroup(params Message) (err error)
	CreateSenderID(params CreateSenderIDParams) (APIResponse[SenderAPIResponse], error)
	GetSenders() (APIResponse[SendersAPIResponse], error)
	GetSMSBalance()
}

type notify struct {
	token       Token
	baseURL     string
	username    string
	password    string
	makeRequest func(method, endpoint string, params io.Reader, opt MakeRequestOptions) ([]byte, error)
}

func NewClient(params NewClientParams) (Client, error) {
	if params.Password == "" || params.Username == "" {
		return nil, ErrMissingCred
	}

	if !validateUsername(params.Username) {
		return nil, ErrInvalidUsername
	}

	client := &notify{
		baseURL:  "https://production.olympusmedia.co.zm/api/v1",
		username: params.Username,
		password: params.Password,
	}

	if params.makeRequest != nil {
		client.makeRequest = params.makeRequest
	} else {
		client.makeRequest = makeRequest
	}

	err := client.authFunc()

	if err != nil {
		panic(err)
	}
	return client, nil
}

func (n *notify) SendToContacts(params Message) error {
	if len(params.Contacts) == 0 {
		return ErrMissingContacts
	}
	payload := Message{
		RecipientType: NOTIFY_RECIPIENT_TYPE_CUSTOM,
		SenderID:      params.SenderID,
		Contacts:      params.Contacts,
		Message:       params.Message,
	}
	jsonBody, _ := json.Marshal(payload)

	return n.sendSMS(jsonBody)

}

func (n *notify) SendToChannel(params Message) (err error) {
	payload := Message{
		RecipientType: NOTIFY_RECIPIENT_TYPE_CHANNEL,
		SenderID:      params.SenderID,
		Channel:       params.Channel,
		Message:       params.Message,
	}
	jsonBody, _ := json.Marshal(payload)

	return n.sendSMS(jsonBody)
}

func (n *notify) SendToContactGroup(params Message) (err error) {
	payload := Message{
		RecipientType: NOTIFY_RECIPIENT_TYPE_CONTACT_GROUP,
		SenderID:      params.SenderID,
		ContactGroup:  params.ContactGroup,
		Message:       params.Message,
	}
	jsonBody, _ := json.Marshal(payload)

	return n.sendSMS(jsonBody)
}

func (n *notify) CreateSenderID(params CreateSenderIDParams) (APIResponse[SenderAPIResponse], error) {
	//TODO implement me
	panic("implement me")
}

func (n *notify) GetSenders() (APIResponse[SendersAPIResponse], error) {
	endpoint := fmt.Sprintf("%s/notify/sender-ids/fetch?error_context=CONTEXT_API_ERROR_JSON", n.baseURL)
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", n.token)
	results, err := n.makeRequest(http.MethodGet, endpoint, nil, MakeRequestOptions{
		Headers: headers,
	})
	var senderRes APIResponse[SendersAPIResponse]

	if err != nil {
		return APIResponse[SendersAPIResponse]{}, err
	}

	if err = json.Unmarshal(results, &senderRes); err != nil {
		log.Println(ErrorPrefix + "failed to unmarshal json")
		return APIResponse[SendersAPIResponse]{}, err
	}

	return senderRes, err
}

func (n *notify) GetSMSBalance() {
	//TODO implement me
	panic("implement me")
}

func (n *notify) authFunc() error {
	endpoint := fmt.Sprintf("%s/authentication/web/login?error_context=CONTEXT_API_ERROR_JSON", n.baseURL)
	bodyInBytes := []byte(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, n.username, n.password))

	bodyReader := bytes.NewReader(bodyInBytes)
	res, err := n.makeRequest(http.MethodPost, endpoint, bodyReader, MakeRequestOptions{})

	var authResponse APIResponse[AuthAPIResponse]

	if err != nil {
		log.Printf(ErrorPrefix+"/%s\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(res, &authResponse)

	if err != nil {
		log.Printf(ErrorPrefix+"/%s\n", err)
		os.Exit(1)
	}

	if !authResponse.Success {
		log.Printf(ErrorPrefix+"/%s\n", err)
		return ErrInvalidCred
	}

	n.token = authResponse.Payload.Token

	return nil
}

func (n *notify) sendSMS(jsonBody []byte) error {
	endpoint := fmt.Sprintf("%s/notify/channels/messages/compose?error_context=CONTEXT_API_ERROR_JSON", n.baseURL)
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", n.token)

	res, err := n.makeRequest(http.MethodPost, endpoint, bytes.NewReader(jsonBody), MakeRequestOptions{
		Headers: headers,
	})

	if err != nil {
		log.Println(err)
		return err
	}

	var parsedBody APIResponse[any]

	_ = json.Unmarshal(res, &parsedBody)

	if !parsedBody.Success {
		err = errors.New(parsedBody.Message)
		if parsedBody.Error != (ErrorResponse{}) {
			err = errors.New(parsedBody.Error.Message)
		}
		log.Printf(ErrorPrefix+"%s\n", string(res))
		log.Printf(ErrorPrefix+"%s\n", err)
		return err
	}

	return nil
}

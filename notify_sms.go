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
	MissingCredErr     = errors.New(ErrorPrefix + "username and password is missing")
	InvalidUsernameErr = errors.New(ErrorPrefix + "username is invalid")
	InvalidCredErr     = errors.New(ErrorPrefix + "failed to authenticate please check your username or password")
	MissingContactsErr = errors.New(ErrorPrefix + "contacts are missing")
	InvalidPayloadErr  = errors.New(ErrorPrefix + "invalid payload")
	MissingAuthErr     = errors.New(ErrorPrefix + "authorization header is missing")
)
var tokenCache = make(map[string]Token)

type NotifySMS interface {
	SendToContacts(params SendSmsToCustomContactsParams) (ok bool, err error)
	SendToChannel(params SendSmsToChannelParams) (ok bool, err error)
	SendToContactGroup(params SendSmsToContactGroup) (ok bool, err error)
	CreateSenderID(params CreateSenderIDParams) (APIResponse[SenderAPIResponse], error)
	GetSenders() (APIResponse[SendersAPIResponse], error)
	GetSMSBalance()
}

type notify struct {
	token       Token
	senderID    SenderID
	baseURL     string
	username    string
	password    string
	makeRequest func(method, endpoint string, params io.Reader, opt MakeRequestOptions) ([]byte, error)
}

func NewClient(params NewClientParams) (NotifySMS, error) {
	if params.Password == "" || params.Username == "" {
		return nil, MissingCredErr
	}

	if !validateUsername(params.Username) {
		return nil, InvalidUsernameErr
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
		return InvalidCredErr
	}

	n.token = authResponse.Payload.Token

	return nil
}

func (n *notify) SendToContacts(params SendSmsToCustomContactsParams) (bool, error) {
	if len(params.Contacts) == 0 {
		return false, MissingContactsErr
	}
	payload := SendSmsToCustomContactsParams{
		RecipientType: "NOTIFY_RECIEPIENT_TYPE_CUSTOM",
		SenderID:      params.SenderID,
		Contacts:      params.Contacts,
		Message:       params.Message,
	}
	jsonBody, _ := json.Marshal(payload)

	return n.sendSMS(jsonBody)

}

func (n *notify) SendToChannel(params SendSmsToChannelParams) (ok bool, err error) {
	payload := SendSmsToChannelParams{
		RecipientType: "NOTIFY_RECIEPIENT_TYPE_CHANNEL",
		SenderID:      params.SenderID,
		Channel:       params.Channel,
		Message:       params.Message,
	}
	jsonBody, _ := json.Marshal(payload)

	return n.sendSMS(jsonBody)
}

func (n *notify) SendToContactGroup(params SendSmsToContactGroup) (ok bool, err error) {
	payload := SendSmsToContactGroup{
		RecipientType: "NOTIFY_RECIEPIENT_TYPE_CONTACT_GROUP",
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

func (n *notify) sendSMS(jsonBody []byte) (bool, error) {
	endpoint := fmt.Sprintf("%s/notify/channels/messages/compose?error_context=CONTEXT_API_ERROR_JSON", n.baseURL)
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", n.token)

	res, err := n.makeRequest(http.MethodPost, endpoint, bytes.NewReader(jsonBody), MakeRequestOptions{
		Headers: headers,
	})

	if err != nil {
		log.Println(err)
		return false, err
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
		return false, err
	}

	return true, nil
}

package notify_sms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

const ErrorPrefix = "NOTIFY_SMS: "

var (
	MissingCredErr = errors.New(ErrorPrefix + "username and password is missing")
	InvalidCredErr = errors.New(ErrorPrefix + "failed to authenticate please check your username or password")
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
	Token    Token
	SenderID SenderID
	baseURL  string
	username string
	password string
}

func NewClient(params NewClientParams) (NotifySMS, error) {
	if params.Password == "" || params.UserName == "" {
		return nil, MissingCredErr
	}
	baseURL := "https://production.olympusmedia.co.zm/api/v1"
	err := auth(params.UserName, params.Password, baseURL)
	if err != nil {
		panic(err)
	}
	return &notify{
		baseURL:  baseURL,
		Token:    tokenCache["token"],
		SenderID: "",
		username: params.UserName,
		password: params.Password,
	}, nil
}

func (n notify) SendToContacts(params SendSmsToCustomContactsParams) (bool, error) {
	payload := SendSmsToCustomContactsParams{
		RecipientType: "NOTIFY_RECIEPIENT_TYPE_CUSTOM",
		SenderID:      params.SenderID,
		Contacts:      params.Contacts,
		Message:       params.Message,
	}
	jsonBody, err := json.Marshal(payload)

	if err != nil {
		log.Println(err)
		return false, err
	}

	if err != nil {
		log.Println(err)
		return false, err
	}

	return n.sendSMS(jsonBody)

}

func (n notify) SendToChannel(params SendSmsToChannelParams) (ok bool, err error) {
	payload := SendSmsToChannelParams{
		RecipientType: "NOTIFY_RECIEPIENT_TYPE_CHANNEL",
		SenderID:      params.SenderID,
		Channel:       params.Channel,
		Message:       params.Message,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		log.Println("NOTIFY_SMS: failed to marshal payload")
	}

	return n.sendSMS(jsonBody)
}

func (n notify) SendToContactGroup(params SendSmsToContactGroup) (ok bool, err error) {
	payload := SendSmsToContactGroup{
		RecipientType: "NOTIFY_RECIEPIENT_TYPE_CONTACT_GROUP",
		SenderID:      params.SenderID,
		ContactGroup:  params.ContactGroup,
		Message:       params.Message,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		log.Println("NOTIFY_SMS: failed to marshal payload")
		return false, err
	}

	return n.sendSMS(jsonBody)
}

func (n notify) CreateSenderID(params CreateSenderIDParams) (APIResponse[SenderAPIResponse], error) {
	//TODO implement me
	panic("implement me")
}

func (n notify) GetSenders() (APIResponse[SendersAPIResponse], error) {
	endpoint := fmt.Sprintf("%s/notify/sender-ids/fetch?error_context=CONTEXT_API_ERROR_JSON", n.baseURL)
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", tokenCache["token"])
	results, err := makeRequest(http.MethodGet, endpoint, nil, MakeRequestOptions{
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

func (n notify) GetSMSBalance() {
	//TODO implement me
	panic("implement me")
}

func (n notify) sendSMS(jsonBody []byte) (bool, error) {
	endpoint := fmt.Sprintf("%s/notify/channels/messages/compose?error_context=CONTEXT_API_ERROR_JSON", n.baseURL)
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", tokenCache["token"])
	res, err := makeRequest(http.MethodPost, endpoint, bytes.NewReader(jsonBody), MakeRequestOptions{
		Headers: headers,
	})

	if err != nil {
		log.Println(err)
		return false, err
	}
	var parsedBody APIResponse[any]

	err = json.Unmarshal(res, &parsedBody)

	if err != nil {
		log.Printf(ErrorPrefix+"/%s\n", err)
		return false, err
	}

	if !parsedBody.Success {
		log.Printf(ErrorPrefix+"/%s\n", err)
		return false, err
	}

	return true, nil
}

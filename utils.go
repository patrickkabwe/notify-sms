package notify_sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func makeRequest(method, endpoint string, params io.Reader, opt MakeRequestOptions) ([]byte, error) {
	req, err := http.NewRequest(method, endpoint, params)
	if err != nil {
		log.Printf(ErrorPrefix+"could not create request: %s\n", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range opt.Headers {
		req.Header.Set(key, value)
	}
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(req)

	if err != nil {
		log.Printf(ErrorPrefix+"error making http request: %s\n", err)
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf(ErrorPrefix+"could not read response body: %s\n", err)
		return nil, err
	}

	return resBody, nil
}

func auth(username, password, baseURL string) error {
	endpoint := fmt.Sprintf("%s/authentication/web/login?error_context=CONTEXT_API_ERROR_JSON", baseURL)
	bodyInBytes := []byte(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, password))

	bodyReader := bytes.NewReader(bodyInBytes)
	res, err := makeRequest(http.MethodPost, endpoint, bodyReader, MakeRequestOptions{})
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

	tokenCache["token"] = authResponse.Payload.Token

	return nil
}

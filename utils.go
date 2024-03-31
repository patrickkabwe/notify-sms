package notify_sms

import (
	"io"
	"log"
	"net/http"
	"regexp"
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

func validateUsername(username string) bool {
	match, _ := regexp.MatchString(`^(\+)?(\d{12})$`, username)
	return username != "" && match
}

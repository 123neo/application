package verve

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

type Service interface {
	Accept(id, optString string, uniqRequests int) (string, error)
}

type service struct {
}

func NewService() Service {
	return &service{}
}

func (s *service) Accept(id, endpoint string, uniqRequests int) (string, error) {
	if len(endpoint) != 0 {
		jsonBody := []byte{}
		bodyReader := bytes.NewReader(jsonBody)
		requestURL := fmt.Sprintf("%s/%d", endpoint, uniqRequests)
		req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
		if err != nil {
			log.Print("error", err)
		}
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Print("error", err)
		}
		log.Print("status code", res.StatusCode)
		log.Print("response body", res.Body)
		defer res.Body.Close()
	}
	return "ok", nil
}

package q3bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

func makeRequest(url, method string, data interface{}) ([]byte, error) {
	body := bytes.NewReader(nil)

	if data != nil {
		j, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		if config.Verbose {
			log.Printf("[HTTPREQUEST] [DEBUG] makeRequest sending data [%s]: %s", url, j)
		}
		body = bytes.NewReader(j)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if config.Verbose {
		log.Printf("[HTTPREQUEST] [DEBUG] makeRequest response: [%s] %s - %s", method, url, resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(b))
	}

	return b, nil
}

package play

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Remote is a Runner that sends each command to the HTTP API at the base URL
// (docs/backend/api.md): the same command contract as the core in process, so the screens
// do not know which one answers. A refusal comes back as the server's body and an error.
func Remote(baseURL string, client *http.Client) Runner {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	var base string = strings.TrimRight(baseURL, "/")
	return func(name string, input []byte) ([]byte, error) {
		response, err := client.Post(base+"/v1/commands/"+name, "application/json", bytes.NewReader(input))
		if err != nil {
			return nil, fmt.Errorf("server %s: %v", base, err)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("server %s: %v", base, err)
		}
		if response.StatusCode == http.StatusOK {
			return body, nil
		}
		var refusal struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &refusal) == nil && refusal.Error != "" {
			return body, fmt.Errorf("%s", refusal.Error)
		}
		return body, fmt.Errorf("server %s: status %d", base, response.StatusCode)
	}
}

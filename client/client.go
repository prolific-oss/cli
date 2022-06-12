package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/benmatselby/prolificli/model"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

// API represents what is allowed to be called on the Prolific client.
type API interface {
	GetMe() (*Me, error)
	GetStudies(status string) (*ListStudiesResponse, error)
	GetSubmissions(id string) (*ListSubmissionsResponse, error)
}

// Client is responsible for interacting with the Prolicif API.
type Client struct {
	Client  *http.Client
	BaseURL string
	Token   string
	Debug   bool
}

// New will return a new Prolific client.
func New() Client {
	client := Client{
		Client:  http.DefaultClient,
		Token:   viper.GetString("PROLIFIC_TOKEN"),
		BaseURL: strings.TrimRight(viper.GetString("PROLIFIC_URL"), "/"),
		Debug:   viper.GetBool("PROLIFIC_DEBUG"),
	}

	return client
}

// Get is the main router for GET requests to the Prolific API.
func (c *Client) Get(url string, response interface{}) (*http.Response, error) {
	request, err := http.NewRequest("GET", c.BaseURL+url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "benmatselby/prolificli")
	request.Header.Set("Authorization", fmt.Sprintf("Token %s", c.Token))

	httpResponse, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s responded with status %d", request.RequestURI, httpResponse.StatusCode)
	}

	if c.Debug {
		body, _ := ioutil.ReadAll(httpResponse.Body)
		fmt.Println(string(body))
	}

	if err := json.NewDecoder(httpResponse.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding JSON response from %s failed: %v", request.URL, err)
	}

	return httpResponse, nil
}

// GetMe will return your user account details.
func (c *Client) GetMe() (*Me, error) {
	var response Me

	url := "/api/v1/users/me"
	_, err := c.Get(url, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetStudies will return you a list of Study objects.
func (c *Client) GetStudies(status string) (*ListStudiesResponse, error) {
	var response ListStudiesResponse

	if !slices.Contains(model.StudyListStatus, status) {
		return nil, fmt.Errorf("%s is not a valid status: %s", status, strings.Join(model.StudyListStatus, ", "))
	}

	statusFragment := ""
	if status == model.StatusUnpublished {
		statusFragment = "published=0"
	} else {
		statusFragment = fmt.Sprintf("%s=1", status)
	}

	url := fmt.Sprintf("/api/v1/studies/?%s", statusFragment)

	_, err := c.Get(url, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

// GetSubmissions will return submission data for a given study.
func (c *Client) GetSubmissions(id string) (*ListSubmissionsResponse, error) {
	var response ListSubmissionsResponse

	url := fmt.Sprintf("/api/v1/studies/%s/submissions/", id)
	_, err := c.Get(url, &response)
	if err != nil {
		return nil, fmt.Errorf("unable to fulfil request %s: %s", url, err)
	}

	return &response, nil
}

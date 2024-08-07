package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/envtrack/envtrack-cli/internal/config"
)

const (
	defaultBaseURL = "https://europe-west1-envtrack-2fd23.cloudfunctions.net"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthToken  string
}

func NewClient(authToken string) *Client {
	baseURL := config.GlobalConf.Get("api_endpoint")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
		AuthToken:  authToken,
	}
}

func (c *Client) sendRequest(method, path string, query url.Values) ([]byte, error) {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, err
	}

	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-carrier-auth", c.AuthToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("RESPONSE DETAILS %v", resp)
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *Client) GetOrganizations() ([]Organization, error) {
	query := url.Values{}

	body, err := c.sendRequest("GET", "/callableGetEnvs", query)
	if err != nil {
		return nil, err
	}

	var org []Organization
	err = json.Unmarshal(body, &org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (c *Client) GetOrganization(orgID string) (*Organization, error) {
	query := url.Values{}
	query.Set("orgId", orgID)

	body, err := c.sendRequest("GET", "/callableGetEnvs", query)
	if err != nil {
		return nil, err
	}

	var org Organization
	err = json.Unmarshal(body, &org)
	if err != nil {
		return nil, err
	}

	return &org, nil
}

func (c *Client) GetProjects(orgID string) ([]Project, error) {
	org, err := c.GetOrganization(orgID)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, proj := range org.Projects {
		projects = append(projects, proj)
	}

	return projects, nil
}

func (c *Client) GetProject(orgID string, prjID string) (*Project, error) {
	org, err := c.GetOrganization(orgID)
	if err != nil {
		return nil, err
	}

	for _, proj := range org.Projects {
		if proj.ID == prjID {
			return &proj, nil
		}
	}

	return nil, fmt.Errorf("project \"%s\" not found", prjID)
}

func (c *Client) GetProjectWithOrganization(orgID string, prjID string) (*Project, *Organization, error) {
	org, err := c.GetOrganization(orgID)
	if err != nil {
		return nil, nil, err
	}

	for _, proj := range org.Projects {
		if proj.ID == prjID {
			return &proj, org, nil
		}
	}

	return nil, nil, fmt.Errorf("project \"%s\" not found", prjID)
}

func (c *Client) GetEnvironments(orgID, projectID string) ([]Environment, error) {
	query := url.Values{}
	query.Set("orgId", orgID)
	query.Set("prjId", projectID)

	body, err := c.sendRequest("GET", "/callableGetEnvs", query)
	if err != nil {
		return nil, err
	}

	var result struct {
		Environments []Environment `json:"environments"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Environments, nil
}

func (c *Client) GetVariables(orgID, projectID, envID string) ([]Variable, error) {
	query := url.Values{}
	query.Set("orgId", orgID)
	query.Set("prjId", projectID)
	query.Set("envId", envID)

	body, err := c.sendRequest("GET", "/callableGetEnvs", query)
	if err != nil {
		return nil, err
	}

	var result Environment
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.ID == envID {
		return result.Variables, nil
	}

	return nil, fmt.Errorf("environment not found")
}

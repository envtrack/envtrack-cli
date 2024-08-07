package api

type Organization struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	ShortName string             `json:"shortName"`
	GithubOrg string             `json:"githubOrg"`
	Projects  map[string]Project `json:"projects,omitempty"`
}

type Project struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	ShortName    string        `json:"shortName"`
	Environments []Environment `json:"environments"`
}

type Environment struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Variables []Variable `json:"variables"`
}

type Variable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

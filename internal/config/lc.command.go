package config

type LocalConfigCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
}

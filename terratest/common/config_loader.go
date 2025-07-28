package common

import (
	"encoding/json"
	"os"
)

type LockSpec struct {
	Level string  `json:"level"`
	Notes *string `json:"notes"`
}
type ResourceGroupSpec struct {
	Location string            `json:"location"`
	Tags     map[string]string `json:"tags"`
	Lock     *LockSpec         `json:"lock"`
}
type RgTestConfig struct {
	ResourceGroups      map[string]ResourceGroupSpec `json:"resource_groups"`
	ExpectedPrincipalID string                       `json:"expected_principal_id"`
	ExpectedRoles       []string                     `json:"expected_roles"`
}

func LoadRgTestConfig(path string) (*RgTestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg RgTestConfig
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

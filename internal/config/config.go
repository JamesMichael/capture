package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Contexts map[string]string `json:"contexts"`
}

func NewFromFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open '%s': %w", path, err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read '%s': %w", path, err)
	}

	var c Config
	if err := json.Unmarshal(bytes, &c); err != nil {
		return nil, fmt.Errorf("cannot parse '%s': %w", path, err)
	}

	return &c, nil
}

func (c *Config) ListContexts() []string {
	cs := make([]string, 0)
	for k, _ := range c.Contexts {
		cs = append(cs, k)
	}
	return cs
}

func (c *Config) FileForContext(context string) string {
	return c.Contexts[context]
}

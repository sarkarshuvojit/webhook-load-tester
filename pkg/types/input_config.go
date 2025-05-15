package types

import (
	"strings"
)

type RootType = int

const (
	RootBody    RootType = iota
	RootHeader  RootType = iota
	RootUnknown RootType = iota
)

type Locator struct {
	Path string `yaml:"path"`
}

func (l Locator) GetRootTypeString() string {
	value, _, _ := strings.Cut(l.Path, ".")
	return value
}

func (l Locator) GetRootType() RootType {
	if strings.HasPrefix(l.Path, "body.") {
		return RootBody
	} else if strings.HasPrefix(l.Path, "headers.") {
		return RootHeader
	}

	return RootUnknown
}

func (l Locator) GetKey() string {
	_, value, _ := strings.Cut(l.Path, ".")
	return value
}

func (l Locator) GetKeys() []string {
	_, value, _ := strings.Cut(l.Path, ".")
	return strings.Split(value, ".")
}

func updateMap(m *map[string]interface{}, keys []string, value string) {
	if len(keys) == 0 {
		return
	}

	key := keys[0]
	if len(keys) == 1 {
		// Last key, set the value
		(*m)[key] = value
		return
	}

	// Intermediate keys, ensure the key exists and is a map
	if _, exists := (*m)[key]; !exists {
		(*m)[key] = make(map[string]any)
	}
	subMap, ok := (*m)[key].(map[string]any)
	if !ok {
		// If the existing value is not a map, we need to handle the conflict.
		// For simplicity, let's clear the existing value and replace it with a map.
		subMap = make(map[string]any)
		(*m)[key] = subMap
	}

	// Recursively update the sub-map
	updateMap(&subMap, keys[1:], value)
}

func (l Locator) SetToLocator(target *map[string]any, value string) {
	updateMap(target, l.GetKeys(), value)
}

func (l Locator) GetByLocator(target *map[string]any) *any {
	return getFromMap(*target, l.GetKeys())
}

// getFromMap recursively retrieves a value from the map based on the provided keys
func getFromMap(m map[string]any, keys []string) *any {
	if len(keys) == 0 {
		return nil
	}

	key := keys[0]
	if len(keys) == 1 {
		// Last key, return the value
		value, exists := m[key]
		if !exists {
			return nil
		}

		return &value
	}

	// Intermediate keys, ensure the key exists and is a map
	val, exists := m[key]
	if !exists {
		return nil
	}

	subMap, ok := val.(map[string]any)
	if !ok {
		// If the value is not a map, the path is invalid
		return nil
	}

	// Recursively get the value from the sub-map
	return getFromMap(subMap, keys[1:])
}

type TestConfig struct {
	Name      string            `yaml:"name"`
	URL       string            `yaml:"url"`
	Body      string            `yaml:"body"`
	Headers   map[string]string `yaml:"headers"`
	Injectors struct {
		ReplyPathInjector     Locator `yaml:"replyPathInjector"`
		CorrelationIDInjector Locator `yaml:"correlationIdInjector"`
	} `yaml:"injectors"`
	Pickers struct {
		CorrelationPicker Locator `yaml:"correlationPicker"`
	} `yaml:"pickers"`
	Timeout int `yaml:"timeout"`
}

type InputConfig struct {
	Version string     `yaml:"version"`
	Server  string     `yaml:"server"`
	Test    TestConfig `yaml:"test"`
	Run     struct {
		Iterations      int `yaml:"iterations"`
		DurationSeconds int `yaml:"durationSeconds"`
	} `yaml:"run"`
	Outputs []struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"outputs"`
}

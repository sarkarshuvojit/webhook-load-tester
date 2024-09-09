package types

import "strings"

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

package types

type TestConfig struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	Body      string `yaml:"body"`
	Injectors struct {
		ReplyPathInjector struct {
			Path string `yaml:"path"`
		} `yaml:"replyPathInjector"`
		CorrelationIDInjector struct {
			Path string `yaml:"path"`
		} `yaml:"correlationIdInjector"`
	} `yaml:"injectors"`
	Pickers struct {
		CorrelationPicker struct {
			Path string `yaml:"path"`
		} `yaml:"correlationPicker"`
	} `yaml:"pickers"`
}

type InputConfig struct {
	Version string       `yaml:"version"`
	Tests   []TestConfig `yaml:"tests"`
	Run     struct {
		Iterations      int `yaml:"iterations"`
		DurationSeconds int `yaml:"durationSeconds"`
	} `yaml:"run"`
	Outputs []struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"outputs"`
}

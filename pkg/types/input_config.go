package types

type InputConfig struct {
	Version string `yaml:"version" json:"version"`
	Tests   []struct {
		Name      string `yaml:"name" json:"name"`
		URL       string `yaml:"url" json:"url"`
		Body      string `yaml:"body" json:"body"`
		Injectors struct {
			CorrelationInjector struct {
				Path string `yaml:"path" json:"path"`
				Use  string `yaml:"use" json:"use"`
			} `yaml:"correlationInjector" json:"correlationInjector"`
		} `yaml:"injectors" json:"injectors"`
		Pickers struct {
			CorrelationPicker struct {
				Path string `yaml:"path" json:"path"`
			} `yaml:"correlationPicker" json:"correlationPicker"`
		} `yaml:"pickers" json:"pickers"`
	} `yaml:"tests" json:"tests"`
	Run struct {
		Iterations      int `yaml:"iterations" json:"iterations"`
		DurationSeconds int `yaml:"durationSeconds" json:"durationSeconds"`
	} `yaml:"run" json:"run"`
	Outputs []struct {
		Type string `yaml:"type" json:"type"`
		Path string `yaml:"path" json:"path"`
	} `yaml:"outputs" json:"outputs"`
}

package lib

type MkTask struct {
	Name    string
	Help    string
	Cmd     string            `yaml:"cmd"`
	Pre     []string          `yaml:"pre"`
	Onerror string            `yaml:"onerror"`
	Env     map[string]string `yaml:"env"`
	Vars    map[string]string `yaml:"vars"`
}
type MkModel struct {
	Env     map[string]string  `yaml:"env"`
	Vars    map[string]string  `yaml:"vars"`
	Tasks   map[string]*MkTask `yaml:"tasks"`
	Default string             `yaml:"default"`
	Stack   map[string]string  `yaml:"-"`
}

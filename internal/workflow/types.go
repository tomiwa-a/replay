package workflow

type Workflow struct {
	Name    string   `yaml:"name" json:"name"`
	Version string   `yaml:"version,omitempty" json:"version,omitempty"`
	Config  Config   `yaml:"config,omitempty" json:"config,omitempty"`
	Steps   []Step   `yaml:"steps" json:"steps"`
	Meta    MetaData `yaml:"meta,omitempty" json:"meta,omitempty"`
}

type MetaData map[string]any

type Config struct {
	HTTP     HTTPConfig     `yaml:"http,omitempty" json:"http,omitempty"`
	Postgres PostgresConfig `yaml:"postgres,omitempty" json:"postgres,omitempty"`
	Redis    RedisConfig    `yaml:"redis,omitempty" json:"redis,omitempty"`
}

type HTTPConfig struct {
	BaseURL string `yaml:"base_url,omitempty" json:"base_url,omitempty"`
}

type PostgresConfig struct {
	DSN string `yaml:"dsn,omitempty" json:"dsn,omitempty"`
}

type RedisConfig struct {
	Addr string `yaml:"addr,omitempty" json:"addr,omitempty"`
}

type StepType string

const (
	StepTypeHTTP  StepType = "http"
	StepTypeDB    StepType = "db"
	StepTypeShell StepType = "shell"
	StepTypePrint StepType = "print"
)

type Step struct {
	Name    string        `yaml:"name" json:"name"`
	Type    StepType      `yaml:"type" json:"type"`
	Request *HTTPRequest  `yaml:"request,omitempty" json:"request,omitempty"`
	DB      *DBRequest    `yaml:"db,omitempty" json:"db,omitempty"`
	Shell   *ShellRequest `yaml:"shell,omitempty" json:"shell,omitempty"`

	// Flat DB shortcuts
	Query   string `yaml:"query,omitempty" json:"query,omitempty"`
	Command any    `yaml:"command,omitempty" json:"command,omitempty"`
	Engine  string `yaml:"engine,omitempty" json:"engine,omitempty"`

	// Flat Shell shortcuts
	Dir     string `yaml:"dir,omitempty" json:"dir,omitempty"`
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	Message string            `yaml:"message,omitempty" json:"message,omitempty"`
	Extract map[string]string `yaml:"extract,omitempty" json:"extract,omitempty"`
	Assert  []AssertRule      `yaml:"assert,omitempty" json:"assert,omitempty"`

	IgnoreError bool `yaml:"ignore_error,omitempty" json:"ignore_error,omitempty"`
}

type HTTPRequest struct {
	Method  string            `yaml:"method" json:"method"`
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body    any               `yaml:"body,omitempty" json:"body,omitempty"`
}

type ShellRequest struct {
	Command  string   `yaml:"command,omitempty" json:"command,omitempty"`
	Commands []string `yaml:"commands,omitempty" json:"commands,omitempty"`
	Dir      string   `yaml:"dir,omitempty" json:"dir,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

type DBEngine string

const (
	DBEnginePostgres DBEngine = "postgres"
	DBEngineRedis    DBEngine = "redis"
)

type DBRequest struct {
	Engine  DBEngine `yaml:"engine" json:"engine"`
	Query   string   `yaml:"query,omitempty" json:"query,omitempty"`
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}

type AssertRule struct {
	Path  string `yaml:"path,omitempty" json:"path,omitempty"`
	Op    string `yaml:"op,omitempty" json:"op,omitempty"`
	Value any    `yaml:"value,omitempty" json:"value,omitempty"`
}

func (r *AssertRule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var slice []any
	if err := unmarshal(&slice); err == nil {
		if len(slice) >= 2 {
			if s, ok := slice[0].(string); ok {
				r.Path = s
			}
			if s, ok := slice[1].(string); ok {
				r.Op = s
			}
			if len(slice) >= 3 {
				r.Value = slice[2]
			}
			return nil
		}
	}

	type Alias AssertRule
	var res Alias
	if err := unmarshal(&res); err != nil {
		return err
	}
	*r = AssertRule(res)
	return nil
}

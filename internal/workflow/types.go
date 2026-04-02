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
	Name    string            `yaml:"name" json:"name"`
	Type    StepType          `yaml:"type" json:"type"`
	Request *HTTPRequest      `yaml:"request,omitempty" json:"request,omitempty"`
	DB      *DBRequest        `yaml:"db,omitempty" json:"db,omitempty"`
	Shell   *ShellRequest     `yaml:"shell,omitempty" json:"shell,omitempty"`
	Message string            `yaml:"message,omitempty" json:"message,omitempty"`
	Extract map[string]string `yaml:"extract,omitempty" json:"extract,omitempty"`
	Assert  []AssertRule      `yaml:"assert,omitempty" json:"assert,omitempty"`
}

type HTTPRequest struct {
	Method  string            `yaml:"method" json:"method"`
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body    any               `yaml:"body,omitempty" json:"body,omitempty"`
}

type ShellRequest struct {
	Command string `yaml:"command" json:"command"`
	Dir     string `yaml:"dir,omitempty" json:"dir,omitempty"`
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
	Path  string `yaml:"path" json:"path"`
	Op    string `yaml:"op" json:"op"`
	Value any    `yaml:"value,omitempty" json:"value,omitempty"`
}

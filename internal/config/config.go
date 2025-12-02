package config

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server   ServerConfig    `yaml:"server" json:"server"`
	Routes   []RouteConfig   `yaml:"routes" json:"routes"`
	Backends []BackendConfig `yaml:"backends" json:"backends"`
	Policies PoliciesConfig  `yaml:"policies" json:"policies"`
}

type ServerConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     string `yaml:"port" json:"port"`
	TLS      bool   `yaml:"tls" json:"tls"`
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

type RouteConfig struct {
	Name       string            `yaml:"name" json:"name"`
	PathPrefix string            `yaml:"path_prefix" json:"path_prefix"`
	Pattern    string            `yaml:"pattern" json:"pattern"`
	Subdomain  string            `yaml:"subdomain" json:"subdomain"`
	Headers    map[string]string `yaml:"headers" json:"headers"`
	Methods    []string          `yaml:"methods" json:"methods"`
	BackendID  string            `yaml:"backend_id" json:"backend_id"`
	Priority   int               `yaml:"priority" json:"priority"`
}

type BackendConfig struct {
	ID            string         `yaml:"id" json:"id"`
	Servers       []string       `yaml:"servers" json:"servers"`
	HealthCheck   HealthConfig   `yaml:"health_check" json:"health_check"`
	LoadBalancing string         `yaml:"load_balancing" json:"load_balancing"`
	Weights       map[string]int `yaml:"weights" json:"weights"`
}

type HealthConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Interval string `yaml:"interval" json:"interval"`
	Timeout  string `yaml:"timeout" json:"timeout"`
	Path     string `yaml:"path" json:"path"`
}

type PoliciesConfig struct {
	RateLimit RateLimitPolicy `yaml:"rate_limit" json:"rate_limit"`
	CORS      CORSPolicy      `yaml:"cors" json:"cors"`
	Auth      AuthPolicy      `yaml:"auth" json:"auth"`
	Cache     CachePolicy     `yaml:"cache" json:"cache"`
}

type RateLimitPolicy struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	MaxRequests int    `yaml:"max_requests" json:"max_requests"`
	Window      string `yaml:"window" json:"window"`
}

type CORSPolicy struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins"`
}

type AuthPolicy struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Type    string `yaml:"type" json:"type"`
	Secret  string `yaml:"secret" json:"secret"`
}

type CachePolicy struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	TTL     string   `yaml:"ttl" json:"ttl"`
	Methods []string `yaml:"methods" json:"methods"`
}

func LoadFromYAML(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func LoadFromJSON(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

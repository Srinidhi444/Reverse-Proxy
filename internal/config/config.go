package config

type Config struct {
	Server   ServerConfig             `yaml:"server"`
	Services map[string]ServiceConfig `yaml:"services"`
	Routes   []RouteConfig            `yaml:"routes"`
}

type ServerConfig struct {
	Address      string `yaml:"address"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
	IdleTimeout  string `yaml:"idle_timeout"`
}

type ServiceConfig struct {
	LoadBalancer string           `yaml:"load_balancer"`
	Endpoints    []EndpointConfig `yaml:"endpoints"`
}

type EndpointConfig struct {
	URL string `yaml:"url"`
}

type RouteConfig struct {
	Path    string `yaml:"path"`
	Match   string `yaml:"match"`
	Service string `yaml:"service"`
}

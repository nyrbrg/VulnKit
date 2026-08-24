package compose

import (
	"fmt"
	"strings"
)

type ServiceConfig struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Version string            `json:"version"`
	Ports   []string          `json:"ports"`
	EnvVars map[string]string `json:"env_vars"`
}

type LabConfig struct {
	Name     string          `json:"name"`
	Services []ServiceConfig `json:"services"`
}

var ServiceDefaults = map[string]ServiceConfig{
	"mysql": {
		Name:  "mysql",
		Image: "mysql",
		Ports: []string{"3306:3306"},
		EnvVars: map[string]string{
			"MYSQL_ROOT_PASSWORD": "vulnkit",
			"MYSQL_DATABASE":      "testdb",
		},
	},
	"postgres": {
		Name:  "postgres",
		Image: "postgres",
		Ports: []string{"5432:5432"},
		EnvVars: map[string]string{
			"POSTGRES_PASSWORD": "vulnkit",
			"POSTGRES_DB":       "testdb",
		},
	},
	"apache": {
		Name:  "apache",
		Image: "httpd",
		Ports: []string{"8080:80"},
	},
	"nginx": {
		Name:  "nginx",
		Image: "nginx",
		Ports: []string{"8081:80"},
	},
	"redis": {
		Name:  "redis",
		Image: "redis",
		Ports: []string{"6379:6379"},
	},
	"mongodb": {
		Name:  "mongodb",
		Image: "mongo",
		Ports: []string{"27017:27017"},
		EnvVars: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "root",
			"MONGO_INITDB_ROOT_PASSWORD": "vulnkit",
		},
	},
}

var ServiceVersions = map[string][]string{
	"mysql":    {"5.5.62", "5.6.51", "5.7.43", "8.0.35", "8.3"},
	"postgres": {"9.6", "10.23", "11.22", "12.19", "13.15", "14.12", "15.7", "16.3"},
	"apache":   {"2.2.34", "2.4.49", "2.4.51", "2.4.57", "2.4.62"},
	"nginx":    {"1.14.2", "1.18.0", "1.20.2", "1.24.0", "1.26.0"},
	"redis":    {"4.0.14", "5.0.14", "6.0.20", "6.2.14", "7.0.15", "7.2.4"},
	"mongodb":  {"3.6", "4.0", "4.2", "4.4", "5.0", "6.0", "7.0"},
}

func Generate(config LabConfig) string {
	var sb strings.Builder

	sb.WriteString("version: '3.8'\n\n")
	sb.WriteString("services:\n")

	for _, svc := range config.Services {
		sb.WriteString(fmt.Sprintf("  %s:\n", svc.Name))
		if strings.HasPrefix(svc.Image, "vulnkit-") || svc.Version == "" {
			sb.WriteString(fmt.Sprintf("    image: %s\n", svc.Image))
		} else {
			sb.WriteString(fmt.Sprintf("    image: %s:%s\n", svc.Image, svc.Version))
		}
		sb.WriteString("    labels:\n")
		sb.WriteString("      - \"managed-by=vulnkit\"\n")
		sb.WriteString(fmt.Sprintf("      - \"vulnkit.lab=%s\"\n", config.Name))

		if len(svc.Ports) > 0 {
			sb.WriteString("    ports:\n")
			for _, p := range svc.Ports {
				sb.WriteString(fmt.Sprintf("      - \"%s\"\n", p))
			}
		}

		if len(svc.EnvVars) > 0 {
			sb.WriteString("    environment:\n")
			for k, v := range svc.EnvVars {
				sb.WriteString(fmt.Sprintf("      %s: %s\n", k, v))
			}
		}

		sb.WriteString("    restart: \"no\"\n")
		sb.WriteString("\n")
	}

	sb.WriteString("networks:\n")
	sb.WriteString("  default:\n")
	sb.WriteString("    name: vulnkit_net\n")

	return sb.String()
}

func GetDefaultConfig(serviceName, version string) ServiceConfig {
	def, ok := ServiceDefaults[serviceName]
	if !ok {
		return ServiceConfig{Name: serviceName, Image: serviceName, Version: version}
	}
	def.Version = version
	return def
}

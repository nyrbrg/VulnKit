package compose_test

import (
	"strings"
	"testing"

	"vulnkit/internal/compose"
)

func TestGenerate_ContainsImageAndVersion(t *testing.T) {
	config := compose.LabConfig{
		Name: "test-lab",
		Services: []compose.ServiceConfig{
			{Name: "mysql", Image: "mysql", Version: "5.6.51", Ports: []string{"3306:3306"}},
		},
	}

	yaml := compose.Generate(config)

	if !strings.Contains(yaml, "mysql:5.6.51") {
		t.Errorf("expected yaml to contain 'mysql:5.6.51', got:\n%s", yaml)
	}
}

func TestGenerate_ContainsPorts(t *testing.T) {
	config := compose.LabConfig{
		Name: "test-lab",
		Services: []compose.ServiceConfig{
			{Name: "redis", Image: "redis", Version: "7.0", Ports: []string{"6379:6379"}},
		},
	}

	yaml := compose.Generate(config)

	if !strings.Contains(yaml, "6379:6379") {
		t.Errorf("expected yaml to contain port mapping, got:\n%s", yaml)
	}
}

func TestGenerate_ContainsVulnkitLabel(t *testing.T) {
	config := compose.LabConfig{
		Name: "test-lab",
		Services: []compose.ServiceConfig{
			{Name: "nginx", Image: "nginx", Version: "1.24"},
		},
	}

	yaml := compose.Generate(config)

	if !strings.Contains(yaml, "managed-by=vulnkit") {
		t.Errorf("expected yaml to contain vulnkit label, got:\n%s", yaml)
	}
}

func TestGenerate_ContainsLabName(t *testing.T) {
	config := compose.LabConfig{
		Name: "my-sqli-lab",
		Services: []compose.ServiceConfig{
			{Name: "mysql", Image: "mysql", Version: "5.6.51"},
		},
	}

	yaml := compose.Generate(config)

	if !strings.Contains(yaml, "my-sqli-lab") {
		t.Errorf("expected yaml to contain lab name, got:\n%s", yaml)
	}
}

func TestGenerate_MultipleServices(t *testing.T) {
	config := compose.LabConfig{
		Name: "multi-lab",
		Services: []compose.ServiceConfig{
			{Name: "mysql", Image: "mysql", Version: "5.6.51", Ports: []string{"3306:3306"}},
			{Name: "apache", Image: "httpd", Version: "2.4.49", Ports: []string{"8080:80"}},
		},
	}

	yaml := compose.Generate(config)

	for _, want := range []string{"mysql:5.6.51", "httpd:2.4.49", "3306:3306", "8080:80"} {
		if !strings.Contains(yaml, want) {
			t.Errorf("expected yaml to contain %q, got:\n%s", want, yaml)
		}
	}
}

func TestGenerate_EnvVars(t *testing.T) {
	config := compose.LabConfig{
		Name: "test-lab",
		Services: []compose.ServiceConfig{
			{
				Name:    "mysql",
				Image:   "mysql",
				Version: "5.6.51",
				EnvVars: map[string]string{
					"MYSQL_ROOT_PASSWORD": "vulnkit",
					"MYSQL_DATABASE":      "testdb",
				},
			},
		},
	}

	yaml := compose.Generate(config)

	if !strings.Contains(yaml, "MYSQL_ROOT_PASSWORD") {
		t.Errorf("expected yaml to contain env vars, got:\n%s", yaml)
	}
}

func TestGenerate_NetworkBlock(t *testing.T) {
	config := compose.LabConfig{
		Name:     "test-lab",
		Services: []compose.ServiceConfig{{Name: "redis", Image: "redis", Version: "7.0"}},
	}

	yaml := compose.Generate(config)

	if !strings.Contains(yaml, "vulnkit_net") {
		t.Errorf("expected yaml to define vulnkit_net network, got:\n%s", yaml)
	}
}

func TestGetDefaultConfig_SetsVersion(t *testing.T) {
	cfg := compose.GetDefaultConfig("mysql", "8.0.35")
	if cfg.Version != "8.0.35" {
		t.Errorf("expected version 8.0.35, got %s", cfg.Version)
	}
}

func TestGetDefaultConfig_UnknownService(t *testing.T) {
	cfg := compose.GetDefaultConfig("unknowndb", "1.0")
	if cfg.Name != "unknowndb" {
		t.Errorf("expected name unknowndb, got %s", cfg.Name)
	}
	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}
}

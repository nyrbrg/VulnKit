package labs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type PackageMapping struct {
	BaseImage   string
	PackageName string
	Port        int
	StartCmd    []string
	EnvVars     map[string]string
	VersionCmd  []string
}

var packageMappings = map[string]map[string]PackageMapping{
	"mysql": {
		"3":   {BaseImage: "ubuntu:14.04", PackageName: "mysql-server-5.5", Port: 3306, StartCmd: []string{"mysqld_safe"}, VersionCmd: []string{"mysql", "--version"}},
		"5.0": {BaseImage: "ubuntu:14.04", PackageName: "mysql-server-5.5", Port: 3306, StartCmd: []string{"mysqld_safe"}, VersionCmd: []string{"mysql", "--version"}},
		"5.1": {BaseImage: "ubuntu:14.04", PackageName: "mysql-server-5.5", Port: 3306, StartCmd: []string{"mysqld_safe"}, VersionCmd: []string{"mysql", "--version"}},
		"5.5": {BaseImage: "ubuntu:14.04", PackageName: "mysql-server-5.5", Port: 3306, StartCmd: []string{"mysqld_safe"}, VersionCmd: []string{"mysql", "--version"}},
		"5.6": {BaseImage: "ubuntu:16.04", PackageName: "mysql-server-5.6", Port: 3306, StartCmd: []string{"mysqld_safe"}},
		"5.7": {BaseImage: "ubuntu:18.04", PackageName: "mysql-server-5.7", Port: 3306, StartCmd: []string{"mysqld_safe"}},
		"8.0": {
			BaseImage:   "ubuntu:20.04",
			PackageName: "mysql-server-8.0",
			Port:        3306,
			StartCmd:    []string{"mysqld_safe"},
			VersionCmd:  []string{"mysql", "--version"},
			EnvVars:     map[string]string{"MYSQL_ROOT_PASSWORD": "vulnkit"},
		},
		"8":   {BaseImage: "ubuntu:20.04", PackageName: "mysql-server-8.0", Port: 3306, StartCmd: []string{"mysqld_safe"}},
		"9":   {BaseImage: "ubuntu:24.04", PackageName: "mysql-server-8.0", Port: 3306, StartCmd: []string{"mysqld_safe"}, VersionCmd: []string{"mysql", "--version"}, EnvVars: map[string]string{"MYSQL_ROOT_PASSWORD": "vulnkit"}},
		"9.0": {BaseImage: "ubuntu:24.04", PackageName: "mysql-server-8.0", Port: 3306, StartCmd: []string{"mysqld_safe"}, VersionCmd: []string{"mysql", "--version"}, EnvVars: map[string]string{"MYSQL_ROOT_PASSWORD": "vulnkit"}},
	},
	"httpd": {
		"2.2": {BaseImage: "ubuntu:12.04", PackageName: "apache2", Port: 80, StartCmd: []string{"apache2ctl", "-D", "FOREGROUND"}},
		"2.4": {
			BaseImage:   "ubuntu:14.04",
			PackageName: "apache2",
			Port:        80,
			StartCmd:    []string{"apache2ctl", "-D", "FOREGROUND"},
			VersionCmd:  []string{"apache2", "-v"},
		}},
	"nginx": {
		"1.14": {BaseImage: "ubuntu:18.04", PackageName: "nginx", Port: 80, StartCmd: []string{"nginx", "-g", "daemon off;"}},
		"1.18": {BaseImage: "ubuntu:20.04", PackageName: "nginx", Port: 80, StartCmd: []string{"nginx", "-g", "daemon off;"}},
	},
	"postgres": {
		"9":  {BaseImage: "ubuntu:14.04", PackageName: "postgresql-9.3", Port: 5432, StartCmd: []string{"postgres"}},
		"10": {BaseImage: "ubuntu:18.04", PackageName: "postgresql-10", Port: 5432, StartCmd: []string{"postgres"}},
	},
	"redis": {
		"4": {BaseImage: "ubuntu:18.04", PackageName: "redis-server", Port: 6379, StartCmd: []string{"redis-server"}},
		"5": {BaseImage: "ubuntu:20.04", PackageName: "redis-server", Port: 6379, StartCmd: []string{"redis-server"}},
	},
}

var versionPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._-]*[a-zA-Z0-9])?$`)

func IsValidVersion(version string) bool {
	return versionPattern.MatchString(version)
}

func FindMapping(software, version string) (*PackageMapping, error) {
	if !IsValidVersion(version) {
		return nil, fmt.Errorf("invalid version format: %q", version)
	}

	swMappings, ok := packageMappings[software]
	if !ok {
		return nil, fmt.Errorf("no mapping for software: %s", software)
	}

	parts := strings.Split(version, ".")

	if len(parts) >= 2 {
		majorMinor := parts[0] + "." + parts[1]
		if m, ok := swMappings[majorMinor]; ok {
			return &m, nil
		}
	}

	if len(parts) >= 1 {
		if m, ok := swMappings[parts[0]]; ok {
			return &m, nil
		}
	}

	targetNum := versionStringToNum(version)
	var bestKey string
	var bestNum int
	for key := range swMappings {
		n := versionStringToNum(key)
		if n <= targetNum && n > bestNum {
			bestNum = n
			bestKey = key
		}
	}
	if bestKey != "" {
		m := swMappings[bestKey]
		return &m, nil
	}

	return nil, fmt.Errorf("no mapping for %s:%s", software, version)
}

func versionStringToNum(v string) int {
	parts := strings.Split(v, ".")
	major, _ := strconv.Atoi(parts[0])
	minor := 0
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	return major*100 + minor
}

func GenerateDockerfile(software, version string, mapping PackageMapping) string {
	startCmd := `["` + strings.Join(mapping.StartCmd, `", "`) + `"]`

	envLines := ""
	for k, v := range mapping.EnvVars {
		envLines += fmt.Sprintf("ENV %s=%s\n", k, v)
	}

	return fmt.Sprintf(`FROM %s
ENV DEBIAN_FRONTEND=noninteractive
# Requested: %s:%s — using closest available package: %s
RUN apt-get update && apt-get install -y %s && apt-get clean
%sEXPOSE %d
CMD %s
`, mapping.BaseImage, software, version, mapping.PackageName, mapping.PackageName, envLines, mapping.Port, startCmd)
}

func getBuildDir(software, version string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	buildDir := filepath.Join(homeDir, ".vulnkit", "builds",
		fmt.Sprintf("%s-%s", software, strings.ReplaceAll(version, ".", "-")))
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", err
	}
	return buildDir, nil
}

func getActualVersion(imageName string, versionCmd []string) (string, error) {
	if len(versionCmd) == 0 {
		return "", fmt.Errorf("no version command")
	}

	args := append([]string{"run", "--rm", imageName}, versionCmd...)
	out, err := exec.Command("docker", args...).Output()
	if err != nil {
		return "", err
	}

	return parseVersion(string(out)), nil
}

func parseVersion(output string) string {
	re := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func BuildImage(ctx context.Context, software, version string, logChan chan<- string) (string, error) {
	mapping, err := FindMapping(software, version)
	if err != nil {
		return "", err
	}

	imageName := fmt.Sprintf("vulnkit-%s-%s", software, strings.ReplaceAll(version, ".", "-"))

	buildDir, err := getBuildDir(software, version)
	if err != nil {
		return "", fmt.Errorf("failed to create build dir: %w", err)
	}

	dockerfile := GenerateDockerfile(software, version, *mapping)
	dockerfilePath := filepath.Join(buildDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
		return "", err
	}

	out, _ := exec.Command("docker", "images", "-q", imageName).Output()
	if strings.TrimSpace(string(out)) != "" {
		logChan <- fmt.Sprintf("✓ Image %s already exists, skipping build", imageName)
		logChan <- fmt.Sprintf("✓ Dockerfile preserved at %s", dockerfilePath)
		return imageName, nil
	}

	logChan <- fmt.Sprintf("✓ Dockerfile saved to %s", dockerfilePath)
	logChan <- fmt.Sprintf("→ Building %s:%s using %s", software, version, mapping.BaseImage)
	logChan <- fmt.Sprintf("→ Package: %s", mapping.PackageName)

	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageName, buildDir)
	cmd.Stdout = &chanWriter{ch: logChan}
	cmd.Stderr = &chanWriter{ch: logChan}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}

	logChan <- fmt.Sprintf("✓ Successfully built %s", imageName)
	logChan <- fmt.Sprintf("✓ Dockerfile preserved at %s", dockerfilePath)

	actualVersion, err := getActualVersion(imageName, mapping.VersionCmd)
	if err == nil && actualVersion != "" {
		logChan <- fmt.Sprintf("✓ Actual installed version: %s", actualVersion)

		os.WriteFile(filepath.Join(buildDir, "actual_version"), []byte(actualVersion), 0644)

		actualImageName := fmt.Sprintf("vulnkit-%s-%s",
			software,
			strings.ReplaceAll(actualVersion, ".", "-"),
		)

		if actualImageName != imageName {
			renameCmd := exec.Command("docker", "tag", imageName, actualImageName)
			if err := renameCmd.Run(); err == nil {
				exec.Command("docker", "rmi", imageName).Run()
				imageName = actualImageName
				logChan <- fmt.Sprintf("✓ Retagged as %s", actualImageName)
			}
		}
	}

	return imageName, nil
}

type chanWriter struct {
	ch chan<- string
}

func (w *chanWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(strings.TrimSpace(string(p)), "\n")
	for _, line := range lines {
		if line != "" {
			w.ch <- line
		}
	}
	return len(p), nil
}

type BuildInfo struct {
	ID               string    `json:"id"`
	Software         string    `json:"software"`
	Version          string    `json:"version"`
	RequestedVersion string    `json:"requested_version"`
	ImageName        string    `json:"image_name"`
	BuildDir         string    `json:"build_dir"`
	CreatedAt        time.Time `json:"created_at"`
}

func resolveBuildVersion(buildDir, requestedVersion string) string {
	actualVersionBytes, err := os.ReadFile(filepath.Join(buildDir, "actual_version"))
	if err == nil {
		if v := strings.TrimSpace(string(actualVersionBytes)); v != "" {
			return v
		}
	}
	return requestedVersion
}

func ListBuilds() ([]BuildInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	buildsDir := filepath.Join(homeDir, ".vulnkit", "builds")

	entries, err := os.ReadDir(buildsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BuildInfo{}, nil
		}
		return nil, err
	}

	var builds []BuildInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		parts := strings.SplitN(name, "-", 2)
		if len(parts) < 2 {
			continue
		}
		software := parts[0]
		requestedVersion := strings.ReplaceAll(parts[1], "-", ".")

		displayVersion := resolveBuildVersion(filepath.Join(buildsDir, name), requestedVersion)

		actualImageName := fmt.Sprintf("vulnkit-%s-%s",
			software,
			strings.ReplaceAll(displayVersion, ".", "-"),
		)

		builds = append(builds, BuildInfo{
			ID:               name,
			Software:         software,
			Version:          displayVersion,
			RequestedVersion: requestedVersion,
			ImageName:        actualImageName,
			BuildDir:         filepath.Join(buildsDir, name),
			CreatedAt:        info.ModTime(),
		})
	}

	sort.Slice(builds, func(i, j int) bool {
		return builds[i].CreatedAt.After(builds[j].CreatedAt)
	})

	return builds, nil
}

func GetDockerfile(id string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dockerfilePath := filepath.Join(homeDir, ".vulnkit", "builds", id, "Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func DeleteBuild(id string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	buildDir := filepath.Join(homeDir, ".vulnkit", "builds", id)

	imageName := fmt.Sprintf("vulnkit-%s", id)
	if parts := strings.SplitN(id, "-", 2); len(parts) == 2 {
		software := parts[0]
		requestedVersion := strings.ReplaceAll(parts[1], "-", ".")
		displayVersion := resolveBuildVersion(buildDir, requestedVersion)
		imageName = fmt.Sprintf("vulnkit-%s-%s", software, strings.ReplaceAll(displayVersion, ".", "-"))
	}

	exec.Command("docker", "rmi", "-f", imageName).Run()

	return os.RemoveAll(buildDir)
}

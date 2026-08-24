package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"vulnkit/internal/compose"
	"vulnkit/internal/docker"
	"vulnkit/internal/labs"
	"vulnkit/internal/presets"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	docker     *docker.Client
	presets    *presets.Store
	projectDir string
}

func NewHandler(d *docker.Client, p *presets.Store, projectDir string) *Handler {
	return &Handler{docker: d, presets: p, projectDir: projectDir}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GET /api/health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	err := h.docker.Ping(r.Context())
	writeJSON(w, 200, map[string]any{
		"status": "ok",
		"docker": err == nil,
	})
}

// POST /api/labs/generate
func (h *Handler) GenerateCompose(w http.ResponseWriter, r *http.Request) {
	var config compose.LabConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if len(config.Services) == 0 {
		writeError(w, 400, "no services specified")
		return
	}
	yaml := compose.Generate(config)
	writeJSON(w, 200, map[string]string{"yaml": yaml})
}

func isPortInUse(port string) (bool, string) {
	out, err := exec.Command("docker", "ps",
		"--format", "{{.Names}}\t{{.Ports}}",
	).Output()
	if err != nil {
		return false, ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ":"+port+"->") {
			parts := strings.Split(line, "\t")
			if len(parts) > 0 {
				return true, parts[0]
			}
			return true, ""
		}
	}
	return false, ""
}

// POST /api/labs/start
func (h *Handler) StartLab(w http.ResponseWriter, r *http.Request) {
	var config compose.LabConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}

	for _, svc := range config.Services {
		for _, p := range svc.Ports {
			hostPort := strings.Split(p, ":")[0]
			if inUse, container := isPortInUse(hostPort); inUse {
				writeError(w, 409, fmt.Sprintf(
					"port %s is already in use by container %s", hostPort, container,
				))
				return
			}
		}
	}

	yaml := compose.Generate(config)

	tmpDir, err := os.MkdirTemp("", "vulnkit-*")
	if err != nil {
		writeError(w, 500, "failed to create temp dir")
		return
	}

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(yaml), 0644); err != nil {
		writeError(w, 500, "failed to write compose file")
		return
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.CommandContext(r.Context(), "docker", "compose", "-f", composePath, "up", "-d")
	out, err := cmd.CombinedOutput()
	if err != nil {
		writeError(w, 500, string(out))
		return
	}

	writeJSON(w, 200, map[string]string{"status": "started", "output": string(out)})
}

// POST /api/labs/stop
func (h *Handler) StopLab(w http.ResponseWriter, r *http.Request) {
	var body struct {
		LabName string `json:"lab_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	if body.LabName == "" {
		writeError(w, 400, "lab_name is required")
		return
	}

	ids, err := h.docker.ContainerIDsForLab(r.Context(), body.LabName)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if len(ids) == 0 {
		writeError(w, 404, fmt.Sprintf("no containers found for lab %q", body.LabName))
		return
	}

	var errs []string
	for _, id := range ids {
		if err := h.docker.RemoveContainer(r.Context(), id); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		writeError(w, 500, strings.Join(errs, "; "))
		return
	}

	writeJSON(w, 200, map[string]string{"status": "stopped"})
}

// GET /api/labs/ports/check?port=3306
func (h *Handler) CheckPort(w http.ResponseWriter, r *http.Request) {
	port := r.URL.Query().Get("port")
	if port == "" {
		writeError(w, 400, "missing port")
		return
	}

	out, err := exec.Command("docker", "ps",
		"--format", "{{.Names}}\t{{.Ports}}",
	).Output()
	if err != nil {
		writeJSON(w, 200, map[string]any{"in_use": false})
		return
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ":"+port+"->") {
			parts := strings.Split(line, "\t")
			containerName := ""
			if len(parts) > 0 {
				containerName = parts[0]
			}
			writeJSON(w, 200, map[string]any{
				"in_use":    true,
				"container": containerName,
			})
			return
		}
	}

	writeJSON(w, 200, map[string]any{"in_use": false})
}

// GET /api/labs/status
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	containers, err := h.docker.ListLabContainers(context.Background())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"containers": containers})
}

// GET /api/labs/logs/{id}
func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "100"
	}
	logs, err := h.docker.GetLogs(r.Context(), id, tail)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"logs": logs})
}

// GET /api/presets
func (h *Handler) ListPresets(w http.ResponseWriter, r *http.Request) {
	list, err := h.presets.List(r.Context())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []presets.Preset{}
	}
	writeJSON(w, 200, map[string]any{"presets": list})
}

// POST /api/presets
func (h *Handler) SavePreset(w http.ResponseWriter, r *http.Request) {
	var p presets.Preset
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}
	saved, err := h.presets.Add(r.Context(), p)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, saved)
}

// DELETE /api/presets/{id}
func (h *Handler) DeletePreset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.presets.Delete(r.Context(), id); err != nil {
		writeError(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// GET /api/hub/search?q=mysql
func (h *Handler) HubSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, 400, "missing query")
		return
	}
	resp, err := http.Get("https://hub.docker.com/v2/search/repositories/?query=" + url.QueryEscape(q) + "&page_size=8")
	if err != nil {
		writeError(w, 502, "docker hub unreachable")
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

// GET /api/hub/tags?image=mysql
func (h *Handler) HubTags(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Query().Get("image")
	if image == "" {
		writeError(w, 400, "missing image")
		return
	}
	if !strings.Contains(image, "/") {
		image = "library/" + image
	}
	resp, err := http.Get("https://hub.docker.com/v2/repositories/" + image + "/tags/?page_size=100&ordering=last_updated")
	if err != nil {
		writeError(w, 502, "docker hub unreachable")
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

// POST /api/labs/containers/{id}/stop
func (h *Handler) StopContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.docker.StopContainer(r.Context(), id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "stopped"})
}

// POST /api/labs/containers/{id}/start
func (h *Handler) StartContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.docker.StartContainer(r.Context(), id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "started"})
}

// DELETE /api/labs/containers/{id}
func (h *Handler) RemoveContainer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.docker.RemoveContainer(r.Context(), id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "removed"})
}

// GET /api/labs/default
func (h *Handler) ListDefaultLabs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"labs": labs.DefaultLabs})
}

// GET /api/labs/builtin/status
func (h *Handler) DefaultLabStatus(w http.ResponseWriter, r *http.Request) {
	out, _ := exec.Command("docker", "ps",
		"--filter", "name=vulnkit-labs",
		"--filter", "status=running",
		"--format", "{{.Names}}",
	).Output()
	running := strings.TrimSpace(string(out)) != ""
	writeJSON(w, 200, map[string]bool{"running": running})
}

// POST /api/labs/builtin/start
func (h *Handler) StartDefaultLab(w http.ResponseWriter, r *http.Request) {
	out, _ := exec.Command("docker", "images", "-q", "vulnkit-vulnkit-labs").Output()
	imageExists := strings.TrimSpace(string(out)) != ""

	if !imageExists {
		cmd := exec.CommandContext(r.Context(), "docker", "compose", "build", "vulnkit-labs")
		cmd.Dir = "/app"
		if out, err := cmd.CombinedOutput(); err != nil {
			writeError(w, 500, "build failed: "+string(out))
			return
		}
	}

	cmd := exec.CommandContext(r.Context(), "docker", "compose", "up", "-d", "vulnkit-labs")
	cmd.Dir = "/app"
	if out, err := cmd.CombinedOutput(); err != nil {
		writeError(w, 500, string(out))
		return
	}

	writeJSON(w, 200, map[string]string{"status": "started"})
}

// POST /api/labs/builtin/stop
func (h *Handler) StopDefaultLab(w http.ResponseWriter, r *http.Request) {
	cmd := exec.CommandContext(r.Context(), "docker", "compose", "stop", "vulnkit-labs")
	cmd.Dir = h.projectDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		writeError(w, 500, fmt.Sprintf("stop failed: %s (dir: %s)", string(out), h.projectDir))
		return
	}
	writeJSON(w, 200, map[string]string{"status": "stopped"})
}

// GET /api/cve/search?q=mysql
func (h *Handler) CVESearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	startIndex := r.URL.Query().Get("startIndex")
	pageSize := r.URL.Query().Get("pageSize")
	if startIndex == "" {
		startIndex = "0"
	}
	if pageSize == "" {
		pageSize = "20"
	}

	nvdURL := fmt.Sprintf(
		"https://services.nvd.nist.gov/rest/json/cves/2.0?keywordSearch=%s&startIndex=%s&resultsPerPage=%s",
		url.QueryEscape(q), startIndex, pageSize,
	)

	req, err := http.NewRequestWithContext(r.Context(), "GET", nvdURL, nil)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	req.Header.Set("Accept", "application/json")

	if apiKey := os.Getenv("NVD_API_KEY"); apiKey != "" {
		req.Header.Set("apiKey", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		writeError(w, 502, "NVD API unreachable")
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

// GET /api/cve/{id}
func (h *Handler) CVEDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	nvdURL := fmt.Sprintf(
		"https://services.nvd.nist.gov/rest/json/cves/2.0?cveId=%s",
		url.QueryEscape(id),
	)

	req, err := http.NewRequestWithContext(r.Context(), "GET", nvdURL, nil)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	req.Header.Set("Accept", "application/json")

	if apiKey := os.Getenv("NVD_API_KEY"); apiKey != "" {
		req.Header.Set("apiKey", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		writeError(w, 502, "NVD API unreachable")
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

// GET /api/labs/dynamic/check?image=mysql&version=3.21
func (h *Handler) DynamicBuildCheck(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Query().Get("image")
	version := r.URL.Query().Get("version")
	if image == "" || version == "" {
		writeError(w, 400, "missing image or version")
		return
	}

	mapping, err := labs.FindMapping(image, version)
	if err != nil {
		writeJSON(w, 200, map[string]any{
			"supported": false,
			"message":   err.Error(),
		})
		return
	}

	writeJSON(w, 200, map[string]any{
		"supported":  true,
		"base_image": mapping.BaseImage,
		"package":    mapping.PackageName,
		"image_name": fmt.Sprintf("vulnkit-%s-%s", image, strings.ReplaceAll(version, ".", "-")),
	})
}

// POST /api/labs/dynamic/build
func (h *Handler) DynamicBuild(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Image   string `json:"image"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "invalid request body")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, 500, "streaming not supported")
		return
	}

	sendEvent := func(eventType, data string) {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
		flusher.Flush()
	}

	logChan := make(chan string, 100)
	errChan := make(chan error, 1)
	imageChan := make(chan string, 1)

	go func() {
		imageName, err := labs.BuildImage(r.Context(), body.Image, body.Version, logChan)
		if err != nil {
			errChan <- err
		} else {
			imageChan <- imageName
		}
		close(logChan)
	}()

	done := false
	for !done {
		select {
		case line, ok := <-logChan:
			if !ok {
				logChan = nil
			} else {
				sendEvent("log", line)
			}
		case imageName := <-imageChan:
			sendEvent("done", imageName)
			done = true
		case err := <-errChan:
			sendEvent("error", err.Error())
			done = true
		case <-r.Context().Done():
			done = true
		}
		if logChan == nil && !done {
			select {
			case imageName := <-imageChan:
				sendEvent("done", imageName)
			case err := <-errChan:
				sendEvent("error", err.Error())
			case <-r.Context().Done():
			}
			done = true
		}
	}
}

// GET /api/labs/dynamic/builds
func (h *Handler) ListDynamicBuilds(w http.ResponseWriter, r *http.Request) {
	builds, err := labs.ListBuilds()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"builds": builds})
}

// GET /api/labs/dynamic/builds/{id}/dockerfile
func (h *Handler) GetBuildDockerfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	content, err := labs.GetDockerfile(id)
	if err != nil {
		writeError(w, 404, "Dockerfile not found")
		return
	}
	writeJSON(w, 200, map[string]string{"dockerfile": content})
}

// DELETE /api/labs/dynamic/builds/{id}
func (h *Handler) DeleteDynamicBuild(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := labs.DeleteBuild(id); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

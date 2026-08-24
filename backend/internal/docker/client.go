package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Client struct{}

func NewClient() (*Client, error) {
	if err := exec.Command("docker", "info").Run(); err != nil {
		return nil, err
	}
	return &Client{}, nil
}

func (c *Client) Close() {}

func (c *Client) Ping(_ context.Context) error {
	return exec.Command("docker", "info").Run()
}

type ContainerStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	State  string `json:"state"`
	Ports  string `json:"ports"`
}

func (c *Client) ListLabContainers(_ context.Context) ([]ContainerStatus, error) {
	out, err := exec.Command("docker", "ps", "-a",
		"--filter", "label=managed-by=vulnkit",
		"--format", "{{json .}}",
	).Output()
	if err != nil {
		return nil, err
	}

	var containers []ContainerStatus
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var raw struct {
			ID     string `json:"ID"`
			Names  string `json:"Names"`
			Image  string `json:"Image"`
			Status string `json:"Status"`
			State  string `json:"State"`
			Ports  string `json:"Ports"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		containers = append(containers, ContainerStatus{
			ID:     raw.ID,
			Name:   strings.TrimPrefix(raw.Names, "/"),
			Image:  raw.Image,
			Status: raw.Status,
			State:  raw.State,
			Ports:  raw.Ports,
		})
	}
	return containers, nil
}

func (c *Client) ContainerIDsForLab(_ context.Context, labName string) ([]string, error) {
	out, err := exec.Command("docker", "ps", "-a",
		"--filter", "label=vulnkit.lab="+labName,
		"--format", "{{.ID}}",
	).Output()
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			ids = append(ids, line)
		}
	}
	return ids, nil
}

func (c *Client) StopContainer(_ context.Context, id string) error {
	return exec.Command("docker", "stop", id).Run()
}

func (c *Client) StartContainer(_ context.Context, id string) error {
	return exec.Command("docker", "start", id).Run()
}

func (c *Client) RemoveContainer(_ context.Context, id string) error {
    out, err := exec.Command("docker", "rm", "-f", id).CombinedOutput()
    if err != nil {
        return fmt.Errorf("%s", out)
    }
    return nil
}

func (c *Client) GetLogs(_ context.Context, id string, tail string) (string, error) {
	out, err := exec.Command("docker", "logs", "--tail", tail, id).CombinedOutput()
	return string(out), err
}
package presets

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"vulnkit/internal/compose"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Preset struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Tags      []string                `json:"tags"`
	Services  []compose.ServiceConfig `json:"services"`
	CreatedAt time.Time               `json:"created_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) List(ctx context.Context) ([]Preset, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, tags, services, created_at
		FROM presets
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var presets []Preset
	for rows.Next() {
		var p Preset
		var tagsJSON, servicesJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &tagsJSON, &servicesJSON, &p.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(tagsJSON, &p.Tags); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(servicesJSON, &p.Services); err != nil {
			return nil, err
		}
		presets = append(presets, p)
	}
	return presets, rows.Err()
}

func newPresetID() (string, error) {
	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), hex.EncodeToString(suffix)), nil
}

func (s *Store) Add(ctx context.Context, p Preset) (Preset, error) {
	id, err := newPresetID()
	if err != nil {
		return Preset{}, err
	}
	p.ID = id
	p.CreatedAt = time.Now()

	tagsJSON, err := json.Marshal(p.Tags)
	if err != nil {
		return Preset{}, err
	}
	servicesJSON, err := json.Marshal(p.Services)
	if err != nil {
		return Preset{}, err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO presets (id, name, tags, services, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, p.ID, p.Name, tagsJSON, servicesJSON, p.CreatedAt)

	return p, err
}

func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.pool.Exec(ctx, `DELETE FROM presets WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("preset not found: %s", id)
	}
	return nil
}

func (s *Store) SeedDefaults(ctx context.Context) error {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM presets`).Scan(&count)
	if err != nil || count > 0 {
		return err
	}

	defaults := []Preset{
		{
			Name: "MySQL 5.6 SQLi labor",
			Tags: []string{"SQLi"},
			Services: []compose.ServiceConfig{
				compose.GetDefaultConfig("mysql", "5.6.51"),
			},
		},
		{
			Name: "Apache path traversal",
			Tags: []string{"RCE"},
			Services: []compose.ServiceConfig{
				compose.GetDefaultConfig("apache", "2.4.49"),
			},
		},
		{
			Name: "PHP XSS playground",
			Tags: []string{"XSS"},
			Services: []compose.ServiceConfig{
				compose.GetDefaultConfig("apache", "2.2.34"),
				compose.GetDefaultConfig("mysql", "5.6.51"),
			},
		},
	}

	for _, p := range defaults {
		if _, err := s.Add(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

package presets_test

import (
	"context"
	"testing"

	"vulnkit/internal/compose"
	"vulnkit/internal/presets"
	"vulnkit/tests"
)

func TestStore_AddAndList(t *testing.T) {
	pool := tests.NewTestDB(t)
	store := presets.NewStore(pool)
	ctx := context.Background()

	p := presets.Preset{
		Name: "Test preset",
		Tags: []string{"SQLi"},
		Services: []compose.ServiceConfig{
			{Name: "mysql", Image: "mysql", Version: "5.6.51"},
		},
	}

	saved, err := store.Add(ctx, p)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if saved.ID == "" {
		t.Error("expected ID to be set after Add")
	}

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(list))
	}
	if list[0].Name != "Test preset" {
		t.Errorf("expected name 'Test preset', got %q", list[0].Name)
	}
}

func TestStore_Delete(t *testing.T) {
	pool := tests.NewTestDB(t)
	store := presets.NewStore(pool)
	ctx := context.Background()

	saved, err := store.Add(ctx, presets.Preset{
		Name:     "To delete",
		Tags:     []string{},
		Services: []compose.ServiceConfig{},
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if err := store.Delete(ctx, saved.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	list, _ := store.List(ctx)
	if len(list) != 0 {
		t.Errorf("expected 0 presets after delete, got %d", len(list))
	}
}

func TestStore_Delete_NotFound(t *testing.T) {
	pool := tests.NewTestDB(t)
	store := presets.NewStore(pool)
	ctx := context.Background()

	err := store.Delete(ctx, "nonexistent-id")
	if err == nil {
		t.Error("expected error when deleting nonexistent preset")
	}
}

func TestStore_List_Empty(t *testing.T) {
	pool := tests.NewTestDB(t)
	store := presets.NewStore(pool)
	ctx := context.Background()

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if list == nil {
		list = []presets.Preset{}
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestStore_Add_PreservesServices(t *testing.T) {
	pool := tests.NewTestDB(t)
	store := presets.NewStore(pool)
	ctx := context.Background()

	services := []compose.ServiceConfig{
		{Name: "mysql", Image: "mysql", Version: "5.6.51", Ports: []string{"3306:3306"}},
		{Name: "apache", Image: "httpd", Version: "2.4.49", Ports: []string{"8080:80"}},
	}

	_, err := store.Add(ctx, presets.Preset{
		Name:     "Multi service",
		Tags:     []string{"RCE"},
		Services: services,
	})
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	list, _ := store.List(ctx)
	if len(list[0].Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(list[0].Services))
	}
	if list[0].Services[0].Version != "5.6.51" {
		t.Errorf("expected version 5.6.51, got %s", list[0].Services[0].Version)
	}
}

func TestStore_SeedDefaults(t *testing.T) {
	pool := tests.NewTestDB(t)
	store := presets.NewStore(pool)
	ctx := context.Background()

	if err := store.SeedDefaults(ctx); err != nil {
		t.Fatalf("SeedDefaults failed: %v", err)
	}

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 default presets, got %d", len(list))
	}

	if err := store.SeedDefaults(ctx); err != nil {
		t.Fatalf("second SeedDefaults failed: %v", err)
	}
	list, _ = store.List(ctx)
	if len(list) != 3 {
		t.Errorf("expected still 3 presets after second seed, got %d", len(list))
	}
}

package tempolite

import (
	"context"
	"testing"
	"time"
)

func TestVersionBasicOperations(t *testing.T) {
	db := NewMemoryDatabase()
	// registry := NewRegistry()

	// Test adding a version directly
	version := &Version{
		EntityID:  1,
		ChangeID:  "test-change",
		Version:   1,
		Data:      map[string]interface{}{"key": "value"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := db.AddVersion(version)
	if err != nil {
		t.Fatalf("Failed to add version: %v", err)
	}
	if id <= 0 {
		t.Error("Expected positive version ID")
	}

	// Test retrieving the version
	retrieved, err := db.GetVersion(id)
	if err != nil {
		t.Fatalf("Failed to retrieve version: %v", err)
	}
	if retrieved.ChangeID != version.ChangeID {
		t.Errorf("Expected ChangeID %s, got %s", version.ChangeID, retrieved.ChangeID)
	}
	if retrieved.Version != version.Version {
		t.Errorf("Expected Version %d, got %d", version.Version, retrieved.Version)
	}
	if retrieved.Data["key"] != version.Data["key"] {
		t.Errorf("Expected Data key value %v, got %v", version.Data["key"], retrieved.Data["key"])
	}
}

func TestVersionInheritance(t *testing.T) {
	db := NewMemoryDatabase()
	registry := NewRegistry()
	ctx := context.Background()
	orchestrator := NewOrchestrator(ctx, db, registry)

	var firstVersion int
	// Create parent workflow that sets initial versions
	parentWorkflow := func(ctx WorkflowContext) error {
		version, err := ctx.GetVersion("feature-flag", 1, 2)
		if err != nil {
			return err
		}
		firstVersion = version
		if version != 2 { // Should get max version on first run
			t.Errorf("Expected version 2, got %d", version)
		}
		return nil
	}

	// Execute parent workflow
	future := orchestrator.Execute(parentWorkflow, nil)

	if err := future.Get(); err != nil {
		t.Fatalf("Parent workflow failed: %v", err)
	}
	parentID := future.WorkflowID()

	// Verify version was set
	versions, err := db.GetVersionsByWorkflowID(parentID)
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("Expected 1 version, got %d", len(versions))
	}
	if versions[0].ChangeID != "feature-flag" {
		t.Errorf("Expected ChangeID 'feature-flag', got '%s'", versions[0].ChangeID)
	}
	if versions[0].Version != 2 {
		t.Errorf("Expected Version 2, got %d", versions[0].Version)
	}

	// Create child workflow that inherits versions
	childWorkflow := func(ctx WorkflowContext) error {
		version, err := ctx.GetVersion("feature-flag", 1, 2)
		if err != nil {
			return err
		}
		if version != firstVersion {
			t.Errorf("Child got version %d, expected %d", version, firstVersion)
		}
		return nil
	}

	// Execute child workflow
	opts := &WorkflowOptions{
		Queue: "default",
	}

	future = orchestrator.Execute(childWorkflow, opts)

	if err = future.Get(); err != nil {
		t.Fatalf("Child workflow failed: %v", err)
	}
	childID := future.WorkflowID()

	// Verify child inherited version
	childVersions, err := db.GetVersionsByWorkflowID(childID)
	if err != nil {
		t.Fatalf("Failed to get child versions: %v", err)
	}
	if len(childVersions) != 1 {
		t.Fatalf("Expected 1 version for child, got %d", len(childVersions))
	}
	if childVersions[0].ChangeID != "feature-flag" {
		t.Errorf("Expected child ChangeID 'feature-flag', got '%s'", childVersions[0].ChangeID)
	}
	if childVersions[0].Version != firstVersion {
		t.Errorf("Expected child Version %d, got %d", firstVersion, childVersions[0].Version)
	}
}

func TestVersionOverrides(t *testing.T) {
	db := NewMemoryDatabase()
	registry := NewRegistry()
	ctx := context.Background()
	orchestrator := NewOrchestrator(ctx, db, registry)

	workflow := func(ctx WorkflowContext) error {
		version, err := ctx.GetVersion("feature-flag", 1, 3)
		if err != nil {
			return err
		}
		if version != 2 { // Should get overridden version
			t.Errorf("Expected version 2 (override), got %d", version)
		}
		return nil
	}

	// Execute with version override
	opts := &WorkflowOptions{
		VersionOverrides: map[string]int{
			"feature-flag": 2,
		},
	}

	future := orchestrator.Execute(workflow, opts)

	if err := future.Get(); err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}
	workflowID := future.WorkflowID()

	// Verify overridden version was set
	versions, err := db.GetVersionsByWorkflowID(workflowID)
	if err != nil {
		t.Fatalf("Failed to get versions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("Expected 1 version, got %d", len(versions))
	}
	if versions[0].ChangeID != "feature-flag" {
		t.Errorf("Expected ChangeID 'feature-flag', got '%s'", versions[0].ChangeID)
	}
	if versions[0].Version != 2 {
		t.Errorf("Expected Version 2 (override), got %d", versions[0].Version)
	}
}

func TestVersionValidation(t *testing.T) {
	db := NewMemoryDatabase()
	registry := NewRegistry()
	ctx := context.Background()
	orchestrator := NewOrchestrator(ctx, db, registry)

	workflow := func(ctx WorkflowContext) error {
		// First call should succeed with max version
		_, err := ctx.GetVersion("feature-flag", 2, 3)
		if err != nil {
			t.Errorf("First GetVersion call failed: %v", err)
			return err
		}

		// Second call should fail as stored version (3) is outside range
		_, err = ctx.GetVersion("feature-flag", 4, 5)
		if err == nil {
			t.Error("Expected error for version outside range, got nil")
		}

		return nil
	}

	future := orchestrator.Execute(workflow, nil)

	if err := future.Get(); err == nil {
		t.Error("Expected workflow to fail due to version validation")
	}
}

func TestContinueAsNewVersionInheritance(t *testing.T) {
	db := NewMemoryDatabase()
	registry := NewRegistry()
	ctx := context.Background()
	orchestrator := NewOrchestrator(ctx, db, registry)

	var continuationID int
	var initialVersion int

	workflow := func(ctx WorkflowContext) error {
		version, err := ctx.GetVersion("feature-flag", 1, 2)
		if err != nil {
			return err
		}
		initialVersion = version
		return ctx.ContinueAsNew(nil)
	}

	future := orchestrator.Execute(workflow, nil)

	if err := future.Get(); err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}
	initialID := future.WorkflowID()

	// Get versions from both workflows
	versions, err := db.GetVersionsByWorkflowID(continuationID)
	if err != nil {
		t.Fatalf("Failed to get continued versions: %v", err)
	}
	originalVersions, err := db.GetVersionsByWorkflowID(initialID)
	if err != nil {
		t.Fatalf("Failed to get original versions: %v", err)
	}

	// Verify versions match
	if len(versions) != len(originalVersions) {
		t.Errorf("Expected same number of versions, got %d vs %d", len(versions), len(originalVersions))
	}
	if len(versions) > 0 && versions[0].Version != initialVersion {
		t.Errorf("Expected version %d, got %d", initialVersion, versions[0].Version)
	}
}

func TestVersionDataPersistence(t *testing.T) {
	db := NewMemoryDatabase()

	metadata := map[string]interface{}{
		"description": "Test feature flag",
		"owner":       "team-a",
		"enabled":     true,
	}

	version := &Version{
		EntityID:  1,
		ChangeID:  "test-change",
		Version:   1,
		Data:      metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add version
	id, err := db.AddVersion(version)
	if err != nil {
		t.Fatalf("Failed to add version: %v", err)
	}

	// Retrieve and verify metadata persisted
	retrieved, err := db.GetVersion(id)
	if err != nil {
		t.Fatalf("Failed to retrieve version: %v", err)
	}

	for key, expected := range metadata {
		if got := retrieved.Data[key]; got != expected {
			t.Errorf("Data[%s]: expected %v, got %v", key, expected, got)
		}
	}

	// Update metadata
	retrieved.Data["enabled"] = false
	retrieved.UpdatedAt = time.Now()
	if err = db.UpdateVersion(retrieved); err != nil {
		t.Fatalf("Failed to update version: %v", err)
	}

	// Verify update persisted
	updated, err := db.GetVersion(id)
	if err != nil {
		t.Fatalf("Failed to retrieve updated version: %v", err)
	}
	if enabled, _ := updated.Data["enabled"].(bool); enabled != false {
		t.Error("Expected 'enabled' to be false after update")
	}
}

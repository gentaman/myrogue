package content

import (
	"testing"
)

func TestRegistry_Validate_UnknownEnemy(t *testing.T) {
	r := NewRegistry()
	r.Enemies = []ActorDef{{ID: "slime", HP: 10}}
	r.EnemyIDMap = map[string]int{"slime": 0}

	r.Floors = []FloorDef{
		{
			Floor: 0,
			EnemyPool: []PoolEntry{
				{ID: "slime", Weight: 100},
				{ID: "dragon", Weight: 10}, // Unknown ID
			},
		},
	}

	err := r.Validate()
	if err == nil {
		t.Fatal("expected error for unknown enemy ID, got nil")
	}
	expected := "enemy pool contains unknown ID dragon"
	if !contains(err.Error(), expected) {
		t.Errorf("expected error to contain %q, got %q", expected, err.Error())
	}
}

func TestRegistry_Validate_DuplicateID(t *testing.T) {
	data := []byte(`[{"id": "slime", "name": "Slime 1"}, {"id": "slime", "name": "Slime 2"}]`)
	_, _, err := loadActors(data)
	if err == nil {
		t.Fatal("expected error for duplicate actor ID, got nil")
	}
	if !contains(err.Error(), "duplicate ID: slime") {
		t.Errorf("expected duplicate ID error, got %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}

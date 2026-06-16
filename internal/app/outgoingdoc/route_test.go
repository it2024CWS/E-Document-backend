package outgoingdoc

import (
	"testing"

	"github.com/google/uuid"
)

func ptr(id uuid.UUID) *uuid.UUID { return &id }

func contains(route []uuid.UUID, id uuid.UUID) bool {
	for _, r := range route {
		if r == id {
			return true
		}
	}
	return false
}

// Owner is approved on the outgoing side and must NOT appear in the recipient route.
func TestBuildRoute_OwnerExcluded(t *testing.T) {
	owner := uuid.New()
	a, b := uuid.New(), uuid.New()

	got := BuildRoute(owner, []uuid.UUID{a, b}, nil)

	if contains(got, owner) {
		t.Errorf("owner must not be in the recipient route: %v", got)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 recipients, got %d: %v", len(got), got)
	}
	if got[0] != a || got[1] != b {
		t.Errorf("selected order not preserved: %v", got)
	}
}

func TestBuildRoute_DirectorLast(t *testing.T) {
	owner := uuid.New()
	a := uuid.New()
	director := uuid.New()

	got := BuildRoute(owner, []uuid.UUID{a}, ptr(director))

	if len(got) != 2 {
		t.Fatalf("expected 2 recipients, got %d: %v", len(got), got)
	}
	if got[len(got)-1] != director {
		t.Errorf("last dept must be director, got %s", got[len(got)-1])
	}
}

func TestBuildRoute_DirectorOmittedWhenNil(t *testing.T) {
	owner := uuid.New()
	a := uuid.New()

	got := BuildRoute(owner, []uuid.UUID{a}, nil)

	for _, id := range got {
		if id == uuid.Nil {
			t.Error("found nil UUID in route")
		}
	}
	if len(got) != 1 {
		t.Errorf("expected 1 recipient (no director), got %d: %v", len(got), got)
	}
}

func TestBuildRoute_OwnerDeduplicatedFromSelected(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()

	// owner also appears in selected — must be excluded
	got := BuildRoute(owner, []uuid.UUID{owner, other}, nil)

	if contains(got, owner) {
		t.Errorf("owner duplicate not excluded: %v", got)
	}
	if len(got) != 1 || got[0] != other {
		t.Errorf("expected [other], got %v", got)
	}
}

func TestBuildRoute_DirectorDeduplicatedFromSelected(t *testing.T) {
	owner := uuid.New()
	director := uuid.New()

	// director also appears in selected — must appear only once, at end
	got := BuildRoute(owner, []uuid.UUID{director}, ptr(director))

	if len(got) != 1 {
		t.Errorf("director duplicate not removed, got %d: %v", len(got), got)
	}
	if got[0] != director {
		t.Errorf("expected [director], got %v", got)
	}
}

func TestBuildRoute_EmptySelected(t *testing.T) {
	owner := uuid.New()
	director := uuid.New()

	got := BuildRoute(owner, nil, ptr(director))

	if len(got) != 1 {
		t.Fatalf("expected 1, got %d: %v", len(got), got)
	}
	if got[0] != director {
		t.Errorf("expected [director], got %v", got)
	}
}

// With no recipients selected and no director, the route is empty (the upload
// flow treats an empty route as an error).
func TestBuildRoute_EmptyWhenOnlyOwner(t *testing.T) {
	owner := uuid.New()

	got := BuildRoute(owner, []uuid.UUID{owner}, nil)

	if len(got) != 0 {
		t.Errorf("expected empty route, got %v", got)
	}
}

func TestBuildRoute_NilUUIDsInSelectedSkipped(t *testing.T) {
	owner := uuid.New()
	valid := uuid.New()

	got := BuildRoute(owner, []uuid.UUID{uuid.Nil, valid, uuid.Nil}, nil)

	if len(got) != 1 {
		t.Errorf("nil UUIDs should be skipped, got %d: %v", len(got), got)
	}
	if got[0] != valid {
		t.Errorf("expected valid UUID, got %s", got[0])
	}
}

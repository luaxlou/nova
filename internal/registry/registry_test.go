package registry

import (
	"errors"
	"testing"
)

type reloadResource struct {
	closed   bool
	closeErr error
}

func (r *reloadResource) Close() error {
	r.closed = true
	return r.closeErr
}

func TestReloadClosesPreviousResourceBeforeBuildingReplacement(t *testing.T) {
	reg := New[*reloadResource]()
	reg.Register("main", func(string) (*reloadResource, error) {
		return &reloadResource{}, nil
	})
	handle := reg.Named("main")

	first, err := handle.Get()
	if err != nil {
		t.Fatalf("get first resource: %v", err)
	}
	if err := handle.Reload(); err != nil {
		t.Fatalf("reload resource: %v", err)
	}
	second, err := handle.Get()
	if err != nil {
		t.Fatalf("get replacement resource: %v", err)
	}

	if !first.closed {
		t.Fatal("previous resource remained open after Reload")
	}
	if first == second {
		t.Fatal("Reload reused previous resource")
	}
	if second.closed {
		t.Fatal("replacement resource is closed")
	}
}

func TestReloadReturnsCloseErrorWithoutBuildingReplacement(t *testing.T) {
	wantErr := errors.New("close failed")
	builds := 0
	reg := New[*reloadResource]()
	reg.Register("main", func(string) (*reloadResource, error) {
		builds++
		return &reloadResource{closeErr: wantErr}, nil
	})
	handle := reg.Named("main")

	first, err := handle.Get()
	if err != nil {
		t.Fatalf("get first resource: %v", err)
	}
	if err := handle.Reload(); !errors.Is(err, wantErr) {
		t.Fatalf("Reload() error = %v, want %v", err, wantErr)
	}

	if !first.closed {
		t.Fatal("previous resource was not closed")
	}
	if builds != 1 {
		t.Fatalf("builder calls = %d, want 1", builds)
	}
}

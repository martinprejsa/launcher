package tests

import (
	"launcher/api/backend"
	"testing"
)

func TestBackend(t *testing.T) {
	index, err := backend.GetWardrobeIndex()
	if err != nil {
		t.Fatal("Failed to retrieve index", err.Error())
	}

	asset := index["test"]
	t.Log(asset.GetPreviewLink())
}

package tests

import (
	"launcher/api/microsoft"
	"testing"
)

func TestLogin(t *testing.T) {
	ms, err := microsoft.MSAuth(false)
	if err != nil {
		t.Error(err)
	} else {
		_, err := ms.GetMinecraftProfile()
		if err != nil {
			t.Error(err)
		}
	}
}

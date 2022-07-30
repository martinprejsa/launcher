package tests

import (
	"launcher/api/microsoft"
	"testing"
)

func TestQuickLogin(t *testing.T) {
	ms, err := microsoft.MSAuth(true)
	if err != nil {
		t.Error(err)
	} else {
		_, err = ms.GetMinecraftProfile()
		if err != nil {
			t.Error(err)
		}
	}
}

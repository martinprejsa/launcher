package tests

import (
	"launcher/api/microsoft"
	"launcher/manager"
	"testing"
)

func Test(t *testing.T) {
	err := manager.InstallTheOnlyProfile("/home/martin/.genecraft")
	if err != nil {
		t.Error(err)
		return
	}

	game := manager.Explore()[0]
	err = game.InstallMinecraft()
	if err != nil {
		t.Error(err)
		return
	}

	ms, err := microsoft.MSAuth(true)
	if err == microsoft.QuickAuthError {
		ms, err = microsoft.MSAuth(false)
		if err != nil {
			t.Error(err)
			return
		}
	}

	profile, err := ms.GetMinecraftProfile()
	if err != nil {
		t.Error(err)
		return
	}

	game.Launch(manager.Auth{
		Username:    profile.Name,
		AccessToken: profile.AccessToken,
		UUID:        profile.ID,
	}, manager.ClientSettings{
		Memory:  0,
		Width:   0,
		Height:  0,
		JvmArgs: "",
	})
}

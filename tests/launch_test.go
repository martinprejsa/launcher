package tests

import (
	"github.com/pkg/errors"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"launcher/api/microsoft"
	"launcher/logging"
	"launcher/manager"
	"testing"
)

func Test(t *testing.T) {
	logging.Logger = logger.NewDefaultLogger()

	t.Log("Installing profile")
	err := manager.InstallTheOnlyProfile("/home/martin/.genecraft")
	if err != nil {
		t.Error(errors.WithMessage(err, "Failed to create profile"))
		return
	}

	t.Log("Exploring...")
	games := manager.Explore()
	if len(games) == 0 {
		t.Error(errors.New("Profile not installed"))
		return
	}
	t.Log("Installing miencraft")
	err = games[0].InstallMinecraft()
	if err != nil {
		t.Error(errors.WithMessage(err, "Failed to install minecraft"))
		return
	}

	t.Log("Trying quickauth")
	ms, err := microsoft.MSAuth(true)
	if err == microsoft.QuickAuthError {
		t.Log("Quickauth unavailable, authenticating")
		ms, err = microsoft.MSAuth(false)
		if err != nil {
			t.Error(errors.WithMessage(err, "Failed to authenticate"))
			return
		}
	}
	t.Log("Fetching minecraft profile")
	profile, err := ms.GetMinecraftProfile()
	if err != nil {
		t.Error(errors.WithMessage(err, "Failed to obtain minecraft profile"))
		return
	}

	t.Log("Launching game")
	err = games[0].Launch(manager.Auth{
		Username:    profile.Name,
		AccessToken: profile.AccessToken,
		UUID:        profile.ID,
	}, manager.ClientSettings{
		Memory:  0,
		Width:   0,
		Height:  0,
		JvmArgs: "",
	})

	if err != nil {
		t.Error(errors.WithMessage(err, "Failed to launch minecraft"))
		return
	}
}

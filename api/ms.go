package api

import (
	"context"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/pkg/errors"
)

const msalClientId = "048f6903-f7d2-47b7-8d7d-47a2fa08b0f7"

func MSGetAccessToken() (string, error) {
	publicClientApp, err := public.New(msalClientId, public.WithAuthority("https://login.microsoftonline.com/consumers"))
	if err != nil {
		return "", errors.Errorf("failed to initialize MSAL client app: %s", err)
	}
	//TODO: maybe offline_access
	//TODO: FIX change redirect
	//TODO: read from cache
	interactive, err := publicClientApp.AcquireTokenInteractive(context.Background(), []string{"XboxLive.signin"})
	if err != nil {
		return "", err
	}
	return interactive.AccessToken, nil
}

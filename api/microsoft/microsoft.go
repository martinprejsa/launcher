package microsoft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

type MinecraftAuthHandle struct {
	AccessToken string
}

type MinecraftProfile struct {
	AccessToken string
	ID          string `json:"id"`
	Name        string `json:"name"`
	Skins       []MinecraftSkin
	Capes       []MinecraftCape
}

type MinecraftSkin struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	URL     string `json:"url"`
	Variant string `json:"variant"`
	Alias   string `json:"alias"`
}

type MinecraftCape struct {
	ID    string `json:"id"`
	State string `json:"state"`
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type xblResponse struct {
	IssueInstant  string
	NotAfter      string
	Token         string
	DisplayClaims xblResponseDisplayClaims
}

type xblResponseDisplayClaims struct {
	Xui []map[string]interface{} `json:"xui"`
}

const msalClientId = "048f6903-f7d2-47b7-8d7d-47a2fa08b0f7"

func (h *MinecraftAuthHandle) GetMinecraftProfile() (MinecraftProfile, error) {
	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Set("Authorization", "Bearer "+h.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return MinecraftProfile{}, err
	}

	b, _ := ioutil.ReadAll(resp.Body)
	var profile MinecraftProfile
	err = json.Unmarshal(b, &profile)
	if err != nil {
		return MinecraftProfile{}, errors.New("couldn't retrieve the minecraft profile") // user doesnt own the game-legacy
	}

	profile.AccessToken = h.AccessToken
	return profile, nil
}

var (
	cacheAccessor = &TokenCache{"login_cache.json"}
)

func MSAuth() (MinecraftAuthHandle, error) {
	publicClientApp, err := public.New(msalClientId, public.WithAuthority("https://login.microsoftonline.com/consumers"), public.WithCache(cacheAccessor))

	var userAccount public.Account
	var accessToken string

	var authenticate = func() (MinecraftAuthHandle, error) {
		if err != nil {
			return MinecraftAuthHandle{}, errors.Errorf("failed to initialize MSAL client app: %s", err)
		}

		interactive, err := publicClientApp.AcquireTokenInteractive(context.Background(), []string{"XboxLive.signin"})
		if err != nil {
			return MinecraftAuthHandle{}, errors.Errorf("token aquisiton failed: %s", err)
		}

		accessToken, err = auth(interactive.AccessToken)
		if !verifyGameOwnership(accessToken) {
			return MinecraftAuthHandle{}, errors.New("failed to verify game-legacy ownership")
		}

		return MinecraftAuthHandle{AccessToken: accessToken}, err
	}

	accounts := publicClientApp.Accounts()
	if len(accounts) > 0 {
		userAccount = accounts[0]
		result, err := publicClientApp.AcquireTokenSilent(context.Background(), []string{"XboxLive.signin"}, public.WithSilentAccount(userAccount))

		if time.Now().After(result.ExpiresOn) {
			fmt.Println("cache expired, authenticating again") //TODO: log, maybe not needed at all
			return authenticate()
		}

		if err != nil {
			fmt.Println("cache invalid, authenticating again") //TODO: log
			return authenticate()
		} else {
			fmt.Println("cache valid") //TODO: log
			token, err := auth(result.AccessToken)
			return MinecraftAuthHandle{token}, err
		}
	} else {
		return authenticate()
	}
}

func verifyGameOwnership(token string) bool {
	return true //TODO implement
}

func auth(accessToken string) (string, error) {
	data := map[string]interface{}{
		"RelyingParty": "http://auth.xboxlive.com", // Must be http
		"TokenType":    "JWT",
		"Properties": map[string]interface{}{
			"AuthMethod": "RPS",
			"SiteName":   "user.auth.xboxlive.com",
			"RpsTicket":  "d=" + accessToken,
		},
	}

	encoded, _ := json.Marshal(data)

	request, _ := http.NewRequest("POST", "https://user.auth.xboxlive.com/user/authenticate", bytes.NewReader(encoded))

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("x-xbl-contract-version", "1")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", errors.Errorf("failed to authenticate with Xbox Live: %s", err)
	} else {
		b, _ := ioutil.ReadAll(response.Body)
		var xblr xblResponse
		_ = json.Unmarshal(b, &xblr)

		body := map[string]interface{}{
			"Properties": map[string]interface{}{
				"SandboxId": "RETAIL",
				"UserTokens": []string{
					xblr.Token,
				},
			},
			"RelyingParty": "rp://api.minecraftservices.com/",
			"TokenType":    "JWT",
		}

		encoded, _ := json.Marshal(body)

		request, _ := http.NewRequest("POST", "https://xsts.auth.xboxlive.com/xsts/authorize", bytes.NewReader(encoded))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json")
		request.Header.Set("x-xbl-contract-version", "1")

		response, err = http.DefaultClient.Do(request)

		if err != nil {
			return "", errors.Errorf("failed to authenticate with Xbox Live security: %s", err)
		} else {
			b, _ := ioutil.ReadAll(response.Body)
			var jsonResponse map[string]interface{}
			json.Unmarshal(b, &jsonResponse)
			token := jsonResponse["Token"].(string)
			uhs := xblr.DisplayClaims.Xui[0]["uhs"].(string)
			return minecraftAuth(token, uhs), nil
		}
	}
}

func minecraftAuth(xstx string, uhs string) string {
	body := map[string]interface{}{}

	body["identityToken"] = "XBL3.0 x=" + uhs + ";" + xstx
	body["ensureLegacyEnabled"] = true
	encoded, _ := json.Marshal(body)

	request, _ := http.NewRequest("POST", "https://api.minecraftservices.com/authentication/login_with_xbox", bytes.NewReader(encoded))
	response, _ := http.DefaultClient.Do(request)

	jsonResponse, _ := ioutil.ReadAll(response.Body)
	var data map[string]interface{}

	json.Unmarshal(jsonResponse, &data)
	return data["access_token"].(string)
}

// i have no idea what ive done, but im glad it works

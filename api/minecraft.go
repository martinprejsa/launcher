package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

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

func GetMinecraftToken() (string, error) {
	publicClientApp, err := public.New(msalClientId, public.WithAuthority("https://login.microsoftonline.com/consumers"))
	if err != nil {
		return "", errors.Errorf("failed to initialize MSAL client app: %s", err)
	}
	//TODO: read from cache
	interactive, err := publicClientApp.AcquireTokenInteractive(context.Background(), []string{"XboxLive.signin"})
	if err != nil {
		return "", errors.Errorf("token aquisiton failed: %s", err)
	}

	return xblAuth(interactive.AccessToken)
}

func xblAuth(accessToken string) (string, error) {
	data := map[string]interface{}{
		"RelyingParty": "http://auth.xboxlive.com",
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
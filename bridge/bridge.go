package bridge

import (
	"context"
	"github.com/pkg/errors"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"launcher/api/microsoft"
	"launcher/events"
	"launcher/manager"
	"launcher/memory"
	"os"
)

type Bridge struct {
	ctx      context.Context
	Profile  microsoft.MinecraftProfile
	GameInfo GameInfo
	Progress float64
}

type ProfileInfo struct {
	Username       string `json:"username"`
	ProfilePicture string `json:"profile_picture"`
}

type HardwareInfo struct {
	MemorySize int `json:"memory_size"`
}

type GameInfo struct {
	IsInstalled bool `json:"isInstalled"`
}

func InitBridge() *Bridge {
	return &Bridge{}
}

/* JS API BEGIN */

type ClientSettings struct {
	KeepOpen         bool `json:"keep_open"`
	Memory           int  `json:"memory"`
	ResolutionWidth  int  `json:"resolution_width"`
	ResolutionHeight int  `json:"resolution_height"`
}

// IsAuthenticated returns the authentication status
func (a *Bridge) IsAuthenticated() bool {
	return false
}

// GetProgress returns the progress, -1 when none
func (a *Bridge) GetProgress() float64 {
	return a.Progress
}

// GetWalletData returns wallet data
func (a *Bridge) GetWalletData() string {
	//TODO: Get from file
	return "[{\"network_name\":\"Solona\",\"network_icon\":\"https://cryptologos.cc/logos/solana-sol-logo.svg?v=022\",\"network_wallets\":[{\"wallet_name\":\"Phantom\",\"wallet_icon\":\"\",\"wallet_link\":\"\"},{\"wallet_name\":\"Coinbase\",\"wallet_icon\":\"https://logosarchive.com/wp-content/uploads/2021/12/Coinbase-icon-symbol-1.svg\",\"wallet_link\":\"\"},{\"wallet_name\":\"Ledger\",\"wallet_icon\":\"\",\"wallet_link\":\"\"},{\"wallet_name\":\"Solfare\",\"wallet_icon\":\"\",\"wallet_link\":\"\"},{\"wallet_name\":\"Soilet\",\"wallet_icon\":\"\",\"wallet_link\":\"\"}]}]"
}

// GetWardrobeData returns wardrobe data
func (a *Bridge) GetWardrobeData() string {
	//TODO: Get from file
	return "[{\"collection_name\":\"BAYC\",\"collection_count\":5,\"collection_items\":[{\"item_id\":8520,\"item_preview\":\"https://lh3.googleusercontent.com/ULjfyo4LJhtV3J9K7lu1xh0YZQBa6WHPp-cwlV2C9sUIyTpgSlv554mh_97fRXsziOIu9xwpukl5NQoDbkE3mlXlWR8zU7qcWQsxVg=w213\"},{\"item_id\":2815,\"item_preview\":\"https://lh3.googleusercontent.com/rhtU2-SYlGptvF9jeyNuANX7BvaY49hQj2hkds7rZAPHV0z1LJpeuRDpJ9GuzcGL9lZ5jsMUH1yz1zMluRU7GoN98ehY09cCW_aR=w213\"},{\"item_id\":3368,\"item_preview\":\"https://lh3.googleusercontent.com/u-2FnHbaJ3U_KCDlmg2McX9Yfo7brsAzOffqihNXCGkHljA89SPPzwdjQiVSWcsvxCoj_ydBcDNCuZvHEekaYekaMEH4XX32k9US=w213\"},{\"item_id\":6226,\"item_preview\":\"https://lh3.googleusercontent.com/Yf4jFbxs54kGI74jY4D_Cmb1jqkLf9kwFo0gHK7Znwrvib-BLUs4cYj6bl4Dzao0Nv-gmGG4K9wJJ3yVhpWk-M09RS7ofxJFgtQF0w=w213\"},{\"item_id\":2672,\"item_preview\":\"https://lh3.googleusercontent.com/obZUkkJf0KQz5BkBBPmhZmqVf5ddEncLetfe6ou8CDvuzV-ldbqpHw1fc-Bwj9aa0m2i2zXuRNgVixHX5Wr0NTpzFF3C5M7hr1K4zQ=w213\"}]},{\"collection_name\":\"MAYC\",\"collection_count\":4,\"collection_items\":[{\"item_id\":3462,\"item_preview\":\"https://lh3.googleusercontent.com/frf9agk8KxH1_YOhN1iagUu14quTtPk7IWQCQM563kEKBem2gp-yGsh8D7LiVo1eNddamVU2VD056w3Pa_OmPKs-Yv7J39pweOB6=w213\"},{\"item_id\":22628,\"item_preview\":\"https://lh3.googleusercontent.com/Gm8mzuE1ckkFrl6PpwXlafCCTkhTlXPsdwYIYV9z9JFSIFvF7MZZTTIw8Neayx-ps4sdFFd6fcLzlfkBnfVyiCRuoyZpfHnS-vn-=w213\"},{\"item_id\":18784,\"item_preview\":\"https://lh3.googleusercontent.com/TXo5IFUopNxzBFhIFrYecx18xQaqVI8kw5i4pZpbZUE0dPbPsnwP7xrmlTufATYm7KXry_uENhT5AeV621s2GrygAA-y4QcvNulY1Q=w213\"},{\"item_id\":19149,\"item_preview\":\"https://lh3.googleusercontent.com/G3YUcbYqAmFUja1H0zLxXxWmwDMe6fJy9vAPfzqwQ_CdY5oYRkD7XNIEY0HLS9ZQuug96vGEt0r2KoabuTicASmTKAEAg7evPJBr8g=w213\"}]},{\"collection_name\":\"CryptoPunks\",\"collection_count\":7,\"collection_items\":[{\"item_id\":8308,\"item_preview\":\"https://lh3.googleusercontent.com/suavwp2FYCd55Pfoz0GSeb60HecKgESJ1CgzMViNWI5S979uks1ncg_QFSFvoqar7nTuHiulGXhlROSBEg-lDKnZoA=w213\"},{\"item_id\":8316,\"item_preview\":\"https://lh3.googleusercontent.com/Syl3Wyp4x40yBYVisp2d2dXCsz6WtcA-70JN-LNiolR0IEe5ybNwoqDaD9W3rI0PH2prY-J7wx6T5En8F4d7qPWgcH-NFMTRC9Cg=w213\"},{\"item_id\":8318,\"item_preview\":\"https://lh3.googleusercontent.com/-8ULNXLRfsHpNDmfjwf2B_ziI6-dQk1IaQtozXKKDPxJvzsxA7JzmI4JAdkjC-vSYmkxI7A7iJcNW1-9_QK6Gets9HFyNHNhi2Jdhw=w213\"},{\"item_id\":8337,\"item_preview\":\"https://lh3.googleusercontent.com/23_2HUfVyx-a4Kqm6LmcvNftbAi5ya82jr4nk1kKX3iLLHLsLK7P0vzOCjUPpnXh_AKASTsI0GizR8uCqmM7-alLnBs2BO1ucjFtKA=w213\"},{\"item_id\":9998,\"item_preview\":\"https://lh3.googleusercontent.com/O0TPreCr-fnuhYTUGwHPfp3gZgqwAogRrdmkm60Aiozg9kTuyMeIKc_A0I_yBNIJfoISRuGllSHsatOjxxMWHMMxMOhMbpOJ43wM8A=w213\"},{\"item_id\":6965,\"item_preview\":\"https://lh3.googleusercontent.com/ClS4KdMPO8_25m1yYx-oUJrgeFi_C85dJbfzOPFUJbR_SRCsyoHd4ZapoMvybh4jLWk3BFiMxUSn8CEA_EHrTsXy=w213\"},{\"item_id\":1321,\"item_preview\":\"https://lh3.googleusercontent.com/hBEz_urvzf15Y6AlCJPe0YCkt_XVNx5qBrnUsGMwPos57dvgQUC-0TEPdCVmiQ_OMpn3SlkWJacE2ZM35u_xBv2Q7dq5zj1crlP-Rw=w213\"}]}]"
}

// Authenticate launches the authentication sequence
func (a *Bridge) Authenticate() (ProfileInfo, error) {
	rsp, err := microsoft.MinecraftAuth()
	if err != nil {
		return ProfileInfo{}, errors.New("failed to authenticate")
	}
	profile, err := rsp.GetMinecraftProfile()
	a.Profile = profile

	if err != nil {
		return ProfileInfo{}, errors.New("failed to obtain minecraft profile")
	}
	return ProfileInfo{profile.Name, ""}, nil
}

// GetGameInfo returns game information
func (a *Bridge) GetGameInfo() GameInfo {
	return a.GameInfo
}

// GetHardwareInfo returns hardware information
func (a *Bridge) GetHardwareInfo() HardwareInfo {
	return HardwareInfo{
		memory.GetMemoryTotal(),
	}
}

// InstallGame installs the game, can be used for reinstall, use GetProgress to monitor
func (a *Bridge) InstallGame() error {
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 0})
	err := manager.CreateProfile("latest")
	games := manager.Explore()
	err = games[0].InstallMinecraft()
	if err != nil {
		return errors.WithMessage(err, "failed to install native minecraft client")
	}
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 100})
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: -1})
	return err
}

// LaunchGame launches the game, use GetProgress to monitor
func (a *Bridge) LaunchGame() error {
	if a.Profile.AccessToken != "" {
		runtime.WindowHide(a.ctx)
		games := manager.Explore()
		games[0].Launch(manager.Auth{
			Username:    a.Profile.Name,
			AccessToken: a.Profile.AccessToken,
			UUID:        a.Profile.ID,
		})
		os.Exit(0)

		return nil
	} else {
		return errors.New("not authorized")
	}
}

/* JS API END */

/* PRIVATE REGION */

func (a *Bridge) Startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
	a.GameInfo.IsInstalled = len(manager.Explore()) > 0
	progressHandler := progressUpdatedNotifier{
		a,
	}
	events.ProgressUpdateEvent.Register(progressHandler)
}

type progressUpdatedNotifier struct {
	b *Bridge
}

func (p progressUpdatedNotifier) Handle(payload events.ProgressUpdateEventPayload) {
	p.b.Progress = payload.Progress
}

package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io/ioutil"
	"launcher/api/backend"
	"launcher/api/microsoft"
	"launcher/events"
	"launcher/logging"
	"launcher/manager"
	"launcher/manager/comp"
	"launcher/memory"
	"os"
	"path/filepath"
	"time"
)

type Bridge struct {
	ctx      context.Context
	profile  microsoft.MinecraftProfile
	gameInfo GameInfo
	progress events.ProgressUpdateEventPayload
	settings manager.LauncherClientSettings
}

type ProfileInfo struct {
	Username       string `json:"username"`
	UUID           string `json:"uuid"`
	ProfilePicture string `json:"profile_picture"`
}

type HardwareInfo struct {
	MemorySize int `json:"memory_size"`
}

type GameInfo struct {
	IsInstalled bool `json:"isInstalled"`
}

type WardrobeData struct {
	Entries []Entry `json:"entries"`
}

type Entry struct {
	Name    string `json:"name"`
	Preview string `json:"preview"`
}

func InitBridge() (*Bridge, error) {
	var bridge Bridge

	file := filepath.Join(comp.GetLauncherRoot(), "launcher-logs", time.Now().Format("2006-01-02-15-04-05")+".log")
	err := os.MkdirAll(filepath.Join(comp.GetLauncherRoot(), "launcher-logs"), os.ModePerm)
	if err != nil {
		return &Bridge{}, errors.WithMessage(err, "failed to create directory launcher-logs in launcher root directory")
	}
	logging.Logger = logger.NewFileLogger(file)
	logging.Logger.Print("Initialized")

	data, err := ioutil.ReadFile(filepath.Join(comp.GetLauncherRoot(), "launcher_config.json"))
	if err == nil {
		var settings manager.LauncherClientSettings
		err := json.Unmarshal(data, &settings)
		if err != nil {
			logging.Logger.Warning("failed to parse launcher_config.json: " + err.Error())
			fmt.Println("eyo2")
		} else {
			bridge.settings = settings
		}
	} else {
		logging.Logger.Warning("failed to read launcher_config.json: " + err.Error())
	}
	return &bridge, nil
}

func (a *Bridge) TerminateBridge(ctx context.Context) bool {
	b, _ := json.Marshal(a.settings)
	err := ioutil.WriteFile(filepath.Join(comp.GetLauncherRoot(), "launcher_config.json"), b, os.ModePerm)
	if err != nil {
		return false
	}
	return false
}

/* JS API BEGIN */

// IsLoginCachePresent returns if cache is present
func (a *Bridge) IsLoginCachePresent() bool {
	return microsoft.IsCachePresent(comp.GetLauncherRoot())
}

// IsAuthenticated returns the authentication status
func (a *Bridge) IsAuthenticated() bool {
	if a.profile.AccessToken != "" {
		return true
	} else {
		return false
	}
}

// GetProgress returns the progress, -1 when none
func (a *Bridge) GetProgress() float64 {
	return a.progress.Progress
}

// GetProgressMessage returns the progress message, empty when none
func (a *Bridge) GetProgressMessage() string {
	return a.progress.Message
}

// GetWalletData returns wallet data
func (a *Bridge) GetWalletData() string {
	//TODO: Get from file
	return "[{\"network_name\":\"Solona\",\"network_icon\":\"https://cryptologos.cc/logos/solana-sol-logo.svg?v=022\",\"network_wallets\":[{\"wallet_name\":\"Phantom\",\"wallet_icon\":\"\",\"wallet_link\":\"\"},{\"wallet_name\":\"Coinbase\",\"wallet_icon\":\"https://logosarchive.com/wp-content/uploads/2021/12/Coinbase-icon-symbol-1.svg\",\"wallet_link\":\"\"},{\"wallet_name\":\"Ledger\",\"wallet_icon\":\"\",\"wallet_link\":\"\"},{\"wallet_name\":\"Solfare\",\"wallet_icon\":\"\",\"wallet_link\":\"\"},{\"wallet_name\":\"Soilet\",\"wallet_icon\":\"\",\"wallet_link\":\"\"}]}]"
}

// GetWardrobeData returns wardrobe data
func (a *Bridge) GetWardrobeData() (WardrobeData, error) {
	i, err := backend.GetWardrobeIndex()
	if err != nil {
		logging.Logger.Error("Failed to retrieve wardrobe index: " + err.Error())
		return WardrobeData{}, errors.New("Failed to retrieve wardrobe index: " + err.Error())
	}
	var w WardrobeData
	for s, entry := range i {
		w.Entries = append(w.Entries, Entry{
			Name:    s,
			Preview: entry.GetPreviewLink(),
		})
	}
	return w, nil
}

// QuickAuth tries the quick authorization using cache
func (a *Bridge) QuickAuth() (ProfileInfo, error) {
	rsp, err := microsoft.MSAuth(true)
	if err == microsoft.QuickAuthError {
		logging.Logger.Error("Quick auth failed")
		return ProfileInfo{}, err
	}
	return a.getProfile(rsp)
}

// Authenticate launches the authentication sequence
func (a *Bridge) Authenticate() (ProfileInfo, error) {
	rsp, err := microsoft.MSAuth(false)
	if err != nil {
		logging.Logger.Error("Failed to authorize: " + err.Error())
		return ProfileInfo{}, errors.New("failed to authorize")
	}

	return a.getProfile(rsp)
}

// GetGameInfo returns game information
func (a *Bridge) GetGameInfo() GameInfo {
	return a.gameInfo
}

// GetHardwareInfo returns hardware information
func (a *Bridge) GetHardwareInfo() HardwareInfo {
	return HardwareInfo{
		memory.GetMemoryTotal(),
	}
}

func (a *Bridge) GetCurrentSettings() manager.LauncherClientSettings {
	return a.settings
}

// InstallGame installs the game, can be used for reinstall, use GetProgress to monitor
func (a *Bridge) InstallGame() error {
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 0, Message: "Creating profile"})
	err := manager.CreateProfile("latest")
	if err != nil {
		logging.Logger.Error("Failed to create profile, caused by: " + err.Error())
		return errors.WithMessage(err, "failed to create profile")
	}
	games := manager.Explore()
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 90, Message: "Installing Minecraft"})
	err = games[0].InstallMinecraft()
	if err != nil {
		logging.Logger.Error("Failed to install minecraft, caused by: " + err.Error())
		return errors.WithMessage(err, "failed to install native minecraft client")
	}
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 100, Message: "Finishing up"})
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: -1})
	return err
}

// LaunchGame launches the game, use GetProgress to monitor
func (a *Bridge) LaunchGame() error {
	if a.profile.AccessToken != "" {
		runtime.WindowHide(a.ctx)
		games := manager.Explore()

		err := games[0].Launch(manager.LauncherAuth{
			Username:    a.profile.Name,
			AccessToken: a.profile.AccessToken,
			UUID:        a.profile.ID,
		}, a.settings)

		if err != nil {
			return errors.WithMessage(err, "failed to launch game")
		} else {
			os.Exit(0)
			return nil
		}

	} else {
		return errors.New("not authorized")
	}
}

func (a *Bridge) SetClientSettings(settings manager.LauncherClientSettings) {
	a.settings = settings
}

/* JS API END */

/* PRIVATE REGION */

func (a *Bridge) getProfile(handle microsoft.MSAuthHandle) (ProfileInfo, error) {
	profile, err := handle.GetMinecraftProfile()
	if err != nil {
		logging.Logger.Error("Failed to obtain minecraft profile info, caused by: " + err.Error())
		return ProfileInfo{}, errors.New("failed to obtain minecraft profile")
	}
	a.profile = profile
	return ProfileInfo{
		Username:       profile.Name,
		UUID:           profile.ID,
		ProfilePicture: "",
	}, nil
}

func (a *Bridge) Startup(ctx context.Context) {
	// Perform your setup here
	_ = os.Chdir(comp.GetLauncherRoot())
	a.ctx = ctx
	a.gameInfo.IsInstalled = len(manager.Explore()) > 0
	progressHandler := progressUpdatedNotifier{
		a,
	}
	events.ProgressUpdateEvent.Register(progressHandler)
}

type progressUpdatedNotifier struct {
	b *Bridge
}

func (p progressUpdatedNotifier) Handle(payload events.ProgressUpdateEventPayload) {
	p.b.progress = payload
}

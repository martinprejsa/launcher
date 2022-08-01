package manager

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"launcher/logging"
	"launcher/manager/comp"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LauncherHandle struct {
	manifest Manifest
}

type LauncherProfile struct {
	Name      string
	Config    string
	JAR       string
	Manifest  Manifest
	Version   Version
	LogCfg    string
	assets    map[string]Asset
	libraries []Library
}

type LauncherAuth struct {
	Username    string
	AccessToken string
	UUID        string
}

type LauncherClientSettings struct {
	Memory  int    `json:"memory"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	JvmArgs string `json:"jvm_args"`
}

func InitLauncher() (LauncherHandle, error) {
	mf, err := GetManifest()
	if err != nil {
		return LauncherHandle{}, errors.WithMessage(err, "failed to initialize launcher")
	}
	return LauncherHandle{mf}, nil
}

func Explore() []LauncherProfile {
	var profiles []LauncherProfile
	dir, _ := ioutil.ReadDir(filepath.Join(comp.GetLauncherRoot(), "versions"))

	for _, profile := range dir {
		if profile.IsDir() {
			mf, _ := GetManifest()
			ver, _ := mf.GetVersion(GlobalMinecraftVersion) //TODO version
			assets, err := ver.GetAssets()
			if err == nil {
				profiles = append(
					profiles, LauncherProfile{
						Name:      profile.Name(),
						Config:    filepath.Join(comp.GetLauncherRoot(), "versions", profile.Name(), profile.Name()+".json"),
						JAR:       filepath.Join(comp.GetLauncherRoot(), "versions", profile.Name(), profile.Name()+".jar"),
						Manifest:  mf,
						Version:   ver,
						LogCfg:    filepath.Join(comp.GetLogCfgsPath(), ver.Logging.Client.File.Url),
						assets:    assets,
						libraries: ver.Libraries,
					})
			} else {
				logging.Logger.Error(fmt.Sprintf("Failed to download assets for profile %s", profile.Name()))
			}
		}
	}

	return profiles
}

// VerifyAssets verifies the game assets, and returns the names of missing or corrupt ones
func (a *LauncherProfile) VerifyAssets() []string {
	var names []string
	for name, a := range a.assets {
		if _, err := os.Stat(filepath.Join(comp.GetAssetsPath(), "objects", a.Hash[0:2], a.Hash)); err == os.ErrNotExist {
			fmt.Println(name + " missing ")
			names = append(names, name)
		} else {
			if !checkSHA1Hash(filepath.Join(comp.GetAssetsPath(), "objects", a.Hash[0:2], a.Hash), a.Hash) {
				fmt.Println(name + " hash bad ")
				names = append(names, name)
			}
		}
	}

	return names
}

// VerifyLibraries verifies the game libraries, and returns the names of missing or corrupt ones
func (a *LauncherProfile) VerifyLibraries() []string {
	var names []string
	for _, a := range a.libraries {
		cont := true
		for _, rule := range a.Rules {
			if !rule.Complies() {
				cont = false
			}
		}
		if cont {
			if _, err := os.Stat(filepath.Join(comp.GetLibraryPath(), a.Downloads.Artifact.Path)); err == os.ErrNotExist {
				fmt.Println(a.Name + " missing ")
				names = append(names, a.Name)
			} else {
				if !checkSHA1Hash(filepath.Join(comp.GetLibraryPath(), a.Downloads.Artifact.Path), a.Downloads.Artifact.SHA1) {
					fmt.Println(a.Name + " hash bad ")
					names = append(names, a.Name)
				}
			}
		}
	}
	return names
}

func (a *LauncherProfile) InstallMinecraft() error {
	h, err := os.Open(a.JAR)
	if err != nil {
		return err
	}
	s, err := h.Stat()
	if err != nil {
		return err
	}

	if s.Size() == 0 {
		err := installMinecraft(a.JAR, a.Version)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateProfile(kind string) error {
	if kind == "fabric-latest" {
		//TODO this kind shit
	}

	return InstallTheOnlyProfile(comp.GetLauncherRoot())
}

func (a *LauncherProfile) Launch(auth LauncherAuth, settings LauncherClientSettings) error {
	if len(a.VerifyAssets())+len(a.VerifyLibraries()) != 0 {
		return errors.New("failed to verify game files, please reinstall")
	}
	parseFabricManifest := func() map[string]interface{} {
		h, err := os.Open(a.Config)
		if err != nil {
			return map[string]interface{}{} // ignore non existing config
		}
		defer h.Close()

		var data map[string]interface{}
		b, _ := ioutil.ReadAll(h)
		_ = json.Unmarshal(b, &data)
		return data
	}

	toPath := func(s string) string {
		seg := strings.Split(s, ":")
		pkg := strings.Split(seg[0], ".")
		return filepath.Join(comp.GetLibraryPath(), filepath.Join(pkg...), seg[len(seg)-2], seg[len(seg)-1], seg[len(seg)-2]+"-"+seg[len(seg)-1]+".jar")
	}

	fabricmf := parseFabricManifest()
	libs := fabricmf["libraries"].([]interface{})
	var extra []string
	for _, l := range libs {
		extra = append(extra, toPath(l.(map[string]interface{})["name"].(string)))
	}

	version := GlobalMinecraftVersion //TODO implement version

	extraJvmArgs := strings.Split(settings.JvmArgs, " ")

	jvm, game := a.Version.CreateCommandLine(a.JAR, LaunchPlaceholders{
		NativesDirectory: ".",
		LauncherName:     "Genecraft Launcher",
		LauncherVersion:  "1.0",
		Username:         auth.Username,
		Version:          version,
		GameDir:          comp.GetLauncherRoot(),
		AssetDir:         comp.GetAssetsPath(),
		AssetIndex:       version,
		UUID:             auth.UUID,
		AccessToken:      auth.AccessToken,
		ClientID:         "",
		XUID:             "",
		UserType:         "msa",
		VersionType:      "release",
		LogCfgPath:       a.LogCfg,
	}, LaunchOptions{
		Width:  settings.Width,
		Height: settings.Height,
		MaxRam: settings.Memory,
	},
		extra, extraJvmArgs)

	args := append(jvm, fabricmf["mainClass"].(string))
	args = append(args, game...)
	cmd := exec.Command("java", args...)
	cmd.Dir = comp.GetLauncherRoot()
	cmd.Stdout = nil
	fmt.Println(cmd.String())

	//TODO: log command
	err := cmd.Run()
	return err
}

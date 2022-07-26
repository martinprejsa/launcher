package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"launcher/manager/comp"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Profile struct {
	Name     string
	Config   string
	JAR      string
	Manifest Manifest
	Version  Version
	LogCfg   string
}

type Auth struct {
	Username    string
	AccessToken string
	UUID        string
}

type ClientSettings struct {
	Memory           int    `json:"memory"`
	ResolutionWidth  int    `json:"resolution_width"`
	ResolutionHeight int    `json:"resolution_height"`
	JvmArgs          string `json:"jvm_args"`
}

func Explore() []Profile {
	var profiles []Profile
	dir, _ := ioutil.ReadDir(filepath.Join(comp.GetLauncherRoot(), "versions"))

	for _, profile := range dir {
		if profile.IsDir() {
			mf, _ := GetManifest()
			ver, _ := mf.GetLatestVersion() //TODO version
			profiles = append(
				profiles, Profile{
					Name:     profile.Name(),
					Config:   filepath.Join(comp.GetLauncherRoot(), "versions", profile.Name(), profile.Name()+".json"),
					JAR:      filepath.Join(comp.GetLauncherRoot(), "versions", profile.Name(), profile.Name()+".jar"),
					Manifest: mf,
					Version:  ver,
					LogCfg:   filepath.Join(comp.GetLogCfgsPath(), ver.Logging.File.ID),
				})
		}
	}

	return profiles
}

func (p *Profile) InstallMinecraft() error {
	h, err := os.Open(p.JAR)
	if err != nil {
		return err
	}
	s, err := h.Stat()
	if err != nil {
		return err
	}

	if s.Size() == 0 {
		err := installMinecraft(p.JAR, p.Version)
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

func (p *Profile) Launch(auth Auth, settings ClientSettings) {
	parseFabricManifest := func() map[string]interface{} {
		h, err := os.Open(p.Config)
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

	version := "1.19" //TODO implement version

	extraJvmArgs := strings.Split(settings.JvmArgs, " ")

	jvm, game := p.Version.CreateCommandLine(p.JAR, LaunchPlaceholders{
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
		Width:            settings.ResolutionWidth,
		Height:           settings.ResolutionHeight,
		MaxRam:           settings.Memory,
		LogCfgPath:       p.LogCfg,
	}, extra, extraJvmArgs)

	args := append(jvm, fabricmf["mainClass"].(string))
	args = append(args, game...)
	cmd := exec.Command("java", args...)
	fmt.Println(cmd.String())
	//TODO: log command
	cmd.Run()
}

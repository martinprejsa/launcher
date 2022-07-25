package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
}

type Auth struct {
	Username    string
	AccessToken string
	UUID        string
}

func Explore() []Profile {
	var profiles []Profile
	dir, _ := ioutil.ReadDir(filepath.Join(GetLauncherRoot(), "versions"))

	for _, profile := range dir {
		if profile.IsDir() {
			mf, _ := GetManifest()
			ver, _ := mf.GetLatestVersion() //TODO version
			profiles = append(
				profiles, Profile{
					Name:     profile.Name(),
					Config:   filepath.Join(GetLauncherRoot(), "versions", profile.Name(), profile.Name()+".json"),
					JAR:      filepath.Join(GetLauncherRoot(), "versions", profile.Name(), profile.Name()+".jar"),
					Manifest: mf,
					Version:  ver,
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

	return InstallTheOnlyProfile(GetLauncherRoot())
}

func (p *Profile) Launch(auth Auth) {
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
		return filepath.Join(GetLibraryPath(), filepath.Join(pkg...), seg[len(seg)-2], seg[len(seg)-1], seg[len(seg)-2]+"-"+seg[len(seg)-1]+".jar")
	} //TODO maybe completely rework how libraries are saved, dont use folders

	fabricmf := parseFabricManifest()
	libs := fabricmf["libraries"].([]interface{})
	var extra = []string{}
	for _, l := range libs {
		extra = append(extra, toPath(l.(map[string]interface{})["name"].(string)))
	}

	version := "1.19"

	jvm, game := p.Version.CreateCommandLine(p.JAR, LaunchPlaceholders{
		".",
		"Genecraft launcher",
		"1.0",
		auth.Username,
		version,
		GetLauncherRoot(),
		GetAssetsPath(),
		version,
		auth.UUID,
		auth.AccessToken,
		"",
		"",
		"msa",
		"release"}, extra)

	args := append(jvm, fabricmf["mainClass"].(string))
	args = append(args, game...)
	cmd := exec.Command("java", args...)
	fmt.Println(cmd.String())
	str, _ := cmd.CombinedOutput()
	fmt.Println(string(str))
	//TODO: log command
	//cmd.Run()
}

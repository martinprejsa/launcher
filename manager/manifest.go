package manager

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"launcher/manager/comp"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const versionManifestUrl = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

type Manifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		Url         string `json:"url"`
		Time        string `json:"time"`
		ReleaseTime string `json:"release_time"`
	} `json:"versions"`
}
type Version struct {
	Arguments struct {
		JVM  []any `json:"jvm"`
		Game []any `json:"game"`
	} `json:"arguments"`
	AssetIndex struct {
		ID        string `json:"id"`
		SHA1      string `json:"sha1"`
		Size      int64  `json:"size"`
		TotalSize int    `json:"totalSize"`
		Url       string `json:"url"`
	} `json:"assetIndex"`
	Downloads map[string]struct {
		SHA1 string `json:"sha1"`
		Size int64  `json:"size"`
		Url  string `json:"url"`
	} `json:"downloads"`
	Assets    string `json:"assets"`
	Libraries []Library
	Logging   struct {
		Client struct {
			Argument string `json:"argument"`
		} `json:"client"`
		File struct {
			ID   string `json:"id"`
			SHA1 string `json:"sha1"`
			Size string `json:"size"`
			Url  string `json:"url"`
		} `json:"file"`
	} `json:"logging"`
}
type Library struct {
	Downloads struct {
		Artifact struct {
			Path string
			SHA1 string
			Size int64
			Url  string
		}
	}
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}
type Rule struct {
	Action string `json:"action"`
	OS     struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Arch    string `json:"arch"`
	} `json:"os"`
}
type Asset struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

type LaunchPlaceholders struct {
	NativesDirectory string `placeholder:"natives_directory"`
	LauncherName     string `placeholder:"launcher_name"`
	LauncherVersion  string `placeholder:"launcher_version"`
	Username         string `placeholder:"auth_player_name"`
	Version          string `placeholder:"version_name"`
	GameDir          string `placeholder:"game_directory"`
	AssetDir         string `placeholder:"assets_root"`
	AssetIndex       string `placeholder:"assets_index_name"`
	UUID             string `placeholder:"auth_uuid"`
	AccessToken      string `placeholder:"auth_access_token"`
	ClientID         string `placeholder:"clientid"`
	XUID             string `placeholder:"auth_xuid"`
	UserType         string `placeholder:"user_type"`
	VersionType      string `placeholder:"version_type"`
	Width            int    `placeholder:"width"`
	Height           int    `placeholder:"height"`
	MaxRam           int    `placeholder:"max_ram"`
	LogCfgPath       string `placeholder:"path"`
}

func GetManifest() (Manifest, error) {
	var mf Manifest
	err := receiveJSONObject(versionManifestUrl, &mf)
	if err != nil {
		return Manifest{}, err
	}
	return mf, nil
}

func (mf *Manifest) GetLatestVersion() (Version, error) {
	for i := range mf.Versions {
		if mf.Versions[i].ID == mf.Latest.Release {
			var ret Version
			err := receiveJSONObject(mf.Versions[i].Url, &ret)
			if err != nil {
				return Version{}, err
			}
			return ret, nil
		}
	}
	return Version{}, errors.Errorf("latest release version \"%s\" not found in the manifest file", mf.Latest.Release)
}
func (v *Version) GetAssets() (map[string]Asset, error) {
	var ret assetIndex
	err := receiveJSONObject(v.AssetIndex.Url, &ret)
	if err != nil {
		return map[string]Asset{}, err
	}
	return ret.Objects, nil
}

func (v *Version) CreateCommandLine(gameJar string, placeholders LaunchPlaceholders, extraLibs []string, extraArgs []string) ([]string, []string) {
	var jvm []string
	var game []string

	replacePlaceholders := func(s string) string {
		rpl := func(s string, key string, value string) string {
			if strings.ContainsRune(value, ' ') {
				return strings.Replace(s, fmt.Sprintf("${%s}", key), fmt.Sprintf("\"%s\"", value), -1) //If contains spaces, enclose in double quotes
			} else {
				return strings.Replace(s, fmt.Sprintf("${%s}", key), value, -1)
			}
		}
		if strings.HasPrefix(s, "--xuid") {
			return string([]byte{0}) //FIXME
		}
		if strings.HasPrefix(s, "--clientId") {
			return string([]byte{0}) //FIXME
		}

		r := reflect.ValueOf(placeholders)
		t := reflect.TypeOf(placeholders)
		for i := 0; i < r.NumField(); i++ {
			s = rpl(s, t.Field(i).Tag.Get("placeholder"), r.Field(i).Interface().(string))
		}
		cp := v.GetLibraryPaths(comp.GetLibraryPath())
		cp = append(cp, extraLibs...)
		cp = append(cp, gameJar)
		return rpl(s, "classpath", strings.Join(cp, string(comp.GetSeparator())))
	}

	jvm = append(jvm, extraArgs...)
	jvm = append(jvm, "-DFabricMcEmu=net.minecraft.client.main.Main") //TODO: dynamic ADD FABRIC JVM OPT
	if placeholders.MaxRam%1024 != 0 {
		placeholders.MaxRam = placeholders.MaxRam - placeholders.MaxRam%1024 // Align to 1024
	}
	jvm = append(jvm, "-Xmx"+strconv.Itoa(placeholders.MaxRam)+"k")
	jvm = append(jvm, v.Logging.Client.Argument)

	for _, a := range v.Arguments.JVM {
		if reflect.TypeOf(a).Kind() == reflect.String {
			jvm = append(jvm, replacePlaceholders(a.(string)))
		} else if reflect.TypeOf(a).Kind() == reflect.Map {
			b, _ := json.Marshal(a.(map[string]interface{})["rules"])
			var rules []Rule
			_ = json.Unmarshal(b, &rules)
			for _, rule := range rules {
				if rule.Complies() {
					val := a.(map[string]interface{})["value"]
					if reflect.TypeOf(val).Kind() == reflect.Slice {
						for _, str := range val.([]interface{}) {
							jvm = append(jvm, replacePlaceholders(str.(string)))
						}
					} else {
						jvm = append(jvm, replacePlaceholders(val.(string)))
					}
				}
			}
		}
	}

	for _, a := range v.Arguments.Game {
		if reflect.TypeOf(a).Kind() == reflect.String {
			game = append(game, replacePlaceholders(a.(string)))
		}
	}

	game = append(game, "--width "+strconv.Itoa(placeholders.Width))
	game = append(game, "--height "+strconv.Itoa(placeholders.Height))

	return jvm, game
}

func (v *Version) GetLibraryPaths(dir string) []string {
	var ret []string
	for _, library := range v.Libraries {
		var cont bool = true
		for _, rule := range library.Rules {
			if !rule.Complies() {
				cont = false
			}
		}
		if cont {
			ret = append(ret, filepath.Join(dir, library.Downloads.Artifact.Path))
		}
	}
	return ret
}

func (r *Rule) Complies() bool {
	if runtime.GOOS == r.OS.Name {
		if r.OS.Arch != "" {
			if runtime.GOARCH == r.OS.Arch {
				return true //TODO figure out how the fuck architecture check works
			} else {
				return false
			}
		} else {
			return true
		}
	} else if runtime.GOOS == "darwin" && r.OS.Name == "osx" {
		if r.OS.Arch != "" {
			if runtime.GOARCH == r.OS.Arch {
				return true //TODO figure out how the fuck architecture check works
			} else {
				return false
			}
		} else {
			return true
		}
	} else {
		return false
	}
}

/* PRIVATE REGION */
type assetIndex struct {
	Objects map[string]Asset `json:"objects"`
}

func receiveJSONObject(address string, a any) error {
	get, err := http.Get(address)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(get.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, a)
}

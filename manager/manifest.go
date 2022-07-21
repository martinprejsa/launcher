package manager

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"reflect"
	"runtime"
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
		Size      int    `json:"size"`
		TotalSize int    `json:"totalSize"`
		Url       string `json:"url"`
	} `json:"assetIndex"`
	Assets          string `json:"assets"`
	ComplianceLevel int    `json:"compliance_level"`
	Libraries       []Library
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
type assetIndex struct {
	Objects map[string]Asset `json:"objects"`
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
func (v *Version) CreateCommandLine() (string, string) {
	var jvm string
	var game string

	for _, a := range v.Arguments.JVM {
		if reflect.TypeOf(a).Kind() == reflect.String {
			jvm += fmt.Sprintf(" %s", a.(string))
		} else if reflect.TypeOf(a).Kind() == reflect.Map {
			b, _ := json.Marshal(a.(map[string]interface{})["rules"])
			var rules []Rule
			_ = json.Unmarshal(b, &rules)
			for _, rule := range rules {
				if rule.Complies() {
					val := a.(map[string]interface{})["value"]
					if reflect.TypeOf(val).Kind() == reflect.Slice {
						for _, str := range val.([]interface{}) {
							jvm += fmt.Sprintf(" %s", str.(string))
						}
					} else {
						jvm += fmt.Sprintf(" %s", val.(string))
					}
				}
			}
		}
	}

	jvm += " -DFabricMcEmu= net.minecraft.client.main.Main"
	//TODO logging

	for _, a := range v.Arguments.Game {
		if reflect.TypeOf(a).Kind() == reflect.String {
			game += fmt.Sprintf(" %s", a.(string))
		}
	}

	// ADD FABRIC JVM OPT
	return jvm, game
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

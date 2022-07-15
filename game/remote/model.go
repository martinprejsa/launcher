package remote

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const versionManifestUrl = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
const resourceUrl = "http://resources.download.minecraft.net/%s/%s"

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
	Arguments  map[string]interface{} `json:"arguments"`
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
		Name string `json:"name"`
	} `json:"os"`
}

type Asset struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

type assetIndex struct {
	Objects map[string]Asset `json:"objects"`
}

type ResourceStatus int8

const (
	Skipped     ResourceStatus = 0
	Downloaded  ResourceStatus = 1
	NotRequired ResourceStatus = 2
	Failed      ResourceStatus = 3
)

func CodeToString(code ResourceStatus) string {
	if code == Skipped {
		return "SKIPPED"
	} else if code == Downloaded {
		return "DOWNLOADED"
	} else if code == NotRequired {
		return "NOT REQUIRED"
	} else if code == Failed {
		return "FAILED"
	} else {
		return "UNKNOWN CODE"
	}
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

func (a *Asset) Download(dir string, name string) (ResourceStatus, error) {
	if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
		if checkHash(filepath.Join(dir, name), a.Hash) {
			return Skipped, nil // Already exists, skip
		}
	}

	r, err := http.Get(fmt.Sprintf(resourceUrl, a.Hash[0:2], a.Hash))
	if err != nil {
		return Failed, err
	}
	//TODO: gzip

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return Failed, err
	}

	err = os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), os.ModePerm)
	if err != nil {
		return Failed, err
	}

	h, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return Failed, err
	}
	defer h.Close()
	//TODO check size
	_, err = h.Write(b)
	if err != nil {
		return Failed, err
	}

	return Downloaded, nil
}

func (lib *Library) Download(dir string) (ResourceStatus, error) {
	dir = filepath.Join(dir, "libraries")
	var skip = false
	for _, rule := range lib.Rules {
		if rule.OS.Name == "windows" && rule.Action == "allow" && runtime.GOOS != "windows" {
			skip = true
		}
		if rule.OS.Name == "linux" && rule.Action == "allow" && runtime.GOOS != "linux" {
			skip = true
		}
		if rule.OS.Name == "osx" && rule.Action == "allow" && runtime.GOOS != "darwin" {
			skip = true
		}
	}

	if skip {
		return NotRequired, nil // Not required on this system, skip
	}

	if _, err := os.Stat(filepath.Join(dir, lib.Downloads.Artifact.Path)); err == nil {
		if checkHash(filepath.Join(dir, lib.Downloads.Artifact.Path), lib.Downloads.Artifact.SHA1) {
			return Skipped, nil // Already exists, skip
		}
	}

	r, err := http.Get(lib.Downloads.Artifact.Url)
	if err != nil {
		return Failed, err
	}

	//TODO: gzip

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return Failed, err
	}

	err = os.MkdirAll(filepath.Dir(filepath.Join(dir, lib.Downloads.Artifact.Path)), os.ModePerm)
	if err != nil {
		return Failed, err
	}

	h, err := os.Create(filepath.Join(dir, lib.Downloads.Artifact.Path))
	if err != nil {
		return Failed, err
	}
	defer h.Close()

	_, err = h.Write(b)

	if !checkHash(filepath.Join(dir, lib.Downloads.Artifact.Path), lib.Downloads.Artifact.SHA1) {
		return Failed, errors.New("failed to verify checksum of downloaded library: " + lib.Name)
	}

	if err != nil {
		return Failed, err
	}
	return Downloaded, nil
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

func checkHash(path string, hash string) bool {
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		h := sha1.New()
		if _, err := io.Copy(h, f); err == nil {
			str := fmt.Sprintf("%x", h.Sum(nil))
			if str == hash {
				return true
			}
		}
	}
	return false
}

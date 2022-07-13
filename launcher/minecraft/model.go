package minecraft

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

const version_manifest_url = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

type Manifest struct {
	Latest   ManifestLatestVersion `json:"latest"`
	Versions []ManifestVersion     `json:"versions"`
}

type ManifestLatestVersion struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type ManifestVersion struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Url         string `json:"url"`
	Time        string `json:"time"`
	ReleaseTime string `json:"release_time"`
}

type Version struct {
	Arguments       map[string]interface{} `json:"arguments"`
	AssetIndex      AssetIndex             `json:"assetIndex"`
	Assets          string                 `json:"assets"`
	ComplianceLevel int                    `json:"compliance_level"`
}

type Library struct {
	Downloads LibraryDownload
	Name      string `json:"name"`
}

type LibraryDownload struct {
	Artifact struct {
		Path string
		SHA1 string
		Size int64
		url  string
	}
}

type AssetIndex struct {
	Objects map[string]Asset `json:"objects"`
}

type Asset struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
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

func GetAssetIndex() (AssetIndex, error) {
	var mf Manifest
	err := receiveJSONObject(version_manifest_url, &mf)
	if err != nil {
		return AssetIndex{}, err
	}

	for i := range mf.Versions {
		if mf.Versions[i].ID == mf.Latest.Release {
			var ret AssetIndex
			err := receiveJSONObject(mf.Versions[i].Url, &ret)
			if err != nil {
				return AssetIndex{}, err
			}
			return ret, nil
		}
	}
	return AssetIndex{}, errors.Errorf("latest release version \"%s\" not found in the manifest file", mf.Latest.Release)
}

func (a *Asset) Download() {

}

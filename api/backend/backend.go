package backend

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

var (
	url         = "http://localhost:8549/asset/"
	indexUrl    = url + "index/"
	previewUrlF = url + "preview/%s/%s"
)

type WardrobeIndex map[string]WardrobeEntry

type WardrobeEntry struct {
	ID string `json:"id"`
}

func GetWardrobeIndex() (WardrobeIndex, error) {
	r, err := http.Get(indexUrl)
	if err != nil {
		return WardrobeIndex{}, errors.New(err.Error())
	}
	b, err := io.ReadAll(r.Body)

	if r.StatusCode != 200 {
		return WardrobeIndex{}, errors.New("failed to contact api")
	}

	var data WardrobeIndex
	_ = json.Unmarshal(b, &data)
	return data, nil
}

func (a *WardrobeEntry) GetPreviewLink() string {
	return fmt.Sprintf(previewUrlF, a.ID[0:2], a.ID)
}

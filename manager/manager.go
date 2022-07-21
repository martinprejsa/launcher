package manager

import (
	"io/ioutil"
	"path/filepath"
)

type Profile struct {
	Name   string
	Config string
	JAR    string
}

func Explore() []Profile {
	path := "/home/martin/.genecraft/launcher"
	var profiles []Profile
	dir, _ := ioutil.ReadDir(filepath.Join(path, "versions"))

	for _, profile := range dir {
		if profile.IsDir() {
			profiles = append(
				profiles, Profile{
					Name:   profile.Name(),
					Config: filepath.Join(path, "version", profile.Name(), profile.Name()+".json"),
					JAR:    filepath.Join(path, "version", profile.Name(), profile.Name()+".jar"),
				})
		}
	}

	return profiles
}

func CreateProfile(dir string, kind string) {
	if kind == "fabric-latest" {
		//TODO this kind shit
	}

	InstallTheOnlyProfile(dir)
}

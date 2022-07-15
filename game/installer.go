package game

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"launcher/game/remote"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const fabricUrl = "https://maven.fabricmc.net/net/fabricmc/fabric-installer/0.11.0/fabric-installer-0.11.0.jar"
const mojangResourceUrl = "https://resources.download.minecraft.net/"

func Install(path string, version string) error {
	installer, err := downloadFabric()
	if err != nil {
		return errors.WithMessage(err, "failed to download fabric")
	}

	err = InstallFabric(installer, path, version)
	if err != nil {
		return errors.WithMessage(err, "failed to install fabric")
		return err
	}

	downloadMojangFiles(path)

	return nil
}

func downloadFabric() (string, error) {
	r, err := http.Get(fabricUrl)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(r.Body)

	h, err := os.CreateTemp("", "fabric-installer")
	if err != nil {
		return "", err
	}
	defer h.Close()
	_, err = h.Write(b)
	if err != nil {
		return "", err
	}
	return h.Name(), nil
}

func InstallFabric(installer string, dir string, version string) error {
	if _, err := os.Stat(dir); err != nil {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	//TODO included java bin
	cmd := exec.Command("java", "-jar", installer, "client", "-dir "+dir, "-mcversion "+version)
	err := cmd.Run()

	if err != nil {
		return err
	}

	return nil
}

func downloadMojangFiles(dir string) error {
	d := filepath.Join(dir, "assets")
	if _, err := os.Stat(d); err != nil {
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			return err
		}
	}

	mf, _ := remote.GetManifest()
	ver, _ := mf.GetLatestVersion()
	asts, _ := ver.GetAssets()
	var counter = 1
	for name, asset := range asts {
		fmt.Printf("asset %s: %d/%d\n", name, counter, len(asts))
		err := asset.Download(d, name)
		if err != nil {
			return err
		}
		counter++
	}

	for _, library := range ver.Libraries {
		err := library.Download(dir)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

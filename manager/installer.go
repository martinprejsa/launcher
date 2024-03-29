package manager

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"launcher/events"
	"launcher/logging"
	"launcher/manager/comp"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const fabricUrl = "https://maven.fabricmc.net/net/fabricmc/fabric-installer/0.11.0/fabric-installer-0.11.0.jar"

type installHandle struct {
	pbar ProgressBar
}

func InstallTheOnlyProfile(dir string) error {
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 1, Message: "Downloading fabric"})
	installer, err := downloadFabric()
	if err != nil {
		return errors.WithMessage(err, "failed to download fabric installer")
	}
	logging.Logger.Print("Downloaded fabric to " + installer)
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 5, Message: "Installing fabric"})

	err = installFabric(installer, dir, GlobalMinecraftVersion)
	logging.Logger.Print("Fabric installed to " + dir)
	//TODO download and install fabric manually
	if err != nil {
		return errors.WithMessage(err, "failed to install fabric")
	}
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 10, Message: "Fetching manifest"})
	mf, _ := GetManifest()
	ver, _ := mf.GetVersion(GlobalMinecraftVersion)
	//TODO error check
	events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: 10, Message: "Downloading logging library"})
	err = downloadLoggingLib(ver)
	if err != nil {
		return errors.WithMessage(err, "failed to download logging library")
	}

	_ = os.MkdirAll(comp.GetIndexesPath(), os.ModePerm)
	h, err := os.Create(filepath.Join(comp.GetIndexesPath(), ver.AssetIndex.ID+".json"))
	var data map[string]interface{}

	err = receiveJSONObject(ver.AssetIndex.Url, &data)
	if err != nil {
		return errors.WithMessage(err, "failed to asset index")
	}

	b, _ := json.Marshal(data)
	_, err = h.Write(b)
	if err != nil {
		return errors.WithMessage(err, "failed to write asset index")
	}
	_ = h.Close()

	_, err = downloadAssets(ver)
	if err != nil {
		return errors.WithMessage(err, "failed to download assets")
	}
	_, err = downloadLibraries(ver)
	if err != nil {
		return errors.WithMessage(err, "failed to download libraries")
	}
	return nil
}

// PRIVATE REGION //

func downloadLoggingLib(version Version) error {
	r, err := http.Get(version.Logging.Client.File.Url)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(r.Body)

	_ = os.MkdirAll(comp.GetLogCfgsPath(), os.ModePerm)

	h, err := os.Create(filepath.Join(comp.GetLogCfgsPath(), version.Logging.Client.File.ID))
	if err != nil {
		return err
	}
	defer h.Close()
	_, err = h.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func installMinecraft(file string, version Version) error {
	r, err := http.Get(version.Downloads["client"].Url)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(r.Body)

	_ = os.MkdirAll(filepath.Dir(file), os.ModePerm)

	h, err := os.OpenFile(file, os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer h.Close()
	_, err = h.Write(b)
	if err != nil {
		return err
	}

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

func checkJava() error {
	required := "17.0.0"
	cmd := exec.Command("java", "--version")
	o, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("invalid java on machine, required version: >=" + required + " current version: " + "n/a")
	}
	str := string(o)
	words := strings.Split(str, " ")
	version := words[1]
	fmt.Println(version)
	if !CompareVersion("17.0.0", version, PrecisionFull) {
		return errors.New("invalid java on machine, required version: >=" + required + " current version: " + version)
	}
	return nil
}

func installFabric(installer string, dir string, version string) error {
	if _, err := os.Stat(dir); err != nil {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	//TODO included java bin
	cmd := exec.Command("java", "-jar", installer, "client", "-dir", dir, "-mcversion", version)

	//b, _ := cmd.CombinedOutput()
	//fmt.Println(string(b))
	if err := checkJava(); err == nil {
		_ = cmd.Run()
		return nil
	} else {
		return err
	}
}

func downloadAssets(version Version) ([]string, error) {
	if _, err := os.Stat(comp.GetAssetsPath()); err != nil {
		err := os.MkdirAll(comp.GetAssetsPath(), os.ModePerm)
		if err != nil {
			return []string{}, err
		}
	}
	asts, err := version.GetAssets()
	if err != nil {
		return []string{}, err
	}
	var paths []string

	piece := float64(40) / float64(len(asts))
	var progress = 10.0

	var counter = 1

	for name, asset := range asts {
		var download = func() {
			res, err := downloadAsset(asset)
			//fmt.Printf("[%d/%d] ASSET: %s %s\n", counter, len(asts), codeToString(res), name) //TODO: log
			if res == Failed {
				logging.Logger.Error("Failed to download asset " + name + "\n\tcaused by: " + err.Error())
			}
			paths = append(paths, filepath.Join(comp.GetAssetsPath(), name))

			progress += piece
			events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: progress, Message: fmt.Sprintf("Downloading asset %d/%d", counter, len(asts))})

			counter++
		}

		download()
	}
	return paths, nil
}

func downloadLibraries(ver Version) ([]string, error) {
	var paths []string

	piece := float64(40) / float64(len(ver.Libraries))
	progress := 55.0

	for i, library := range ver.Libraries {
		res, err := downloadLibrary(library)
		if err != nil {
			//TODO: log
			return []string{}, err
		}
		//fmt.Printf("[%d/%d] LIBRARY: %s %s \n", i+1, len(ver.Libraries), codeToString(res), library.Name) //TODO: log
		if res == Failed {
			logging.Logger.Error("Failed to download library " + library.Name + "\n\tcaused by: " + err.Error())
		}
		progress += piece

		events.ProgressUpdateEvent.Trigger(events.ProgressUpdateEventPayload{Progress: progress, Message: fmt.Sprintf("Downloading library %d/%d", i+1, len(ver.Libraries))})
		paths = append(paths, filepath.Join(comp.GetLibraryPath(), library.Downloads.Artifact.Path))
	}
	return paths, nil
}

const resourceUrl = "https://resources.download.minecraft.net/%s/%s"

type resourceStatus int8

const (
	Skipped     resourceStatus = 0
	Downloaded  resourceStatus = 1
	NotRequired resourceStatus = 2
	Failed      resourceStatus = 3
)

func codeToString(code resourceStatus) string {
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

func downloadAsset(a Asset) (resourceStatus, error) {
	dir := comp.GetAssetsPath()
	if _, err := os.Stat(filepath.Join(dir, "objects", a.Hash[0:2], a.Hash)); err != os.ErrNotExist {
		if checkSHA1Hash(filepath.Join(dir, "objects", a.Hash[0:2], a.Hash), a.Hash) {
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

	err = os.MkdirAll(filepath.Join(dir, "objects", a.Hash[0:2]), os.ModePerm)
	if err != nil {
		return Failed, err
	}

	h, err := os.Create(filepath.Join(dir, "objects", a.Hash[0:2], a.Hash))
	if err != nil {
		return Failed, err
	}
	defer h.Close()
	_, err = h.Write(b)
	if err != nil {
		return Failed, err
	}

	return Downloaded, nil
}

func downloadLibrary(lib Library) (resourceStatus, error) {
	dir := comp.GetLibraryPath()
	var skip = false
	for _, rule := range lib.Rules {
		if !rule.Complies() {
			skip = true
		}
	}

	if skip {
		return NotRequired, nil // Not required on this system, skip
	}

	if _, err := os.Stat(filepath.Join(dir, lib.Downloads.Artifact.Path)); err != os.ErrNotExist {
		if checkSHA1Hash(filepath.Join(dir, lib.Downloads.Artifact.Path), lib.Downloads.Artifact.SHA1) {
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

	if !checkSHA1Hash(filepath.Join(dir, lib.Downloads.Artifact.Path), lib.Downloads.Artifact.SHA1) {
		return Failed, errors.New("failed to verify checksum of downloaded library: " + lib.Name)
	}

	if err != nil {
		return Failed, err
	}
	return Downloaded, nil
}

func checkSHA1Hash(path string, hash string) bool {
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

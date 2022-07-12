package game

import "github.com/pkg/errors"

type LaunchOptions struct {
	JavaBin      string
	Libraries    []string
	Username     string
	Version      string
	GameDir      string
	AssetsDir    string
	AssetIndex   string
	Uuid         string
	AccessToken  string
	ClientId     string
	Xuid         string
	UserType     string
	VersionType  string
	OtherOptions map[string]string
}

func Verify() bool {
	return true
}

func Start(options LaunchOptions) error {
	if !Verify() {
		return errors.New("failed to verify game files")
	} else {
		return nil
	}
}

func createStartCommand(options LaunchOptions) string {
	var command string = ""

	var append = func(thing string) {
		command += thing
		command += " "
	}

	append(options.JavaBin)
	// ..
	append("--username " + options.Username)
	append("--version " + options.Version)
	append("--gameDir " + options.GameDir)
	append("--assetIndex " + options.AssetIndex)
	append("--uuid " + options.Uuid)
	append("--accessToken " + options.AccessToken)
	append("--clientID " + options.ClientId)
	append("--xuid " + options.Xuid)
	append("--userType" + options.UserType)
	append("--versionType" + options.VersionType)

	return command
}

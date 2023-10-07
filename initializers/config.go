package initializers

import (
	"errors"
	"flag"
	"os"
	"path"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type Config struct {
	BaseURL string
	Port    int
	Key     string
	Secret  string
}

// NewConfig handles all of the command line arguments and config file values
func NewConfig() (Config, error) {
	var region = flag.String("r", "com", "Region of Veracode instance (com/eu/us)")
	var port = flag.Int("p", 8080, "Local port to run on (default is 8080)")
	var shouldSave = flag.Bool("s", false, "Should save values provided (default is FALSE)")
	flag.Parse()

	credFilePath, err := getCredentialsFilePath()
	if err != nil {
		return Config{}, err
	}

	profile, cfg, err := getProfile(credFilePath)
	if err != nil {
		return Config{}, err
	}

	if !isFlagPassed("r") && profile.HasKey("region") {
		*region = profile.Key("region").Value()
	}

	if !isFlagPassed("p") && profile.HasKey("port") {
		*port = profile.Key("port").MustInt()
		if *port == 0 {
			*port = 8080
		}
	}

	key, secret := profile.Key("veracode_api_key_id").String(), profile.Key("veracode_api_key_secret").String()
	if key == "" || secret == "" {
		err := errors.New("failed to load Veracode API credentials from file. Please refer to documentation: https://docs.veracode.com/r/c_httpie_tool")
		logrus.WithFields(logrus.Fields{"Function": "NewConfig()"}).Error(err)
		return Config{}, err
	}

	if *shouldSave {
		if profile.HasKey("region") {
			profile.Key("region").SetValue(*region)
		} else {
			profile.NewKey("region", *region)
		}

		if profile.HasKey("port") {
			profile.Key("port").SetValue(strconv.Itoa(*port))
		} else {
			profile.NewKey("port", strconv.Itoa(*port))
		}

		cfg.SaveTo(credFilePath)
	}

	logrus.WithFields(logrus.Fields{"Function": "NewConfig()"}).Infof("Region(%s) Port(%d) Profile(%s)", *region, *port, profile.Name())

	return Config{
		BaseURL: parseRegion(*region),
		Port:    *port,
		Key:     key,
		Secret:  secret,
	}, nil
}

// isFlagPassed checks whether flag has been explicitly set
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// parseRegion takes a region and returns the base URL for that region
func parseRegion(region string) string {
	switch region {
	case "eu":
		return "https://api.veracode.eu/api/authn/v2"
	case "com":
		return "https://api.veracode.com/api/authn/v2"
	case "us":
		return "https://api.veracode.us/api/authn/v2"
	default:
		return "https://api.veracode.com/api/authn/v2"
	}
}

// getCredentialsFilePath gets the Veracode API credentials file path
func getCredentialsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.WithFields(logrus.Fields{"Function": "GetProfile()"}).Error(err)
		return "", err
	}

	return path.Join(homeDir, ".veracode", "credentials"), nil
}

// GetProfile returns a pointer to the section of the credentials file that the user is using
func getProfile(filePath string) (*ini.Section, *ini.File, error) {
	profile := os.Getenv("VERACODE_API_PROFILE")

	cfg, err := ini.Load(filePath)
	if err != nil {
		logrus.WithFields(logrus.Fields{"Function": "GetProfile()"}).Error(err)
		return nil, nil, err
	}

	lenSections := len(cfg.Sections())

	if lenSections >= 2 && profile == "" {
		profile = "default"
	}

	//logrus.WithFields(logrus.Fields{"Function": "GetProfile()"}).Infof("Using profile: %s", profile)

	section, err := cfg.GetSection(profile)

	return section, cfg, err
}

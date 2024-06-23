package config

import (
	"errors"
	"os"
	"path"
	"sync"

	"github.com/DanCreative/veracode-go/veracode"
	"gopkg.in/yaml.v3"
)

var (
	errNoClient   = errors.New("veracode client is not configured")
	errNoProfiles = errors.New("no veracode credentials profiles set")

	profileMu = sync.Mutex{}
	configMu  = sync.Mutex{}
)

type Profile struct {
	Name                 string `schema:"name"`
	VeracodeApiKeyId     string `schema:"key"`
	VeracodeApiKeySecret string `schema:"secret"`
}

type Config struct {
	Region  string `yaml:"region" schema:"region"`
	Profile string `yaml:"profile" schema:"profile"`
}

func WriteConfig(config Config, appFolderPath string) error {
	configMu.Lock()
	defer configMu.Unlock()

	d, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(appFolderPath, os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(path.Join(appFolderPath, "config.yaml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(d)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfig(appFolderPath string) (Config, error) {
	configMu.Lock()
	defer configMu.Unlock()

	configFilePath := path.Join(appFolderPath, "config.yaml")

	file, err := os.Open(configFilePath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	c := Config{}
	if err := d.Decode(&c); err != nil {
		return Config{}, err
	}

	return c, nil
}

func LoadProfiles(baseFolder string) (map[string]Profile, error) {
	profileMu.Lock()
	defer profileMu.Unlock()

	profiles, err := veracode.GetProfiles(path.Join(baseFolder, "credentials"))
	if err != nil {
		return nil, err
	}

	rprofiles := make(map[string]Profile)

	for k, profile := range profiles {
		rprofiles[k] = Profile(profile)
	}

	if len(rprofiles) == 0 {
		return nil, errNoProfiles
	}

	return rprofiles, nil
}

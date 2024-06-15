package config

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"sync"
	"time"

	"github.com/DanCreative/veracode-go/veracode"
)

var (
	errProfileDoesNotExist = errors.New("profile provided in the config does not exist in the credentials file")
)

type applicationConfigService struct {
	baseFolder string
	rwmu       sync.RWMutex
	client     *veracode.Client
}

func NewApplicationConfigService(baseFolder string) *applicationConfigService {
	return &applicationConfigService{
		baseFolder: baseFolder,
	}
}

func (a *applicationConfigService) GetClient() (*veracode.Client, error) {
	if a.client == nil {
		return nil, errNoClient
	}
	a.rwmu.RLock()
	defer a.rwmu.RUnlock()
	return a.client, nil
}

func (a *applicationConfigService) GetConfig() (Config, error) {
	return LoadConfig(path.Join(a.baseFolder, "veracode_admin_plus"))
}

func (a *applicationConfigService) GetProfiles() (map[string]Profile, error) {
	return LoadProfiles(a.baseFolder)
}

func (a *applicationConfigService) SetClient() error {
	if err := os.MkdirAll(path.Join(a.baseFolder, "veracode_admin_plus"), os.ModePerm); err != nil {
		return err
	}

	config, err := a.GetConfig()
	if err != nil {
		return err
	}

	profiles, err := a.GetProfiles()
	if err != nil {
		return err
	}

	selectedProfile, ok := profiles[config.Profile]

	if !ok {
		return errProfileDoesNotExist
	}

	rateTransport, err := veracode.NewRateTransport(nil, 2/time.Second, 10)
	if err != nil {
		return err
	}

	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return err
	}

	httpClient := &http.Client{
		Transport: rateTransport,
		Jar:       jar,
	}

	client, err := veracode.NewClient(veracode.Region(config.Region), httpClient, selectedProfile.VeracodeApiKeyId, selectedProfile.VeracodeApiKeySecret)
	if err != nil {
		return err
	}

	a.rwmu.Lock()
	a.client = client
	defer a.rwmu.Unlock()

	return nil
}

func (a *applicationConfigService) CreateCredentialsFile(profile Profile) error {
	f, err := os.OpenFile(path.Join(a.baseFolder, "credentials"), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("[%s]\nveracode_api_key_id     = %s\nveracode_api_key_secret = %s", profile.Name, profile.VeracodeApiKeyId, profile.VeracodeApiKeySecret))
	if err != nil {
		return err
	}
	return nil
}

func (a *applicationConfigService) UpdateConfig(config Config) error {
	err := WriteConfig(config, path.Join(a.baseFolder, "veracode_admin_plus"))
	if err != nil {
		return err
	}
	return nil
}

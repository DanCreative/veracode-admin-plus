package veracode

import (
	"errors"
	"net/http"
	"os"
	"path"

	vmac "github.com/DanCreative/veracode-hmac-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

type AuthTransport struct {
	key       string
	secret    string
	Transport http.RoundTripper
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	bearer, err := vmac.CalculateAuthorizationHeader(req.URL, req.Method, t.key, t.secret)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")

	return t.transport().RoundTrip(req)
}

func (t *AuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

func (t *AuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func NewAuthTransport(rt http.RoundTripper) (AuthTransport, error) {
	t := AuthTransport{
		Transport: rt,
	}

	err := t.loadCredentials()
	if err != nil {
		return AuthTransport{}, err
	}

	return t, nil
}

func (t *AuthTransport) loadCredentials() error {
	profile := os.Getenv("VERACODE_API_PROFILE")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.WithFields(logrus.Fields{"Function": "loadCredentials()"}).Error(err)
		return err
	}

	cfg, err := ini.Load(path.Join(homeDir, ".veracode", "credentials"))
	if err != nil {
		logrus.WithFields(logrus.Fields{"Function": "loadCredentials()"}).Error(err)
		return err
	}

	lenSections := len(cfg.Sections())

	if lenSections >= 2 && profile == "" {
		profile = "default"
	}

	logrus.WithFields(logrus.Fields{"Function": "loadCredentials()"}).Infof("Profile(%s) out of %d profiles", profile, lenSections)

	t.key, t.secret = cfg.Section(profile).Key("veracode_api_key_id").String(), cfg.Section(profile).Key("veracode_api_key_secret").String()
	if t.key == "" || t.secret == "" {
		err := errors.New("failed to load Veracode API credentials from file. Please refer to documentation: https://docs.veracode.com/r/c_httpie_tool")
		logrus.WithFields(logrus.Fields{"Function": "loadCredentials()"}).Error(err)
		return err
	}

	logrus.WithFields(logrus.Fields{"Function": "loadCredentials()"}).Debugf("Key: %s, Secret: %s", t.key, t.secret)
	return nil
}

package veracode

import (
	"net/http"

	vmac "github.com/DanCreative/veracode-hmac-go"
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

func NewAuthTransport(rt http.RoundTripper, key, secret string) (AuthTransport, error) {
	t := AuthTransport{
		Transport: rt,
		key:       key,
		secret:    secret,
	}

	// // err := t.loadCredentials()
	// if err != nil {
	// 	return AuthTransport{}, err
	// }

	return t, nil
}

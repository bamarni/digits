package digits

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type key int

const Key key = 0

type Identity struct {
	PhoneNumber      string `json:"phone_number"`
	Id               int    `json:"id"`
	IdStr            string `json:"id_str"`
	VerificationType string `json:"verification_type"`
	CreatedAt        string `json:"created_at"`
	// TODO "access_token":{"token":"XXXXXXX-XXXXXX","secret":"XXXXXXXXXX"}
}

type errorHandler func(w http.ResponseWriter, r *http.Request, err error)

type Digits struct {
	ProviderHeader    string
	CredentialsHeader string
	Whitelist         []string
	Client            *http.Client
	ErrorHandler      errorHandler
	PhoneNumber       string
}

func Default() *Digits {
	return &Digits{
		ProviderHeader:    "X-Auth-Service-Provider",
		CredentialsHeader: "X-Verify-Credentials-Authorization",
		Whitelist:         []string{"api.digits.com", "api.twitter.com"},
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		ErrorHandler: defaultErrorHandler,
	}
}

func (dig *Digits) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	identity, err := dig.FromRequest(r)
	if err != nil {
		dig.ErrorHandler(w, r, err)
		return
	}

	ctx := context.WithValue(r.Context(), Key, identity)
	r = r.WithContext(ctx)

	next(w, r)
}

func (dig *Digits) FromRequest(r *http.Request) (*Identity, error) {
	if dig.PhoneNumber != "" {
		return &Identity{PhoneNumber: dig.PhoneNumber}, nil
	}
	provider := r.Header.Get(dig.ProviderHeader)
	u, err := url.Parse(provider)
	if err != nil {
		return nil, err
	}

	matched := false
	for _, domain := range dig.Whitelist {
		if matched, _ = regexp.MatchString(domain, u.Host); matched == true {
			break
		}
	}
	if matched == false {
		return nil, errors.New("unauthorized service provider")
	}

	return Verify(provider, r.Header.Get(dig.CredentialsHeader), dig.Client)
}

func Verify(serviceProvider, credentials string, client *http.Client) (*Identity, error) {
	req, err := http.NewRequest("GET", serviceProvider, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", credentials)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("unsuccessful response")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	identity := &Identity{}
	err = json.Unmarshal(body, identity)

	return identity, err
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

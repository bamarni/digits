package digits

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var defaultClient = &http.Client{
	Timeout: 10 * time.Second,
}

type key int

const Key key = 0

type Options struct {
	ProviderHeader    string
	CredentialsHeader string
	Client            *http.Client
	ErrorHandler      ErrorHandler
	PhoneNumber       string
	Debug             bool
	Store             Store
}

type AccessToken struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

type Identity struct {
	PhoneNumber      string      `json:"phone_number"`
	Id               int         `json:"id"`
	IdStr            string      `json:"id_str"`
	VerificationType string      `json:"verification_type"`
	CreatedAt        string      `json:"created_at"`
	AccessToken      AccessToken `json:"access_token"`
}

type Store interface {
	Load(credentials string) *Identity
	Save(credentials string, identity *Identity)
}

type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

type Digits struct {
	Logger            *log.Logger
	providerHeader    string
	credentialsHeader string
	whitelist         []string
	client            *http.Client
	errorHandler      ErrorHandler
	phoneNumber       string
	store             Store
}

func New(options Options) *Digits {
	dig := &Digits{
		whitelist: []string{"api.digits.com", "api.twitter.com"},
		store:     options.Store,
	}

	if options.ProviderHeader == "" {
		dig.providerHeader = "X-Auth-Service-Provider"
	} else {
		dig.providerHeader = options.ProviderHeader
	}

	if options.CredentialsHeader == "" {
		dig.credentialsHeader = "X-Verify-Credentials-Authorization"
	} else {
		dig.credentialsHeader = options.CredentialsHeader
	}

	if options.Client == nil {
		dig.client = defaultClient
	} else {
		dig.client = options.Client
	}

	if options.ErrorHandler == nil {
		dig.errorHandler = defaultErrorHandler
	} else {
		dig.errorHandler = options.ErrorHandler
	}

	if options.PhoneNumber != "" {
		dig.phoneNumber = options.PhoneNumber
	}

	if options.Debug {
		dig.Logger = log.New(os.Stdout, "[digits] ", log.LstdFlags)
	}

	return dig
}

func Default() *Digits {
	return New(Options{})
}

func (dig *Digits) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	identity, err := dig.FromRequest(r)
	if err != nil {
		if dig.Logger != nil {
			dig.Logger.Printf("Verification failed : %s", err)
		}
		dig.errorHandler(w, r, err)
		return
	}

	ctx := context.WithValue(r.Context(), Key, identity)
	r = r.WithContext(ctx)

	next(w, r)
}

func (dig *Digits) FromRequest(r *http.Request) (*Identity, error) {
	if dig.phoneNumber != "" {
		return &Identity{PhoneNumber: dig.phoneNumber}, nil
	}
	provider := r.Header.Get(dig.providerHeader)
	u, err := url.Parse(provider)
	if err != nil {
		return nil, err
	}

	matched := false
	for _, domain := range dig.whitelist {
		if domain == u.Host {
			matched = true
			break
		}
	}

	if matched == false {
		return nil, errors.New("unauthorized service provider")
	}

	credentials := r.Header.Get(dig.credentialsHeader)
	if dig.Logger != nil {
		dig.Logger.Printf(
			"Verifying '%s' with provider '%s'",
			credentials,
			provider,
		)
	}
	if dig.store != nil {
		if identity := dig.store.Load(credentials); identity != nil {
			if dig.Logger != nil {
				dig.Logger.Print("Identity loaded from store")
			}
			return identity, nil
		}
	}

	identity, err := Verify(provider, credentials, dig.client)
	if dig.store != nil && err == nil {
		dig.store.Save(credentials, identity)
	}
	return identity, err
}

func Verify(serviceProvider, credentials string, client *http.Client) (*Identity, error) {
	req, err := http.NewRequest("GET", serviceProvider, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", credentials)

	if nil == client {
		client = defaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("unsuccessful response")
	}

	identity := &Identity{}
	err = json.NewDecoder(resp.Body).Decode(identity)

	return identity, err
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

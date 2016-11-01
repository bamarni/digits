package digits

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
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
	Provider          string
	CredentialsHeader string
	Client            *http.Client
	ErrorHandler      errorHandler
	PhoneNumber       string
}

func Default() *Digits {
	return &Digits{
		Provider:          "https://api.digits.com/1.1/sdk/account.json",
		CredentialsHeader: "X-Verify-Credentials-Authorization",
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		ErrorHandler: defaultErrorHandler,
	}
}

func (dig *Digits) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	identity = Identity{}
	if dig.PhoneNumber == "" {
		identity, err := Verify(dig.Provider, r.Header.Get(dig.CredentialsHeader), dig.Client)
		if err != nil {
			dig.ErrorHandler(w, r, err)
			return
		}
	} else {
		identity.PhoneNumber = dig.PhoneNumber
	}

	ctx := context.WithValue(r.Context(), Key, identity)
	r = r.WithContext(ctx)

	next(w, r)
}

func Verify(serviceProvider, credentials string, client *http.Client) (identity Identity, err error) {
	req, err := http.NewRequest("GET", serviceProvider, nil)
	if err != nil {
		return identity, err
	}
	req.Header.Set("Authorization", credentials)

	resp, err := client.Do(req)
	if err != nil {
		return identity, err
	}
	if resp.StatusCode != 200 {
		return identity, errors.New("unsuccessful response")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return identity, err
	}

	err = json.Unmarshal(body, &identity)

	return identity, err
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}

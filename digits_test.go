package digits

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var httpHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprint(w, `{"phone_number":"+331234567890","access_token":{"token":"XXXXXXX-XXXXXXYZ","secret":"XXXXXXXXYZ"},"id_str":"716736250749784014","verification_type":"sms","id":716736250749784014,"created_at":"Tue Nov 01 07:34:53 +0000 2016"}`)
})

func TestSuccess(t *testing.T) {
	ts := httptest.NewServer(httpHandler)
	defer ts.Close()

	identity, err := Verify(ts.URL, "Dummy", http.DefaultClient)

	assert.NoError(t, err)

	assert.Equal(t, "+331234567890", identity.PhoneNumber)
	assert.Equal(t, "sms", identity.VerificationType)
	assert.Equal(t, "716736250749784014", identity.IdStr)
	assert.Equal(t, 716736250749784014, identity.Id)
	assert.Equal(t, "Tue Nov 01 07:34:53 +0000 2016", identity.CreatedAt) // TODO : parse as time.Time? Maybe use another field for this.
	assert.Equal(t, "XXXXXXX-XXXXXXYZ", identity.AccessToken.Token)
	assert.Equal(t, "XXXXXXXXYZ", identity.AccessToken.Secret)
}

func TestUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusUnauthorized)
	}))
	defer ts.Close()

	_, err := Verify(ts.URL, "Dummy", nil)

	assert.Error(t, err)
}

func TestFromRequestStatic(t *testing.T) {
	dig := New(Options{PhoneNumber: "+33123456789"})

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	identity, err := dig.FromRequest(req)

	assert.NoError(t, err)
	assert.Equal(t, identity.PhoneNumber, "+33123456789")
}

func TestFromRequestSuccess(t *testing.T) {
	ts := httptest.NewServer(httpHandler)
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	dig := Default()
	dig.whitelist = []string{u.Host}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(dig.providerHeader, ts.URL)

	_, err := dig.FromRequest(req)

	assert.NoError(t, err)
}

func TestFromRequestWhitelistUnauthorized(t *testing.T) {
	dig := Digits{
		whitelist: []string{"api.example2.com"},
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	req.Header.Set(dig.providerHeader, "http://api.example3.com")

	_, err := dig.FromRequest(req)

	assert.Error(t, err, "unauthorized service provider")
}

func TestFromRequestWithStore(t *testing.T) {
	ts := httptest.NewServer(httpHandler)
	defer ts.Close()

	dig := New(Options{Store: NewMemoryStore()})
	u, _ := url.Parse(ts.URL)
	dig.whitelist = []string{u.Host}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(dig.providerHeader, ts.URL)
	credentials := "Dummy"
	req.Header.Set(dig.credentialsHeader, credentials)

	identity, err := dig.FromRequest(req)

	assert.NoError(t, err)
	assert.Equal(t, identity, dig.store.Load(credentials))
}

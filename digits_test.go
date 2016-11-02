package digits

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		fmt.Fprint(w, `{"phone_number":"+331234567890","access_token":{"token":"XXXXXXX-XXXXXXYZ","secret":"XXXXXXXXYZ"},"id_str":"716736250749784014","verification_type":"sms","id":716736250749784014,"created_at":"Tue Nov 01 07:34:53 +0000 2016"}`)
	}))
	defer ts.Close()

	identity, err := Verify(ts.URL, "Dummy", http.DefaultClient)

	assert.NoError(t, err)

	assert.Equal(t, "+331234567890", identity.PhoneNumber)
	assert.Equal(t, "sms", identity.VerificationType)
	assert.Equal(t, "716736250749784014", identity.IdStr)
	assert.Equal(t, 716736250749784014, identity.Id)
	assert.Equal(t, "Tue Nov 01 07:34:53 +0000 2016", identity.CreatedAt) // TODO : parse as time.Time? Maybe use another field for this.
}

func TestUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusUnauthorized)
	}))
	defer ts.Close()

	_, err := Verify(ts.URL, "Dummy", http.DefaultClient)

	assert.Error(t, err)
}

func TestFromRequestStatic(t *testing.T) {
	dig := &Digits{PhoneNumber: "+33123456789"}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	identity, err := dig.FromRequest(req)

	assert.NoError(t, err)
	assert.Equal(t, identity.PhoneNumber, "+33123456789")
}

func TestFromRequestSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		fmt.Fprint(w, `{"phone_number":"+331234567890","access_token":{"token":"XXXXXXX-XXXXXXYZ","secret":"XXXXXXXXYZ"},"id_str":"716736250749784014","verification_type":"sms","id":716736250749784014,"created_at":"Tue Nov 01 07:34:53 +0000 2016"}`)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	dig := &Digits{
		Provider:  ts.URL,
		Whitelist: []string{u.Host},
		Client:    http.DefaultClient,
	}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	_, err := dig.FromRequest(req)

	assert.NoError(t, err)
}

func TestFromRequestWhitelistUnauthorized(t *testing.T) {
	dig := &Digits{
		Whitelist: []string{"*.example2.com"},
	}

	req := httptest.NewRequest("GET", "http://example.com/", nil)
	_, err := dig.FromRequest(req)

	assert.Error(t, err)
}

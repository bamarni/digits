# digits [![Build Status](https://travis-ci.org/bamarni/digits.svg?branch=master)](https://travis-ci.org/bamarni/digits)

Library for [Twitter Digits](https://get.digits.com/), it acts as a **delegator** (for more information : https://dev.twitter.com/oauth/echo).

This package requires Go +1.7

## Usage

### Standalone

``` go
package main

import (
	"net/http"

	"github.com/bamarni/digits"
)

func main() {
	identity, err := digits.Verify(
		"https://api.digits.com/1.1/sdk/account.json",
		"OAuth oauth_consumer_key=[...] oauth_*=[...]",
		nil,
	)
	// println(identity.PhoneNumber)
	// ...
}
```

### Negroni middleware

The default handler throws a HTTP 401 response in case of authentication failure.
Otherwise, user identity can be retrieved through the request context :

``` go
package main

import (
	"net/http"

	"github.com/bamarni/digits"
	"github.com/urfave/negroni"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		identity := r.Context().Value(digits.Key).(*digits.Identity)
		fmt.Fprintf(w, "Can I reach you at %s?", identity.PhoneNumber)
	})

	n := negroni.Classic()
	n.Use(digits.Default())
	n.UseHandler(mux)

	http.ListenAndServe(":3000", n)
}
```

In dev environment, you can skip authentication to Digits API and use a static
phone number instead :

``` go
digitsMiddleware := digits.New(digits.Options{
	PhoneNumber: "+33123456789",
})
n := negroni.New(digitsMiddleware)
// ...
```

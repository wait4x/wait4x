// Copyright 2019-2025 The Wait4X Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package http provides the HTTP checker for the Wait4X application.
package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"wait4x.dev/v3/checker"
)

// TestMain is the main function for the HTTP checker.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// TestHttpInvalidAddress tests the HTTP checker with an invalid address.
func TestHttpInvalidAddress(t *testing.T) {
	hc := New("http://not-exists.tld", WithTimeout(time.Second))
	assert.Error(t, hc.Check(context.TODO()))
}

// TestHttpValidAddress tests the HTTP checker with a valid address.
func TestHttpValidAddress(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	hc := New(ts.URL)
	identity, err := hc.Identity()

	assert.Nil(t, err)
	assert.Nil(t, hc.Check(context.TODO()))
	assert.Equal(t, ts.URL, identity)
}

// TestHttpInvalidStatusCode tests the HTTP checker with an invalid status code.
func TestHttpInvalidStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectStatusCode(http.StatusCreated))

	var expectedError *checker.ExpectedError
	assert.ErrorAs(t, hc.Check(context.TODO()), &expectedError)
}

// TestHttpValidStatusCode tests the HTTP checker with a valid status code.
func TestHttpValidStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectStatusCode(http.StatusOK))

	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpInvalidTLS tests the HTTP checker with an invalid TLS certificate.
func TestHttpInvalidTLS(t *testing.T) {
	hc := New("https://expired.badssl.com", WithInsecureSkipTLSVerify(true))
	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpNoRedirect tests the HTTP checker with no redirect.
func TestHttpNoRedirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "https://wait4x.dev")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))

	defer ts.Close()
	hc := New(ts.URL, WithExpectStatusCode(http.StatusTemporaryRedirect), WithNoRedirect(true))

	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpRedirect tests the HTTP checker with a redirect.
func TestHttpRedirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "https://wait4x.dev")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))

	defer ts.Close()
	hc := New(ts.URL, WithExpectStatusCode(http.StatusOK))

	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpInvalidBody tests the HTTP checker with an invalid body.
func TestHttpInvalidBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Wait4X"))
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectBodyRegex("FooBar"))

	var expectedError *checker.ExpectedError
	assert.ErrorAs(t, hc.Check(context.TODO()), &expectedError)
}

// TestHttpValidBody tests the HTTP checker with a valid body.
func TestHttpValidBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Wait4X is the best CLI tools. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla volutpat arcu malesuada lacus vulputate feugiat. Etiam vitae sem quis ligula consequat euismod. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus fringilla sapien non lacus volutpat sollicitudin. Donec sollicitudin sit amet purus ac rutrum. Nam nunc orci, luctus a sagittis."))
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectBodyRegex("Wait4X.+best.+tools"))

	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpValidBodyJSON tests the HTTP checker with a valid body JSON.
func TestHttpValidBodyJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user": {"name": "test"}, "is_active": true}`))
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectBodyJSON("user"))
	assert.Nil(t, hc.Check(context.TODO()))

	hc = New(ts.URL, WithExpectBodyJSON("user.name"))
	assert.Nil(t, hc.Check(context.TODO()))

	hc = New(ts.URL, WithExpectBodyJSON("is_active"))
	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpInvalidBodyJSON tests the HTTP checker with an invalid body JSON.
func TestHttpInvalidBodyJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user": {"name": "test"}, "is_active": true}`))
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectBodyJSON("test"))

	var expectedError *checker.ExpectedError
	assert.ErrorAs(t, hc.Check(context.TODO()), &expectedError)
}

// TestHttpInvalidBodyXPath tests the HTTP checker with an invalid body XPath.
func TestHttpInvalidBodyXPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<div><code id='ip'>127.0.0.1</code></div>"))
	}))
	defer ts.Close()

	var expectedError *checker.ExpectedError

	hc := New(ts.URL, WithExpectBodyXPath("//hello"))
	assert.ErrorAs(t, hc.Check(context.TODO()), &expectedError)

	hc = New(ts.URL, WithExpectBodyXPath("//code[@id='test']"))
	assert.ErrorAs(t, hc.Check(context.TODO()), &expectedError)
}

// TestHttpValidBodyXPath tests the HTTP checker with a valid body XPath.
func TestHttpValidBodyXPath(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<div><code id='ip'>127.0.0.1</code></div>"))
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectBodyXPath("//div/code"))
	assert.Nil(t, hc.Check(context.TODO()))

	hc = New(ts.URL, WithExpectBodyXPath("//code[@id='ip']"))
	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpValidHeader tests the HTTP checker with a valid header.
func TestHttpValidHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Test-Header", "test-value")
		w.Header().Add("Test-Header-New", "test-value-new")
		w.Header().Add("Authorization", "Token 1234")
		w.Header().Add("X-Foo", "")
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectHeader("Test-Header"))
	assert.Nil(t, hc.Check(context.TODO()))

	hc = New(ts.URL, WithExpectHeader("X-Foo"))
	assert.Nil(t, hc.Check(context.TODO()))

	hc = New(ts.URL, WithExpectHeader("X-Foo=.*"))
	assert.Nil(t, hc.Check(context.TODO()))

	// Regex.
	hc = New(ts.URL, WithExpectHeader("Test-Header=test-.+"))
	assert.Nil(t, hc.Check(context.TODO()))

	hc = New(ts.URL, WithExpectHeader("Authorization=^Token\\s.+"))
	assert.Nil(t, hc.Check(context.TODO()))

	// Key value.
	hc = New(ts.URL, WithExpectHeader("Test-Header=test-value"))
	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpInvalidHeader tests the HTTP checker with an invalid header.
func TestHttpInvalidHeader(t *testing.T) {
	var expectedError *checker.ExpectedError

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Test-Header", "test-value")
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectHeader("Test-Header-New"))
	assert.ErrorAs(t, hc.Check(context.TODO()), &expectedError)

	hc = New(ts.URL, WithExpectHeader("Test-.+=test-value"))
	assert.ErrorAs(t, hc.Check(context.TODO()), &expectedError)

	hc = New(ts.URL, WithExpectHeader("Test-Header=[A-Z]"))
	assert.ErrorAs(t, hc.Check(context.TODO()), &expectedError)
}

// TestHttpRequestHeaders tests the HTTP checker with request headers.
func TestHttpRequestHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := new(bytes.Buffer)
		for key, value := range r.Header {
			fmt.Fprintf(resp, "%s=%s,", key, value)
		}
		w.Write(resp.Bytes())
	}))
	defer ts.Close()

	hc := New(
		ts.URL,
		WithRequestHeaders(http.Header{"Authorization": []string{"Token 123"}}),
		WithRequestHeader("Foo", []string{"test1 test2"}),
		WithExpectBodyRegex("(.*Authorization=\\[Token 123\\].*Foo=\\[test1 test2\\].*)|(.*Foo=\\[test1 test2\\].*Authorization=\\[Token 123\\].*)"),
	)
	assert.Nil(t, hc.Check(context.TODO()))
}

// TestHttpInvalidCombinationFeatures tests the HTTP checker with invalid combination features.
func TestHttpInvalidCombinationFeatures(t *testing.T) {
	var expectedError *checker.ExpectedError

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Test-Header", "test-value")
		w.Write([]byte("Wait4X"))
	}))
	defer ts.Close()

	hc := New(ts.URL, WithExpectStatusCode(http.StatusCreated), WithExpectBodyRegex("FooBar"))
	err := hc.Check(context.TODO())
	assert.ErrorAs(t, err, &expectedError)
	assert.Equal(t, "the body doesn't expect", err.Error())

	hc = New(ts.URL, WithExpectStatusCode(http.StatusCreated), WithExpectBodyRegex("Wait4X"), WithExpectHeader("X-Foo"))
	err = hc.Check(context.TODO())
	assert.ErrorAs(t, err, &expectedError)
	assert.Equal(t, "the http header key doesn't expect", err.Error())

	hc = New(ts.URL, WithExpectStatusCode(http.StatusOK), WithExpectBodyRegex("Wait4X"), WithExpectHeader("Test-Header"))
	err = hc.Check(context.TODO())
	assert.ErrorAs(t, err, &expectedError)
	assert.Equal(t, "the status code doesn't expect", err.Error())
}

// TestHttpRequestBody tests the HTTP checker with a request body.
func TestHttpRequestBody(t *testing.T) {
	var expectedError *checker.ExpectedError

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		w.Write(buf.Bytes())
	}))
	defer ts.Close()

	hc := New(
		ts.URL,
		WithRequestBody(strings.NewReader("name=test&score=1")), WithExpectBodyRegex("something"),
	)
	err := hc.Check(context.TODO())
	assert.ErrorAs(t, err, &expectedError)

	hc = New(
		ts.URL,
		WithRequestBody(strings.NewReader("name=test&score=1")), WithExpectBodyRegex("name=test&score=1"),
	)
	err = hc.Check(context.TODO())
	assert.Nil(t, err)
}

// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package http_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"gfx.cafe/open/jrpc"
	"gfx.cafe/open/jrpc/contrib/jmux"
	"gfx.cafe/open/jrpc/pkg/codec"
	"gfx.cafe/open/jrpc/pkg/server"

	"gfx.cafe/open/jrpc/pkg/jrpctest"

	. "gfx.cafe/open/jrpc/contrib/codecs/http"
)

const contentType = "application/json"
const maxRequestContentLength = 1024 * 1024 * 5
const respLength = maxRequestContentLength * 3

func confirmStatusCode(t *testing.T, got, want int) {
	t.Helper()
	if got == want {
		return
	}
	if gotName := http.StatusText(got); len(gotName) > 0 {
		if wantName := http.StatusText(want); len(wantName) > 0 {
			t.Fatalf("response status code: got %d (%s), want %d (%s)", got, gotName, want, wantName)
		}
	}
	t.Fatalf("response status code: got %d, want %d", got, want)
}

func confirmRequestValidationCode(t *testing.T, method, contentType, body string, expectedStatusCode int) {
	t.Helper()
	request := httptest.NewRequest(method, "http://url.com", strings.NewReader(body))
	if len(contentType) > 0 {
		request.Header.Set("Content-Type", contentType)
	}
	code, err := ValidateRequest(request)
	if code == 0 {
		if err != nil {
			t.Errorf("validation: got error %v, expected nil", err)
		}
	} else if err == nil {
		t.Errorf("validation: code %d: got nil, expected error", code)
	}
	confirmStatusCode(t, code, expectedStatusCode)
}

func TestHTTPErrorResponseWithDelete(t *testing.T) {
	confirmRequestValidationCode(t, http.MethodDelete, contentType, "", http.StatusMethodNotAllowed)
}

func TestHTTPErrorResponseWithPut(t *testing.T) {
	confirmRequestValidationCode(t, http.MethodPut, contentType, "", http.StatusMethodNotAllowed)
}

func TestHTTPErrorResponseWithMaxContentLength(t *testing.T) {
	body := make([]rune, maxRequestContentLength+1)
	confirmRequestValidationCode(t,
		http.MethodPost, contentType, string(body), http.StatusRequestEntityTooLarge)
}

//NOTE: this test is not needed since we no longer check this
//
//func TestHTTPErrorResponseWithEmptyContentType(t *testing.T) {
//	confirmRequestValidationCode(t, http.MethodPost, "", "", http.StatusUnsupportedMediaType)
//}

func TestHTTPErrorResponseWithValidRequest(t *testing.T) {
	confirmRequestValidationCode(t, http.MethodPost, contentType, "", 0)
}

func confirmHTTPRequestYieldsStatusCode(t *testing.T, method, contentType, body string, expectedStatusCode int) {
	t.Helper()
	s := server.NewServer(jmux.NewMux())
	ts := httptest.NewServer(&Server{Server: s})
	defer ts.Close()

	request, err := http.NewRequest(method, ts.URL, strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create a valid HTTP request: %v", err)
	}
	if len(contentType) > 0 {
		request.Header.Set("Content-Type", contentType)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	confirmStatusCode(t, resp.StatusCode, expectedStatusCode)
}

func TestHTTPResponseWithEmptyGet(t *testing.T) {
	confirmHTTPRequestYieldsStatusCode(t, http.MethodGet, "", "", http.StatusOK)
}

// This checks that maxRequestContentLength is not applied to the response of a request.
func TestHTTPRespBodyUnlimited(t *testing.T) {
	s := jrpctest.NewServer()
	ts := httptest.NewServer(&Server{Server: s})
	defer ts.Close()

	c, err := DialHTTP(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	var r string
	if err := c.Do(nil, &r, "large_largeResp", nil); err != nil {
		t.Fatal(err)
	}
	if len(r) != respLength {
		t.Fatalf("response has wrong length %d, want %d", len(r), respLength)
	}
}

// Tests that an HTTP error results in an HTTPError instance
// being returned with the expected attributes.
func TestHTTPErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error has occurred!", http.StatusTeapot)
	}))
	defer ts.Close()

	c, err := DialHTTP(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	var r string
	err = c.Do(nil, &r, "test_method", nil)
	if err == nil {
		t.Fatal("error was expected")
	}

	httpErr, ok := err.(*codec.HTTPError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}

	if httpErr.StatusCode != http.StatusTeapot {
		t.Error("unexpected status code", httpErr.StatusCode)
	}
	if httpErr.Status != "418 I'm a teapot" {
		t.Error("unexpected status text", httpErr.Status)
	}
	if body := string(httpErr.Body); body != "error has occurred!\n" {
		t.Error("unexpected body", body)
	}

	if errMsg := httpErr.Error(); errMsg != "418 I'm a teapot: error has occurred!\n" {
		t.Error("unexpected error message", errMsg)
	}
}

func TestHTTPPeerInfo(t *testing.T) {
	s := jrpctest.NewServer()
	ts := httptest.NewServer(&Server{Server: s})
	defer ts.Close()

	c, err := DialHTTP(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	c.SetHeader("user-agent", "ua-testing")
	c.SetHeader("x-forwarded-for", "origin.example.com")

	// Request peer information.
	var info codec.PeerInfo
	if err := c.Do(nil, &info, "test_peerInfo", nil); err != nil {
		t.Fatal(err)
	}

	if info.RemoteAddr == "" {
		t.Error("RemoteAddr not set")
	}
	if info.Transport != "http" {
		t.Errorf("wrong Transport %q", info.Transport)
	}
	if info.HTTP.Version != "HTTP/1.1" {
		t.Errorf("wrong HTTP.Version %q", info.HTTP.Version)
	}
	if info.HTTP.UserAgent != "ua-testing" {
		t.Errorf("wrong HTTP.UserAgent %q", info.HTTP.UserAgent)
	}
	if info.HTTP.Origin != "origin.example.com" {
		t.Errorf("wrong HTTP.Origin %q", info.HTTP.UserAgent)
	}
}
func TestClientHTTP(t *testing.T) {
	s := jrpctest.NewServer()
	ts := httptest.NewServer(&Server{Server: s})
	defer ts.Close()
	c, err := DialHTTP(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// Launch concurrent requests.
	var (
		results    = make([]jrpctest.EchoResult, 100)
		errc       = make(chan error, len(results))
		wantResult = jrpctest.EchoResult{String: "a", Int: 0, Args: new(jrpctest.EchoArgs)}
	)
	defer ts.Close()
	for i := range results {
		i := i
		go func() {
			errc <- jrpc.CallInto(nil, c, &results[i], "test_echo", wantResult.String, wantResult.Int, wantResult.Args)
		}()
	}

	// Wait for all of them to complete.
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
	for i := range results {
		select {
		case err := <-errc:
			if err != nil {
				t.Fatal(err)
			}
		case <-timeout.C:
			t.Fatalf("timeout (got %d/%d) results)", i+1, len(results))
		}
	}

	// Check results.
	for i := range results {
		if !reflect.DeepEqual(results[i], wantResult) {
			t.Errorf("result %d mismatch: got %#v, want %#v", i, results[i], wantResult)
		}
	}
}

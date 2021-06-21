// Copyright 2021 Andrei Costescu <andrei@costescu.no>
// Copyright 2020 Oz Tiram <oz.tiram@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package netbox

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestQuery(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(
		`{"count":1, "results":[{"address": "10.0.0.2/25", "dns_name": "my_host"}]}`)

	want := "10.0.0.2"
	ctx := context.Background()
	u, _ := url.Parse("https://example.org")
	n := Netbox{
		URL:           u,
		Token:         "mytoken",
		CacheDuration: time.Millisecond * 100,
	}
	got := n.query(ctx, "my_host")
	if got != want {
		t.Fatalf("Expected %s but got %s", want, got)
	}

}

func TestNoSuchHost(t *testing.T) {

	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "NoSuchHost"}).Reply(
		200).BodyString(`{"count":0,"next":null,"previous":null,"results":[]}`)

	want := ""
	ctx := context.Background()
	u, _ := url.Parse("https://example.org")
	n := Netbox{
		URL:           u,
		Token:         "mytoken",
		CacheDuration: time.Millisecond * 100,
	}
	got := n.query(ctx, "NoSuchHost")
	if got != want {
		t.Fatalf("Expected empty string but got %s", got)
	}

}

func TestLocalCache(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(
		`{"count":1, "results":[{"address": "10.0.0.2/25", "dns_name": "my_host"}]}`)

	ipAddress := ""

	ctx := context.Background()
	u, _ := url.Parse("https://example.org")
	n := Netbox{
		URL:           u,
		Token:         "mytoken",
		CacheDuration: time.Millisecond * 100,
	}
	got := n.query(ctx, "my_host")

	item, err := localCache.Get("my_host")
	if err == nil {
		ipAddress = item.Value().(string)
	}

	assert.Equal(t, got, ipAddress, "local cache item didn't match")

}

func TestLocalCacheExpiration(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	gock.New("https://example.org/api/ipam/ip-addresses/").MatchParams(
		map[string]string{"dns_name": "my_host"}).Reply(
		200).BodyString(
		`{"count":1, "results":[{"address": "10.0.0.2/25", "dns_name": "my_host"}]}`)

	ctx := context.Background()
	u, _ := url.Parse("https://example.org")
	n := Netbox{
		URL:           u,
		Token:         "mytoken",
		CacheDuration: time.Millisecond * 100,
	}
	n.query(ctx, "my_host")
	<-time.After(101 * time.Millisecond)
	item, err := localCache.Get("my_host")
	if err != nil {
		t.Fatalf("Expected errors, but got: %v", item)
	}
}

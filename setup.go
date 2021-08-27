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
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/coredns/caddy"
)

// init registers this plugin.
func init() { plugin.Register("netbox", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {

	netboxPlugin, err := newNetBox(c)
	if err != nil {
		return plugin.Error("netbox", err)
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		netboxPlugin.Next = next
		return netboxPlugin
	})

	// All OK, return a nil error.
	return nil
}

func contains(lst []string, v string) bool {
	for _, e := range lst {
		if e == v {
			return true
		}
	}
	return false
}

func newNetBox(c *caddy.Controller) (Netbox, error) {

	nbURL := ""
	token := ""
	localCacheDuration := ""
	duration := time.Second
	var err error
	nb := Netbox{}
	allowedPriorities := []string{"dns_name", "virtual_machine", "device"}

	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "url":
				if !c.NextArg() {
					return nb, c.ArgErr()
				}
				nbURL = c.Val()
			case "token":
				if !c.NextArg() {
					return nb, c.ArgErr()
				}
				token = c.Val()
			case "localCacheDuration":
				if !c.NextArg() {
					return nb, c.ArgErr()
				}
				localCacheDuration = c.Val()
				duration, err = time.ParseDuration(localCacheDuration)
				if err != nil {
					localCacheDuration = ""
				}
			case "priority":
				nb.Priority = c.RemainingArgs()
				if len(nb.Priority) == 0 {
					nb.Priority = allowedPriorities
				} else {
					for _, v := range nb.Priority {
						if !contains(allowedPriorities, v) {
							return nb, c.Errf("unknown priority: %s", v)
						}
					}
				}
			case "stop_when_found":
				if !c.NextArg() {
					return nb, c.ArgErr()
				}
				nb.StopFound, err = strconv.ParseBool(c.Val())
				if err != nil {
					return nb, err
				}
			default:
				return nb, c.Errf("unknown property: %q", c.Val())
			}
		}
	}

	if nbURL == "" || token == "" || localCacheDuration == "" {
		return nb, errors.New("could not parse netbox config")
	}
	u, err := url.Parse(nbURL)
	if err != nil {
		return nb, err
	}
	nb.URL = u
	nb.Token = token
	nb.CacheDuration = duration
	return nb, nil

}

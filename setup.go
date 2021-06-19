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
		return *netboxPlugin
	})

	// All OK, return a nil error.
	return nil
}

func newNetBox(c *caddy.Controller) (*Netbox, error) {

	url := ""
	token := ""

	for c.Next() {
		if c.NextBlock() {
			for {
				switch c.Val() {
				case "url":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					url = c.Val()

				case "token":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					token = c.Val()
				}

				if !c.Next() {
					break
				}
			}
		}

	}

	if url == "" || token == "" {
		return nil, errors.New("could not parse netbox config")
	}

	return &Netbox{Url: url, Token: token}, nil

}

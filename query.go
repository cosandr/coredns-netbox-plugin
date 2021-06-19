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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Records struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Records  []IP    `json:"results"`
}

type IP struct {
	ID        int       `json:"id"`
	Address   string    `json:"address"`
	Interface Interface `json:"interface"`
	Family    Family    `json:"family"`
}

type Family struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

type Interface struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	Device         *Device         `json:"device,omitempty"`
	VirtualMachine *VirtualMachine `json:"virtual_machine,omitempty"`
}

type VirtualMachine struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Device struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

var localCache map[string]string
var client = &http.Client{Timeout: 5 * time.Second}

func doRequest(ctx context.Context, url, token string) (*Records, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("limit", "300")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

	log.Debugf("GET %s", req.URL.String())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	records := Records{}
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response %v", err)
	}
	return &records, nil
}

func updateCache(ctx context.Context, url, token string) error {
	records, err := doRequest(ctx, fmt.Sprintf("%s/api/ipam/ip-addresses/", url), token)
	if err != nil {
		return err
	}
	// Reset cache
	localCache = make(map[string]string)
	// Get all results
	for {
		for _, ip := range records.Records {
			// Only IPv4
			if ip.Family.Value != 4 {
				continue
			}
			addr := strings.Split(ip.Address, "/")[0]
			if ip.Interface.VirtualMachine != nil {
				localCache[ip.Interface.VirtualMachine.Name] = addr
			} else if ip.Interface.Device != nil {
				dev := ip.Interface.Device
				// Add both name and display if they're different
				if dev.Name != "" {
					localCache[dev.Name] = addr
				}
				if dev.DisplayName != "" && dev.Name != dev.DisplayName {
					localCache[dev.DisplayName] = addr
				}
			} else {
				log.Debugf("IP %d [%s] has no device or virtual machine associated with its interface", ip.ID, addr)
			}
		}
		if records.Next == nil {
			break
		}
		// Get next page
		records, err = doRequest(ctx, *records.Next, token)
		if err != nil {
			log.Errorf("cannot get next page: %v", err)
			break
		}
	}
	return nil
}

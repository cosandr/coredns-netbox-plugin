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
	"net"
	"net/http"
	"path"
	"time"

	"github.com/imkira/go-ttlmap"
)

type IPAddress struct {
	ID                 int    `json:"id"`
	Display            string `json:"display"`
	Address            string `json:"address"`
	AssignedObjectType string `json:"assigned_object_type"`
	AssignedObjectId   int    `json:"assigned_object_id"`
	DnsName            string `json:"dns_name"`
}

type IPAddressResults struct {
	Count   int         `json:"count"`
	Results []IPAddress `json:"results"`
}

var localCache = ttlmap.New(nil)
var client = &http.Client{Timeout: 5 * time.Second}

func (n Netbox) runRequest(ctx context.Context, endpoint string, params map[string]string, out interface{}) error {
	var resp *http.Response
	u := *n.URL
	u.Path = path.Join(u.Path, "api/", endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", u.String()+"/", nil)
	if err != nil {
		log.Errorf("cannot create request %v", err)
		return err
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", n.Token))

	log.Debugf("GET %s", req.URL.String())
	resp, err = client.Do(req)
	if err != nil {
		log.Errorf("HTTP Error %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("invalid response code %d", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		log.Errorf("could not unmarshal response %v", err)
		return err
	}
	return nil
}

func (n Netbox) getAddresses(ctx context.Context, name string) ([]string, error) {
	// search dns_name, device, virtual_machine
	ret := make([]string, 0)
	for _, prio := range n.Priority {
		params := map[string]string{
			prio: name,
		}
		results := IPAddressResults{}
		err := n.runRequest(ctx, "ipam/ip-addresses/", params, &results)
		if err != nil {
			return ret, err
		}
		if results.Count == 0 {
			log.Debugf("%s %s not found", prio, name)
			continue
		}
		log.Debugf("found %d %s(s)", results.Count, prio)
		for _, r := range results.Results {
			ip, _, err := net.ParseCIDR(r.Address)
			if err != nil {
				log.Warning(err)
				continue
			}
			ret = append(ret, ip.String())
			log.Debugf("added %s for %s %s", ip.String(), prio, name)
		}
		if n.StopFound {
			break
		}
	}
	return ret, nil
}

func (n Netbox) query(ctx context.Context, dnsName string) []string {
	item, err := localCache.Get(dnsName)
	if err == nil {
		log.Debugf("found in local cache %s", dnsName)
		return item.Value().([]string)
	}
	addr, err := n.getAddresses(ctx, dnsName)
	if err != nil || len(addr) == 0 {
		return []string{}
	}
	err = localCache.Set(dnsName, ttlmap.NewItem(addr, ttlmap.WithTTL(n.CacheDuration)), nil)
	if err != nil {
		log.Warning(err)
	}
	return addr
}

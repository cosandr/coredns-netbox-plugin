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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/imkira/go-ttlmap"
)

type IP struct {
	Address string `json:"address"`
}

type Record struct {
	PrimaryIP4 IP     `json:"primary_ip4"`
	Name       string `json:"name,omitempty"`
}

type RecordsList struct {
	Records []Record `json:"results"`
}

var localCache = ttlmap.New(nil)
var client = &http.Client{Timeout: 5 * time.Second}

func query(url, token, dns_name string, duration time.Duration) string {
	item, err := localCache.Get(dns_name)
	if err == nil {
		clog.Debugf("found in local cache %s", dns_name)
		return item.Value().(string)
	}
	records := RecordsList{}
	var resp *http.Response
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/dcim/devices/", url), nil)
	if err != nil {
		clog.Errorf("cannot create request %v", err)
		return ""
	}
	q := req.URL.Query()
	q.Add("name", dns_name)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

	clog.Infof("GET %s", req.URL.String())
	resp, err = client.Do(req)
	if err != nil {
		clog.Errorf("HTTP Error %v", err)
		return ""
	}

	if resp.StatusCode != http.StatusOK {
		clog.Errorf("invalid response code %d", resp.StatusCode)
		return ""
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		clog.Errorf("could not unmarshal response %v", err)
		return ""
	}

	if len(records.Records) == 0 {
		clog.Info("record not found")
		return ""
	}

	ip_address := strings.Split(records.Records[0].PrimaryIP4.Address, "/")[0]
	clog.Infof("record found %s", ip_address)
	localCache.Set(dns_name, ttlmap.NewItem(ip_address, ttlmap.WithTTL(duration)), nil)
	return ip_address
}

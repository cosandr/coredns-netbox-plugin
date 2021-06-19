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
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("netbox")

type Netbox struct {
	Url   string
	Token string
	Next  plugin.Handler
}

func (n Netbox) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	search := strings.TrimRight(state.QName(), ".")

	log.Debugf("searching for '%s', %d names in cache", search, len(localCache))
	ipAddress, ok := localCache[search]
	// Not in cache, update it
	if !ok {
		log.Debugf("'%s' not in cache", search)
		err := updateCache(ctx, n.Url, n.Token)
		if err != nil {
			log.Error(err)
			return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
		}
		ipAddress, ok = localCache[search]
		if !ok {
			log.Debugf("did not find %s", search)
			return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
		}
	} else {
		log.Debugf("'%s' in cache", search)
	}
	log.Debugf("found %s: %s", search, ipAddress)
	// no IP is found in netbox pass processing to the next plugin
	if len(ipAddress) == 0 {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}

	// Export metric with the server label set to the current
	// server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	rec := new(dns.A)
	rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600}
	rec.A = net.ParseIP(ipAddress)
	m := new(dns.Msg)
	m.Answer = []dns.RR{rec}
	m.SetReply(r)
	err := w.WriteMsg(m)

	if err != nil {
		log.Error(err)
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}

	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (n Netbox) Name() string { return "netbox" }

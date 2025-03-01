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
	"net/url"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("netbox")

type Netbox struct {
	URL           *url.URL
	Token         string
	CacheDuration time.Duration
	Priority      []string
	Next          plugin.Handler
	StopFound     bool
}

func (n Netbox) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	state := request.Request{W: w, Req: r}
	search := strings.TrimRight(state.QName(), ".")
	// Remove domain
	s := strings.Split(search, ".")
	// Only change if we have a domain
	if len(s) > 1 {
		search = strings.Join(s[:len(s)-1], ".")
	}

	ips := n.query(ctx, search)
	// no IPs found in netbox pass processing to the next plugin
	if len(ips) == 0 {
		return plugin.NextOrFailure(n.Name(), n.Next, ctx, w, r)
	}

	// Export metric with the server label set to the current
	// server handling the request.
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	var records []dns.RR
	for _, ipAddress := range ips {
		rec := new(dns.A)
		rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600}
		rec.A = net.ParseIP(ipAddress)
		records = append(records, rec)
	}

	m := new(dns.Msg)
	m.Answer = records
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

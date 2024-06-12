package dnssrv

import (
	"errors"
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/dns/model"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

func dnsHandler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	// 是否设置为权威服务器
	// m.Authoritative = true
	// m.RecursionDesired = r.RecursionDesired

	for _, q := range r.Question {
		global.Log.Infof("Dns query record: %s lookup [%s %s %s]", w.RemoteAddr(), q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass])

		queryZone, queryHost, err := parseDomain(q.Name)
		if err != nil {
			global.Log.Errorf("Failed to fetch dns zone: %v", err)
			return
		}
		// global.Log.Debugf("q.Name: %s; qZone: %s; qHost: %s", q.Name, queryZone, queryHost)

		var dnsZones model.DnsZones
		if err := dnsZones.FindAll(); err != nil {
			global.Log.Errorf("Failed to fetch dns zone: %v", err)
			return
		}

		if dz := findMatchingDnsZone(queryZone, dnsZones); dz != nil {
			handleLocalQuery(q, m, dz, queryHost)
		} else {
			handleForwardQuery(r, m)
		}
	}

	// global.Log.Debugf("Dns Message: \n\n%+v\n ", m)
	err := w.WriteMsg(m)
	if err != nil {
		global.Log.Errorf("Failed to write DNS response: %v", err)
	}
}

func handleLocalQuery(q dns.Question, m *dns.Msg, dz *model.DnsZone, queryHost string) {
	dnsRecords, err := queryDomain(dz.ID, queryHost)
	if err != nil {
		global.Log.Errorf("Query domain error: %v", err)
		m.Rcode = dns.RcodeServerFailure
		return
	}
	if len(dnsRecords) == 0 {
		m.Rcode = dns.RcodeNameError
		return
	}

	for _, dr := range dnsRecords {
		switch q.Qtype {
		case dns.TypeA:
			if dr.Type == "CNAME" {
				processDnsRecord(m, dz, dr, dns.TypeCNAME)
				recursiveARecord(m, dr.Value)
			} else if dr.Type == "A" {
				ip := net.ParseIP(dr.Value)
				if ip == nil {
					global.Log.Errorf("Failed to parse IP address: %s", dr.Value)
					continue
				}
				processDnsRecord(m, dz, dr, dns.TypeA)
			}
		case dns.TypeCNAME:
			if dr.Type == "CNAME" {
				processDnsRecord(m, dz, dr, dns.TypeCNAME)
			}
		case dns.TypeTXT:
			if dr.Type == "TXT" {
				processDnsRecord(m, dz, dr, dns.TypeTXT)
			}
		}
	}
}

func handleForwardQuery(r *dns.Msg, m *dns.Msg) {
	forwardClient := &dns.Client{
		Net:     "udp",
		Timeout: 5 * time.Second,
	}
	resp, _, err := forwardClient.Exchange(r, "1.1.1.1:53")
	if err != nil {
		global.Log.Errorf("Failed to forward DNS query %v", err)
		return
	}
	m.Answer = append(m.Answer, resp.Answer...)
	m.Ns = append(m.Ns, resp.Ns...)
	m.Extra = append(m.Extra, resp.Extra...)
}

func queryDomain(zone_id uint, host string) (dnsRecords model.DnsRecords, err error) {
	filter := map[string]interface{}{
		"zone_id": zone_id,
		"host":    host,
	}
	if err = dnsRecords.Find(filter); err != nil {
		global.Log.Errorf("Failed to fetch DNS records: %v", err)
		return
	}
	return
}

// ensureFQDN 确保域名以点结尾
func ensureFQDN(domain string) string {
	if domain[len(domain)-1] != '.' {
		return domain + "."
	}
	return domain
}
func parseDomain(domain string) (zone string, host string, err error) {
	// 验证输入是否为空
	if domain == "" {
		return "", "", errors.New("domain cannot be empty")
	}

	domain = ensureFQDN(domain)

	domainParts := strings.Split(domain, ".")
	// 验证域名是是否为二级域名, eg: example.com
	if len(domainParts) < 3 {
		return "", "", errors.New("domain must contain at least two parts")
	}

	// 二级域名处理
	if len(domainParts) == 3 {
		return domain, "@", nil
	}

	// 一般情况处理，构造zone和host
	zone = strings.ToLower(domainParts[len(domainParts)-3] + "." + domainParts[len(domainParts)-2] + ".")
	host = strings.ToLower(domain[:len(domain)-len(zone)-1])

	// 对于非法的zone或host，应进行进一步的检查和错误处理
	// 这里假设zone和host的格式总是合法的，根据实际情况可能需要调整
	return zone, host, nil
}

func findMatchingDnsZone(queryZone string, dnsZones model.DnsZones) *model.DnsZone {
	for _, dz := range dnsZones {
		dz.Name = ensureFQDN(dz.Name)
		// global.Log.Debugf("queryZone: %s, intercepted dnsZone: %v", queryZone, dz.Name)
		if queryZone == dz.Name {
			return dz
		}
	}
	return nil
}

// 处理DNS记录的函数，减少代码重复
func processDnsRecord(m *dns.Msg, dz *model.DnsZone, dr *model.DnsRecord, recordType uint16) {
	var dn string
	if dr.Host == "@" {
		dn = dz.Name
	} else {
		dn = fmt.Sprintf("%s.%s", dr.Host, dz.Name)
	}
	buildDnsMessage(m, dn, dns.TypeToString[recordType], dr.Value, dr.Ttl)
}

func buildDnsMessage(m *dns.Msg, dname string, dtype string, drecord string, dttl uint32) {
	switch dtype {
	case "A":
		ip := net.ParseIP(drecord)
		if ip != nil {
			aRR := &dns.A{
				Hdr: dns.RR_Header{
					Name:   ensureFQDN(dname),
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    dttl,
				},
				A: ip,
			}
			m.Answer = append(m.Answer, aRR)
		}
	case "CNAME":
		cnameRR := &dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   ensureFQDN(dname),
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    dttl,
			},
			Target: ensureFQDN(drecord),
		}
		m.Answer = append(m.Answer, cnameRR)

	case "TXT":
		txtRR := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   ensureFQDN(dname),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    dttl,
			},
			Txt: []string{drecord},
		}
		m.Answer = append(m.Answer, txtRR)
	default:
		global.Log.Warnf("Unsupported query type: %s", dtype)
	}
}

// recursiveARecord 递归查询 A 记录
func recursiveARecord(m *dns.Msg, domain string) {
	queryZone, queryHost, err := parseDomain(domain)
	if err != nil {
		global.Log.Errorf("Failed to fetch dns zone: %v", err)
		return
	}
	// global.Log.Debugf("recurisiveDomain: %s; qZone: %s; qHost: %s", domain, queryZone, queryHost)

	var dnsZones model.DnsZones
	if err := dnsZones.FindAll(); err != nil {
		global.Log.Errorf("Failed to fetch dns zone: %v", err)
		return
	}

	var dnsRecords model.DnsRecords
	if dz := findMatchingDnsZone(queryZone, dnsZones); dz != nil {
		dnsRecords, err = queryDomain(dz.ID, queryHost)
		if err != nil {
			global.Log.Errorf("Query domain error: %v", err)
			m.Rcode = dns.RcodeServerFailure
			return
		}
		if len(dnsRecords) == 0 {
			m.Rcode = dns.RcodeNameError
			return
		}
		for _, dr := range dnsRecords {
			if dr.Type == "CNAME" {
				buildDnsMessage(m, domain, dns.TypeToString[dns.TypeCNAME], dr.Value, dr.Ttl)
				// 递归查询 A 记录
				recursiveARecord(m, dr.Value)
			} else if dr.Type == "A" {
				buildDnsMessage(m, domain, dns.TypeToString[dns.TypeA], dr.Value, dr.Ttl)
			}
		}
	} else {
		r := new(dns.Msg)
		r.SetQuestion(ensureFQDN(domain), dns.TypeA)
		handleForwardQuery(r, m)
	}
}

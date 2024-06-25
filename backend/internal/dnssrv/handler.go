package dnssrv

import (
	"errors"
	"fmt"
	"micro-net-hub/internal/config"
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

		dnsZones := getLocalZones()

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
	dnsRecords, err := getLocalZoneRecords(dz.ID, queryHost)
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
			dn := genDomain(dz, dr)
			handleATypeDnsRecord(m, dr, dn, 1)
		case dns.TypeCNAME:
			handleDnsRecord(m, dz, dr, dns.TypeCNAME)
		case dns.TypeTXT:
			handleDnsRecord(m, dz, dr, dns.TypeTXT)
		}
	}
}

func handleForwardQuery(r *dns.Msg, m *dns.Msg) {
	forwardClient := &dns.Client{
		Net:     "udp",
		Timeout: 5 * time.Second,
	}
	resp, _, err := forwardClient.Exchange(r, config.Conf.Dns.ForwardAddr)
	if err != nil {
		global.Log.Errorf("Failed to forward DNS query %v", err)
		return
	}
	m.Answer = append(m.Answer, resp.Answer...)
	m.Ns = append(m.Ns, resp.Ns...)
	m.Extra = append(m.Extra, resp.Extra...)
}

func getLocalZones() (dnsZones model.DnsZones) {
	if dnsZones = model.CacheDnsZonesGet(); dnsZones != nil {
		return
	}

	if err := dnsZones.FindAll(); err != nil {
		global.Log.Errorf("Failed to fetch dns zone: %v", err)
		return
	}

	model.CacheDnsZonesSet(dnsZones)
	return
}

func getLocalZoneRecords(zone_id uint, host string) (dnsRecords model.DnsRecords, err error) {
	if dnsRecords = model.CacheDnsRecordGet(zone_id, host); dnsRecords != nil {
		return
	}

	filter := map[string]interface{}{
		"zone_id": zone_id,
		"host":    host,
	}
	if err = dnsRecords.Find(filter); err != nil {
		global.Log.Errorf("Failed to fetch DNS records: %v", err)
		return
	}

	model.CacheDnsRecordSet(zone_id, host, dnsRecords)
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

func genDomain(dz *model.DnsZone, dr *model.DnsRecord) (domain string) {
	if dr.Host == "@" {
		domain = dz.Name
	} else {
		domain = fmt.Sprintf("%s.%s", dr.Host, dz.Name)
	}
	return
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

// 递归查询 A 记录
func handleATypeDnsRecord(m *dns.Msg, dr *model.DnsRecord, domain string, recursionDepth int) {
	// global.Log.Debug("handleATypeDnsRecord dr: %+v", dr)
	// 递归查询次数限制, 防止 DNS 放大攻击
	if recursionDepth >= config.Conf.Dns.MaxRecursionDepth {
		global.Log.Warnf("Exceeded maximum recursion depth for domain: %s", domain)
		m.Rcode = dns.RcodeServerFailure
		return
	}

	if dr.Type == "CNAME" {
		buildDnsMessage(m, domain, dns.TypeToString[dns.TypeCNAME], dr.Value, dr.Ttl)

		queryZone, queryHost, err := parseDomain(dr.Value)
		if err != nil {
			global.Log.Errorf("Failed to fetch dns zone: %v", err)
			return
		}
		// global.Log.Debugf("recurisiveDomain: %s; qZone: %s; qHost: %s", dr.Value, queryZone, queryHost)

		dnsZones := getLocalZones()
		var dnsRecords model.DnsRecords
		if dz := findMatchingDnsZone(queryZone, dnsZones); dz != nil {
			dnsRecords, err = getLocalZoneRecords(dz.ID, queryHost)
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
				handleATypeDnsRecord(m, dr, genDomain(dz, dr), recursionDepth+1)
			}
		} else {
			r := new(dns.Msg)
			r.SetQuestion(ensureFQDN(dr.Value), dns.TypeA)
			handleForwardQuery(r, m)
		}
	} else if dr.Type == "A" {
		ip := net.ParseIP(dr.Value)
		if ip != nil {
			buildDnsMessage(m, domain, dns.TypeToString[dns.TypeA], dr.Value, dr.Ttl)
		} else {
			global.Log.Errorf("Failed to parse IP address: %s", dr.Value)
		}
	}
}

func handleDnsRecord(m *dns.Msg, dz *model.DnsZone, dr *model.DnsRecord, recordType uint16) {
	if dr.Type == dns.TypeToString[recordType] {
		dn := genDomain(dz, dr)
		buildDnsMessage(m, dn, dns.TypeToString[recordType], dr.Value, dr.Ttl)
	}
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

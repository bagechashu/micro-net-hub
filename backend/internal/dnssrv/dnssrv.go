package dnssrv

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type Question struct {
	qname  string
	qtype  string
	qclass string
}

func (q *Question) String() string {
	return q.qname + " " + q.qclass + " " + q.qtype
}
func UnFqdn(s string) string {
	if dns.IsFqdn(s) {
		return s[:len(s)-1]
	}
	return s
}

// func main() {
// 	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
// 		var resp dns.Msg
// 		resp.SetReply(req)
// 		for _, q := range req.Question {
// 			a := dns.A{
// 				Hdr: dns.RR_Header{
// 					Name:   q.Name,
// 					Rrtype: dns.TypeA,
// 					Class:  dns.ClassINET,
// 					Ttl:    0,
// 				},
// 				A: net.ParseIP("127.0.0.1").To4(),
// 			}
// 			resp.Answer = append(resp.Answer, &a)
// 		}
// 		w.WriteMsg(&resp)
// 	})
// 	log.Fatal(dns.ListenAndServe(":53", "udp", nil))
// }

func handler(w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]
	// global.Log.Infof("req.Question: %+v", q)

	queryInfo := Question{UnFqdn(q.Name), dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}
	global.Log.Infof("Dns query record: %s lookup [%s]", w.RemoteAddr(), queryInfo.String())

	// Query hosts
	m := new(dns.Msg)
	m.SetReply(req)
	m.RecursionDesired = req.RecursionDesired

	// Check for local domains and handle them
	queryDomain := strings.ToLower(q.Name)
	if queryDomain == "baidu.com." || queryDomain == "google.com." {
		switch q.Qtype {
		case dns.TypeA:
			// Handling A records
			ipRecords := []net.IP{net.ParseIP("127.0.0.1").To4(), net.ParseIP("127.0.0.2").To4()}
			for _, ip := range ipRecords {
				aRR := &dns.A{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 10}, A: ip}
				m.Answer = append(m.Answer, aRR)
			}
		case dns.TypeCNAME:
			// Handling CNAME records
			cnameRecords := []string{"cname.example.com"}
			for _, cname := range cnameRecords {
				// important: domain must be fully qualified, so add trailing dot if not present.
				if !strings.HasSuffix(cname, ".") {
					cname += "."
				}
				cnameRR := &dns.CNAME{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 10}, Target: cname}
				m.Answer = append(m.Answer, cnameRR)
			}
		case dns.TypeTXT:
			// Handling TXT records
			txtRecords := []string{"example txt record"}
			for _, txt := range txtRecords {
				txtRR := &dns.TXT{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 10}, Txt: []string{txt}}
				m.Extra = append(m.Extra, txtRR)
			}
		default:
			// Unhandled query types
			global.Log.Warnf("Unsupported query type: %s", dns.TypeToString[q.Qtype])
		}
	} else {
		// Forward the query to 1.1.1.1 or 8.8.8.8
		forwardIP := net.ParseIP("1.1.1.1") // You can choose between 1.1.1.1 and 8.8.8.8
		forwardClient := &dns.Client{Net: "udp"}
		resp, _, err := forwardClient.Exchange(req, forwardIP.String()+":53")
		if err != nil {
			global.Log.Errorf("Failed to forward DNS query: %v", err)
			return
		}
		m = resp
	}

	err := w.WriteMsg(m)
	if err != nil {
		global.Log.Errorf("Failed to write DNS response: %v", err)
	}
}

func NewDnsServerUdp() *dns.Server {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", handler)

	server := &dns.Server{
		Addr:         config.Conf.Dns.ListenAddr,
		Net:          "udp",
		Handler:      mux,
		UDPSize:      65535,
		ReadTimeout:  time.Duration(config.Conf.Dns.ReadTimeoutSecond) * time.Second,
		WriteTimeout: time.Duration(config.Conf.Dns.WriteTimeoutSecond) * time.Second,
	}

	global.Log.Infof("New dns server on UDP: %s", config.Conf.Dns.ListenAddr)
	return server
}

func RunUdp() (err error) {
	udpServer := NewDnsServerUdp()
	return udpServer.ListenAndServe()
}

func NewDnsServerTcp() *dns.Server {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", handler)

	server := &dns.Server{
		Addr:         config.Conf.Dns.ListenAddr,
		Net:          "tcp",
		Handler:      mux,
		ReadTimeout:  time.Duration(config.Conf.Dns.ReadTimeoutSecond) * time.Second,
		WriteTimeout: time.Duration(config.Conf.Dns.WriteTimeoutSecond) * time.Second,
	}

	global.Log.Infof("New dns server on TCP: %s", config.Conf.Dns.ListenAddr)
	return server
}

func RunTcp() (err error) {
	tcpServer := NewDnsServerTcp()
	return tcpServer.ListenAndServe()
}

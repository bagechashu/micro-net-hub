package dnssrv

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"time"

	"github.com/miekg/dns"
)

func NewDnsServerUdp() *dns.Server {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", dnsHandler)

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

func RunUdp() error {
	udpServer := NewDnsServerUdp()
	return udpServer.ListenAndServe()
}

func NewDnsServerTcp() *dns.Server {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", dnsHandler)

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

func RunTcp() error {
	tcpServer := NewDnsServerTcp()
	return tcpServer.ListenAndServe()
}

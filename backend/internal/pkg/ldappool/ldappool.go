package ldappool

import (
	ldap "github.com/go-ldap/ldap/v3"

	"math/rand"
	"net"
	"sync"
	"time"
)

type LdapPool struct {
	conns     []*ldap.Conn
	reqConns  map[uint64]chan *ldap.Conn
	openConn  int
	maxOpen   int
	url       string
	adminDN   string
	adminPass string
}

var muLock sync.Mutex

func NewLdapPool(maxConn int, url, adminDN, adminPass string) *LdapPool {
	return &LdapPool{
		conns:     make([]*ldap.Conn, 0),
		reqConns:  make(map[uint64]chan *ldap.Conn),
		openConn:  0,
		maxOpen:   maxConn,
		url:       url,
		adminDN:   adminDN,
		adminPass: adminPass,
	}
}

func (lp *LdapPool) GetConn() (*ldap.Conn, error) {
	muLock.Lock()
	// 判断当前连接池内是否存在连接
	connNum := len(lp.conns)
	if connNum > 0 {
		lp.openConn++
		conn := lp.conns[0]
		copy(lp.conns, lp.conns[1:])
		lp.conns = lp.conns[:connNum-1]

		muLock.Unlock()
		// 发现连接已经 close 重新获取连接
		if conn.IsClosing() {
			return lp.createConn()
		}
		return conn, nil
	}

	// 当现有连接池为空时，并且当前超过最大连接限制
	if lp.maxOpen != 0 && lp.openConn > lp.maxOpen {
		// 创建一个等待队列
		req := make(chan *ldap.Conn, 1)
		reqKey := lp.nextRequestKeyLocked()
		lp.reqConns[reqKey] = req
		muLock.Unlock()

		// 等待请求归还
		return <-req, nil
	}

	lp.openConn++
	muLock.Unlock()
	return lp.createConn()
}

// 放回了一个 LDAP 连接
func (lp *LdapPool) PutConn(conn *ldap.Conn) {
	muLock.Lock()
	defer muLock.Unlock()

	// 先判断是否存在等待的队列
	if num := len(lp.reqConns); num > 0 {
		var req chan *ldap.Conn
		var reqKey uint64
		for reqKey, req = range lp.reqConns {
			break
		}
		delete(lp.reqConns, reqKey)
		req <- conn
		return
	}

	lp.openConn--
	if !conn.IsClosing() {
		lp.conns = append(lp.conns, conn)
	}
}

// TODO: TLS url conn
// https://cybernetist.com/2020/05/18/getting-started-with-go-ldap/
// ldap.DialURL(ldapURL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
func (lp *LdapPool) createConn() (*ldap.Conn, error) {
	// log.Printf("ldap dsn: %s@tcp(%s)", lp.adminDN, lp.url)
	conn, err := ldap.DialURL(lp.url, ldap.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}))
	if err != nil {
		return nil, err
	}
	err = conn.Bind(lp.adminDN, lp.adminPass)
	if err != nil {
		return nil, err
	}
	lp.PutConn(conn)

	return conn, err
}

// 获取下一个请求令牌
func (lp *LdapPool) nextRequestKeyLocked() uint64 {
	for {
		reqKey := rand.Uint64()
		if _, ok := lp.reqConns[reqKey]; !ok {
			return reqKey
		}
	}
}

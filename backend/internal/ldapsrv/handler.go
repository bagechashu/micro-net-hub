package ldapsrv

import (
	"fmt"
	"log"
	"micro-net-hub/internal/global"
	"sync"
	"time"

	"github.com/merlinz01/ldapserver"
)

type LdapSrvHandler struct {
	ldapserver.BaseHandler
	abandonment     map[ldapserver.MessageID]bool
	abandonmentLock sync.Mutex
}

var theOnlyAuthorizedUser = ldapserver.MustParseDN("cn=admin,dc=example,dc=com")

func (ls *LdapSrvHandler) checkPassword(dn ldapserver.DN, password string) bool {
	if dn.Equal(theOnlyAuthorizedUser) {
		return password == "123456"
	}
	return false
}

func getAuth(conn *ldapserver.Conn) ldapserver.DN {
	var auth ldapserver.DN
	if conn.Authentication != nil {
		if authdn, ok := conn.Authentication.(ldapserver.DN); ok {
			auth = authdn
		}
	}
	// To make sure we can do a nil check
	if len(auth) == 0 {
		auth = nil
	}
	global.Log.Debug("Currently authenticated as:", auth)
	return auth
}

func (ls *LdapSrvHandler) Abandon(conn *ldapserver.Conn, msg *ldapserver.Message, messageID ldapserver.MessageID) {
	global.Log.Debug("Abandon request")
	ls.abandonmentLock.Lock()
	if _, exists := ls.abandonment[messageID]; exists {
		ls.abandonment[messageID] = true
	}
	ls.abandonmentLock.Unlock()
}
func (ls *LdapSrvHandler) Bind(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.BindRequest) {
	global.Log.Debug("Bind request")
	res := &ldapserver.BindResult{}
	// if !conn.IsTLS() {
	// 	global.Log.Debug("Rejecting Bind on non-TLS connection")
	// 	res.ResultCode = ldapserver.ResultConfidentialityRequired
	// 	res.DiagnosticMessage = "TLS is required for the Bind operation"
	// 	conn.SendResult(msg.MessageID, nil, ldapserver.TypeBindResponseOp, res)
	// 	return
	// }
	dn, err := ldapserver.ParseDN(req.Name)
	if err != nil {
		global.Log.Debug("Error parsing DN:", err)
		res.ResultCode = ldapserver.ResultInvalidDNSyntax
		res.DiagnosticMessage = "the provided DN is invalid"
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeBindResponseOp, res)
		return
	}
	start := time.Now()
	switch req.AuthType {
	case ldapserver.AuthenticationTypeSimple:
		log.Printf("Simple authentication from %s for \"%s\"\n", conn.RemoteAddr(), req.Name)
		if ls.checkPassword(dn, req.Credentials.(string)) {
			log.Printf("Successful Bind from %s for \"%s\"\n", conn.RemoteAddr(), dn)
			conn.Authentication = dn
			res.ResultCode = ldapserver.ResultSuccess
		} else {
			log.Printf("Invalid credentials from %s for \"%s\"\n", conn.RemoteAddr(), dn)
			conn.Authentication = nil
			res.ResultCode = ldapserver.ResultInvalidCredentials
		}
	case ldapserver.AuthenticationTypeSASL:
		creds := req.Credentials.(*ldapserver.SASLCredentials)
		log.Printf("SASL authentication from %s for \"%s\" using mechanism %s", conn.RemoteAddr(), dn, creds.Mechanism)
		switch creds.Mechanism {
		case "CRAM-MD5":
			// Put verification code in here
			log.Printf("CRAM-MD5 authentication from %s for \"%s\"\n", conn.RemoteAddr(), dn)
			conn.Authentication = nil
			res.ResultCode = ldapserver.ResultAuthMethodNotSupported
			res.DiagnosticMessage = "the CRAM-MD5 authentication method is not supported"
		default:
			log.Printf("Unsupported SASL mechanism from %s for \"%s\"\n", conn.RemoteAddr(), dn)
			conn.Authentication = nil
			res.ResultCode = ldapserver.ResultAuthMethodNotSupported
			res.DiagnosticMessage = "the SASL authentication method requested is not supported"
		}
	default:
		log.Printf("Unsupported authentication method from %s for \"%s\"\n", conn.RemoteAddr(), dn)
		res.ResultCode = ldapserver.ResultAuthMethodNotSupported
		res.DiagnosticMessage = "the authentication method requested is not supported by this server"
	}
	// Make sure the response takes at least a second in order to prevent timing attacks
	sofar := time.Since(start)
	if sofar < time.Second {
		time.Sleep(time.Second - sofar)
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeBindResponseOp, res)
}

func (ls *LdapSrvHandler) Compare(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.CompareRequest) {
	global.Log.Debug("Compare request")
	// Allow cancellation
	ls.abandonment[msg.MessageID] = false
	defer func() {
		ls.abandonmentLock.Lock()
		delete(ls.abandonment, msg.MessageID)
		ls.abandonmentLock.Unlock()
	}()
	auth := getAuth(conn)
	if !auth.Equal(theOnlyAuthorizedUser) {
		global.Log.Debug("Not an authorized connection!", auth)
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeCompareResponseOp,
			ldapserver.ResultInsufficientAccessRights.AsResult(
				"the connection is not authorized to perform the requested operation"))
		return
	}
	// Pretend to take a while
	time.Sleep(time.Second * 2)
	global.Log.Debug("Compare DN:", req.Object)
	global.Log.Debug("  Attribute:", req.Attribute)
	global.Log.Debug("  Value:", req.Value)
	if ls.abandonment[msg.MessageID] {
		global.Log.Debug("Abandoning compare request")
		return
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeCompareResponseOp,
		ldapserver.ResultCompareTrue.AsResult(""))
}

func (ls *LdapSrvHandler) Search(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.SearchRequest) {
	global.Log.Debug("Search request")
	// Allow cancellation
	ls.abandonment[msg.MessageID] = false
	defer func() {
		ls.abandonmentLock.Lock()
		delete(ls.abandonment, msg.MessageID)
		ls.abandonmentLock.Unlock()
	}()

	auth := getAuth(conn)
	if !auth.Equal(theOnlyAuthorizedUser) {
		global.Log.Debug("Not an authorized connection!", auth)
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeModifyResponseOp,
			ldapserver.ResultInsufficientAccessRights.AsResult(
				"the connection is not authorized to perform the requested operation"))
		return
	}
	global.Log.Debug("Base object:", req.BaseObject)
	switch req.Scope {
	case ldapserver.SearchScopeBaseObject:
		global.Log.Debug("Scope: base object")
	case ldapserver.SearchScopeSingleLevel:
		global.Log.Debug("Scope: single level")
	case ldapserver.SearchScopeWholeSubtree:
		global.Log.Debug("Scope: whole subtree")
	}
	switch req.DerefAliases {
	case ldapserver.AliasDerefNever:
		global.Log.Debug("Never deref aliases")
	case ldapserver.AliasDerefFindingBaseObj:
		global.Log.Debug("Deref aliases finding base object")
	case ldapserver.AliasDerefInSearching:
		global.Log.Debug("Deref aliases in searching")
	case ldapserver.AliasDerefAlways:
		global.Log.Debug("Always deref aliases")
	}
	global.Log.Debug("Size limit:", req.SizeLimit)
	global.Log.Debug("Time limit:", req.TimeLimit)
	global.Log.Debug("Types only:", req.TypesOnly)
	global.Log.Debug("Filter:", req.Filter)
	global.Log.Debug("Attributes:", req.Attributes)

	// Return some entries
	for i := 0; i < 5; i++ {
		if ls.abandonment[msg.MessageID] {
			global.Log.Debug("Abandoning search request after", i, "requests")
			return
		}
		// Pretend to take a while
		// time.Sleep(time.Second * 1)
		entry := &ldapserver.SearchResultEntry{
			ObjectName: fmt.Sprintf("uid=jdoe%d,%s", i, req.BaseObject),
			Attributes: []ldapserver.Attribute{
				{Description: "uid", Values: []string{fmt.Sprintf("jdoe%d", i)}},
				{Description: "givenname", Values: []string{fmt.Sprintf("John %d", i)}},
				{Description: "sn", Values: []string{"Doe"}},
			},
		}
		global.Log.Debug("Sending entry", i)
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultEntryOp, entry)
	}

	conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp,
		ldapserver.ResultSuccess.AsResult(""))
}

// below method is not implemented
func (ls *LdapSrvHandler) Add(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.AddRequest) {
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeAddResponseOp, res)
}

func (ls *LdapSrvHandler) Delete(conn *ldapserver.Conn, msg *ldapserver.Message, dn string) {
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeDeleteResponseOp, res)
}

func (ls *LdapSrvHandler) Modify(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.ModifyRequest) {
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeModifyResponseOp, res)
}

func (ls *LdapSrvHandler) ModifyDN(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.ModifyDNRequest) {
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeModifyDNResponseOp, res)
}

func (ls *LdapSrvHandler) Extended(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.ExtendedRequest) {
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeExtendedResponseOp, res)
}

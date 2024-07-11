package ldapsrv

import (
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"sync"
	"time"

	"micro-net-hub/internal/module/account/model"
	totpModel "micro-net-hub/internal/module/totp/model"

	"github.com/merlinz01/ldapserver"
	"gorm.io/gorm"
)

type LdapSrvHandler struct {
	ldapserver.BaseHandler
	abandonment     map[ldapserver.MessageID]bool
	abandonmentLock sync.Mutex
}

// checkPassword:
// 普通用户使用 "密码+totp" 校验
// 用于三方软件 bind 的账户, 仅使用 "密码" 校验.
func (ls *LdapSrvHandler) checkPassword(dn ldapserver.DN, password string) bool {
	u, err := getUserInfo(dn)
	if err != nil {
		global.Log.Errorf("ldap DN [%v] get user info error: %v", dn, err)
		return false
	}

	pwd := tools.NewParsePasswd(u.Password)
	// totp 验证开关
	if !config.Conf.LdapServer.TotpEnable {
		return pwd == password
	}
	for _, r := range u.Roles {
		if r.Keyword == config.Conf.LdapServer.BindDNRoleKeyword {
			return pwd == password
		}
	}
	pwd = pwd + totpModel.GetGoogleTotp(u.Totp.Secret)
	return pwd == password
}

func checkBindRuleUser(auth ldapserver.DN) bool {
	u, err := getUserInfo(auth)
	if err != nil {
		return false
	}
	for _, r := range u.Roles {
		if r.Keyword == config.Conf.LdapServer.BindDNRoleKeyword {
			return true
		}
	}
	return false
}

func getUserInfo(dn ldapserver.DN) (*model.User, error) {
	uc := model.CacheUserDNGet(dn.String())
	if uc != nil {
		return uc, nil
	}

	var u model.User
	err := u.Find(map[string]interface{}{"user_dn": dn.String()})
	if err != nil {
		return nil, err
	}
	global.Log.Debugf("UserInfo get from cache: %+v", u)
	model.CacheUserDNSet(dn.String(), &u)
	return &u, nil
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
	// global.Log.Debug("ldapserver auth dn:", auth)
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
	res := &ldapserver.BindResult{}
	dn, err := ldapserver.ParseDN(req.Name)
	if err != nil {
		global.Log.Errorf("Error parsing DN: %s", err)
		res.ResultCode = ldapserver.ResultInvalidDNSyntax
		res.DiagnosticMessage = "the provided DN is invalid"
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeBindResponseOp, res)
		return
	}
	start := time.Now()
	switch req.AuthType {
	case ldapserver.AuthenticationTypeSimple:
		if ls.checkPassword(dn, req.Credentials.(string)) {
			conn.Authentication = dn
			res.ResultCode = ldapserver.ResultSuccess
		} else {
			global.Log.Errorf("Invalid credentials from %s for \"%s\"", conn.RemoteAddr(), dn)
			conn.Authentication = nil
			res.ResultCode = ldapserver.ResultInvalidCredentials
		}
	case ldapserver.AuthenticationTypeSASL:
		global.Log.Errorf("Unsupported SASL mechanism from %s for \"%s\"", conn.RemoteAddr(), dn)
		conn.Authentication = nil
		res.ResultCode = ldapserver.ResultAuthMethodNotSupported
		res.DiagnosticMessage = "the SASL authentication method requested is not supported"
	default:
		global.Log.Errorf("Unsupported authentication method from %s for \"%s\"", conn.RemoteAddr(), dn)
		res.ResultCode = ldapserver.ResultAuthMethodNotSupported
		res.DiagnosticMessage = "the authentication method requested is not supported by this server"
	}
	// Make sure the response takes at least a second in order to prevent timing attacks
	sofar := time.Since(start)
	if sofar < 500*time.Millisecond {
		time.Sleep(500*time.Millisecond - sofar)
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeBindResponseOp, res)
}

func (ls *LdapSrvHandler) Search(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.SearchRequest) {
	ls.abandonmentLock.Lock()
	ls.abandonment[msg.MessageID] = false
	defer func() {
		delete(ls.abandonment, msg.MessageID)
		ls.abandonmentLock.Unlock()
	}()

	auth := getAuth(conn)
	if !checkBindRuleUser(auth) {
		global.Log.Debug("Not an authorized connection!", auth)
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeModifyResponseOp,
			ldapserver.ResultInsufficientAccessRights.AsResult(
				"the connection is not authorized to perform the requested operation"))
		return
	}

	if ls.abandonment[msg.MessageID] {
		return
	}

	query, err := parseLdapQuery(req.Filter.String())
	if err != nil {
		global.Log.Errorf("parseLdapQuery error: %s", err)
	}

	if query.MemberOf != "" {
		find, err := model.UserExistsInGroup(query.Uid, query.MemberOf)
		if err != nil {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultOperationsError.AsResult(err.Error()))
			return
		}
		if !find {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
	}
	udn := ldapserver.MustParseDN(fmt.Sprintf("uid=%s,ou=%s,%s", query.Uid, query.ObjectClass, req.BaseObject))
	u, err := getUserInfo(udn)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
		global.Log.Errorf("ldap DN [%v] get user info error: %v", udn, err)
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultOperationsError.AsResult(""))
		return
	}
	// global.Log.Debugf("ldapserver search result user: %+v", u)
	entry := &ldapserver.SearchResultEntry{
		ObjectName: u.UserDN,
		Attributes: []ldapserver.Attribute{
			{Description: "objectClass", Values: []string{"inetOrgPerson"}},
			{Description: "uid", Values: []string{query.Uid}},
			{Description: "cn", Values: []string{query.Uid}},
			{Description: "sn", Values: []string{query.Uid}},
			{Description: "displayName", Values: []string{u.Nickname}},
			{Description: "givenName", Values: []string{u.GivenName}},
			{Description: "mail", Values: []string{u.Mail}},
			{Description: "mobile", Values: []string{u.Mobile}},
		},
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultEntryOp, entry)
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultSuccess.AsResult(""))
}

func (ls *LdapSrvHandler) Compare(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.CompareRequest) {
	global.Log.Debug("Compare request")
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeCompareResponseOp, res)
}

// below method is not implemented
func (ls *LdapSrvHandler) Add(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.AddRequest) {
	global.Log.Debug("Add request")
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeAddResponseOp, res)
}

func (ls *LdapSrvHandler) Delete(conn *ldapserver.Conn, msg *ldapserver.Message, dn string) {
	global.Log.Debug("Delete request")
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeDeleteResponseOp, res)
}

func (ls *LdapSrvHandler) Modify(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.ModifyRequest) {
	global.Log.Debug("Modify request")
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeModifyResponseOp, res)
}

func (ls *LdapSrvHandler) ModifyDN(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.ModifyDNRequest) {
	global.Log.Debug("ModifyDN request")
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeModifyDNResponseOp, res)
}

func (ls *LdapSrvHandler) Extended(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.ExtendedRequest) {
	global.Log.Debug("Extended request")
	res := &ldapserver.Result{
		ResultCode: ldapserver.ResultOperationsError,
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeExtendedResponseOp, res)
}

package ldapsrv

import (
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
	totpEnable      bool
}

// checkPassword:
// 普通用户使用 "密码+totp" 校验
// 用于三方软件 bind 的账户, 仅使用 "密码" 校验.
func (ls *LdapSrvHandler) checkPassword(dn ldapserver.DN, password string) bool {
	d, err := ParseLdapDN(dn)
	if err != nil {
		return false
	}

	u, err := getUserInfo(d.Rdn)
	if err != nil {
		return false
	}

	pwd := tools.NewParsePasswd(u.Password)
	// totp 验证开关
	if !ls.totpEnable {
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
	d, err := ParseLdapDN(auth)
	if err != nil {
		return false
	}
	u, err := getUserInfo(d.Rdn)
	if err != nil {
		return false
	}
	// 判断用户是否有 Bind DN 权限
	for _, r := range u.Roles {
		if r.Keyword == config.Conf.LdapServer.BindDNRoleKeyword {
			return true
		}
	}
	return false
}

func getUserInfo(username string) (*model.User, error) {
	uc := model.CacheUserInfoGet(username)
	if uc != nil {
		return uc, nil
	}

	var u model.User
	err := u.Find(map[string]interface{}{"username": username})
	if err != nil {
		return nil, err
	}
	global.Log.Debugf("ldapserver get UserInfo from database: %+v", username)
	model.CacheUserInfoSet(username, &u)
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
	// global.Log.Debugf("LDAP Bind Req: %+v", req)

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

// TODO: &{BaseObject: Scope:0 DerefAliases:0 SizeLimit:0 TimeLimit:0 TypesOnly:false Filter:(objectClass=*) Attributes:[altServer namingContexts supportedCapabilities supportedControl supportedExtension supportedFeatures supportedLdapVersion supportedSASLMechanisms]}
//
// LdapSrvHandler Search method
// Below Search Req can response successfully.
// [normal search]:
// &{BaseObject:dc=example,dc=com Scope:2 DerefAliases:0 SizeLimit:1 TimeLimit:0 TypesOnly:false Filter:(&(uid=test21)(memberOf:=cn=t1,ou=allhands,dc=example,dc=com)) Attributes:[]}
// &{BaseObject:dc=example,dc=com Scope:2 DerefAliases:0 SizeLimit:0 TimeLimit:3 TypesOnly:false Filter:(&(objectClass=people)(uid=test01)(memberOf:=cn=t1,ou=allhands,dc=example,dc=com)) Attributes:[]}
// [gitlab]:
// &{BaseObject:uid=test01,ou=people,dc=example,dc=com Scope:0 DerefAliases:0 SizeLimit:0 TimeLimit:0 TypesOnly:false Filter:(memberOf:=cn=t1,ou=allhands,dc=example,dc=com) Attributes:[dn uid cn mail email userPrincipalName sAMAccountName userid]}
func (ls *LdapSrvHandler) Search(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.SearchRequest) {

	global.Log.Debugf("LDAP Search Req: %+v", req)

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
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
		return
	}

	if ls.abandonment[msg.MessageID] {
		return
	}

	// adapt "gitlab" ldap search request
	if req.BaseObject != "" && req.BaseObject != config.Conf.LdapServer.BaseDN {
		d, err := ParseLdapDN(ldapserver.MustParseDN(req.BaseObject))
		if err != nil {
			global.Log.Errorf("parse req BaseObject to LdapDN error: %s", err)
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
		if d.Rdn == "" {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultSuccess.AsResult(""))
			return
		}

		entry, err := genEntry(d.Rdn)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
				return
			}
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultOperationsError.AsResult(""))
			return
		}

		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultEntryOp, entry)
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultSuccess.AsResult(""))
		return
	}
	// adapt "normal" ldap search request
	query, err := ParseLdapQuery(req.Filter.String())
	if err != nil {
		global.Log.Errorf("parseLdapQuery error: %s", err)
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
		return
	}
	if query.Uid == "" {
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultSuccess.AsResult(""))
		return
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

	entry, err := genEntry(query.Uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultOperationsError.AsResult(""))
		return
	}

	conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultEntryOp, entry)
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultSuccess.AsResult(""))
}

func genEntry(username string) (entry *ldapserver.SearchResultEntry, err error) {
	u, err := getUserInfo(username)
	if err != nil {
		return
	}

	entry = &ldapserver.SearchResultEntry{
		ObjectName: u.UserDN,
		Attributes: []ldapserver.Attribute{
			{Description: "objectClass", Values: []string{"inetOrgPerson"}},
			{Description: "uid", Values: []string{username}},
			{Description: "userid", Values: []string{username}},
			{Description: "cn", Values: []string{username}},
			{Description: "sn", Values: []string{username}},
			{Description: "displayName", Values: []string{u.Nickname}},
			{Description: "givenName", Values: []string{u.GivenName}},
			{Description: "mail", Values: []string{u.Mail}},
			{Description: "email", Values: []string{u.Mail}},
			{Description: "mobile", Values: []string{u.Mobile}},
		},
	}
	return
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

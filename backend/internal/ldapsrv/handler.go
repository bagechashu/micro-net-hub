package ldapsrv

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"strings"
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
	if err != nil || d.Rdn == "" {
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
	if err != nil || d.Rdn == "" {
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

func getGroupInfo(groupname string) (*model.Group, error) {
	var g model.Group
	err := g.Find(map[string]interface{}{"group_name": groupname})
	if err != nil {
		return nil, err
	}
	global.Log.Debugf("ldapserver get GroupInfo from database: %+v", groupname)
	return &g, nil
}

func getGroupsInfo() ([]*model.Group, error) {
	var gs model.Groups
	err := gs.Find(map[string]interface{}{"group_type": GroupOfUniqueNamesFields})
	if err != nil {
		return nil, err
	}
	return gs, nil
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

// ### TODO: search req NOT supported: ###
// [normal Wildcards search]
// &{BaseObject: Scope:0 DerefAliases:0 SizeLimit:0 TimeLimit:0 TypesOnly:false Filter:(objectClass=*) Attributes:[altServer namingContexts supportedCapabilities supportedControl supportedExtension supportedFeatures supportedLdapVersion supportedSASLMechanisms]}
// [nexus user mapping]:
// x &{BaseObject:ou=people,dc=example,dc=com Scope:2 DerefAliases:3 SizeLimit:20 TimeLimit:0 TypesOnly:false Filter:(&(objectClass=inetOrgPerson)(uid=*)) Attributes:[uid cn mail labeledUri memberOf:=cn=t1,ou=allhands,dc=example,dc=com]}

// ### search req supported: ###
// [normal search]:
// √ &{BaseObject:dc=example,dc=com Scope:2 DerefAliases:0 SizeLimit:1 TimeLimit:0 TypesOnly:false Filter:(&(uid=test21)(memberOf:=cn=t1,ou=allhands,dc=example,dc=com)) Attributes:[]}
// √ &{BaseObject:dc=example,dc=com Scope:2 DerefAliases:0 SizeLimit:0 TimeLimit:3 TypesOnly:false Filter:(&(objectClass=inetOrgPerson)(uid=test01)(memberOf:=cn=t1,ou=allhands,dc=example,dc=com)) Attributes:[]}
// [gitlab]:
// √ &{BaseObject:uid=test01,ou=people,dc=example,dc=com Scope:0 DerefAliases:0 SizeLimit:0 TimeLimit:0 TypesOnly:false Filter:(memberOf:=cn=t1,ou=allhands,dc=example,dc=com) Attributes:[dn uid cn mail email userPrincipalName sAMAccountName userid]}
// [nexus]:
// √ &{BaseObject:dc=example,dc=com Scope:2 DerefAliases:3 SizeLimit:1 TimeLimit:0 TypesOnly:false Filter:(&(objectClass=inetOrgPerson)(uid=test01)) Attributes:[uid cn mail labeledUri memberOf:=cn=t1,ou=allhands,dc=example,dc=com]}
// √ &{BaseObject:ou=people,dc=example,dc=com Scope:2 DerefAliases:3 SizeLimit:1 TimeLimit:0 TypesOnly:false Filter:(&(objectClass=inetOrgPerson)(uid=test01)) Attributes:[uid cn mail labeledUri memberOf:=cn=t1,ou=allhands,dc=example,dc=com]}

// LdapSrvHandler Search method
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

	// parse ldap search request
	query, err := ParseLdapQuery(req.Filter.String())
	if err != nil {
		global.Log.Errorf("parseLdapQuery error: %s", err)
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
		return
	}

	ocs := strings.Split(query.ObjectClass, "|")
	for _, oc := range ocs {
		switch oc {
		case "inetOrgPerson":
			searchUser(conn, msg, req, query)
		case "groupOfUniqueNames", "groupOfNames", "organizationalUnit":
			searchGroup(conn, msg, req, query)
		default:
			searchUser(conn, msg, req, query)
		}
	}
}

func searchUser(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.SearchRequest, query *Query) {
	var user string
	var group string

	// adapt "gitlab" ldap search request:
	// if request is "gitlab" Req, it cat get "user" from the BaseObject DN.
	// if request is normal Req, "user" is empty.
	if req.BaseObject != "" && req.BaseObject != config.Conf.LdapServer.BaseDN {
		bodn, err := ldapserver.ParseDN(req.BaseObject)
		if err != nil {
			global.Log.Errorf("Error parsing DN: %s", req.BaseObject)
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
		d, err := ParseLdapDN(bodn)
		if err != nil {
			global.Log.Errorf("parse req BaseObject to LdapDN error: %s", err)
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
		if d.Rdn != "" {
			user = d.Rdn
		}
	}

	// adapt "normal" ldap search request
	if user == "" {
		if query.Uid != "" {
			user = query.Uid
		} else {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultSuccess.AsResult(""))
			return
		}

	}

	if query.MemberOf != "" {
		group = query.MemberOf
	} else {
		// adapt "Nexus" Dynamic groups filter:
		// "Nexus" Dynamic groups filter set "memeberOf" at Attributes.
		// So we need to get "memberOf" from Attributes.
		attr := parseLdapAttributes(req.Attributes)
		if attr.MemberOf != "" {
			group = attr.MemberOf
		}
	}

	// adapt Search with group filter:
	// if group is not empty, we need to check if user is in group.
	if group != "" {
		find, err := model.UserExistsInGroup(user, group)
		if err != nil {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultOperationsError.AsResult(err.Error()))
			return
		}
		if !find {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
	}

	// generate entry of user
	entry, err := genUserEntry(user)
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

func genUserEntry(username string) (entry *ldapserver.SearchResultEntry, err error) {
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

func searchGroup(conn *ldapserver.Conn, msg *ldapserver.Message, req *ldapserver.SearchRequest, query *Query) {
	var group string

	// adapt "harbor" ldap search request:
	// SRCH base="cn=t1,ou=allhands,dc=example,dc=com" scope=2 deref=0 filter="(&(objectClass=groupOfUniqueNames)(cn=*))"
	if req.BaseObject != "" && req.BaseObject != config.Conf.LdapServer.BaseDN {
		bodn, err := ldapserver.ParseDN(req.BaseObject)
		if err != nil {
			global.Log.Errorf("Error parsing DN: %s", req.BaseObject)
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
		d, err := ParseLdapDN(bodn)
		if err != nil {
			global.Log.Errorf("parse req BaseObject to LdapDN error: %s", err)
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
		if d.Rdn != "" {
			group = d.Rdn
		}
	}

	// if group is not empty, response the entry that queryed.
	// if group is empty, response all group entrys.
	if group != "" {
		entry, err := genGroupEntry(group)
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

	entrys, err := genGroupEntrys()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultNoSuchObject.AsResult(""))
			return
		}
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultOperationsError.AsResult(""))
		return
	}
	for _, entry := range entrys {
		conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultEntryOp, entry)
	}
	conn.SendResult(msg.MessageID, nil, ldapserver.TypeSearchResultDoneOp, ldapserver.ResultSuccess.AsResult(""))
}

func genGroupEntry(groupname string) (entry *ldapserver.SearchResultEntry, err error) {
	g, err := getGroupInfo(groupname)
	if err != nil {
		return
	}

	entry = &ldapserver.SearchResultEntry{
		ObjectName: g.GroupDN,
		Attributes: []ldapserver.Attribute{
			{Description: "cn", Values: []string{g.GroupName}},
			{Description: "description", Values: []string{g.Remark}},
		},
	}
	return
}

func genGroupEntrys() (entrys []*ldapserver.SearchResultEntry, err error) {
	gs, err := getGroupsInfo()
	if err != nil {
		return
	}

	for _, g := range gs {
		entrys = append(entrys, &ldapserver.SearchResultEntry{
			ObjectName: g.GroupDN,
			Attributes: []ldapserver.Attribute{
				{Description: "cn", Values: []string{g.GroupName}},
				{Description: "description", Values: []string{g.Remark}},
			},
		})
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

const custom = {
  common: {
    areyousure: "你确定吗?",
    confirm: "确定",
    cancel: "取消",
    delete: "删除",
    edit: "编辑",
    add: "添加",
    save: "保存",
    close: "关闭",
    reset: "重置",
    search: "搜索",
    index: "首页",
    total: "共计",
    refresh: "刷新",
    closeOthers: "关闭其他",
    closeAll: "关闭所有",
    nodata: "暂无数据"
  },
  valid: {
    length: "长度在 {0} 到 {1} 个字符",
    mustInput: "必须填写",
    pleaseInput: "请输入查询项"
  },
  tips: {
    switchSizeSuccess: "切换表格尺寸成功",
    formValidFailed: "表单验证失败",
    notFound: "未找到查询的内容",
    notFoundAndRetry: "未找到, 请重试",
    foundSome: "查询到 {0} 条内容",
    copySuccess: "复制成功",
    copyFailed: "复制失败, {0}",
    addBookmarkInfo: "如果未添加成功, 请按 Ctrl+D 或 Command+D 将此页面添加至书签"
  },
  sidebar: {
    Sitenav: "网站导航",
    UserManage: "人员管理",
    User: "用户管理",
    Group: "分组管理",
    FieldRelation: "字段关系管理",
    System: "系统管理",
    Role: "角色管理",
    Menu: "菜单管理",
    Api: "接口管理",
    SitenavManager: "导航配置",
    DnsManager: "Dns配置",
    NoticeManager: "公告管理",
    Log: "日志管理",
    OperationLog: "操作日志",
    Profile: "个人中心"
  },
  loginform: {
    login: "登录",
    logout: "退出登录",
    username: "用户名",
    password: "密码",
    forgetPassword: "忘记密码?",
    oldpass: "原密码",
    newpass: "新密码",
    confirmNewPass: "确认密码",
    oldpassTips: "请输入原密码",
    newpassTips: "请输入新密码",
    confirmNewPassTips: "请再次输入新密码",
    oldpassErr: "请输入旧密码",
    newpassErr: "新密码不能为空",
    newpassValidErrLenMustThan8: "密码长度必须大于8位",
    newpassValidErrMustComplex: "密码要包含数字、大写字母、小写字母、特殊字符中的至少三种",
    confirmNewPassErr: "两次输入密码不一致",
    changePasswordSuccess: "修改密码成功，请重新登录",
    changePasswordErr: "修改密码失败"
  },
  profile: {
    profile: "个人中心",
    aboutme: "关于我",
    changePassword: "修改个人密码",
    changeTotp: "修改 TOTP 秘钥",
    changeTotpTips: "请输入OTP",
    changeTotpTipsValid: "OTP 为6位数字码"
  },
  totpNotice: {
    notice: "注意:",
    content: [
      "重置后, 之前的TOTP秘钥将失效.",
      "二维码只显示一次, 刷新/切换页面后将消失."
    ]
  },
  sitenav: {
    searchTips: "内网地址搜索",
    visit: "跳转",
    doc: "相关文档",
    copyURL: "拷贝网址",
    addBookmark: "加入书签",
    totolInfo: "共计：{0} 个站点"
  }
};

export default custom;

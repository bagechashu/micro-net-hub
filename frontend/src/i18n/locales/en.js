const custom = {
  common: {
    areyousure: "Are you sure?",
    confirm: "Confirm",
    cancel: "Cancel",
    delete: "Delete",
    edit: "Edit",
    add: "Add",
    save: "Save",
    close: "Close",
    reset: "Reset",
    search: "Search",
    index: "Home",
    total: "Total",
    refresh: "Refresh",
    closeOthers: "Close Others",
    closeAll: "Close All"
  },
  message: {
    switchSizeSuccess: "Table size switch successful"
  },
  sidebar: {
    Sitenav: "Site Navigation",
    UserManage: "Account Manage",
    User: "User",
    Group: "Group",
    FieldRelation: "Field Relation",
    System: "System Manage",
    Role: "Role",
    Menu: "Menu",
    Api: "API",
    SitenavManager: "Site Navigation Manage",
    DnsManager: "DNS Manage",
    NoticeManager: "Notice Manage",
    Log: "Log Manage",
    OperationLog: "Operation Log",
    Profile: "Profile"
  },
  tagview: {},
  loginform: {
    login: "Login",
    logout: "Logout",
    username: "Username",
    password: "Password",
    forgetPassword: "Forgot Password?",
    oldpass: "Old Password",
    newpass: "New Password",
    confirmNewPass: "Retype New Password",
    oldpassTips: "Please enter your old password",
    newpassTips: "Please enter your new password",
    confirmNewPassTips: "Please re-enter your new password",
    oldpassErr: "Please enter your old password",
    newpassErr: "New password cannot be empty",
    newpassValidErrLenMustThan8: "length must be greater than 8 characters",
    newpassValidErrMustComplex: "must include at least 3 type characters",
    confirmNewPassErr: "Passwords do not match",
    changePasswordSuccess: "Password changed successfully, please log in again",
    changePasswordErr: "Failed to change password",
    changePasswordValidErr: "Failed to validate the Change Password form."
  },
  profile: {
    profile: "Profile",
    aboutme: "About Me",
    changePassword: "Change Password",
    changeTotp: "Change TOTP Key",
    changeTotpTips: "Please enter the OTP",
    changeTotpTipsValid: "OTP must be a 6-digit numeric code.",
    changeTotpTipsValidErr: "Failed to validate the TOTP reset form."
  },
  totpNotice: {
    notice: "Note:",
    content: [
      "After resetting, the previous TOTP key will become invalid.",
      "The QR code will be displayed only once and will disappear after refreshing or switching the page."
    ]
  },
  valid: {
    length: "Must be between {0} and {1} characters"
  }
};

export default custom;

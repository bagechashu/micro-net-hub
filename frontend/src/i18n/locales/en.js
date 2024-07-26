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
    closeAll: "Close All",
    nodata: "No data available"
  },
  valid: {
    length: "Must be between {0} and {1} characters",
    mustInput: "This field is required",
    pleaseInput: "Please enter a search item"
  },
  tips: {
    switchSizeSuccess: "Table size switch successful",
    formValidFailed: "Form validation failed",
    notFound: "No content found for your query",
    notFoundAndRetry: "No content found, please try again",
    foundSome: "{0} items found",
    copySuccess: "Copy successful",
    copyFailed: "Copy failed, {0}",
    addBookmarkInfo: "If adding was unsuccessful, please press Ctrl+D or Command+D to bookmark this page"
  },
  sidebar: {
    Sitenav: "Site Navigation",
    UserManage: "Account Management",
    User: "User",
    Group: "Group",
    FieldRelation: "Field Relation",
    System: "System Management",
    Role: "Role",
    Menu: "Menu",
    Api: "API",
    SitenavManager: "Site Navigation Management",
    DnsManager: "DNS Management",
    NoticeManager: "Notice Management",
    Log: "Log Management",
    OperationLog: "Operation Log",
    Profile: "Profile"
  },
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
    changePasswordErr: "Failed to change password"
  },
  profile: {
    profile: "Profile",
    aboutme: "About Me",
    changePassword: "Change Password",
    changeTotp: "Change TOTP Key",
    changeTotpTips: "Please enter the OTP",
    changeTotpTipsValid: "OTP must be a 6-digit numeric code."
  },
  totpNotice: {
    notice: "Note:",
    content: [
      "After resetting, the previous TOTP key will become invalid.",
      "The QR code will be displayed only once and will disappear after refreshing or switching the page."
    ]
  },
  sitenav: {
    searchTips: "Intranet address search",
    visit: "Visit",
    doc: "Related Documents",
    copyURL: "Copy URL",
    addBookmark: "Add to Bookmarks",
    totolInfo: "Total: {0} sites"
  }
};

export default custom;

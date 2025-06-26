package setup

import (
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"regexp"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ch_translations "github.com/go-playground/validator/v10/translations/zh"
)

// 初始化Validator数据校验
func InitValidate() {
	chinese := zh.New()
	uni := ut.New(chinese, chinese)
	trans, _ := uni.GetTranslator("zh")
	global.Trans = trans
	global.Validate = validator.New()
	_ = ch_translations.RegisterDefaultTranslations(global.Validate, global.Trans)
	_ = global.Validate.RegisterValidation("checkMobile", checkMobile)
	_ = global.Validate.RegisterValidation("checkSecondLevelDomain", checkSecondLevelDomain)
	global.Log.Infof("初始化validator.v10数据校验器完成")
}

func checkMobile(fl validator.FieldLevel) bool {
	return tools.CheckMobile(fl.Field().String())
}

// ValidateSecondLevelDomain 校验是否为二级域名
func checkSecondLevelDomain(fl validator.FieldLevel) bool {
	domain := fl.Field().String()

	// 基本正则匹配（不含子域名的二级域名), 仅允许末尾不带.的域名
	// 支持 a-z0-9- 的域名，后缀至少 2 位字符
	reg := regexp.MustCompile(`^(?i)[a-zA-Z0-9-]+\.[a-z]{2,}$`)
	return reg.MatchString(domain)
}

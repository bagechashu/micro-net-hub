package setup

import (
	"micro-net-hub/internal/global"
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
	global.Log.Infof("初始化validator.v10数据校验器完成")
}

// Match Chinese and international phone numbers
func checkMobile(fl validator.FieldLevel) bool {
	reg := `^(\+|00)??(\d{1,3})??((1|0)\d{8,10})??$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(fl.Field().String())
}

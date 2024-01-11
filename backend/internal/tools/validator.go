package tools

import (
	"regexp"
)

// 校验邮箱格式
func CheckEmail(email string) bool {
	// 邮箱格式的正则表达式
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	rgx := regexp.MustCompile(regex)

	return rgx.MatchString(email)
}

// Match Chinese and international phone numbers
func CheckMobile(mobileNo string) bool {
	reg := `^(\+|00)??(\d{1,3})??((1|0)\d{8,10})??$`
	rgx := regexp.MustCompile(reg)

	return rgx.MatchString(mobileNo)
}

// qq号校验
func CheckQQNo(qqNo string) bool {
	// 判断qq号码是否合法的正则表达式
	regex := `^[1-9][0-9]{4,10}$`
	rgx := regexp.MustCompile(regex)

	return rgx.MatchString(qqNo)
}

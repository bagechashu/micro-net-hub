package helper

import (
	"fmt"
	"micro-net-hub/internal/global"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var ReqAssertErr = NewRspError(SystemErr, fmt.Errorf("请求异常"))

const (
	SystemErr    = 500
	MySqlErr     = 501
	LdapErr      = 505
	OperationErr = 506
	ConfigErr    = 507
	ValidatorErr = 412
)

type RspError struct {
	code int
	err  error
}

func (re *RspError) Error() string {
	return re.err.Error()
}

func (re *RspError) Code() int {
	return re.code
}

// NewRspError New
func NewRspError(code int, err error) *RspError {
	return &RspError{
		code: code,
		err:  err,
	}
}

// NewMySqlError mysql错误
func NewMySqlError(err error) *RspError {
	return NewRspError(MySqlErr, err)
}

// NewValidatorError 验证错误
func NewValidatorError(err error) *RspError {
	return NewRspError(ValidatorErr, err)
}

// NewLdapError ldap错误
func NewLdapError(err error) *RspError {
	return NewRspError(LdapErr, err)
}

// NewOperationError 操作错误
func NewOperationError(err error) *RspError {
	return NewRspError(OperationErr, err)
}

// NewConfigError 操作错误
func NewConfigError(err error) *RspError {
	return NewRspError(ConfigErr, err)
}

// ReloadErr 重新加载错误
func ReloadErr(err interface{}) *RspError {
	rspErr, ok := err.(*RspError)
	if !ok {
		rspError, ok := err.(error)
		if !ok {
			return &RspError{
				code: SystemErr,
				err:  fmt.Errorf("unknow error"),
			}
		}
		return &RspError{
			code: SystemErr,
			err:  rspError,
		}
	}
	return rspErr
}

func BindAndValidateRequest(c *gin.Context, reqStruct interface{}) {
	var err error

	if reqStruct == nil {
		return
	}

	// bind struct
	err = c.Bind(reqStruct)
	if err != nil {
		Err(c, NewValidatorError(err), nil)
		return
	}
	//
	err = global.Validate.Struct(reqStruct)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			global.Log.Errorf("bind reqStruct err: \n\treqStruct: %+v\n\terr: %s", reqStruct, err)
			Err(c, NewValidatorError(fmt.Errorf(err.Translate(global.Trans))), nil)
			return
		}
	}

}

func HandleResponse(c *gin.Context, data interface{}, rspError interface{}) {
	if rspError != nil {
		Err(c, ReloadErr(rspError), data)
		return
	}
	Success(c, data)
}

// Success http 成功
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

// Err http 错误
func Err(c *gin.Context, err *RspError, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": err.Code(),
		"msg":  err.Error(),
		"data": data,
	})
}

// 返回前端
func Response(c *gin.Context, httpStatus int, code int, data gin.H, message string) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"data":    data,
		"message": message,
	})
}

// 返回前端-成功
func SuccessWithMessage(c *gin.Context, data gin.H, message string) {
	Response(c, http.StatusOK, 200, data, message)
}

// 返回前端-失败
func FailWithMessage(c *gin.Context, data gin.H, message string) {
	Response(c, http.StatusBadRequest, 400, data, message)
}

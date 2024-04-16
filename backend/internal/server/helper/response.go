package helper

import (
	"fmt"
	"micro-net-hub/internal/global"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var ReqAssertErr = NewRspError(RspCodeSystemErr, fmt.Errorf("请求异常"))

const (
	RspCodeSuccess       = 200
	RspCodeBadRequestErr = 400
	RspCodeValidatorErr  = 412
	RspCodeSystemErr     = 500
	RspCodeMySqlErr      = 501
	RspCodeLdapErr       = 505
	RspCodeOperationErr  = 506
	RspCodeConfigErr     = 507
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
	return NewRspError(RspCodeMySqlErr, err)
}

// NewValidatorError 验证错误
func NewValidatorError(err error) *RspError {
	return NewRspError(RspCodeValidatorErr, err)
}

// NewLdapError ldap错误
func NewLdapError(err error) *RspError {
	return NewRspError(RspCodeLdapErr, err)
}

// NewOperationError 操作错误
func NewOperationError(err error) *RspError {
	return NewRspError(RspCodeOperationErr, err)
}

// NewConfigError 操作错误
func NewConfigError(err error) *RspError {
	return NewRspError(RspCodeConfigErr, err)
}

// ReloadErr 重新加载错误
func ReloadErr(err interface{}) *RspError {
	rspErr, ok := err.(*RspError)
	if !ok {
		rspError, ok := err.(error)
		if !ok {
			return &RspError{
				code: RspCodeSystemErr,
				err:  fmt.Errorf("unknow error"),
			}
		}
		return &RspError{
			code: RspCodeSystemErr,
			err:  rspError,
		}
	}
	return rspErr
}

type HandlerLogic func(c *gin.Context, reqStructInstance interface{}) (result interface{}, respError interface{})

type EmptyStruct struct{}

func HandleRequest(c *gin.Context, reqStructInstance interface{}, fn HandlerLogic) {
	var err error

	// bind struct
	if err = c.Bind(reqStructInstance); err != nil {
		Err(c, NewValidatorError(err), nil)
		return
	}

	// reqStruct validate check
	if err = global.Validate.Struct(reqStructInstance); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			global.Log.Errorf("bind reqStruct err: \n\treqStruct: %+v\n\terr: %s", reqStructInstance, err)
			Err(c, NewValidatorError(fmt.Errorf(err.Translate(global.Trans))), nil)
			return
		}
	}

	// exec HandlerLogic
	data, rspError := fn(c, reqStructInstance)
	if rspError != nil {
		Err(c, ReloadErr(rspError), data)
		return
	}
	Success(c, data)
}

func BindAndValidateRequest(c *gin.Context, reqStructInstance interface{}) error {
	// bind struct
	if err := c.Bind(reqStructInstance); err != nil {
		ErrV2(c, NewValidatorError(err))
		return err
	}

	// reqStruct validate check
	if errs := global.Validate.Struct(reqStructInstance); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			global.Log.Errorf("bind reqStruct err: \n\treqStruct: %+v\n\terr: %s", reqStructInstance, err)
			ErrV2(c, NewValidatorError(fmt.Errorf(err.Translate(global.Trans))))
			return err
		}
	}
	return nil
}

// Success http 成功
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": RspCodeSuccess,
		"msg":  "success",
		"data": data,
	})
}

// Err Response
func ErrV2(c *gin.Context, err *RspError) {
	c.JSON(http.StatusOK, gin.H{
		"errcode": err.Code(),
		"msg":     err.Error(),
	})
}

// Err Response with data
func ErrV2WithData(c *gin.Context, err *RspError, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"errcode": err.Code(),
		"msg":     err.Error(),
		"data":    data,
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
	Response(c, http.StatusOK, RspCodeSuccess, data, message)
}

// 返回前端-失败
func FailWithMessage(c *gin.Context, data gin.H, message string) {
	Response(c, http.StatusBadRequest, RspCodeBadRequestErr, data, message)
}

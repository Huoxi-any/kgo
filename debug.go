package kgo

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// DumpPrint 打印调试变量.
func (ks *LkkDebug) DumpPrint(vs ...interface{}) {
	dumpPrint(vs...)
}

// DumpStacks 打印堆栈信息.
func (kd *LkkDebug) DumpStacks() {
	buf := make([]byte, 16384)
	buf = buf[:runtime.Stack(buf, true)]
	fmt.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===\n\n", buf)
}

// GetCallName 获取调用的方法名称;f为目标方法;onlyFun为true时仅返回方法,不包括包名.
func (kd *LkkDebug) GetCallName(f interface{}, onlyFun bool) string {
	var funcObj *runtime.Func
	r := reflectPtr(reflect.ValueOf(f))
	switch r.Kind() {
	case reflect.Invalid:
		// Skip this function, and fetch the PC and file for its parent
		pc, _, _, _ := runtime.Caller(1)
		// Retrieve a Function object this functions parent
		funcObj = runtime.FuncForPC(pc)
	case reflect.Func:
		funcObj = runtime.FuncForPC(r.Pointer())
	default:
		return ""
	}

	name := funcObj.Name()
	if onlyFun {
		// extract just the function name (and not the module path)
		return strings.TrimPrefix(filepath.Ext(name), ".")
	}

	return name
}

// GetCallFile 获取调用方法的文件路径.
func (kd *LkkDebug) GetCallFile() string {
	_, file, _, _ := runtime.Caller(1)
	return file
}

// GetCallDir 获取调用方法的文件目录.
func (kd *LkkDebug) GetCallDir() string {
	return filepath.Dir(kd.GetCallFile())
}

// GetCallLine 获取调用方法的行号.
func (kd *LkkDebug) GetCallLine() int {
	// Skip this function, and fetch the PC and file for its parent
	_, _, line, _ := runtime.Caller(1)
	return line
}

// GetCallPackage 获取调用方法或调用文件的包名.callFile为调用文件路径.
func (kd *LkkDebug) GetCallPackage(callFile ...string) string {
	var sourceFile string
	if len(callFile) == 0 {
		sourceFile = kd.GetCallFile()
	} else {
		sourceFile = callFile[0]
	}

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, sourceFile, nil, parser.PackageClauseOnly)
	if err != nil || astFile.Name == nil {
		return ""
	}

	return astFile.Name.Name
}

// HasMethod 检查对象t是否具有method方法.
func (kd *LkkDebug) HasMethod(t interface{}, method string) bool {
	_, err := methodExists(t, method)
	return err == nil
}

// GetMethod 获取对象t中的method方法.
// 注意:返回的方法中的第一个参数是接收者.
// 所以,调用返回的方法时,必须将接收者作为第一个参数传递.
func (kd *LkkDebug) GetMethod(t interface{}, method string) interface{} {
	return getMethod(t, method)
}

// CallMethod 调用对象的方法.
// 若执行成功,则结果是该方法的返回结果;
// 否则返回(nil, error).
func (kd *LkkDebug) CallMethod(t interface{}, method string, args ...interface{}) ([]interface{}, error) {
	m := kd.GetMethod(t, method)
	if m == nil {
		return nil, fmt.Errorf("[CallMethod] The %#v have no method: %s", t, method)
	}

	return CallFunc(m, args...)
}

// GetFuncNames 获取变量的所有(公开的)函数名.
func (kd *LkkDebug) GetFuncNames(obj interface{}) (res []string) {
	return getFuncNames(obj)
}

// WrapError 错误包裹.
func (ks *LkkDebug) WrapError(err error, args ...interface{}) (res error) {
	num := len(args)
	if err == nil && num == 0 {
		res = errors.New("[WrapError] parameter error")
	} else if err != nil && num == 0 {
		res = err
	} else {
		var msg []string
		for _, v := range args {
			msg = append(msg, toStr(v))
		}

		if err != nil {
			msg = append(msg, fmt.Sprintf("last error: %s", err.Error()))
		}

		res = errors.New(strings.Join(msg, "\r\n"))
	}

	return
}

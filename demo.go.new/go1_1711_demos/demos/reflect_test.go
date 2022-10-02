package demos

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//
// Demo: reflect
//

type GetOkrDetailResp struct {
	OkrId   int64
	UInfo   *UserInfo
	ObjList []*ObjInfo
}

func (resp *GetOkrDetailResp) PrettyPrint() {
	objStrs := make([]string, 0, len(resp.ObjList))
	for _, obj := range resp.ObjList {
		objStrs = append(objStrs, fmt.Sprintf("id=%d,content=[%s]", obj.ObjId, obj.Content))
	}
	fmt.Printf("id=%d,list=[%s]\n", resp.OkrId, strings.Join(objStrs, "|"))
}

type ObjInfo struct {
	ObjId   int64
	Content string
}

type UserInfo struct {
	Name         string
	Age          int
	IsLeader     bool
	Salary       float64
	privateFiled int
}

func (uInfo *UserInfo) PrettyString() string {
	return fmt.Sprintf("name=%s,age=%d,is_leader=%v,salary=%.2f\n", uInfo.Name, uInfo.Age, uInfo.IsLeader, uInfo.Salary)
}

// NewUserInfoByReflect 利用反射创建结构体。
func NewUserInfoByReflect(req interface{}) *UserInfo {
	if req == nil {
		return nil
	}

	reqType := reflect.TypeOf(req)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}
	return reflect.New(reqType).Interface().(*UserInfo)
}

// ModifyOkrDetailRespData 修改struct字段值。
func ModifyOkrDetailRespData(req interface{}) error {
	reqValue := reflect.ValueOf(req).Elem()
	if !reqValue.CanSet() {
		return fmt.Errorf("value cannot be set")
	}

	uType := reqValue.FieldByName("UInfo").Type().Elem() // UInfo 是指针类型 *UserInfo
	uInfo := reflect.New(uType)
	reqValue.FieldByName("UInfo").Set(uInfo)
	return nil
}

// FilterOkrRespData 读取struct字段值，并根据条件进行过滤。
func FilterOkrRespData(reqData interface{}, objId int64) {
	valueOf := reflect.ValueOf(reqData).Elem()
	for i := 0; i < valueOf.NumField(); i++ {
		fieldValue := valueOf.Field(i)
		if fieldValue.Kind() != reflect.Slice {
			continue
		}

		fieldType := fieldValue.Type()                      // type: []*ObjInfo
		sliceType := fieldType.Elem()                       // type: *ObjInfo
		slicePtr := reflect.New(reflect.SliceOf(sliceType)) // 创建一个指向 slice 的指针
		slice := slicePtr.Elem()
		slice.Set(reflect.MakeSlice(reflect.SliceOf(sliceType), 0, 0))
		for i := 0; i < fieldValue.Len(); i++ {
			if fieldValue.Index(i).Elem().FieldByName("ObjId").Int() != objId {
				continue
			}
			slice = reflect.Append(slice, fieldValue.Index(i))
		}
		fieldValue.Set(slice)
	}
}

func TestReflectOkrResp(t *testing.T) {
	// 利用反射创建一个新的对象
	var uInfo *UserInfo
	uInfo = NewUserInfoByReflect((*UserInfo)(nil))
	assert.NotNil(t, uInfo)
	fmt.Printf("new user info: %+v\n", uInfo)

	uInfo = NewUserInfoByReflect(uInfo)
	assert.NotNil(t, uInfo)
	fmt.Printf("new user info: %+v\n", uInfo)
	fmt.Println()

	// 修改 resp 返回值里面的 user info 字段（初始化）
	reqData1 := new(GetOkrDetailResp)
	assert.Nil(t, reqData1.UInfo)
	err := ModifyOkrDetailRespData(reqData1)
	assert.NoError(t, err)
	fmt.Printf("modified user info: %+v\n", reqData1.UInfo)
	fmt.Println()

	// 对 respData 进行过滤操作
	reqData := &GetOkrDetailResp{OkrId: 123}
	for i := 0; i < 10; i++ {
		reqData.ObjList = append(reqData.ObjList, &ObjInfo{ObjId: int64(i), Content: fmt.Sprint(i)})
	}
	fmt.Println("before filter:")
	reqData.PrettyPrint()

	FilterOkrRespData(reqData, 6)
	fmt.Println("after filter:")
	reqData.PrettyPrint()
}

//
// Demo: 从切片中过滤指定元素，不修改原切片
//

func DeleteSliceElms(s interface{}, elms ...interface{}) interface{} {
	m := make(map[interface{}]struct{})
	for _, e := range elms {
		m[e] = struct{}{}
	}

	v := reflect.ValueOf(s)
	res := reflect.MakeSlice(reflect.TypeOf(s), 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		if _, ok := m[v.Index(i).Interface()]; !ok {
			res = reflect.Append(res, v.Index(i))
		}
	}
	return res.Interface()
}

func TestDeleteSliceElms(t *testing.T) {
	slice := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	elms := []interface{}{uint64(1), uint64(3), uint64(5), uint64(7), uint64(9)}
	res := DeleteSliceElms(slice, elms...)
	t.Logf("results: %v", res)
}

//
// Demo: get func name and run by reflect
//

type caller func(string)

func sayHello(name string) {
	fmt.Println("Hello:", name)
}

func exec(c interface{}, params ...interface{}) {
	typeOf := reflect.TypeOf(c)
	fmt.Println("type:", typeOf.Kind())
	if typeOf.Kind() != reflect.Func {
		fmt.Println("not caller")
		return
	}

	// get func name
	valueOf := reflect.ValueOf(c)
	name := runtime.FuncForPC(valueOf.Pointer()).Name()
	pkgName, funcName := getFuncName(name)
	fmt.Printf("exec: pkg=%s, func=%s()\n", pkgName, funcName)

	// run func()
	paramValues := make([]reflect.Value, 0, len(params))
	for _, param := range params {
		paramValues = append(paramValues, reflect.ValueOf(param))
	}
	valueOf.Call(paramValues)

	_, ok := valueOf.Interface().(caller)
	fmt.Println("is caller:", ok)
}

func getFuncName(fullName string) (pkgName, funcName string) {
	items := strings.Split(fullName, ".")
	return items[0], strings.Join(items[1:], ".")
}

func TestGetFuncNameByReflect(t *testing.T) {
	exec(sayHello, "foo")
	t.Log("done")
}

//
// Demo: function proxy by reflect
//

type personInterface interface {
	String() string
	SayHello(string)
}

type personImpl struct {
	Name string
	Age  int
}

func (p *personImpl) String() string {
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	return fmt.Sprintf("person: name=%s,age=%d", p.Name, p.Age)
}

func (p *personImpl) SayHello(name string) {
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	fmt.Printf("hello: %s\n", name)
}

type personProxy struct {
	String_   func() string
	SayHello_ func(string)
}

func (p *personProxy) String() string {
	return p.String_()
}

func (p *personProxy) SayHello(tag string) {
	p.SayHello_(tag)
}

func createProxy(impl, proxy interface{}) (interface{}, error) {
	if _, _, err := checkParam(impl); err != nil {
		return nil, err
	}
	valueOf, typeOf, err := checkParam(proxy)
	if err != nil {
		return nil, err
	}

	for i := 0; i < valueOf.NumField(); i++ { // use struct here, get fields
		fieldTypeOf := typeOf.Field(i) // fieldTypeOf: field type info (name,type) when define struct
		fmt.Printf("walk field: type=%s,name=%s\n", fieldTypeOf.Type.Kind(), fieldTypeOf.Name)
		if !strings.HasSuffix(fieldTypeOf.Name, "_") {
			continue
		}

		fieldValueOf := valueOf.Field(i) // fieldValueOf: field value info (name,value) when init struct
		if fieldValueOf.Kind() == reflect.Func && fieldValueOf.IsValid() && fieldValueOf.CanSet() {
			fmt.Printf("wrap func: %s()\n", fieldTypeOf.Name)
			funcName := strings.TrimSuffix(fieldTypeOf.Name, "_")
			implFunc := reflect.ValueOf(impl).MethodByName(funcName) // use pointer here, get impl func
			if implFunc.IsNil() {
				return nil, fmt.Errorf("impl func name is not found: %s", funcName)
			}
			fieldValueOf.Set(reflect.MakeFunc(fieldTypeOf.Type, wrapFunctionWithDebug(implFunc)))
		}
	}

	return proxy, nil
}

func checkParam(param interface{}) (reflect.Value, reflect.Type, error) {
	typeOf := reflect.TypeOf(param)
	if typeOf.Kind() != reflect.Ptr {
		return reflect.Value{}, nil, fmt.Errorf("parameter must be pointer")
	}

	valueOf := reflect.ValueOf(param).Elem()
	typeOf = valueOf.Type()
	if typeOf.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("pointer must be ref to struct")
	}
	return valueOf, typeOf, nil
}

// wrapFunctionWithDebug adds debug info when run function.
func wrapFunctionWithDebug(f reflect.Value) func(in []reflect.Value) []reflect.Value {
	return func(in []reflect.Value) []reflect.Value {
		fmt.Printf("\ncall func: %s()\n", getCallingMethodName(3))
		printFuncInOutValues(in, "param")
		start := time.Now()
		out := f.Call(in)
		fmt.Printf("\tprofile: %d milliseconds\n", time.Since(start).Milliseconds())
		printFuncInOutValues(out, "result")
		return out
	}
}

func getCallingMethodName(skip int) string {
	pc := make([]uintptr, 1)
	runtime.Callers(skip, pc)
	fullName := runtime.FuncForPC(pc[0]).Name()
	_, funcName := getFuncName(fullName)
	return funcName
}

func printFuncInOutValues(values []reflect.Value, tag string) {
	if len(values) == 0 {
		fmt.Printf("\t%s: nil\n", tag)
		return
	}
	for _, val := range values {
		fmt.Printf("\t%s: type=%s,value=[%v]\n", tag, val.Type().Name(), val.Interface())
	}
}

func TestCreateProxyForPerson(t *testing.T) {
	rand.Seed(time.Now().Unix())
	impl := &personImpl{Name: "foo", Age: 31}
	proxy, err := createProxy(impl, &personProxy{})
	assert.NoError(t, err)

	p, ok := proxy.(personInterface)
	assert.True(t, ok)
	p.SayHello("bar")
	fmt.Println(p.String())
}

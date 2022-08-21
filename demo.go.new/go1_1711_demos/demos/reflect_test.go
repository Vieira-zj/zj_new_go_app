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

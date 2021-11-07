package demos

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

type ReflectUser01 struct {
	Gender int
}

//
// 变量类型和属性操作 TypeOf
//

func TestReflectTypeOfDemo01(t *testing.T) {
	// 获取类型
	u := struct {
		name string
	}{}

	typeOf := reflect.TypeOf(u)
	fmt.Println(typeOf)
	fmt.Println("Name:", typeOf.Name())
	fmt.Println("Kind:", typeOf.Kind()) // 获取变量的底层类型
	fmt.Println()

	/* 指针变量 */
	user := &ReflectUser01{
		Gender: 1,
	}

	typeOf = reflect.TypeOf(user)
	fmt.Println("Name:", typeOf.Name())
	fmt.Println("Kind:", typeOf.Kind())

	elem := typeOf.Elem() // 获取指针变量的值
	fmt.Println("Name:", elem.Name())
	fmt.Println("Kind:", elem.Kind())
}

func TestReflectTypeOfDemo02(t *testing.T) {
	// 结构体
	u := struct {
		name    string
		Age     int
		Address struct {
			City    string
			Country string
		} `json:"address"`
		ReflectUser01
	}{}

	/* 获取反射对象信息 */
	typeOf := reflect.TypeOf(u)
	fmt.Println("NumField:", typeOf.NumField())
	fmt.Println("Field 0:", typeOf.Field(0))
	if structField, ok := typeOf.FieldByName("name"); ok {
		fmt.Println("FieldByName:", structField)
	}
	// 获取第2个结构体的第0个元素
	fmt.Println("FieldByIndex:", typeOf.FieldByIndex([]int{2, 0}))
	println()

	/* StructField 字段对象内容 */
	structField := typeOf.Field(2)
	fmt.Println("Name:", structField.Name)
	// 字段的可访问的包名
	// 大写字母打头的公共字段, 都可以访问, 故此值为空
	fmt.Println("PkgPath:", structField.PkgPath)
	fmt.Println("Type:", structField.Type)
	fmt.Println("Tag:", structField.Tag)
	fmt.Println("Index:", structField.Index)
	fmt.Println("Anonymous:", structField.Anonymous)
	println()

	/* StructTag 标签内容 */
	tag := structField.Tag
	fmt.Println("Tag Get:", tag.Get("json")) // 若不存在, 返回空字符串
	if value, ok := tag.Lookup("json"); ok {
		fmt.Println("Tag Lookup:", value)
	}
}

func TestReflectTypeOfDemo03(t *testing.T) {
	// 数组
	l := []int{1, 2, 3}

	typeOf := reflect.TypeOf(l)
	// 返回空, 因为数组其实就是一个指向首地址的指针
	fmt.Println("Name:", typeOf.Name())
	// 获取变量的底层类型
	fmt.Println("Kind:", typeOf.Kind())

	// 获取数组元素的类型
	elem := typeOf.Elem()
	fmt.Println("Elem Name:", elem.Name())
	fmt.Println("Elem Kind:", elem.Kind())
}

func TestReflectTypeOfDemo04(t *testing.T) {
	// map
	m := map[string]int{
		"one": 1,
	}

	typeOf := reflect.TypeOf(m)
	fmt.Println("Name:", typeOf.Name()) // 指针, 返回空
	fmt.Println("Kind:", typeOf.Kind())
	// 获取 map key 的类型
	fmt.Println("Key Kind:", typeOf.Key().Kind())
	// 获取 map value 的类型
	fmt.Println("Value Kind:", typeOf.Elem().Kind())
}

//
// 获取属性值 ValueOf
//

func TestReflectValueOfDemo01(t *testing.T) {
	// 基础类型
	a := int64(3)
	valueOf := reflect.ValueOf(a)
	fmt.Println("Value:", valueOf.Int())
}

func TestReflectValueOfDemo02(t *testing.T) {
	// 结构体
	u := struct {
		Name string
		Age  int
	}{"xiao ming", 20}

	valueOf := reflect.ValueOf(u)
	fmt.Println("Value:", valueOf.Field(0).String())
}

func TestReflectValueOfDemo03(t *testing.T) {
	// 数组
	l := []int{1, 2, 3}
	valueOf := reflect.ValueOf(l)
	fmt.Println("Len:", valueOf.Len())
	fmt.Println("Index 0 Value:", valueOf.Index(0))
}

func TestReflectValueOfDemo04(t *testing.T) {
	// map
	m := map[string]string{
		"one": "1",
		"two": "2",
	}

	valueOf := reflect.ValueOf(m)
	fmt.Println("Len:", valueOf.Len())
	fmt.Println("MapIndex:", valueOf.MapIndex(reflect.ValueOf("one")))
	fmt.Println("MapIndex:", valueOf.MapIndex(reflect.ValueOf("three")))
	fmt.Println("MapKeys:", valueOf.MapKeys())
	fmt.Println()

	mapIter := valueOf.MapRange()
	for {
		if !mapIter.Next() {
			break
		}
		fmt.Println("Key:", mapIter.Key())
		fmt.Println("Value:", mapIter.Value())
	}
}

//
// 属性赋值 Set
//
// 这里注意, 只有指针类型的的变量才能被赋值. 其实很好理解, 值类型在方法调用时是通过复制传值的, 只有传递指针才能够找到原始值的内存地址进行修改.
// 因此, 我们在赋值之前, 要调用Kind方法对其类型进行判断.
//

func TestReflectSetDemo01(t *testing.T) {
	// 基础类型
	a := int64(3)
	valueOf := reflect.ValueOf(&a)
	fmt.Println("CanSet:", valueOf.CanSet())
	valueOf.Elem().SetInt(19)
	fmt.Println("Value:", a)
}

func TestReflectSetDemo02(t *testing.T) {
	// 结构体
	// 注意：结构体只有公共字段才能够通过反射进行赋值, 若赋值给一个私有字段, 会抛出异常
	u := struct {
		Name string
		Age  int
	}{"xiao ming", 20}

	valueOf := reflect.ValueOf(&u)
	valueOf.Elem().Field(0).SetString("foo")
	fmt.Println(u)
}

func TestReflectSetDemo03(t *testing.T) {
	// 数组
	l := []int{1, 2, 3}

	valueOf := reflect.ValueOf(&l)
	setValueOf := reflect.ValueOf([]int{4, 5})
	valueOf.Elem().Set(setValueOf)
	fmt.Println("Value:", l)

	valueOf.Elem().Index(0).SetInt(9)
	fmt.Println("Value:", l)
}

func TestReflectSetDemo04(t *testing.T) {
	// map
	m := map[string]string{
		"one": "1",
	}
	valueOf := reflect.ValueOf(&m)
	valueOf.Elem().SetMapIndex(reflect.ValueOf("two"), reflect.ValueOf("2"))
	fmt.Println("Value:", m)
}

//
// Call
//

func add(a, b int) int {
	return a + b
}

func TestReflectCallDemo01(t *testing.T) {
	// 普通方法
	valueOf := reflect.ValueOf(add)
	paramList := []reflect.Value{reflect.ValueOf(2), reflect.ValueOf(3)}
	retList := valueOf.Call(paramList)
	fmt.Println("Call Results:", retList[0].Int())
}

type ReflectUser02 struct {
	Name string
}

func (u ReflectUser02) GetName() string {
	return u.Name
}

func (u *ReflectUser02) SetName(name string) {
	u.Name = name
}

func TestReflectCallDemo02(t *testing.T) {
	// 结构体方法
	u := ReflectUser02{}
	typeOf := reflect.TypeOf(&u)
	fmt.Println("NumMethod:", typeOf.NumMethod())
	fmt.Println("Method 0:", typeOf.Method(0))
	if method, ok := typeOf.MethodByName("GetName"); ok {
		fmt.Println("MethodByName:", method)
	}
	fmt.Println()

	/* Method 对象 */
	setNameFunc, _ := typeOf.MethodByName("SetName")
	fmt.Println("Name:", setNameFunc.Name)
	fmt.Println("Type:", setNameFunc.Type) // 方法签名
	fmt.Println("Index:", setNameFunc.Index)
	fmt.Println("PkgPath:", setNameFunc.PkgPath)
	fmt.Println()

	/* 调用方法 */
	params := []reflect.Value{reflect.ValueOf(&u), reflect.ValueOf("xiao ming")}
	setNameFunc.Func.Call(params)
	fmt.Println("Value:", u)

	getNameFunc, _ := typeOf.MethodByName("GetName")
	name := getNameFunc.Func.Call([]reflect.Value{reflect.ValueOf(&u)})[0]
	fmt.Println("Call Results:", name)
}

//
// Example
//

func myGetType(arg interface{}) string {
	// 不能处理指针类型
	switch arg.(type) {
	case nil:
		return "nil"
	case int:
		return "int"
	case int64:
		return "int64"
	case string:
		return "string"
	case ReflectUser02:
		return "User"
	default:
		return fmt.Sprintf("unexpected type %T: %v", arg, arg)
	}
}

func TestMyGetType(t *testing.T) {
	fmt.Println("types:")
	p := "foo"
	u := ReflectUser02{Name: "bar"}
	for _, item := range []interface{}{3, int64(101), nil, "hello", u, &p} {
		fmt.Println(myGetType(item))
	}
}

func myToString(obj interface{}) string {
	var objType string
	typeOf := reflect.TypeOf(obj)
	if typeOf.Kind() == reflect.Struct {
		objType = typeOf.Name()
	} else {
		objType = typeOf.Kind().String() // 基础类型
	}

	fmt.Println("handle kind:", objType)
	valueOf := reflect.ValueOf(obj)
	switch objType {
	case "string":
		return valueOf.String()
	case "int":
		intValue := valueOf.Int()
		return strconv.Itoa(int(intValue))
		// return strconv.Itoa(obj.(int))
	case "ptr":
		return valueOf.Elem().FieldByName("Name").String()
	case "ReflectUser02":
		return valueOf.FieldByName("Name").String()
	default:
		return fmt.Sprintln("not support type:", objType)
	}
}

func TestMyToString(t *testing.T) {
	fmt.Println(myToString("hello"))
	fmt.Println(myToString(3))

	u := ReflectUser02{Name: "bar"}
	fmt.Println(myToString(u))
	u = ReflectUser02{Name: "foo"}
	fmt.Println(myToString(&u))
}

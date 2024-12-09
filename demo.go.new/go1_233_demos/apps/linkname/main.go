package main

import (
	"log"

	_ "unsafe"
)

// go:linkname usage:
// go:linkname localname importpath.name
//
// 将本地的变量或方法 (localname) 链接到导入的变量或方法 (importpath.name).
// 由于该指令破坏了类型系统和包的模块化原则, 只有在引入 unsafe 包的前提下才能使用这个指令.
//
// 可以使用 go:linkname 来引用第三方包中私有的变量和方法.
//

func main() {
	log.Println("call linked func")
	Foo()
}

//go:linkname Foo main.myFoo
func Foo()

func myFoo() {
	log.Println("myFoo called")
}

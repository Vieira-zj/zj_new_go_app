# Func Diff

Diff funcs between src and dst go file, and output:

```text
[fnDel]:del
[fnAdd]:add
[fnHello]:same
[fnChange]:change
[fnConditional]:same
```

## Test Data

- `go.mod`

```text
module demo.funcdiff

go 1.16
```

- `src1/main.go`

```golang
package main

import (
    "log"
)

func fnHello(name string) {
    log.Println("hello: " + name)
}

func fnChange() {
    log.Println("func to change")
}

func fnDel() {
    log.Println("func to del")
}

func fnConditional(cond bool) {
    if cond {
        log.Println("cond: true")
    } else {
        log.Println("cond: false")
    }
}

func main() {
    fnHello("foo")
    fnConditional(true)
}
```

- `src2/main.go`

```golang
package main

import (
    "log"
)

func fnHello(name string) {
    log.Println("hello: " + name)
}

func fnAdd() {
    log.Println("func to add")
}

func fnChange() {
    log.Println("func is changed")
}

func fnConditional(cond bool) {
    if cond {
        log.Println("cond: true")
    } else {
        log.Println("cond: false")
    }
}

func main() {
    fnHello("foo")
    fnConditional(true)
}
```


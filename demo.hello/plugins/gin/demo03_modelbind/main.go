package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Person .
type Person struct {
	Name       string    `form:"name"`
	Address    string    `form:"address"`
	Birthday   time.Time `form:"birthday" time_format:"2006-01-02" time_utc:"1"`
	CreateTime time.Time `form:"createTime" time_format:"unixNano"`
	UnixTime   time.Time `form:"unixTime" time_format:"unix"`
}

// Student .
type Student struct {
	ID   string `uri:"id" binding:"required,uuid"`
	Name string `uri:"name" binding:"required"`
}

// TestHeader .
type TestHeader struct {
	Rate   int    `header:"Rate"`
	Domain string `header:"Domain"`
}

// Login .
type Login struct {
	User     string `form:"user" json:"user" xml:"user"  binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

// Booking .
type Booking struct {
	CheckIn  time.Time `form:"check_in" binding:"required,bookabledate" time_format:"2006-01-02"`
	CheckOut time.Time `form:"check_out" binding:"required,gtfield=CheckIn" time_format:"2006-01-02"`
}

var bookableDate validator.Func = func(fl validator.FieldLevel) bool {
	if date, ok := fl.Field().Interface().(time.Time); ok {
		today := time.Now()
		if today.After(date) {
			return false
		}
	}
	return true
}

func main() {
	// Model binding and validation
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("bookabledate", bookableDate)
	}

	// Bind Query String or Post Data
	// curl "http://localhost:8081/person?name=foo&address=bar&birthday=1992-03-15&createTime=1562400033000000123&unixTime=1562400033"
	router.Any("/person", func(c *gin.Context) {
		var person Person
		if err := c.ShouldBind(&person); err == nil {
			log.Println("====== Only Bind By Query String ======")
			log.Println("name:", person.Name)
			log.Println("address:", person.Address)
			log.Println("birthday:", person.Birthday)
			log.Println("createTime:", person.CreateTime)
			log.Println("unixTime:", person.UnixTime)
		}
		c.String(http.StatusOK, "Success")
	})

	// Bind Uri
	// curl 'http://localhost:8081/student/foo/987fbc97-4bed-5078-9f07-9141ba07c9f3' | jq .
	// curl 'http://localhost:8081/student/bar/not-uuid' | jq .
	router.GET("/student/:name/:id", func(c *gin.Context) {
		var student Student
		if err := c.ShouldBindUri(&student); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": student.Name, "uuid": student.ID})
	})

	// Bind Header
	// curl 'http://localhost:8081/testing' -H "rate:300" -H "domain:music" | jq .
	router.GET("/testing", func(c *gin.Context) {
		h := TestHeader{}
		if err := c.ShouldBindHeader(&h); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		}
		fmt.Printf("bind headers: %#v\n", h)
		c.JSON(http.StatusOK, gin.H{"Rate": h.Rate, "Domain": h.Domain})
	})

	// Bind JSON
	// curl -XPOST http://localhost:8081/loginJSON -H 'content-type: application/json' \
	//   -d '{"user": "manu", "password": "123"}'
	//   -d '{"user": "manu"}'
	router.POST("/loginJSON", func(c *gin.Context) {
		var login Login
		if err := c.ShouldBindJSON(&login); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if login.User != "manu" || login.Password != "123" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		}
		c.JSON(http.StatusOK, gin.H{"status": "you are logged in"})
	})

	// Custom Validators
	// curl "http://localhost:8081/bookable?check_in=2030-04-16&check_out=2030-04-17" | jq .
	// curl "http://localhost:8081/bookable?check_in=2030-03-10&check_out=2030-03-09" | jq .
	// curl "http://localhost:8081/bookable?check_in=2000-03-09&check_out=2000-03-10" | jq .
	router.GET("/bookable", getBookable)

	router.Run(":8081")
}

func getBookable(c *gin.Context) {
	var b Booking
	if err := c.ShouldBindWith(&b, binding.Query); err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Booking dates are valid"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

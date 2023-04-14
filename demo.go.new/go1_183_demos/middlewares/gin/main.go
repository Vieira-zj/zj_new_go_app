package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
api test:

curl http://localhost:8081/

curl -v -XPOST http://localhost:8081/user -d '{"birthday":"10/07","timezone":"Asia/Shanghai"}'
*/

func main() {
	r := initServer()
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}

// Router

func initServer() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.NoMethod(HandleNotFound)
	r.NoRoute(HandleNotFound)

	r.GET("/", HandlePing)

	// validate middleware should be before CreateUser
	r.POST("/user", ValidateJsonBody[CreateUserHttpBody](), HandleCreateUser)

	return r
}

// Handle

func HandleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func HandlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Pong",
	})
}

// Create User Handle

type CreateUserHttpBody struct {
	Birthday string `json:"birthday" binding:"required,datetime=01/02"`
	Timezone string `json:"timezone" binding:"omitempty,timezone"`
}

func HandleCreateUser(c *gin.Context) {
	httpBody := GetJsonBody[CreateUserHttpBody](c)
	log.Printf("create user: birthday: %s, timezone: %s", httpBody.Birthday, httpBody.Timezone)
	c.JSON(http.StatusOK, gin.H{
		"message": "success created",
	})
}

const keyJsonBody = "jsonBody"

// ValidateJsonBody a middleware to validate request body by generic.
func ValidateJsonBody[BodyType any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body BodyType
		if err := c.ShouldBindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Set(keyJsonBody, body)
		c.Next()
	}
}

func GetJsonBody[BodyType any](c *gin.Context) BodyType {
	return c.MustGet(keyJsonBody).(BodyType)
}

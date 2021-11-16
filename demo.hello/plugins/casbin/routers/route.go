package routers

import (
	"net/http"

	"demo.hello/plugins/casbin/middleware"
	"demo.hello/plugins/casbin/utils"
	"github.com/gin-gonic/gin"
)

// R .
var R *gin.Engine

func init() {
	R = gin.Default()
	R.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "bad Request"})
	})
	api()
}

func api() {
	auth := R.Group("/api")
	{
		// 添加一条 policy 策略
		auth.POST("acs", func(c *gin.Context) {
			utils.C = c
			// mock data
			subject := "tom"
			object := "/api/routers"
			action := "POST"
			cacheName := subject + object + action
			result := utils.Enforcer.AddPolicy(subject, object, action)
			if result {
				if err := utils.GlobalCache.Delete(cacheName); err != nil {
					utils.Error("delete cache error: " + err.Error())
					c.Abort()
					return
				}
				utils.Success("add policy success")
			} else {
				utils.Error("add policy failed")
			}
		})

		// 删除一条 policy 策略
		auth.DELETE("acs/:id", func(c *gin.Context) {
			utils.C = c
			// mock data
			subject := "tom"
			object := "/api/routers"
			action := "POST"
			cacheName := subject + object + action
			result := utils.Enforcer.RemovePolicy(subject, object, action)
			if result {
				if err := utils.GlobalCache.Delete(cacheName); err != nil {
					utils.Error("delete cache error: " + err.Error())
					c.Abort()
					return
				}
				utils.Success("delete policy success")
			} else {
				utils.Error("delete Policy fail")
			}
		})

		// 获取路由列表
		auth.POST("/routers", middleware.Privilege(), func(c *gin.Context) {
			type data struct {
				Method string `json:"method"`
				Path   string `json:"path"`
			}
			var datas []data

			routers := R.Routes()
			for _, r := range routers {
				datas = append(datas, data{Method: r.Method, Path: r.Path})
			}
			utils.C = c
			utils.Success(datas)
		})
	}

	user := R.Group("/api/v1")
	user.Use(middleware.Privilege())
	{
		user.POST("user", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "user Add success"})
		})
		user.GET("user/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "user Get success " + id})
		})
		user.DELETE("user/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "user Delete success " + id})
		})
		user.PUT("user/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "user Update success " + id})
		})
	}
}

package middleware

import (
	"log"

	"demo.hello/plugins/casbin/utils"
	"github.com/gin-gonic/gin"
)

// Privilege .
func Privilege() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.C = c
		userName := c.GetHeader("userName")
		if len(userName) == 0 {
			utils.Error("header miss userName")
			c.Abort()
			return
		}

		path := c.Request.URL.Path
		method := c.Request.Method
		cacheName := userName + path + method
		// 从缓存中读取
		entry, err := utils.GlobalCache.Get(cacheName)
		if err == nil && entry != nil {
			if string(entry) == "true" {
				c.Next()
			} else {
				utils.Error("access denied")
				c.Abort()
			}
			return
		}

		// 从数据库中读取
		if err := utils.Enforcer.LoadPolicy(); err != nil {
			log.Println("loadPolicy error")
			panic(err)
		}
		// 验证策略规则
		result, err := utils.Enforcer.EnforceSafe(userName, path, method)
		if err != nil {
			utils.Error("No permission found")
			c.Abort()
			return
		}

		if !result {
			if err := utils.GlobalCache.Set(cacheName, []byte("false")); err != nil {
				log.Println("set cache error:", err)
			}
			utils.Error("access denied")
			c.Abort()
			return
		}
		if err := utils.GlobalCache.Set(cacheName, []byte("true")); err != nil {
			log.Println("set cache error:", err)
		}
		c.Next()
	}
}

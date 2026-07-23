package auth

import (
	"fmt"
	"net/http"
	"strings"
	"word/common"
	"word/config"
	"word/database"

	"github.com/gin-gonic/gin"
)

func Authorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从前端传来的请求头中获得tokenString
		tokenString := ctx.GetHeader("Authorization")
		// 先检查tokenString格式
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			fmt.Println("token格式不正确, tokenString是：", tokenString)
			ctx.JSON(401, gin.H{"code": 2, "info": "Invalid or expired JWT"})
			ctx.Abort()
			return
		}
		// 去除tokenString中无意义的前7位
		tokenString = tokenString[7:]
		// 解析tokenString，检查是否有效
		token, claim, err := common.ParseToken(tokenString)
		if err != nil || !token.Valid {
			fmt.Println("解析出来的token无效！错误是：", err.Error())
			ctx.JSON(401, gin.H{"code": 2, "info": "Invalid or expired JWT"})
			ctx.Abort()
			return
		}

		// 从claim中获取该token对应的用户id，并查找对应用户
		cur_id := claim.Id
		user_db := database.GetDB()
		var tmp_user config.UserInfo
		user_db.Table("user_infos").First(&tmp_user, cur_id)

		// 如果没查着对应用户
		if tmp_user.UserName == "" {
			fmt.Println("用户不存在")
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "权限不足"})
			ctx.Abort()
			return
		}

		// 用户存在
		ctx.Set("user", tmp_user)

		// 中间件继续
		ctx.Next()
	}
}

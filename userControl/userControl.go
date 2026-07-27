package usercontrol

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"time"
	"word/common"
	"word/config"
	"word/database"
	"word/redis"

	"github.com/gin-gonic/gin"
)

type loginJson struct {
	UserName string `form:"userName"`
	Password string `form:"password"`
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 获得前端传来的json表单，得到用户的名字和密码
		var loginInput loginJson
		ctx.ShouldBind(&loginInput)

		// 读取请求体
		// 读取 body
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			fmt.Printf("读取 body 失败: %v", err)
			return
		}

		// 重新赋值 body，因为读取后 body 为空了
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		fmt.Printf("Body: %s\n", string(bodyBytes))

		// 先检查用户名和密码是否有效
		if loginInput.Password == "" || loginInput.UserName == "" {
			if loginInput.Password == "" {
				fmt.Println("401错误，错误是：", "密码为空")
			}
			if loginInput.UserName == "" {
				fmt.Println("401错误，错误是：", "用户名为空")
			}

			ctx.JSON(401, gin.H{"code": 2, "info": "empty password or username"})
			return
		}

		// 确认输入有效，检查这个用户是不是已经注册过
		user_db := database.GetDB()
		var tmp_user config.UserInfo
		user_db.Table("user_infos").Where("user_name = ?", loginInput.UserName).First(&tmp_user)

		// 如果用户还没有注册
		if tmp_user.UserName == "" {
			new_user := config.UserInfo{
				UserName: loginInput.UserName,
				Password: loginInput.Password,
			}
			user_db.Table("user_infos").Create(&new_user)
			token, err := common.ReleaseToken(int(new_user.UserID))
			if err != nil {
				ctx.JSON(500, gin.H{"code": -4, "info": "token发放失败"})
				return
			}

			// 将redis内容也进行初始化
			new_user_id := new_user.UserID
			// 构建不同的几个redis key
			score_key := "score:" + strconv.Itoa(int(new_user_id))
			progress_key := "progress:" + strconv.Itoa(int(new_user_id))
			// 对于几个redis进行初始化
			rctx := context.Background()
			rdb := redis.GetRedis()
			rdb.Set(rctx, score_key, 0, time.Second*3600*24*3650) // 分数初始化为0
			rdb.HSet(rctx, progress_key, "cee", 0, "cet4", 0)     // 两张表的学习进度初始化为0

			ctx.JSON(200, gin.H{"code": 0, "info": "Succeed", "token": token})
			return
		}

		// 用户已经注册了
		// 如果密码错误
		if tmp_user.Password != loginInput.Password {
			ctx.JSON(401, gin.H{"code": 2, "info": "Wrong password"})
			return
		}

		// 密码正确，可以登录
		token, err := common.ReleaseToken(int(tmp_user.UserID))
		if err != nil {
			ctx.JSON(500, gin.H{"code": -4, "info": "token发放失败"})
			return
		}
		ctx.JSON(200, gin.H{"code": 0, "info": "Succeed", "token": token})
	}
}

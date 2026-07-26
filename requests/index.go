package requests

import (
	"context"
	"strconv"
	"word/config"
	"word/redis"

	"github.com/gin-gonic/gin"
)

func GetIndex() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 初始化rdb
		rdb := redis.GetRedis()
		rctx := context.Background()

		// 获取当前用户
		user, _ := ctx.Get("user")
		userID := user.(config.UserInfo).UserID

		// 构建redis key
		progress_key := "progress:" + strconv.Itoa(int(userID))
		cee_progress_str := rdb.HGet(rctx, progress_key, "cee").Val()
		cet4_progress_str := rdb.HGet(rctx, progress_key, "cet4").Val()
		cee_progress_int, _ := strconv.Atoi(cee_progress_str)
		cet4_progress_int, _ := strconv.Atoi(cet4_progress_str)

		// 返回信息
		ctx.JSON(200, gin.H{
			"code":          0,
			"info":          "获取用户进度成功",
			"cee_progress":  cee_progress_int,
			"cee_total":     3933,
			"cet4_progress": cet4_progress_int,
			"cet4_total":    2674,
		})

	}
}

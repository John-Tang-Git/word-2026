package requests

import (
	"context"
	"strconv"
	"time"
	"word/config"
	"word/redis"

	"github.com/gin-gonic/gin"
)

// 返回内容：用户名、积分、连续签到天数、两张表的学习进度
func GetInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取rdb
		rdb := redis.GetRedis()
		rctx := context.Background()

		// 获取当前用户
		user, _ := ctx.Get("user")
		userID := user.(config.UserInfo).UserID
		userName := user.(config.UserInfo).UserName

		// 获取积分
		score_key := "score:" + strconv.Itoa(int(userID))
		userScore := rdb.Get(rctx, score_key).Val()

		// 获取连续签到天数
		cur_date := time.Now().Format("20060102")
		ym := cur_date[:6]
		day_str := cur_date[6:]
		day_int, _ := strconv.Atoi(day_str)
		constant_days := ConstantSignedDays(int(userID), day_int, ym)

		// 获取学习进度
		progress_key := "progress:" + strconv.Itoa(int(userID))
		cee_progress := rdb.HGet(rctx, progress_key, "cee").Val()
		cet4_progress := rdb.HGet(rctx, progress_key, "cet4").Val()

		fmt.Printf("[GetInfo] userID=%d, progress_key=%s, cee=%q, cet4=%q\n",
			userID, progress_key, cee_progress, cet4_progress)

		// 统一转换为数字格式
		user_score_return, _ := strconv.Atoi(userScore)
		cee_progress_return, _ := strconv.Atoi(cee_progress)
		cet4_progress_return, _ := strconv.Atoi(cet4_progress)

		// 统一返回
		ctx.JSON(200, gin.H{
			"code":     0,
			"info":     "获取该用户信息成功",
			"用户名":      userName,
			"积分":       user_score_return,
			"连续签到天数":   constant_days,
			"高考词汇学习进度": cee_progress_return,
			"四级词汇学习进度": cet4_progress_return,
		})
	}
}

package requests

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"word/config"
	"word/database"
	"word/redis"

	"github.com/gin-gonic/gin"
)

// 无论是主页进入，还是在学习中点击“下一个”，都是同一个请求
func GetStudy() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 获取rdb
		rdb := redis.GetRedis()
		rctx := context.Background()

		// 获取当前用户
		user, _ := ctx.Get("user")
		userID := user.(config.UserInfo).UserID

		// 获取当前是哪一套单词表
		cur_alphabet := ctx.Param("alphabet")

		// 先获取此前有哪些不会的单词，复习全部
		unknown_key := "unknown:" + strconv.Itoa(int(userID)) + ":" + cur_alphabet
		unknown_len := rdb.SCard(rctx, unknown_key).Val()
		fmt.Println("unknown_key:", unknown_key)
		fmt.Println("目前还不会的词：", unknown_len)

		// 如果仍有不会的词，直接把这个词取出来
		not_review_key := "no_review:" + strconv.Itoa(int(userID)) + time.Now().Format("20060102")
		if vals := rdb.SDiff(rctx, unknown_key, not_review_key).Val(); len(vals) != 0 {
			word_id := vals[0]
			rdb.SRem(rctx, unknown_key, word_id)
			// 根据val（wordID）从redis中取出对应的词汇，注意，不能是当天刚刚不会的单词！
			word_key := "alphabet:" + cur_alphabet + ":" + word_id
			unknown_english := rdb.HGet(rctx, word_key, "english").Val()
			unknown_chinese := rdb.HGet(rctx, word_key, "chinese").Val()
			fmt.Println("存在不会的词！")
			// 同步操作：如果这个word_id对应的unknown记录，在mysql中也存过，这里就一并删除了
			db := database.GetDB()
			result := db.Table("user_unknowns").
				Where("user_id = ? AND word_id = ?", userID, word_id).
				Delete(&config.UserUnknown{})
			if result.Error != nil {
				fmt.Printf("删除失败: %v", result.Error)
			}

			ctx.JSON(200, gin.H{
				"code":     0,
				"info":     "复习词汇",
				"word_id":  word_id,
				"english":  unknown_english,
				"chinese":  unknown_chinese,
				"isReview": true,
			})
			return
		}

		// 如果全部单词都复习完，直接获取当前用户当前单词表的学习进度
		fmt.Println("目前没有需要复习的单词！")
		progress_key := "progress:" + strconv.Itoa(int(userID))
		cur_progress := rdb.HGet(rctx, progress_key, cur_alphabet).Val()
		cur_progress_int := 0
		if cur_progress != "" {
			cur_progress_int, _ = strconv.Atoi(cur_progress)
		}
		cur_progress_int++ //肯定要比当前进度多一个

		// 根据学习进度，给出下一个需要学习的单词的英文
		word_key := "alphabet:" + cur_alphabet + ":" + strconv.Itoa(cur_progress_int)
		fmt.Println("当前的redis_key:", word_key)
		cur_english := rdb.HGet(rctx, word_key, "english").Val()
		cur_chinese := rdb.HGet(rctx, word_key, "chinese").Val()
		ctx.JSON(200, gin.H{
			"code":     0,
			"info":     "新词汇",
			"word_id":  cur_progress_int,
			"english":  cur_english,
			"chinese":  cur_chinese,
			"isReview": false,
		})
	}
}

// 处理单词
type WordInput struct {
	WordID     uint `json:"WordID"`
	Understand bool `json:"Understand"`
}

func PostStudy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取rdb
		rdb := redis.GetRedis()
		rctx := context.Background()

		// 获取当前用户
		user, _ := ctx.Get("user")
		userID := user.(config.UserInfo).UserID

		// 先解析前端传来的信息：wordID，会还是不会
		var wordInput WordInput
		ctx.ShouldBindJSON(&wordInput)

		// 解析，当前是什么单词表
		cur_alphabet := ctx.Param("alphabet")

		// 如果这个单词不会，就需要存入unknown队列了
		if !wordInput.Understand {
			unknown_key := "unknown:" + strconv.Itoa(int(userID)) + ":" + cur_alphabet
			rdb.SAdd(rctx, unknown_key, wordInput.WordID)

			// 注意！每一天新不会的单词，不要当天复习
			not_review_key := "no_review:" + strconv.Itoa(int(userID)) + time.Now().Format("20060102")
			rdb.SAdd(rctx, not_review_key, wordInput.WordID)
			rdb.Expire(rctx, not_review_key, time.Second*3600*24) // 这个不复习的单词表是临时的，24小时后过期
		}

		// 无论会不会，都不影响记录进度, 而且只有在当前学习的word_id比redis中记录的进度大时，才需要更新进度
		progress_key := "progress:" + strconv.Itoa(int(userID))
		progress_in_redis := rdb.HGet(rctx, progress_key, cur_alphabet).Val()
		progress_in_redis_int, _ := strconv.Atoi(progress_in_redis)
		if int(wordInput.WordID) > progress_in_redis_int {
			rdb.HSet(rctx, progress_key, cur_alphabet, wordInput.WordID)
		}

		// 获取该词对应的中英文
		word_key := "alphabet:" + cur_alphabet + ":" + strconv.Itoa(int(wordInput.WordID))
		fmt.Println("当前的词语redis key:", word_key)
		english := rdb.HGet(rctx, word_key, "english").Val()
		chinese := rdb.HGet(rctx, word_key, "chinese").Val()

		// 返回时，连带该词的中文解释一并返回
		ctx.JSON(200, gin.H{
			"code":     0,
			"info":     "已记录进度",
			"progress": wordInput.WordID,
			"english":  english,
			"chinese":  chinese,
		})
	}
}

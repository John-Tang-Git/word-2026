package redis

import (
	"context"
	"strconv"
	"word/config"
	"word/database"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

// 初次建立redis时，一定要判断是否把单词表导入redis了
func InitAlphabetRedis() {
	rctx := context.Background()
	db := database.GetDB()

	// 高考单词表
	// 先抽取单词2000号，如果有，说明存储过
	cee_word_key := "alphabet:" + "cee" + ":2000"
	exists, _ := rdb.HExists(rctx, cee_word_key, "english").Result()
	if !exists { // 没存储过，就得从mysql里面拿了
		// 先获取高考词汇
		var cee_words []config.CeeInfo
		db.Find(&cee_words)
		// 再把这些单词存入redis
		for _, cee_word := range cee_words {
			cur_cee_word_key := "alphabet:" + "cee" + ":" + strconv.Itoa(int(cee_word.WordID))
			rdb.HSet(rctx, cur_cee_word_key, "english", cee_word.English)
			rdb.HSet(rctx, cur_cee_word_key, "chinese", cee_word.Chinese)
		}
	}

	// 四级单词表
	// 先抽取单词2000号，如果有，说明存储过
	cet4_word_key := "alphabet:" + "cet4" + ":2000"
	exists, _ = rdb.HExists(rctx, cet4_word_key, "english").Result()
	if !exists { // 没存储过，就得从mysql里面拿了
		// 先获取高考词汇
		var cet4_words []config.CetFourInfo
		db.Find(&cet4_words)
		// 再把这些单词存入redis
		for _, cet4_word := range cet4_words {
			cur_cet4_word_key := "alphabet:" + "cet4" + ":" + strconv.Itoa(int(cet4_word.WordID))
			rdb.HSet(rctx, cur_cet4_word_key, "english", cet4_word.English)
			rdb.HSet(rctx, cur_cet4_word_key, "chinese", cet4_word.Chinese)
		}
	}
}

func InitRedis() *redis.Client {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	InitAlphabetRedis()

	return rdb
}

func GetRedis() *redis.Client {
	return rdb
}

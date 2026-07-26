package datasync

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"word/config"
	"word/database"
	"word/redis"
	"word/requests"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

var wg sync.WaitGroup

// 将签到信息存入mysql
func SignToMysql() {

	defer wg.Done()

	// 初始化mysql和redis
	db := database.GetDB()
	rctx := context.Background()
	rdb := redis.GetRedis()

	// 从redis中取出相应的所有key
	iter := rdb.Scan(rctx, 0, "sign:*", 0).Iterator()

	for iter.Next(rctx) {
		key := iter.Val()
		// 从key中解析userID与年月
		key_infos := strings.Split(key, ":")
		userID_str := key_infos[1]
		ym := key_infos[2]
		userID, _ := strconv.Atoi(userID_str)
		// 获取当月有多少天
		day_str := requests.HowManyDays(ym)
		day_int, _ := strconv.Atoi(day_str)
		// 获取签到记录
		result, _ := rdb.BitField(rctx, key, "GET", "u"+day_str, 0).Result()
		bits := result[0]
		bitsStr := fmt.Sprintf("%0*b", day_int, bits) // 等到当月签到0101格式字符串
		// 先判断数据库里有没有这一条
		var existing config.UserSign
		db_result := db.Table("user_signs").Where("`user_id`=? AND `year_month`=?", userID, ym).First(&existing)
		if db_result.Error == gorm.ErrRecordNotFound {
			// 如果这一个月的存储记录还没存入
			fmt.Println("当前这条签到记录还没存入！key:", key)
			tmp_sign := config.UserSign{
				UserID:     uint(userID),
				YearMonth:  ym,
				SignedBits: bitsStr,
			}
			db.Table("user_signs").Create(&tmp_sign)
		} else {
			// 这一个月的存储记录已经存入，更新即可
			fmt.Println("当前这条签到记录已存在，只需更新！key:", key)
			db.Model(&existing).Updates(map[string]interface{}{
				"signed_bits": bitsStr,
			})
		}
	}
}

// 存储用户积分
func ScoreToMysql() {
	defer wg.Done()

	// 初始化mysql和redis
	db := database.GetDB()
	rctx := context.Background()
	rdb := redis.GetRedis()

	// 从redis中取出相应的所有key
	iter := rdb.Scan(rctx, 0, "score:*", 0).Iterator()

	for iter.Next(rctx) {
		key := iter.Val()
		key_infos := strings.Split(key, ":")
		userID_str := key_infos[1]
		userID, _ := strconv.Atoi(userID_str)
		// 从redis中取得相应用户的积分
		cur_score := rdb.Get(rctx, key).Val()
		cur_score_int, _ := strconv.Atoi(cur_score)
		// 检查是不是在数据库中
		var existing config.UserScore
		db_result := db.Table("user_scores").Where("`user_id`=?", userID).First(&existing)
		if db_result.Error == gorm.ErrRecordNotFound {
			// 没存储过
			tmp_score := config.UserScore{
				UserID: uint(userID),
				Score:  uint(cur_score_int),
			}
			db.Table("user_scores").Create(&tmp_score)
		} else {
			// 存储过，更新
			db.Model(&existing).Updates(map[string]interface{}{
				"score": uint(cur_score_int),
			})
		}
	}
}

// 存储用户进度
func ProgressToMysql() {

	defer wg.Done()

	// 初始化mysql和redis
	db := database.GetDB()
	rctx := context.Background()
	rdb := redis.GetRedis()

	// 从redis中取出相应的所有key
	iter := rdb.Scan(rctx, 0, "progress:*", 0).Iterator()

	for iter.Next(rctx) {
		key := iter.Val()
		key_infos := strings.Split(key, ":")
		userID_str := key_infos[1]
		userID, _ := strconv.Atoi(userID_str)
		// 从redis中取得相应用户的进度
		cee_progress_str := rdb.HGet(rctx, key, "cee").Val()
		cet4_progress_str := rdb.HGet(rctx, key, "cet4").Val()
		cee_progress_int, _ := strconv.Atoi(cee_progress_str)
		cet4_progress_int, _ := strconv.Atoi(cet4_progress_str)
		// 检查是不是在数据库中
		var existing config.UserProgress
		db_result := db.Table("user_progresses").Where("`user_id`=?", userID).First(&existing)
		if db_result.Error == gorm.ErrRecordNotFound {
			// 没存储过
			tmp_progress := config.UserProgress{
				UserID:  uint(userID),
				Cee:     uint(cee_progress_int),
				CetFour: uint(cet4_progress_int),
			}
			db.Table("user_progresses").Create(&tmp_progress)
		} else {
			// 存储过，更新
			db.Model(&existing).Updates(map[string]interface{}{
				"cee":      uint(cee_progress_int),
				"cet_four": uint(cet4_progress_int),
			})
		}
	}
}

// 将用户不会的词一条一条存入mysql中
func UnknownToMysql() {

	defer wg.Done()

	// 初始化mysql和redis
	db := database.GetDB()
	rctx := context.Background()
	rdb := redis.GetRedis()

	// 从redis中取出相应的所有key
	iter := rdb.Scan(rctx, 0, "unknown:*", 0).Iterator()

	for iter.Next(rctx) {
		key := iter.Val()
		key_infos := strings.Split(key, ":")
		// userID
		userID_str := key_infos[1]
		userID, _ := strconv.Atoi(userID_str)
		// 哪一张单词表
		cur_alphabet := key_infos[2]
		// 从redis中取得相应用户相应单词表不会的词
		unknown_word_ids := rdb.SMembers(rctx, key).Val()

		// 每一个词分别检查
		for _, word_id := range unknown_word_ids {
			// 检查一下有没有存储过这条不会的记录
			var exists config.UserUnknown
			result := db.Table("user_unknowns").Where("`user_id`=? AND `word_id`=? AND `alphabet`=?", userID, word_id, cur_alphabet).First(&exists)
			if result.Error == gorm.ErrRecordNotFound {
				// 确实没存过，就存一下
				word_id_int, _ := strconv.Atoi(word_id)
				tmp_unknown := config.UserUnknown{
					UserID:   uint(userID),
					WordID:   uint(word_id_int),
					Alphabet: cur_alphabet,
				}
				db.Table("user_unknowns").Create(&tmp_unknown)
			}
		}
	}
}

func RedisToMysql() {
	// 在退出函数前，不要因错误中断函数
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			fmt.Println(string(debug.Stack()))
		}
	}()

	wg.Add(4)

	go SignToMysql()
	go ScoreToMysql()
	go ProgressToMysql()
	go UnknownToMysql()

	wg.Wait()
}

func StartCronScheduler() {

	crn := cron.New()

	fmt.Println("Cron scheduler initialized, waiting for first execution...")

	_, err := crn.AddFunc("*/15 * * * *", func() {
		RedisToMysql()
		fmt.Println("Cron: Starting scheduled sync...")
	})

	if err != nil {
		fmt.Println("定时存储mysql错误，错误是", err)
	}

	crn.Start()
	fmt.Println("Cron scheduler started")
	// 保持主程序运行
}

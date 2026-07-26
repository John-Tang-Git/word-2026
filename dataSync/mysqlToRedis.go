package datasync

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"
	"word/config"
	"word/database"
	"word/redis"
)

var WG sync.WaitGroup

// 将mysql中存储的用户签到信息存入redis
func SignToRedis() {
	defer WG.Done()

	// 初始化
	db := database.GetDB()
	rdb := redis.GetRedis()
	rctx := context.Background()

	// 从mysql中取得所有用户签到信息
	var all_signs []config.UserSign
	db.Table("user_signs").Find(&all_signs)

	// 把这些用户签到信息分别存入redis
	for _, cur_sign := range all_signs {
		// 先获取信息
		cur_user_id := cur_sign.UserID
		cur_year_month := cur_sign.YearMonth
		cur_bits_str := cur_sign.SignedBits
		// 获取当月有多少天
		day_int := len(cur_bits_str)
		day_str := strconv.Itoa(day_int)
		// 将bits转化为int64整数
		cur_bits, _ := strconv.ParseInt(cur_bits_str, 2, day_int)
		// 构造redis key
		cur_sign_key := "sign:" + strconv.Itoa(int(cur_user_id)) + ":" + cur_year_month
		// 检查rdb中已经存储的当月签到记录，和mysql中的数据进行对比
		bits_in_redis := rdb.BitField(rctx, cur_sign_key, "GET", "u"+day_str, 0).Val()[0]
		// 假设mysql中存储的签到记录天数更多，就更新redis
		if bits_in_redis < cur_bits {
			rdb.BitField(rctx, cur_sign_key, "SET", "u"+day_str, 0, int64(cur_bits))
		}
	}
}

// 把积分信息存入redis
func ScoreToRedis() {
	defer WG.Done()

	// 初始化
	db := database.GetDB()
	rdb := redis.GetRedis()
	rctx := context.Background()

	// 从mysql中取得所有用户积分信息
	var all_scores []config.UserScore
	db.Table("user_scores").Find(&all_scores)

	// 把这些用户签到信息分别存入redis
	for _, cur_score := range all_scores {
		// 先获取信息
		cur_user_id := cur_score.UserID
		cur_score := cur_score.Score
		// 构造redis key
		cur_score_key := "score:" + strconv.Itoa(int(cur_user_id))
		// 检查redis和mysql中存储的积分哪个大
		score_in_redis := rdb.Get(rctx, cur_score_key).Val()
		score_in_redis_int, _ := strconv.Atoi(score_in_redis)
		// 只有在mysql存储的积分值更大时，才刷新redis
		if score_in_redis_int < int(cur_score) {
			rdb.Set(rctx, cur_score_key, strconv.Itoa(int(cur_score)), 0)
		}
	}
}

func ProgressToRedis() {
	defer WG.Done()

	// 初始化
	db := database.GetDB()
	rdb := redis.GetRedis()
	rctx := context.Background()

	// 从mysql中取得所有用户积分信息
	var all_progresses []config.UserProgress
	db.Table("user_progresses").Find(&all_progresses)

	// 把这些用户签到信息分别存入redis
	for _, cur_progress := range all_progresses {
		// 先获取信息
		cur_user_id := cur_progress.UserID
		cur_cee_progress := cur_progress.Cee
		cur_cet4_progress := cur_progress.CetFour
		// fmt.Printf("用户id：%d, MySQL中高考进度%d, 四级进度%d", cur_user_id, cur_cee_progress, cur_cet4_progress)
		// 构造redis key
		cur_progress_key := "progress:" + strconv.Itoa(int(cur_user_id))
		// 检查mysql和redis中哪个进度更靠前
		progress_cee_in_redis := rdb.HGet(rctx, cur_progress_key, "cee").Val()
		progress_cet4_in_redis := rdb.HGet(rctx, cur_progress_key, "cet4").Val()
		progress_cee_in_redis_int, _ := strconv.Atoi(progress_cee_in_redis)
		progress_cet4_in_redis_int, _ := strconv.Atoi(progress_cet4_in_redis)
		// 先检查cee进度谁更靠前
		if progress_cee_in_redis_int < int(cur_cee_progress) {
			rdb.HSet(rctx, cur_progress_key, "cee", cur_cee_progress)
		}
		// 再检查cet4进度谁更靠前
		//fmt.Printf("在这里检查用户%d的四级进度是否同步\n", cur_user_id)
		// fmt.Printf("redis中存储的四级进度是%d, mysql中存的四级进度是%d\n", progress_cet4_in_redis_int, cur_cet4_progress)
		if progress_cet4_in_redis_int < int(cur_cet4_progress) {

			rdb.HSet(rctx, cur_progress_key, "cet4", cur_cet4_progress)
		}

	}
}

func UnknownToRedis() {
	defer WG.Done()

	// 初始化
	db := database.GetDB()
	rdb := redis.GetRedis()
	rctx := context.Background()

	// 从mysql中取得所有用户积分信息
	var all_unknowns []config.UserUnknown
	db.Table("user_unknowns").Find(&all_unknowns)

	// 把这些用户签到信息分别存入redis
	for _, cur_unknown := range all_unknowns {
		// 先获取信息
		cur_user_id := cur_unknown.UserID
		cur_alphabet := cur_unknown.Alphabet
		cur_word_id := cur_unknown.WordID
		// 构造redis key
		cur_unknown_key := "unknown:" + strconv.Itoa(int(cur_user_id)) + ":" + cur_alphabet
		rdb.SAdd(rctx, cur_unknown_key, cur_word_id)
	}
}

func MysqlToRedis() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			fmt.Println(string(debug.Stack()))
		}
	}()

	WG.Add(4)

	go SignToRedis()
	go ScoreToRedis()
	go ProgressToRedis()
	go UnknownToRedis()

	WG.Wait()
}

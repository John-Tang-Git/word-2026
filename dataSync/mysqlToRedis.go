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
		bits, _ := strconv.ParseInt(cur_bits_str, 2, day_int)
		// 构造redis key
		cur_sign_key := "sign:" + strconv.Itoa(int(cur_user_id)) + ":" + cur_year_month
		rdb.BitField(rctx, cur_sign_key, "SET", "u"+day_str, 0, int64(bits))
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
		rdb.Set(rctx, cur_score_key, strconv.Itoa(int(cur_score)), 0)
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
	db.Table("user_scores").Find(&all_progresses)

	// 把这些用户签到信息分别存入redis
	for _, cur_progress := range all_progresses {
		// 先获取信息
		cur_user_id := cur_progress.UserID
		cur_cee_progress := cur_progress.Cee
		cur_cet4_progress := cur_progress.CetFour
		// 构造redis key
		cur_progress_key := "progress:" + strconv.Itoa(int(cur_user_id))
		rdb.HSet(rctx, cur_progress_key, "cee", cur_cee_progress, "cet4", cur_cet4_progress)
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
	db.Table("user_scores").Find(&all_unknowns)

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

package database

import (
	"context"
	"fmt"
	"time"
	"tournament/config"

	"github.com/go-redis/redis/v8"
)

const (
	tournamentStatusKey = "tournament:%s:status"
	participantScoreKey = "tournament:%s:participant:%s:score"
	leaderboardKey      = "tournament:%s:leaderboard"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	return client
}

// Функции для работы с турнирами в Redis
func SetTournamentStatus(client *redis.Client, tournamentID string, status string) error {
	ctx := context.Background()
	key := fmt.Sprintf(tournamentStatusKey, tournamentID)
	return client.Set(ctx, key, status, 0).Err()
}

func GetTournamentStatus(client *redis.Client, tournamentID string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf(tournamentStatusKey, tournamentID)
	return client.Get(ctx, key).Result()
}

func SetParticipantScore(client *redis.Client, tournamentID string, userID string, score int) error {
	ctx := context.Background()
	key := fmt.Sprintf(participantScoreKey, tournamentID, userID)
	leaderboardKey := fmt.Sprintf(leaderboardKey, tournamentID)

	pipe := client.Pipeline()
	pipe.Set(ctx, key, score, 0)
	pipe.ZAdd(ctx, leaderboardKey, &redis.Z{
		Score:  float64(score),
		Member: userID,
	})
	_, err := pipe.Exec(ctx)
	return err
}

func GetParticipantScore(client *redis.Client, tournamentID string, userID string) (int, error) {
	ctx := context.Background()
	key := fmt.Sprintf(participantScoreKey, tournamentID, userID)
	score, err := client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return score, err
}

func GetLeaderboard(client *redis.Client, tournamentID string, start, stop int64) ([]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf(leaderboardKey, tournamentID)
	return client.ZRevRange(ctx, key, start, stop).Result()
}

func GetParticipantRank(client *redis.Client, tournamentID string, userID string) (int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf(leaderboardKey, tournamentID)
	return client.ZRevRank(ctx, key, userID).Result()
}

func ClearTournamentData(client *redis.Client, tournamentID string) error {
	ctx := context.Background()
	keys := []string{
		fmt.Sprintf(tournamentStatusKey, tournamentID),
		fmt.Sprintf(leaderboardKey, tournamentID),
	}

	// Получаем все ключи участников
	pattern := fmt.Sprintf(participantScoreKey, tournamentID, "*")
	iter := client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return client.Del(ctx, keys...).Err()
	}
	return nil
}

func GetTopParticipants(client *redis.Client, tournamentID string, count int64) ([]redis.Z, error) {
	ctx := context.Background()
	return client.ZRevRangeWithScores(ctx, "tournament:"+tournamentID+":scores", 0, count-1).Result()
}

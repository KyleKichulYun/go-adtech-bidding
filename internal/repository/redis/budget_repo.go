// internal/repository/redis/budget_repo.go
package redis

import (
	"context"
	"fmt"
	"go-adtech-bidding/scripts/lua" // 방금 생성한 lua 패키지 임포트

	"github.com/redis/go-redis/v9"
)

type BudgetRepository struct {
	client *redis.Client
	script *redis.Script
}

func NewBudgetRepository(client *redis.Client) *BudgetRepository {
	return &BudgetRepository{
		client: client,
		script: redis.NewScript(lua.DeductBudget), // embed된 변수 직접 사용
	}
}

func (r *BudgetRepository) DeductBudget(ctx context.Context, campaignID string, bidAmount int64) (bool, error) {
	key := fmt.Sprintf("campaign:budget:%s", campaignID)

	result, err := r.script.Run(ctx, r.client, []string{key}, bidAmount).Int()
	if err != nil {
		return false, fmt.Errorf("failed to execute lua script: %w", err)
	}

	return result == 1, nil
}

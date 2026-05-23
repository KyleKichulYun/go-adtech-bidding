// internal/repository/redis/budget_repo.go
package redis

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/redis/go-redis/v9"
)

//go:embed ../../../scripts/lua/deduct_budget.lua
var deductBudgetLua string

// BudgetRepository는 캠페인 예산 관리를 담당합니다.
type BudgetRepository struct {
	client *redis.Client
	script *redis.Script
}

// NewBudgetRepository는 생성자 함수입니다.
func NewBudgetRepository(client *redis.Client) *BudgetRepository {
	return &BudgetRepository{
		client: client,
		script: redis.NewScript(deductBudgetLua), // EVALSHA를 위한 스크립트 객체 생성
	}
}

// DeductBudget은 Lua 스크립트를 통해 원자적으로 예산을 차감합니다.
func (r *BudgetRepository) DeductBudget(ctx context.Context, campaignID string, bidAmount int64) (bool, error) {
	key := fmt.Sprintf("campaign:budget:%s", campaignID)
	
	// script.Run은 내부적으로 EVALSHA를 먼저 시도하고, 스크립트가 캐싱되어 있지 않으면 EVAL로 Fallback 합니다.
	result, err := r.script.Run(ctx, r.client, []string{key}, bidAmount).Int()
	if err != nil {
		return false, fmt.Errorf("failed to execute lua script: %w", err)
	}

	return result == 1, nil
}

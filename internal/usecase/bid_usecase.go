// internal/usecase/bid_usecase.go
package usecase

import (
	"context"
	"go-adtech-bidding/internal/domain"
	"time"
)

type BudgetRepository interface {
	DeductBudget(ctx context.Context, campaignID string, bidAmount int64) (bool, error)
}

type bidUsecase struct {
	budgetRepo BudgetRepository
	eventRepo  domain.BidEventRepository // Kafka 리포지토리 추가
}

func NewBidUsecase(repo BudgetRepository, eventRepo domain.BidEventRepository) domain.BidUsecase {
	return &bidUsecase{
		budgetRepo: repo,
		eventRepo:  eventRepo,
	}
}

func (u *bidUsecase) ProcessBid(ctx context.Context, req *domain.BidRequest) (*domain.BidResponse, error) {
	// 1. 원자적 예산 차감 (Redis Lua Script)
	isSuccess, err := u.budgetRepo.DeductBudget(ctx, req.CampaignID, req.Price)
	if err != nil {
		return nil, err
	}

	// 2. 불변의 비딩 이벤트 생성
	event := &domain.BidEvent{
		BidID:      req.ID,
		CampaignID: req.CampaignID,
		Price:      req.Price,
		DeviceID:   req.DeviceID,
		IsSuccess:  isSuccess,
		Timestamp:  time.Now().UnixMilli(),
	}

	// 3. 🌟 고루틴을 활용하여 핵심 경로에서 Kafka 전송 지연을 완전히 격리
	// HTTP 컨텍스트가 종료되더라도 전송이 유실되지 않도록 Background 컨텍스트 사용
	go func(e *domain.BidEvent) {
		// 비동기 전송 중 에러는 메인 스레드를 블로킹하지 않고 별도 로깅 처리 프로세스로 위임
		_ = u.eventRepo.PublishBidEvent(context.Background(), e)
	}(event)

	// 4. 즉시 응답 반환 (15ms 제약 준수)
	return &domain.BidResponse{
		ID:         req.ID,
		CampaignID: req.CampaignID,
		BidPrice:   req.Price,
		IsSuccess:  isSuccess,
	}, nil
}
package usecase

import (
	"context"
	"go-adtech-bidding/internal/domain"
)

type BudgetRepository interface {
	DeductBudget(ctx context.Context, campaignID string, bidAmount int64) (bool, error)
}

type bidUsecase struct {
	budgetRepo BudgetRepository
}

func NewBidUsecase(repo BudgetRepository) domain.BidUsecase {
	return &bidUsecase{budgetRepo: repo}
}

func (u *bidUsecase) ProcessBid(ctx context.Context, req *domain.BidRequest) (*domain.BidResponse, error) {
	// 1. 원자적 예산 차감 시도 (Lua 스크립트 실행)
	isSuccess, err := u.budgetRepo.DeductBudget(ctx, req.CampaignID, req.Price)
	if err != nil {
		return nil, err
	}

	// 2. 응답 객체 생성
	return &domain.BidResponse{
		ID:         req.ID,
		CampaignID: req.CampaignID,
		BidPrice:   req.Price,
		IsSuccess:  isSuccess,
	}, nil
}

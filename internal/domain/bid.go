package domain

import "context"

// BidRequest는 광고 익스체인지로부터 들어오는 입찰 요청 구조체입니다.
type BidRequest struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	Price      int64  `json:"price"`
	DeviceID   string `json:"device_id"`
}

// BidResponse는 입찰 결과를 반환하는 구조체입니다.
type BidResponse struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	BidPrice   int64  `json:"bid_price"`
	IsSuccess  bool   `json:"is_success"`
}

// BidUsecase는 입찰 비즈니스 로직의 인터페이스입니다.
type BidUsecase interface {
	ProcessBid(ctx context.Context, req *BidRequest) (*BidResponse, error)
}

// BidEvent는 Kafka로 발행할 불변의 기록 데이터 모델입니다.
type BidEvent struct {
	BidID      string `json:"bid_id"`
	CampaignID string `json:"campaign_id"`
	Price      int64  `json:"price"`
	DeviceID   string `json:"device_id"`
	IsSuccess  bool   `json:"is_success"`
	Timestamp  int64  `json:"timestamp"`
}

// BidEventRepository는 Kafka 이벤트 발행을 정의하는 인터페이스입니다.
type BidEventRepository interface {
	PublishBidEvent(ctx context.Context, event *BidEvent) error
	Close() error
}

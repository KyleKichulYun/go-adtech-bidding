// internal/repository/kafka/bid_producer.go
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"go-adtech-bidding/internal/domain"
	"time"

	"github.com/segmentio/kafka-go"
)

type BidEventProducer struct {
	writer *kafka.Writer
}

func NewBidEventProducer(brokers []string, topic string) domain.BidEventRepository {
	return &BidEventProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},   // 트래픽 분산을 위한 밸런서
			BatchSize:    100,                   // 100개의 이벤트가 모이면 한 번에 전송
			BatchTimeout: 10 * time.Millisecond, // 10ms가 지나면 배치가 차지 않아도 전송 (초저지연 유지)
			Async:        true,                  // 🌟 핵심 경로 블로킹 방지를 위한 비동기 모드 활성화
		},
	}
}

func (p *BidEventProducer) PublishBidEvent(ctx context.Context, event *domain.BidEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal bid event: %w", err)
	}

	// Async: true 설정으로 인해 WriteMessages는 브로커의 응답을 대기하지 않고
	// 내부 메모리 버퍼에 메시지를 넣은 후 즉시 리턴됩니다. (레이턴시 < 1ms)
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.CampaignID), // 캠페인별 순서 보장을 위해 캠페인ID를 키로 지정
		Value: payload,
	})
	if err != nil {
		return fmt.Errorf("failed to queue kafka message: %w", err)
	}

	return nil
}

func (p *BidEventProducer) Close() error {
	return p.writer.Close()
}

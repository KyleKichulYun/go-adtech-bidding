// cmd/bidder/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-adtech-bidding/internal/delivery/http"
	repoKafka "go-adtech-bidding/internal/repository/kafka"
	repoRedis "go-adtech-bidding/internal/repository/redis"
	"go-adtech-bidding/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. 설정 및 인프라 초기화 (로컬 인프라 환경 주소 구성)
	redisAddr := "localhost:6379"
	kafkaBrokers := []string{"localhost:9092"}
	kafkaTopic := "bid-events"

	// Redis 클라이언트 생성
	// [튜닝 1] Redis 커넥션 풀 확장 (2000 동시 접속 대비)
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		PoolSize:     2500, // 최대 VUs(2000)보다 넉넉하게 설정
        MinIdleConns: 500,  // 유휴 커넥션을 많이 유지하여 맺는 시간 단축
	})

	// Redis 연결 확인을 위한 Ping (정상 구동 상태 확인 타임아웃 2초 지정)
	initCtx, initCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer initCancel()
	if err := rdb.Ping(initCtx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	log.Println("Connected to Redis successfully.")

	// Kafka 비동기 프로듀서 초기화
	kafkaProducer := repoKafka.NewBidEventProducer(kafkaBrokers, kafkaTopic)
	log.Println("Initialized Kafka Async Producer.")

	// 2. 의존성 주입 및 레이어 조립 (Dependency Injection)
	budgetRepo := repoRedis.NewBudgetRepository(rdb)
	bidUsecase := usecase.NewBidUsecase(budgetRepo, kafkaProducer)

	// 초저지연 전용 타임아웃 튜닝이 반영된 Fiber 어플리케이션 선언
    // [튜닝 2 & 3] HTTP 서버 소켓 설정 완화
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Second * 2,  // 네트워크 패킷 읽기 타임아웃은 넉넉하게
		WriteTimeout: time.Second * 2,  // 응답 쓰기 타임아웃도 넉넉하게
		IdleTimeout:  time.Second * 10, // Keep-Alive 유휴 커넥션 유지 시간
		DisableKeepalive: false,        // Keep-Alive 강제 활성화
	})
	// API 엔드포인트 그룹화 및 핸들러 등록
	apiV1 := app.Group("/api/v1")
	http.NewBidHandler(apiV1, bidUsecase)

	// 3. 그레이스풀 셧다운 (Graceful Shutdown) 구조화
	// 커널 종료 시그널(Ctrl+C, k8s의 종료 요청 등)을 수신할 채널 정의
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	// 메인 HTTP 서버는 논블로킹 고루틴으로 실행하여 시그널 감지와 격리
	go func() {
		log.Printf("Bidder service starting on port :8080...")
		if err := app.Listen(":8080"); err != nil {
			log.Printf("Server listen stopping reason: %v", err)
		}
	}()

	// 시스템 종료 시그널 대기 (Blocking)
	<-stopSignal
	log.Println("Shutdown signal received. Cleansing resources gracefully...")

	// 셧다운 진행 시 최대 처리 한계 시간(Grace Period) 5초 할당
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// 인프라 자원 안전하게 Close (순서대로 닫기)
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Fiber HTTP server shutdown error: %v", err)
	}
	
	if err := rdb.Close(); err != nil {
		log.Printf("Redis client connection pool close error: %v", err)
	}
	
	if err := kafkaProducer.Close(); err != nil {
		log.Printf("Kafka producer flush & close error: %v", err)
	}

	log.Println("Bidder service cleanly stopped.")
}

.PHONY: compose-up compose-down run clean

# 도커 인프라 (Redis, Kafka) 백그라운드 실행
compose-up:
	docker-compose up -d

# 도커 인프라 종료 및 리소스 정리
compose-down:
	docker-compose down -v

# Go 입찰 서버 실행
run:
	go run cmd/bidder/main.go

.PHONY: load-test

# k6 부하 테스트 실행
load-test:
	k6 run scripts/k6/load_test.js
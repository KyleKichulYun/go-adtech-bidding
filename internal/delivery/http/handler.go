package http

import (
	"context"
	"go-adtech-bidding/internal/domain"
	"time"

	"github.com/gofiber/fiber/v2"
)

type BidHandler struct {
	usecase domain.BidUsecase
}

func NewBidHandler(router fiber.Router, uc domain.BidUsecase) {
	handler := &BidHandler{usecase: uc}
	router.Post("/bid", handler.HandleBid)
}

func (h *BidHandler) HandleBid(c *fiber.Ctx) error {
	// 15ms 타임아웃 컨텍스트 생성 (AdTech 핵심 제약 사항)
	ctx, cancel := context.WithTimeout(c.Context(), 15*time.Millisecond)
	defer cancel()

	req := new(domain.BidRequest)
	// Fiber의 파싱 최적화 메커니즘 활용
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	resp, err := h.usecase.ProcessBid(ctx, req)
	if err != nil {
		// 타임아웃 발생 시 408 혹은 내부 에러 반환
		if ctx.Err() == context.DeadlineExceeded {
			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"error": "timeout"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if !resp.IsSuccess {
		return c.Status(fiber.StatusNoContent).Send(nil) // 예산 부족 시 입찰 참여 안 함 (204)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

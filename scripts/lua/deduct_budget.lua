-- scripts/lua/deduct_budget.lua
-- KEYS[1]: 캠페인 예산 Redis Key (예: campaign:budget:1001)
-- ARGV[1]: 입찰 요청 금액 (Bid Amount)

local budget_key = KEYS[1]
local bid_amount = tonumber(ARGV[1])

-- 현재 예산 조회 (값이 없으면 0으로 간주)
local current_budget = tonumber(redis.call('GET', budget_key) or '0')

-- 잔여 예산이 입찰가보다 크거나 같으면 차감 진행
if current_budget >= bid_amount then
    redis.call('DECRBY', budget_key, bid_amount)
    return 1 -- 성공 (True)
else
    return 0 -- 예산 부족 (False)
end

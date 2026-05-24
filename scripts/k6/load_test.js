import http from 'k6/http';
import { check, sleep } from 'k6';

// ❌ 기존의 외부 URL import 코드는 삭제합니다.
// import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// ✅ IDE 경고를 없애고 네트워크 의존성을 제거하기 위해 순수 JS로 직접 구현합니다.
function randomIntBetween(min, max) {
    return Math.floor(Math.random() * (max - min + 1) + min);
}

function randomString(length) {
    const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += charset[Math.floor(Math.random() * charset.length)];
    }
    return result;
}

// 테스트 옵션: 점진적으로 트래픽을 늘려 최고조에 달한 후 감소시킴
export const options = {
    stages: [
        { duration: '10s', target: 200 },  // 10초 동안 500명의 유저로 웜업
        { duration: '30s', target: 500 }, // 30초 동안 2000명의 동시 유저 부하 유지
        { duration: '10s', target: 0 },    // 10초 동안 쿨다운
    ],
    thresholds: {
        // 성공적인 HTTP 요청 비율이 99.9% 이상이어야 함
        http_req_failed: ['rate<0.001'],
        // 전체 요청의 99%가 15ms 이내에 처리되어야 함 (AdTech 핵심 제약)
        http_req_duration: ['p(99)<15'], 
    },
};

export default function () {
    const url = 'http://localhost:8080/api/v1/bid';

    // 랜덤 입찰 데이터 생성
    const payload = JSON.stringify({
        id: `bid_${randomString(10)}`,
        campaign_id: `camp_${randomIntBetween(1, 100)}`, // 100개의 캠페인 키 분산 접근 (Redis 경합 테스트)
        price: randomIntBetween(1, 10),
        device_id: `dev_${randomString(15)}`,
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(url, payload, params);

    // 응답 검증: 예산이 있어서 입찰 성공(200) 하거나 예산 부족으로 스킵(204) 되어야 정상 처리로 간주
    check(res, {
        'status is 200 or 204': (r) => r.status === 200 || r.status === 204,
        'latency is under 15ms': (r) => r.timings.duration < 15,
    });

    sleep(1); // 각 가상 유저는 1초마다 요청
}

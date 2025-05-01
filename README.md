# Coupon Service

대규모 트래픽에 대응할 수 있는 **쿠폰 발급 시스템**입니다.  
Redis, SQS, Lambda, Postgres 등을 조합하여  
높은 동시성, 일관성, 확장성을 고려해 설계되었습니다.

---

## 서버 실행 방법

### 1. 로컬 실행 (Docker Compose)

```
docker compose -f local/docker-compose.yml up --build -d
```

- Redis, Postgres, LocalStack(SQS) 포함 모두 실행됩니다.
- 초기 DB 스키마 및 샘플 데이터는 `init.sh`에서 자동 적용됩니다.

### 2. 개별 모듈 실행

#### Campaign Manager (캠페인 관리 서버)

```
go run cmd/campaign-manage/main.go
```

#### Scheduler (큐 → SQS 전송 스케줄러)

```
go run cmd/scheduler/main.go
```

#### Lambda Processor (SQS → DB 소비 Lambda)

```
go run cmd/lambda/main.go
```

---

## API 요청 방법

### ✅ 캠페인 생성

```
curl -X POST http://localhost:8080/coupon.v1.CouponService/CreateCampaign \
  -H "Content-Type: application/json" \
  -d '{
    "name": "봄맞이 이벤트",
    "coupon_count": 1000,
    "start_at": "2025-05-01T00:00:00Z",
    "end_at": "2025-05-31T23:59:59Z"
  }'
```

### ✅ 쿠폰 발급 요청

```
curl -X POST http://localhost:8080/coupon.v1.CouponService/IssueCoupon \
  -H "Content-Type: application/json" \
  -d '{
    "campaign_id": 1,
    "user_id": 101
  }'
```

### ✅ 캠페인 조회 (발급된 쿠폰 코드 포함)

```
curl -X POST http://localhost:8080/coupon.v1.CouponService/GetCampaign \
  -H "Content-Type: application/json" \
  -d '{
    "campaign_id": 1
  }'
```

---

## 설계

### 📊 아키텍처 설계

```
사용자 → Campaign Manager → Redis 대기열 → Scheduler → SQS → Lambda → Postgres (최종 발급)
```

- **Redis**: 캠페인별 사용자 요청 대기열 (ZSET) 및 중복 요청 필터링 (SET)
- **Scheduler**: Redis에서 1초마다 일정 수의 요청을 가져와 SQS로 전달
- **SQS**: 고가용성 큐로 Lambda와 연결
- **Lambda Processor**: SQS 메시지를 받아 최종적으로 쿠폰을 발급 (Postgres insert), 성공 시 Redis 요청 기록 제거
- **Postgres**: 캠페인, 쿠폰의 영속적 저장소 (ACID 보장)

### 📂 프로젝트 구조

```
/cmd
  ├── campaign-manage    # 캠페인 및 API 서버
  ├── scheduler          # Redis → SQS 스케줄러
  └── lambda             # SQS → DB 소비자

/internal
  ├── api                # ConnectRPC 핸들러
  ├── service            # 비즈니스 로직
  ├── repository         # DB, Redis 접근
  ├── db                 # Postgres 연결
  └── redis              # Redis 클라이언트

/gen
  └── go                 # proto → Go 코드

/proto
  └── coupon/v1          # proto 정의 파일
```

### 🗂️ ERD

- **campaigns**
  | id (PK) | name | coupon_count | start_at | end_at |
  |---------|-------|--------------|----------|--------|

- **coupons**
  | id (PK) | code (UNIQUE) | user_id | campaign_id (FK) | issued_at | expired_at (nullable) |

---

## ⚙️ 설계 시 고려한 부분

✅ **대규모 요청 처리**: Redis로 캠페인별 사용자 요청을 비동기 대기열로 관리, Scheduler → SQS → Lambda로 백그라운드 발급 처리, API 서버는 최대한 빠르게 요청 접수 후 응답

✅ **일관성**: Lambda에서 발급 성공 후 Redis에서 요청 기록 제거, Postgres의 UNIQUE INDEX로 최종 발급 충돌 방지

✅ **확장성**: 캠페인 관리, 발급 스케줄러, Lambda 모듈별로 독립 배포 가능 (ECS, Lambda 이미지), SQS와 Redis를 통해 모듈 간 강한 결합 해소

✅ **고유한 쿠폰 코드**: UUID 기반 한글+숫자 치환 방식으로 충돌 가능성 최소화, 16글자 고정

✅ **클린한 레이어드 설계**: handler → service → repository, repository 분리로 테스트 용이

---

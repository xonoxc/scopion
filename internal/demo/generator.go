package demo

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store"
)

var (
	services  = []string{"api", "worker", "webhook", "cron", "scheduler", "auth", "payment"}
	endpoints = []string{
		"GET /users", "POST /login", "GET /orders", "POST /webhook",
		"ProcessPayment", "SendNotification", "CleanupSessions",
		"ScheduleTask", "ValidateToken", "ChargeCard",
	}
	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		"PostmanRuntime/7.32.2",
		"curl/7.68.0",
	}
)

func generateCustomData(service, name string) map[string]any {
	data := make(map[string]any)

	switch service {
	case "api":
		if strings.Contains(name, "GET /") {
			data["method"] = "GET"
			data["path"] = strings.TrimPrefix(name, "GET ")
			data["user_id"] = fmt.Sprintf("user_%d", rand.Intn(1000))
			data["response_size"] = rand.Intn(5000) + 100
			data["status_code"] = 200
			data["user_agent"] = userAgents[rand.Intn(len(userAgents))]
		} else if strings.Contains(name, "POST /") {
			data["method"] = "POST"
			data["path"] = strings.TrimPrefix(name, "POST ")
			data["ip_address"] = fmt.Sprintf("192.168.1.%d", rand.Intn(255))
			data["request_size"] = rand.Intn(10000) + 100
			data["content_type"] = "application/json"
			if rand.Float64() < 0.1 {
				data["failed_reason"] = "validation_error"
			}
		} else if strings.Contains(name, "PUT /") || strings.Contains(name, "DELETE /") {
			method := "PUT"
			if strings.Contains(name, "DELETE") {
				method = "DELETE"
			}
			data["method"] = method
			data["path"] = strings.TrimPrefix(strings.TrimPrefix(name, method+" "), "")
			data["user_id"] = fmt.Sprintf("user_%d", rand.Intn(1000))
			data["status_code"] = 200
		}
	case "worker":
		switch name {
		case "ProcessPayment":
			data["amount"] = float64(rand.Intn(10000)) / 100
			data["currency"] = "USD"
			data["payment_method"] = []string{"credit_card", "paypal", "bank_transfer"}[rand.Intn(3)]
			data["transaction_id"] = fmt.Sprintf("txn_%s", randomID())
			data["processing_time_ms"] = rand.Intn(5000) + 100
		case "SendEmail":
			data["recipient_count"] = rand.Intn(10) + 1
			data["email_type"] = []string{"welcome", "notification", "marketing", "reset"}[rand.Intn(4)]
			data["template_id"] = fmt.Sprintf("template_%d", rand.Intn(100))
			data["priority"] = []string{"high", "normal", "low"}[rand.Intn(3)]
		}
	case "payment":
		data["amount"] = float64(rand.Intn(50000)) / 100
		data["currency"] = []string{"USD", "EUR", "GBP"}[rand.Intn(3)]
		data["gateway"] = []string{"stripe", "paypal", "braintree"}[rand.Intn(3)]
		data["card_type"] = []string{"visa", "mastercard", "amex"}[rand.Intn(3)]
		data["region"] = []string{"us-east", "us-west", "eu-west", "ap-south"}[rand.Intn(4)]
		if rand.Float64() < 0.2 {
			data["error_code"] = fmt.Sprintf("ERR_%d", rand.Intn(1000))
		}
	case "auth":
		data["auth_type"] = []string{"jwt", "oauth", "basic", "session"}[rand.Intn(4)]
		data["user_id"] = fmt.Sprintf("user_%d", rand.Intn(10000))
		data["ip_address"] = fmt.Sprintf("10.0.%d.%d", rand.Intn(255), rand.Intn(255))
		data["device_fingerprint"] = fmt.Sprintf("fp_%s", randomID())
		if rand.Float64() < 0.1 {
			data["suspicious_activity"] = true
			data["risk_score"] = rand.Float64()
		}
	case "webhook":
		data["source"] = []string{"stripe", "github", "slack", "twilio"}[rand.Intn(4)]
		data["event_type"] = []string{"payment.succeeded", "user.created", "order.updated"}[rand.Intn(3)]
		data["webhook_id"] = fmt.Sprintf("wh_%s", randomID())
		data["payload_size"] = rand.Intn(10000) + 100
		data["signature_valid"] = rand.Float64() > 0.05
	default:
		data["operation_id"] = fmt.Sprintf("op_%s", randomID())
		data["duration_ms"] = rand.Intn(10000) + 10
		data["resource_count"] = rand.Intn(100) + 1
	}

	data["request_id"] = fmt.Sprintf("req_%s", randomID())
	data["timestamp_ns"] = time.Now().UnixNano()

	return data
}

func Start(ctx context.Context, s store.Storage, b *live.Broadcaster, log *slog.Logger) {
	generateHistoricalData(s, log)

	go generateLoop(ctx, "api", s, b, log, 50, 250, 0.1)
	go generateLoop(ctx, "worker", s, b, log, 100, 500, 0.2)
	go generateLoop(ctx, "webhook", s, b, log, 200, 800, 0.15)
	go generateLoop(ctx, "cron", s, b, log, 500, 1000, 0.05)
	go generateLoop(ctx, "scheduler", s, b, log, 200, 600, 0.08)
	go generateLoop(ctx, "auth", s, b, log, 100, 350, 0.12)
	go generateLoop(ctx, "payment", s, b, log, 150, 400, 0.25)
}

var serviceEndpoints = map[string][]string{
	"api":       {"GET /users", "POST /login", "GET /orders", "PUT /profile", "DELETE /session"},
	"worker":    {"ProcessPayment", "SendEmail", "GenerateReport", "CleanupData"},
	"webhook":   {"POST /webhook/payment", "POST /webhook/order", "POST /webhook/user"},
	"cron":      {"CleanupSessions", "ArchiveOldData", "GenerateDailyReport", "SyncExternalData"},
	"scheduler": {"ScheduleTask", "QueueJob", "ProcessQueue", "UpdateMetrics"},
	"auth":      {"ValidateToken", "RefreshToken", "PasswordReset", "UserLogin"},
	"payment":   {"ChargeCard", "RefundPayment", "ValidatePayment", "ProcessRefund"},
}

func generateLoop(ctx context.Context, service string, s store.Storage, b *live.Broadcaster, log *slog.Logger, minDelay, maxDelay int, errorRate float64) {
	endpoints := serviceEndpoints[service]
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		traceID := randomID()
		endpoint := endpoints[rand.Intn(len(endpoints))]

		emitWithTime(s, b, log, service, endpoint, traceID, "info", time.Now())

		if rand.Float64() < errorRate {
			emitWithTime(s, b, log, service, "db operation", traceID, "error", time.Now())
		}

		delay := time.Duration(minDelay+rand.Intn(maxDelay-minDelay)) * time.Millisecond
		time.Sleep(delay)
	}
}

func generateHistoricalData(s store.Storage, log *slog.Logger) {
	now := time.Now()
	for range 200 {
		pastTime := now.Add(-time.Duration(rand.Intn(86400)) * time.Second)

		traceID := randomID()
		service := services[rand.Intn(len(services))]
		endpoint := endpoints[rand.Intn(len(endpoints))]

		level := "info"
		if rand.Float64() < 0.15 {
			level = "error"
		}

		emitWithTime(s, nil, log, service, endpoint, traceID, level, pastTime)

		if rand.Float64() < 0.6 {
			pastTime2 := pastTime.Add(time.Duration(rand.Intn(500)) * time.Millisecond)
			emitWithTime(s, nil, log, service, "db operation", traceID, level, pastTime2)
		}
	}
}

func emitWithTime(s store.Storage, b *live.Broadcaster, log *slog.Logger, service, name, traceID, level string, timestamp time.Time) {
	e := model.Event{
		ID:          uuid.New().String(),
		Timestamp:   timestamp,
		Service:     service,
		Name:        name,
		TraceID:     traceID,
		SpanID:      fmt.Sprintf("span_%s", randomID()),
		Environment: []string{"production", "staging", "development"}[rand.Intn(3)],
		Level:       level,
		DurationMs:  float64(rand.Intn(500) + 10),
	}

	if rand.Float64() < 0.7 {
		e.Data = generateCustomData(service, name)
	}

	duration := time.Duration(e.DurationMs) * time.Millisecond
	spanStatus := "OK"
	if level == "error" {
		spanStatus = "ERROR"
	}
	span := model.Span{
		TraceID:   traceID,
		SpanID:    e.SpanID,
		Name:      name,
		Service:   service,
		StartTime: timestamp,
		EndTime:   timestamp.Add(duration),
		Status:    spanStatus,
	}

	for retries := range 3 {
		err := s.Append(e)
		if err == nil {
			break
		}
		if retries == 2 {
			log.Error("failed to append event after retries", "error", err, "service", service, "name", name)
		}
		time.Sleep(time.Duration(retries+1) * 10 * time.Millisecond)
	}

	for retries := range 3 {
		err := s.AppendSpan(span)
		if err == nil {
			break
		}
		if retries == 2 {
			log.Error("failed to append span after retries", "error", err, "service", service, "name", name)
		}
		time.Sleep(time.Duration(retries+1) * 10 * time.Millisecond)
	}

	if b != nil {
		b.Publish(e)
	}
}

func randomID() string {
	return fmt.Sprintf("%s_%s", time.Now().Format("150405.000000"), uuid.New().String()[:8])
}

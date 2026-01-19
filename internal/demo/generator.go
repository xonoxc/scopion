package demo

import (
	"fmt"
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
	_          = []string{"GET /users", "POST /login", "GET /orders", "PUT /profile", "DELETE /session"}
	_          = []string{"ProcessPayment", "SendEmail", "GenerateReport", "CleanupData"}
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
		if name == "ProcessPayment" {
			data["amount"] = float64(rand.Intn(10000)) / 100
			data["currency"] = "USD"
			data["payment_method"] = []string{"credit_card", "paypal", "bank_transfer"}[rand.Intn(3)]
			data["transaction_id"] = fmt.Sprintf("txn_%s", randomID())
			data["processing_time_ms"] = rand.Intn(5000) + 100
		} else if name == "SendEmail" {
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

func Start(store *store.Store, live *live.Broadcaster) {
	generateHistoricalData(store)

	go generateAPIRequests(store, live)
	go generateWorkerTasks(store, live)
	go generateWebhookEvents(store, live)
	go generateCronJobs(store, live)
	go generateSchedulerTasks(store, live)
	go generateAuthEvents(store, live)
	go generatePaymentEvents(store, live)
}

func generateHistoricalData(store *store.Store) {
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

		emitWithTime(store, nil, service, endpoint, traceID, level, pastTime)

		if rand.Float64() < 0.6 {
			pastTime2 := pastTime.Add(time.Duration(rand.Intn(500)) * time.Millisecond)
			emitWithTime(store, nil, service, "db operation", traceID, level, pastTime2)
		}
	}
}

func generateAPIRequests(store *store.Store, live *live.Broadcaster) {
	apiEndpoints := []string{"GET /users", "POST /login", "GET /orders", "PUT /profile", "DELETE /session"}
	for {
		traceID := randomID()
		endpoint := apiEndpoints[rand.Intn(len(apiEndpoints))]

		emit(store, live, "api", endpoint, traceID, "info")
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

		if rand.Float64() < 0.1 {
			emit(store, live, "api", "db query", traceID, "error")
		} else {
			emit(store, live, "api", "db query", traceID, "info")
		}

		time.Sleep(time.Duration(200+rand.Intn(800)) * time.Millisecond)
	}
}

func generateWorkerTasks(store *store.Store, live *live.Broadcaster) {
	workerTasks := []string{"ProcessPayment", "SendEmail", "GenerateReport", "CleanupData"}
	for {
		traceID := randomID()
		task := workerTasks[rand.Intn(len(workerTasks))]

		emit(store, live, "worker", task, traceID, "info")
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

		if rand.Float64() < 0.2 {
			emit(store, live, "worker", "db update", traceID, "error")
		} else {
			emit(store, live, "worker", "db update", traceID, "info")
		}

		time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
	}
}

func generateWebhookEvents(store *store.Store, live *live.Broadcaster) {
	webhookEvents := []string{"POST /webhook/payment", "POST /webhook/order", "POST /webhook/user"}
	for {
		traceID := randomID()
		event := webhookEvents[rand.Intn(len(webhookEvents))]

		emit(store, live, "webhook", event, traceID, "info")
		time.Sleep(time.Duration(200+rand.Intn(500)) * time.Millisecond)

		if rand.Float64() < 0.15 {
			emit(store, live, "webhook", "process webhook", traceID, "error")
		}

		time.Sleep(time.Duration(1000+rand.Intn(3000)) * time.Millisecond)
	}
}

func generateCronJobs(store *store.Store, live *live.Broadcaster) {
	cronJobs := []string{"CleanupSessions", "ArchiveOldData", "GenerateDailyReport", "SyncExternalData"}
	for {
		traceID := randomID()
		job := cronJobs[rand.Intn(len(cronJobs))]

		emit(store, live, "cron", job, traceID, "info")
		time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

		if rand.Float64() < 0.05 {
			emit(store, live, "cron", "file operation", traceID, "error")
		}

		time.Sleep(time.Duration(30000+rand.Intn(60000)) * time.Millisecond)
	}
}

func generateSchedulerTasks(store *store.Store, live *live.Broadcaster) {
	schedulerTasks := []string{"ScheduleTask", "QueueJob", "ProcessQueue", "UpdateMetrics"}
	for {
		traceID := randomID()
		task := schedulerTasks[rand.Intn(len(schedulerTasks))]

		emit(store, live, "scheduler", task, traceID, "info")
		time.Sleep(time.Duration(200+rand.Intn(400)) * time.Millisecond)

		if rand.Float64() < 0.08 {
			emit(store, live, "scheduler", "queue operation", traceID, "error")
		}

		time.Sleep(time.Duration(2000+rand.Intn(5000)) * time.Millisecond)
	}
}

func generateAuthEvents(store *store.Store, live *live.Broadcaster) {
	authEvents := []string{"ValidateToken", "RefreshToken", "PasswordReset", "UserLogin"}
	for {
		traceID := randomID()
		event := authEvents[rand.Intn(len(authEvents))]

		emit(store, live, "auth", event, traceID, "info")
		time.Sleep(time.Duration(100+rand.Intn(300)) * time.Millisecond)

		if rand.Float64() < 0.12 {
			emit(store, live, "auth", "db lookup", traceID, "error")
		}

		time.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)
	}
}

func generatePaymentEvents(store *store.Store, live *live.Broadcaster) {
	paymentEvents := []string{"ChargeCard", "RefundPayment", "ValidatePayment", "ProcessRefund"}
	for {
		traceID := randomID()
		event := paymentEvents[rand.Intn(len(paymentEvents))]

		emit(store, live, "payment", event, traceID, "info")
		time.Sleep(time.Duration(150+rand.Intn(350)) * time.Millisecond)

		if rand.Float64() < 0.25 {
			emit(store, live, "payment", "payment gateway", traceID, "error")
		}

		time.Sleep(time.Duration(800+rand.Intn(2000)) * time.Millisecond)
	}
}

func emit(store *store.Store, live *live.Broadcaster, service, name, traceID, level string) {
	emitWithTime(store, live, service, name, traceID, level, time.Now())
}

func emitWithTime(store *store.Store, live *live.Broadcaster, service, name, traceID, level string, timestamp time.Time) {
	e := model.Event{
		ID:        uuid.New().String(),
		Timestamp: timestamp,
		Service:   service,
		Name:      name,
		TraceID:   traceID,
		Level:     level,
	}

	if rand.Float64() < 0.7 {
		e.Data = generateCustomData(service, name)
	}

	for retries := range 3 {
		err := store.Append(e)
		if err == nil {
			if live != nil {
				live.Publish(e)
			}
			return
		}

		if retries == 2 {
			println("Failed to append event after retries:", err.Error())
		}
		time.Sleep(time.Duration(retries+1) * 10 * time.Millisecond)
	}
}

func randomID() string {
	return time.Now().Format("150405.000000")
}

package demo

import (
	"math/rand"
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
)

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

		e := model.Event{
			ID:        uuid.New().String(),
			Timestamp: pastTime,
			Service:   service,
			Name:      endpoint,
			TraceID:   traceID,
			Level:     level,
		}

		if err := store.Append(e); err != nil {
			// Silently continue on error during historical data generation
			continue
		}

		// Add a second event for some traces to simulate spans
		if rand.Float64() < 0.6 {
			e2 := model.Event{
				ID:        uuid.New().String(),
				Timestamp: pastTime.Add(time.Duration(rand.Intn(500)) * time.Millisecond),
				Service:   service,
				Name:      "db operation",
				TraceID:   traceID,
				Level:     level,
			}
			store.Append(e2) // Ignore error for second event
		}
	}
}

func generateAPIRequests(store *store.Store, live *live.Broadcaster) {
	apiEndpoints := []string{"GET /users", "POST /login", "GET /orders", "PUT /profile", "DELETE /session"}
	for {
		traceID := randomID()
		endpoint := apiEndpoints[rand.Intn(len(apiEndpoints))]

		// API request
		emit(store, live, "api", endpoint, traceID, "info")
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

		// Sometimes add an error
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

		// Higher error rate for workers
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

		// Cron jobs run less frequently
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

		// Payments have higher error rates
		if rand.Float64() < 0.25 {
			emit(store, live, "payment", "payment gateway", traceID, "error")
		}

		time.Sleep(time.Duration(800+rand.Intn(2000)) * time.Millisecond)
	}
}

func emit(store *store.Store, live *live.Broadcaster, service, name, traceID, level string) {
	e := model.Event{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Service:   service,
		Name:      name,
		TraceID:   traceID,
		Level:     level,
	}

	for retries := range 3 {
		err := store.Append(e)
		if err == nil {
			live.Publish(e)
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

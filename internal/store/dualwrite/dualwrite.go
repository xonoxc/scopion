package dualwrite

import (
	"log"

	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store"
)

/**
*
* this is the store outer methods will be wiriting
* to while the system is in DUAL_WRITE state
**/
type DualWriteStore struct {
	primary   store.Storage
	secondary store.Storage
}

func New(primary, secondary store.Storage) *DualWriteStore {
	return &DualWriteStore{
		primary:   primary,
		secondary: secondary,
	}
}

func (d *DualWriteStore) Primary() store.Storage {
	return d.primary
}

func (d *DualWriteStore) Secondary() store.Storage {
	return d.secondary
}

func (d *DualWriteStore) GetStats() (*model.Stats, error) {
	return d.primary.GetStats()
}

func (d *DualWriteStore) Append(event model.Event) error {
	if err := d.primary.Append(event); err != nil {
		return err
	}

	if err := d.secondary.Append(event); err != nil {
		log.Printf("warning: failed to write to secondary store: %v", err)
	}

	return nil
}

func (d *DualWriteStore) Recent(n int) ([]model.Event, error) {
	return d.primary.Recent(n)
}

func (d *DualWriteStore) GetServices() ([]model.ServiceInfo, error) {
	return d.primary.GetServices()
}

func (d *DualWriteStore) GetErrorsByService(hours int) ([]model.ErrorByService, error) {
	return d.primary.GetErrorsByService(hours)
}

func (d *DualWriteStore) GetTraces(limit int) ([]model.TraceInfo, error) {
	return d.primary.GetTraces(limit)
}

func (d *DualWriteStore) GetEventsByTraceID(traceID string) ([]model.Event, error) {
	return d.primary.GetEventsByTraceID(traceID)
}

func (d *DualWriteStore) SearchEvents(query string, limit int) ([]model.Event, error) {
	return d.primary.SearchEvents(query, limit)
}

func (d *DualWriteStore) GetThroughput(hours int) ([]model.ThroughputData, error) {
	return d.primary.GetThroughput(hours)
}

func (d *DualWriteStore) Close() error {
	if err := d.primary.Close(); err != nil {
		return err
	}
	return d.secondary.Close()
}

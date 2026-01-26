package store

import (
	"github.com/xonoxc/scopion/internal/model"
)

/*
*this is the interface for the storage services
*this will help to switch between different storage services
*****/

type Storage interface {
	Append(event model.Event) error

	Recent(n int) ([]model.Event, error)

	/*
		stats related methods
	*/
	GetStats() (*model.Stats, error)

	/*
		services related methods
	*/
	GetServices() ([]model.ServiceInfo, error)

	GetErrorsByService(hours int) ([]model.ErrorByService, error)

	/*
		trace related methods
	*/
	GetTraces(limit int) ([]model.TraceInfo, error)

	GetEventsByTraceID(traceID string) ([]model.Event, error)

	/*
		search related methods
	*/
	SearchEvents(query string, limit int) ([]model.Event, error)

	/*
		throughput related methods
	*/
	GetThroughput(hours int) ([]model.ThroughputData, error)

	/*
	*closing the storage service
	 */
	Close() error
}

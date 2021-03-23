package errands

import (
	"context"

	log "github.com/sirupsen/logrus"
)

const (
	ContextLogger    = "logger"
	ContextRequestID = "request-id"
)

// ErrandContext implements api_context.Context for Errands.
type ErrandContext struct {
	context.Context

	ID     string
	logger *log.Entry
}

func NewErrandContext(parentCtx context.Context, errandID string) *ErrandContext {
	return &ErrandContext{
		Context: parentCtx,
		ID:      errandID,
		logger:  log.WithField("errand_id", errandID),
	}
}

func (c *ErrandContext) RequestID() string {
	return c.ID
}

func (c *ErrandContext) Logger() *log.Entry {
	return c.logger
}

func (c *ErrandContext) AddFieldsToLogger(fields log.Fields) {
	c.logger = c.logger.WithFields(fields)
}

func (c *ErrandContext) AddFieldToLogger(key string, value interface{}) {
	c.logger = c.logger.WithField(key, value)
}

func (c *ErrandContext) Value(key interface{}) interface{} {
	switch key {
	case ContextLogger:
		return c.Logger()
	case ContextRequestID:
		return c.RequestID()
	default:
		return c.Context.Value(key)
	}
}

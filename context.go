package errands

import (
	"context"

	log "github.com/sirupsen/logrus"
)

const (
	ContextLogger    = "logger"
	ContextRequestID = "request-id"
)

// Context implements api_context.Context for Errands.
type Context struct {
	context.Context

	ID     string
	logger *log.Entry
}

func NewContext(parentCtx context.Context, errandID string) *Context {
	return &Context{
		Context: parentCtx,
		ID:      errandID,
		logger:  log.WithField("errand_id", errandID),
	}
}

func (c *Context) RequestID() string {
	return c.ID
}

func (c *Context) Logger() *log.Entry {
	return c.logger
}

func (c *Context) AddFieldsToLogger(fields log.Fields) {
	c.logger = c.logger.WithFields(fields)
}

func (c *Context) AddFieldToLogger(key string, value interface{}) {
	c.logger = c.logger.WithField(key, value)
}

func (c *Context) Value(key interface{}) interface{} {
	switch key {
	case ContextLogger:
		return c.Logger()
	case ContextRequestID:
		return c.RequestID()
	default:
		return c.Context.Value(key)
	}
}

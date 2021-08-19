package utils

import "github.com/tendermint/tendermint/libs/log"

// NopLogger is a logger that doesn't do anything
type NopLogger struct{}

// Interface assertions
var _ log.Logger = (*NopLogger)(nil)

// NewNopLogger returns a logger that doesn't do anything.
func NewNopLogger() log.Logger { return &NopLogger{} }

func (NopLogger) Info(string, ...interface{})  {}
func (NopLogger) Debug(string, ...interface{}) {}
func (NopLogger) Error(string, ...interface{}) {}

func (l *NopLogger) With(...interface{}) log.Logger {
	return l
}

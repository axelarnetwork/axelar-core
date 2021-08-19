package utils

import "github.com/tendermint/tendermint/libs/log"

// NOPLogger is a logger that doesn't do anything
type NOPLogger struct{}

// Interface assertions
var _ log.Logger = (*NOPLogger)(nil)

// NewNOPLogger returns a logger that doesn't do anything.
func NewNOPLogger() log.Logger { return &NOPLogger{} }

func (NOPLogger) Info(string, ...interface{})  {}
func (NOPLogger) Debug(string, ...interface{}) {}
func (NOPLogger) Error(string, ...interface{}) {}

func (l *NOPLogger) With(...interface{}) log.Logger {
	return l
}

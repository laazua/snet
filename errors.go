package snet

import "errors"

var (
	ErrCertFileNotFound       = errors.New("cert files do not exist")
	ErrClientConnected        = errors.New("client already connected")
	ErrClientNotConnected     = errors.New("client not connected")
	ErrMagicNumberInvalid     = errors.New("invalid magic number")
	ErrPacketTooLarge         = errors.New("packet too large")
	ErrPacketIvalid           = errors.New("invalid packet")
	ErrWorkerPoolClosed       = errors.New("worker pool is closed")
	ErrWorkerPoolQueueFull    = errors.New("worker pool queue is full")
	ErrServerHandlerNotSet    = errors.New("server handler not set")
	ErrServerWorkerPoolNotSet = errors.New("server worker pool not set")
)

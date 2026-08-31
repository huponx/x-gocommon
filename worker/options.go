package worker

import "go.uber.org/zap"

type options struct {
	logger *zap.Logger
}

func defaultOptions() options {
	return options{logger: zap.NewNop()}
}

type Option func(*options)

func WithLogger(log *zap.Logger) Option {
	return func(o *options) {
		if log != nil {
			o.logger = log
		}
	}
}

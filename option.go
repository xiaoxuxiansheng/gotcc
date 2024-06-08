package gotcc

import "time"

type Options struct {
	// 事务执行时长限制
	Timeout time.Duration
	// 轮询监控任务间隔时长
	MonitorTick time.Duration
}

type Option func(*Options)

func WithTimeout(timeout time.Duration) Option {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return func(o *Options) {
		o.Timeout = timeout
	}
}

func WithMonitorTick(tick time.Duration) Option {
	if tick <= 0 {
		tick = 10 * time.Second
	}

	return func(o *Options) {
		o.MonitorTick = tick
	}
}

func repair(o *Options) {
	if o.MonitorTick <= 0 {
		o.MonitorTick = 10 * time.Second
	}

	if o.Timeout <= 0 {
		o.Timeout = 5 * time.Second
	}
}

package handler

import (
	"github.com/go-gost/gost/pkg/bypass"
	"github.com/go-gost/gost/pkg/logger"
	"github.com/go-gost/gost/pkg/resolver"
)

type Options struct {
	Bypass   bypass.Bypass
	Resolver resolver.Resolver
	Logger   logger.Logger
}

type Option func(opts *Options)

func LoggerOption(logger logger.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

func BypassOption(bypass bypass.Bypass) Option {
	return func(opts *Options) {
		opts.Bypass = bypass
	}
}

func ResolverOption(resolver resolver.Resolver) Option {
	return func(opts *Options) {
		opts.Resolver = resolver
	}
}

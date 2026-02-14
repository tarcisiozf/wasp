package tests

import "wasp/wasp"

type EnvironmentOption func(*Environment)

func WithInstanceOptions(opts ...wasp.InstanceOption) EnvironmentOption {
	return func(env *Environment) {
		env.instanceOptions = append(env.instanceOptions, opts...)
	}
}

type Environment struct {
	instanceOptions []wasp.InstanceOption
}

func NewEnvironment(opts ...EnvironmentOption) *Environment {
	env := &Environment{}
	for _, opt := range opts {
		opt(env)
	}
	return env
}

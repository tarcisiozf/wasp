package tests

import (
	"testing"

	"github.com/tarcisiozf/wasp"
)

type EnvironmentOption func(*Environment)

func WithInstanceOptions(opts ...wasp.InstanceOption) EnvironmentOption {
	return func(env *Environment) {
		env.instanceOptions = append(env.instanceOptions, opts...)
	}
}

func WithWasmtimeBuilder() EnvironmentOption {
	builder := &WasmtimeWatBuilder{}
	return func(env *Environment) {
		env.watBuilder = builder
	}
}

type WatBuilder interface {
	Build(dir, wat string) (Build, error)
}

type Environment struct {
	instanceOptions []wasp.InstanceOption
	watBuilder      WatBuilder
}

func (env *Environment) BuildWat(t *testing.T, wat string) (Build, error) {
	dir := t.TempDir()
	return env.watBuilder.Build(dir, wat)
}

func NewEnvironment(opts ...EnvironmentOption) *Environment {
	env := &Environment{}
	for _, opt := range opts {
		opt(env)
	}
	if env.watBuilder == nil {
		env.watBuilder = &Wat2WasmBuilder{}
	}
	return env
}

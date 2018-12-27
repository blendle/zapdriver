package zapdriver

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const serviceContextKey = "serviceContext"

func ServiceContext(name string) zap.Field {
	return zap.Object(serviceContextKey, newServiceContext(name))
}

type serviceContext struct {
	Name string `json:"service"`
}

// MarshalLogObject implements zapcore.ObjectMarshaller interface.
func (service_context serviceContext) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("service", service_context.Name)

	return nil
}

func newServiceContext(name string) *serviceContext {
	return &serviceContext{
		Name: name,
	}
}

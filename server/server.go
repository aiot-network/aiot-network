package server

type IService interface {
	Name() string
	Start() error
	Stop() error
	Info() map[string]interface{}
}

type IServer interface {
	Register(IService)
	Start() error
	Stop() error
}

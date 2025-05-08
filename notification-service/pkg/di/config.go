package di

import (
	"go.uber.org/dig"
)

type Container struct {
	Container *dig.Container
}

func InitContainer() *Container {
	c := dig.New()
	return &Container{Container: c}
}

func (c *Container) GetContainer() *dig.Container {
	return c.Container
}

func (c *Container) RegisterService(service interface{}) {
	c.Container.Provide(service)
}

func (c *Container) InvokeService(service interface{}) {
	c.Container.Invoke(service)
}

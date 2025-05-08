package app

import "go.uber.org/dig"

var Container *dig.Container

func InitDeps() {
	Container = dig.New()

}

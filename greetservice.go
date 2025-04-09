package main

import (
	"os"
	"wails-react-3-demo/update"
)

type GreetService struct{}

func (g *GreetService) Greet(name string) string {
	err := update.AppUpdate()
	if err != nil {
		return err.Error()
	}
	return "更新成功，重新启动应用即可!"
}

func (g *GreetService) GetCurrentPath() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	return executable
}

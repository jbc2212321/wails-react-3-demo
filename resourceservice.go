package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type ResourceService struct{}

func (r *ResourceService) InitResource() {

}

func NewResourceService() *ResourceService {
	return &ResourceService{}
}

func (r *ResourceService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	fmt.Println("Starting Service.......")
	return nil
}

func (r *ResourceService) OnStartup(ctx context.Context, options application.ServiceOptions) error {
	fmt.Println("......Start Service......")

	r.InitResource()

	path, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	}
	dir := filepath.Dir(path)
	fmt.Println("path:", path) // for example /home/user/main
	fmt.Println("dir:", dir)   // for example /home/user

	return nil
}

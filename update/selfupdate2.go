package update

import (
	"fmt"
	"github.com/sanbornm/go-selfupdate/selfupdate"
	"net/http"
	"os"

	go_update "github.com/inconshreveable/go-update"
)

func update() {
	var updater = &selfupdate.Updater{
		CurrentVersion: version,                                                                    // 你的应用程序当前版本
		ApiURL:         "http://updates.yourdomain.com/",                                           // 更新API地址
		BinURL:         "https://github.com/jbc2212321/wails-react-3-demo/releases/tag/build-mac/", // 全量二进制下载地址
		DiffURL:        "http://updates.yourdomain.com/",                                           // 补丁下载地址
		Dir:            "update/",                                                                  // 存储临时状态文件的目录
		CmdName:        "wails3-react-3-demo",                                                      // 你的应用程序名称
		OnSuccessfulUpdate: func() {
			os.Exit(0)
		},
	}

	// 开启后台更新检查
	go func() {
		err := updater.BackgroundRun()
		if err != nil {
			panic(err)
		}
	}()

}

func doUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = go_update.Apply(resp.Body, go_update.Options{})
	if err != nil {
		// error handling
		fmt.Println(err)
	}
	return err
}

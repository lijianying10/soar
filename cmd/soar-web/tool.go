package main

import (
	"fmt"
	"os"

	"github.com/XiaoMi/soar/common"
)

// initConfig load config from default->file->cmdFlag
func initConfig() {
	// 加载配置文件，处理命令行参数
	err = common.ParseConfig("./cfg.ini")
	// 检查配置文件及命令行参数是否正确
	if common.CheckConfig && err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	common.LogIfWarn(err, "")
}

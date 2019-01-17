package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/XiaoMi/soar/advisor"
	"github.com/XiaoMi/soar/common"
	"github.com/XiaoMi/soar/database"
	"github.com/XiaoMi/soar/env"
)

var err error

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

func shutdown(vEnv *env.VirtualEnv, rEnv *database.Connector) {
	if common.Config.DropTestTemporary {
		vEnv.CleanUp()
	}
	err := vEnv.Conn.Close()
	common.LogIfWarn(err, "")
	err = rEnv.Conn.Close()
	common.LogIfWarn(err, "")
	os.Exit(0)
}

// initQuery
func initQuery(query string) string {
	// 读入待优化 SQL ，当配置文件或命令行参数未指定 SQL 时从管道读取
	if query == "" {
		// check stdin is pipe or terminal
		// https://stackoverflow.com/questions/22744443/check-if-there-is-something-to-read-on-stdin-in-golang
		stat, err := os.Stdin.Stat()
		if stat == nil {
			common.Log.Critical("os.Stdin.Stat Error: %v", err)
			os.Exit(1)
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// stdin is from a terminal
			fmt.Println("Args format error, use --help see how to use it!")
			os.Exit(1)
		}
		// read from pipe
		var data []byte
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			common.Log.Critical("ioutil.ReadAll Error: %v", err)
		}
		common.Log.Debug("initQuery get query from os.Stdin")
		return string(data)
	}

	if _, err := os.Stat(query); err == nil {
		var data []byte
		data, err = ioutil.ReadFile(query)
		if err != nil {
			common.Log.Critical("ioutil.ReadFile Error: %v", err)
		}
		common.Log.Debug("initQuery get query from file: %s", query)
		return string(data)
	}

	return query
}

// reportTool tools in report type
func reportTool(sql string, bom []byte) (isContinue bool, exitCode int) {
	switch common.Config.ReportType {
	case "html":
		// HTML 格式输入 CSS 加载
		fmt.Println(common.MarkdownHTMLHeader())
		return true, 0
	case "md2html":
		// markdown2html 转换小工具
		fmt.Println(common.MarkdownHTMLHeader())
		fmt.Println(common.Markdown2HTML(sql))
		return false, 0
	case "explain-digest":
		// 当用户输入为 EXPLAIN 信息，只对 Explain 信息进行分析
		// 注意： 这里只能处理一条 SQL 的 EXPLAIN 信息，用户一次反馈多条 SQL 的 EXPLAIN 信息无法处理
		advisor.DigestExplainText(sql)
		return false, 0
	case "chardet":
		// Get charset of input
		charset := common.CheckCharsetByBOM(bom)
		if charset == "" {
			charset = common.Chardet([]byte(sql))
		}
		fmt.Println(charset)
		return false, 0
	case "remove-comment":
		fmt.Println(database.RemoveSQLComments(sql))
		return false, 0
	default:
		return true, 0
	}
}

func verboseInfo() {
	if !common.Config.Verbose {
		return
	}
	// syntax check verbose mode, add output for success!
	if common.Config.OnlySyntaxCheck {
		fmt.Println("Syntax check OK!")
		return
	}
	switch common.Config.ReportType {
	case "markdown":
		if common.Config.TestDSN.Disable || common.Config.OnlineDSN.Disable {
			fmt.Println("MySQL environment verbose info")
			// TestDSN
			if common.Config.TestDSN.Disable {
				fmt.Println("* test-dsn:", common.Config.TestDSN.Addr, "is disable, please check log.")
			}
			// OnlineDSN
			if common.Config.OnlineDSN.Disable {
				fmt.Println("* online-dsn:", common.Config.OnlineDSN.Addr, "is disable, please check log.")
			}
		}
	}
}

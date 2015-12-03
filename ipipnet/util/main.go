package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/yangchenxing/foochow/ipipnet"
	"github.com/yangchenxing/foochow/logging"
)

var (
	dataPath = flag.String("data", "data/ipipnet.dat", "数据文件路径")
	idsPath  = flag.String("ids", "data/ipipnet.ids", "ID文件路径")
	url      = flag.String("url", "", "数据下载路径")
	interval = flag.Duration("interval", time.Minute, "检查更新周期")
)

func init() {
	ipipnet.UnknownPlaceCallback = func(text string) {
		logging.Warn("未知地点: %q", text)
	}
	ipipnet.UnknownISPCallback = func(text string) {
		logging.Warn("未知ISP: %q", text)
	}
}

func main() {
	flag.Parse()
	if err := ipipnet.Initialize(*idsPath, *dataPath, *url, *interval); err != nil {
		logging.Error("初始化IPIP.net失败: %s", err.Error())
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(ipipnet.Locate(net.ParseIP(scanner.Text())))
	}
}

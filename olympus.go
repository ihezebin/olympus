package olympus

import (
	"fmt"
	"time"

	"github.com/ihezebin/olympus/util"
)

func init() {
	now := time.Now()
	beforeTime := now.Format(time.DateTime)
	//默认初始化程序时区为东8区
	time.Local = util.DefaultTZ
	afterTime := now.In(time.Local).Format(time.DateTime)
	fmt.Printf("[Olympus] Now default TimeZone: %s, Set to %s\n", beforeTime, afterTime)
}

// UseUTC 支持显示调用，设置程序时区为UTC时区
func UseUTC() {
	time.Local = time.UTC
}

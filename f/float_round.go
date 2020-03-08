package f

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Round 四舍五入 小数点后 n 位
func Round(v float64, n int) float64 {
	shift := math.Pow(10, float64(n))
	f := 0.0000000001 + v // 对浮点数产生.xxx999999999 计算不准进行处理
	return math.Floor(f*shift+.5) / shift
}

// Floor 小数点后 n 位 - 舍去
func Floor(v float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n+1)+"f", v)
	temp := strings.Split(floatStr, ".")
	var newFloat string
	if len(temp) < 2 || n >= len(temp[1]) {
		newFloat = floatStr
	} else {
		newFloat = temp[0] + "." + temp[1][:n]
	}
	f, _ := strconv.ParseFloat(newFloat, 64)
	return f
}

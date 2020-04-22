package f

import (
	"os/user"
	"testing"
)

func TestPrint(t *testing.T) {
	// 当前系统登录用户
	usr, _ := user.Current()
	t.Log("\n", "usr.Username:", usr.Username, "\n", "usr.HomeDir:", usr.HomeDir)

	// 输出可打印字符
	t.Logf("%8c", 65)   // %c=Unicode字符
	t.Logf("%8x", 65)   // %x=16进制
	t.Logf("%#8o", 65)  // %x=8进制
	t.Logf("%#8x", 65)  // %x=16进制 补0双字节
	t.Logf("%08U", 'A') // %U=Unicode
	t.Logf("%08x", 'A') // %x=Hex 补0对齐字符
	t.Logf("%#U", '国')  // Unicode编码
	t.Logf("%+q", "祖国") // Ascii编码
}

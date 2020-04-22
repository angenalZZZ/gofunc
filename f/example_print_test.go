package f_test

import (
	"fmt"
	"os/user"
)

// go test -v -run ^ExamplePrintTest$ github.com/angenalZZZ/gofunc/f
func ExamplePrintTest() {
	// 当前系统登录用户
	usr, _ := user.Current()
	fmt.Println("01.usr.Username:", usr.Username)
	fmt.Println("02.usr.HomeDir:", usr.HomeDir)

	// 输出可打印字符
	fmt.Printf("03.%8c\n", 65)    // %c=Unicode字符
	fmt.Printf("04.%8x\n", 65)    // %x=16进制
	fmt.Printf("05.%#8o\n", 65)   // %x=8进制
	fmt.Printf("06.%#8x\n", 65)   // %x=16进制 补0双字节
	fmt.Printf("07.%08U\n", 'A')  // %U=Unicode
	fmt.Printf("08.%08x\n", 'A')  // %x=Hex 补0对齐字符
	fmt.Printf("09.%#U\n", '国')   // Unicode编码
	fmt.Printf("10.% 0x\n", "祖国") // 16进制 补空格
	fmt.Printf("11.%+q\n", "祖国")  // Ascii编码

	// Output:
	// 01.usr.Username: 0R0VMR1DWV35XWH\Administrator
	// 02.usr.HomeDir: C:\Users\Administrator
	// 03.       A
	// 04.      41
	// 05.    0101
	// 06.    0x41
	// 07.  U+0041
	// 08.00000041
	// 09.U+56FD '国'
	// 10.e7 a5 96 e5 9b bd
	// 11."\u7956\u56fd"
}

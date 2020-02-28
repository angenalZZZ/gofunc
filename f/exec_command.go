package f

import (
	"log"
	"os/exec"
	"runtime/debug"
)

// ExecCommandOutput 执行命令并获取输出结果.
func ExecCommandOutput(command string, arg ...string) (string, error) {
	out, err := exec.Command(command, arg...).Output()
	if err != nil {
		log.Println("callCommand failed!")
		log.Println("")
		log.Println(string(debug.Stack()))
		return "", err
	}
	return string(out), nil
}

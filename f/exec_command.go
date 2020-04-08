package f

import (
	"bufio"
	"log"
	"os/exec"
	"runtime/debug"
)

// ExecCommandOutput 执行命令后获取输出结果.
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

// ExecCommandOutputScanner 执行命令后一行一行的扫描输出结果.
func ExecCommandOutputScanner(lineScanner func([]byte) bool, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	bs := bufio.NewScanner(out)
	for bs.Scan() {
		line := bs.Bytes()
		if lineScanner(line) == false {
			break
		}
	}
	return bs.Err()
}

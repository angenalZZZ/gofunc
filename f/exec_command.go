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
	return String(out), nil
}

// ExecCommandOutputScanner 执行命令后一行一行的扫描输出结果.
func ExecCommandOutputScanner(command string, arg []string, lineScanner func([]byte) bool, splits ...bufio.SplitFunc) error {
	cmd := exec.Command(command, arg...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	bs := bufio.NewScanner(out)
	for _, split := range splits {
		bs.Split(split)
	}
	for bs.Scan() {
		line := bs.Bytes()
		if lineScanner(line) == false {
			break
		}
	}
	return bs.Err()
}

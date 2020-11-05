package js

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/angenalZZZ/gofunc/f"
)

// JobJs use cron job worker in javascript.
type JobJs struct {
	/*
		CRON Expression Format
		----------     ----------   --------------    --------------------------
		e.g. "@daily", "@hourly", "@every 1h30m",
			"30 * * * *"                     every hour on the half hour
			"30 3-6,20-23 * * *"             in the range 3-6am, 8-11pm
			"CRON_TZ=Asia/Tokyo 30 04 * * *" at 04:30 Tokyo time every day
		----------     ----------   --------------    --------------------------
		Field name   | Mandatory? | Allowed values  | Allowed special characters
		----------   | ---------- | --------------  | --------------------------
		Minutes      | Yes        | 0-59            | * / , -
		Hours        | Yes        | 0-23            | * / , -
		Day of month | Yes        | 1-31            | * / , - ?
		Month        | Yes        | 1-12 or JAN-DEC | * / , -
		Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?
		----------   | ---------- | --------------  | --------------------------
		特定字符的含义如下：
		* 表示匹配该域的任意值，假如在minute域使用*, 即表示每分钟都会触发事件
		/ 表示起始时间开始触发，然后每隔固定时间触发一次，例如在minute域使用5/20,则意味着在5分的时候开始触发一次，而25，45等分别触发一次
		, 表示列出枚举值。例如：在minute域使用5,20，则意味着在5和20分每分钟触发一次
		- 表示范围，例如在minute域使用5-20，表示从5分到20分每分钟触发一次
		? 字符仅被用于“日”和“周”两个子表达式，表示不指定值，当两个子表达式其中之一被指定了值以后，为了避免冲突，需要将另一个子表达式的值设为?
	*/
	Spec string
	// the javascript content
	Script string
	// the javascript filepath
	File string
	// the javascript file modify time
	FileModTime time.Time
	// the last run time
	LastRunTime time.Time
	// the job name
	Name string
}

// Init the javascript job.
func (j *JobJs) Init() error {
	if j.File != "" && !j.FileIsMod() {
		return os.ErrNotExist
	}
	if err := j.FileMod(); err != nil {
		return err
	}
	return nil
}

// Run implementation cron.Job interface.
func (j *JobJs) Run() {
	if Runtime == nil {
		return
	}

	j.LastRunTime = time.Now()
	fmt.Printf("%s run script job %q ", j.LastRunTime.Format("15:04:05.000"), j.Name)
	res, err := Runtime.RunString(j.Script)
	fmt.Printf("takes %s ", time.Now().Sub(j.LastRunTime))
	if err != nil {
		fmt.Printf("error: %v", err)
	} else if res != nil {
		fmt.Printf("return: %+v", res.Export())
	} else {
		fmt.Print("complete")
	}
	fmt.Println()

	return
}

// FileIsMod check javascript file is modify.
func (j *JobJs) FileIsMod() bool {
	if j.File == "" {
		return false
	}
	info, err := os.Stat(j.File)
	if os.IsNotExist(err) {
		return false
	}
	if t := info.ModTime(); t.Unix() != j.FileModTime.Unix() {
		j.FileModTime = t
		return true
	}
	return false
}

// FileMod read updated javascript file.
func (j *JobJs) FileMod() error {
	if j.File == "" {
		return os.ErrNotExist
	}
	script, err := f.ReadFile(j.File)
	if err != nil {
		return err
	}
	j.Script = strings.TrimSpace(string(script))
	return err
}

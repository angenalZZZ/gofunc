package js

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/dop251/goja"
)

// NewJobs create javascript jobs.
func NewJobs(r *GoRuntime, script string, parentName string, name string) ([]*JobJs, error) {
	var (
		filename    string
		fileModTime time.Time
		prev        = fmt.Sprintf("init %q jobs, ", parentName)
	)
	if r == nil {
		return nil, errors.New(prev + "js runtime object is not initialized")
	}
	if script == "" {
		return nil, errors.New(prev + "js script content can't be empty")
	}
	if strings.HasSuffix(script, ".js") {
		filename = script
		info, err := os.Stat(filename)
		if os.IsNotExist(err) || info.IsDir() {
			return nil, errors.New(prev + "js script file does not exist")
		}
		buf, err1 := f.ReadFile(filename)
		if err1 != nil {
			return nil, errors.New(prev + err1.Error())
		}
		script, fileModTime = strings.TrimSpace(string(buf)), info.ModTime()
	}
	if script == "" {
		return nil, errors.New(prev + "js script content can't be empty")
	}
	if parentName == "" {
		return nil, errors.New(prev + "js script var name can't be empty")
	}
	if _, err := r.RunString(script); err != nil {
		return nil, err
	}

	var newErr = func(errs ...string) error {
		err := prev + "load script error:"
		for _, s := range errs {
			err += fmt.Sprintf(" %+v", s)
		}
		return errors.New(err)
	}

	self := r.Runtime.Get(parentName)
	objs, ok := self.Export().([]interface{})
	if !ok {
		return nil, newErr(parentName, "must be an array")
	}

	jobs, find := make([]*JobJs, 0, len(objs)), name != ""
	for i, obj := range objs {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			return nil, newErr(parentName, "array item be an object")
		}

		item := new(JobJs)
		if item.Name, ok = objMap["name"].(string); !ok {
			return nil, newErr(parentName, fmt.Sprintf("array item[%d]'s name not found", i))
		} else if item.Name == "" {
			return nil, newErr(parentName, fmt.Sprintf("array item[%d]'s name can't be empty", i))
		}
		if item.Spec, ok = objMap["spec"].(string); !ok {
			return nil, newErr(parentName, fmt.Sprintf("array item[%d]'s spec not found", i))
		} else if item.Spec == "" {
			return nil, newErr(parentName, fmt.Sprintf("array item[%d]'s spec can't be empty", i))
		}
		if item.Func, ok = objMap["func"].(func(goja.FunctionCall) goja.Value); !ok {
			return nil, newErr(parentName, fmt.Sprintf("array item[%d]'s func not found", i))
		}

		item.Script, item.File, item.FileModTime = script, filename, fileModTime
		item.Self, item.ParentName = r.ToValue(obj), parentName

		if find {
			if name != item.Name {
				continue
			}
			jobs = append(jobs, item)
			break
		}

		jobs = append(jobs, item)
	}

	for _, item := range jobs {
		item.Parent = jobs
	}

	return jobs, nil
}

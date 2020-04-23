package app

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hyahm/golog"
)

type Remote struct {
	Islocal bool // 是否是本地用户
	Src     string
	Dst     string //
	Host    string
	Port    int
	User    string
	Owner   string
}

func (r *Remote) shell(command string) {
	var cmd *exec.Cmd
	golog.Info(command)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("/bin/bash", "-c", command)
	}
	out, err := cmd.Output()
	if err != nil {
		log.Println(string(out))
		golog.Error(err)
		log.Fatal(err)
	}
	log.Println(string(out))
}

func (r *Remote) makedir(path string) {
	golog.UpFunc()
	path = filepath.Join(r.Dst, path)
	path = strings.ReplaceAll(path, "\\", "/")
	command := fmt.Sprintf(`ssh -p %d %s@%s "mkdir -p %s"`, r.Port, r.User, r.Host, path)
	r.shell(command)
	if r.Owner != r.User {
		command := fmt.Sprintf(`ssh -p %d %s@%s "chown -R %s:%s %s"`, r.Port, r.User, r.Host, r.Owner, r.Owner, path)
		r.shell(command)
	}
}

func (r *Remote) copyfile(path string) {
	remotefile := filepath.Join(r.Dst, path)
	remotefile = strings.ReplaceAll(remotefile, "\\", "/")
	localfile := filepath.Join(r.Src, path)
	command := fmt.Sprintf(`scp -P %d %s %s@%s:%s`, r.Port, localfile, r.User, r.Host, remotefile)
	r.shell(command)
	if r.Owner != r.User {
		command := fmt.Sprintf(`ssh -p %d %s@%s "chown %s:%s %s"`, r.Port, r.User, r.Host, r.Owner, r.Owner, remotefile)
		r.shell(command)
	}
}

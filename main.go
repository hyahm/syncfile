package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"log"

	"github.com/hyahm/goconfig"
	"github.com/hyahm/golog"
)

const INTERVAL = 1 * time.Second

func (fo *Info) Load() error {
	file, err := os.Open(goconfig.ReadString("server.gob", "gob.txt"))
	if err != nil {
		golog.Error(err)
		return err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	err = dec.Decode(fo)
	if err != nil {
		log.Println("加载失败")
		golog.Error(err)
		return err
	}
	fmt.Println("已从文件加载")
	return nil
}

func (fo *Info) Save() {
	file, err := os.OpenFile(goconfig.ReadString("server.gob", "gob.txt"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		golog.Error(err)
		log.Fatal(err)
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	err = enc.Encode(fo)
	if err != nil {
		log.Println("保存失败， 数据将会重新传输")
		golog.Error(err)
		log.Fatal(err)
	}
}

var exitChan chan os.Signal

func exitHandle(fo *Info) {

	s := <-exitChan
	fmt.Println("收到退出信号", s)
	fo.Save()
	os.Exit(1) //如果ctrl+c 关不掉程序，使用os.Exit强行关掉
}

func main() {
	goconfig.InitConf("sync.ini", goconfig.INI)
	golog.InitLogger(goconfig.ReadString("log.path"), goconfig.ReadInt64("log.size"), goconfig.ReadBool("log.every"))
	src := goconfig.ReadString("server.src")
	dst := goconfig.ReadString("remote.dst")
	islocal := goconfig.ReadBool("remote.islocal")
	var r *Remote
	if !islocal {
		host := goconfig.ReadString("remote.host")
		if host == "" {
			log.Fatal("if islocal is false, host must be need")
		}
		r = &Remote{
			Islocal: islocal,
			Dst:     dst,
			Src:     src,
			Host:    host,
			Port:    goconfig.ReadInt("remote.port", 22),
			User:    goconfig.ReadString("remote.user", "root"),
		}
	} else {
		r = &Remote{
			Islocal: islocal,
			Dst:     dst,
			Src:     src,
		}
	}

	golog.Info(src)

	fo := NewInfo(src, r)
	golog.Info(fo.Local)
	checkPath(src)
	checkPath(dst)

	exitChan = make(chan os.Signal)
	signal.Notify(exitChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	go exitHandle(fo)

	// fo.Load()

	go fo.cleanfile()
	go fo.cleandir()
	golog.Infof("%s", fo.Local)
	for {
		fi, err := os.Stat(fo.Local)
		if err != nil {
			golog.Error(err)
			log.Fatal(err)
		}
		if fi.ModTime().UnixNano() != fo.Root {
			fo.LoopDir("")
			fo.Root = fi.ModTime().UnixNano()
		}

		time.Sleep(fo.Interval)
	}

}

type Remote struct {
	Islocal bool // 是否是本地用户
	Src     string
	Dst     string //
	Host    string
	Port    int
	User    string
}

// 这个dir 是相对目录
func (fo *Info) MakeDir(dir string) {
	dir = filepath.Join(fo.Dst.Dst, dir)
	golog.Infof("make dir %s", dir)
	fi, err := os.Stat(dir)
	if err == nil && fi.IsDir() {
		golog.Infof("have dir: %s", dir)
		golog.Error(err)
		return
	}
	if fo.Dst.Islocal {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			golog.Error(err)
			log.Fatal(err)
		}
	} else {

		fo.Dst.makedir(dir)
	}

}

func NewInfo(local string, remote *Remote, interval ...time.Duration) *Info {
	thisInterval := INTERVAL
	if len(interval) > 0 {
		thisInterval = interval[0]
	}

	golog.Info(local)
	return &Info{
		Interval: thisInterval,
		Local:    local,
		Dst:      remote,
		Dir:      make(map[string]int64),
		File:     make(map[string]int64),
	}
}

type Info struct {
	Interval time.Duration
	Local    string
	Dst      *Remote
	Dir      map[string]int64 // 保存文件夹最后修改时间
	Root     int64            // 根文件夹时间
	File     map[string]int64 // 所有的文件 信息
}

func checkPath(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		golog.Error(err)
		os.MkdirAll(path, 0755)
		return
	}

	if !fi.IsDir() {
		golog.Error(err)
		log.Fatalf("%s is a file", path)
	}
}

// 需要上传的目录
func (fo *Info) LoopDir(dir string) {
	golog.Info("start")
	thisDir := filepath.Join(fo.Local, dir)
	ll, err := ioutil.ReadDir(thisDir)
	if err != nil {
		golog.Error(err)
		log.Fatal(err)
	}
	for _, v := range ll {
		newname := strings.ReplaceAll(v.Name(), " ", "")
		newname = strings.ReplaceAll(newname, "\\", "")
		path := filepath.Join(fo.Local, dir, newname)
		middir := filepath.Join(dir, newname)
		abcdir := filepath.Join(thisDir, v.Name())
		os.Rename(abcdir, path)

		// 去掉所有文件名的空格和\

		if v.IsDir() {
			thismtime := v.ModTime().UnixNano()
			if mtime, ok := fo.Dir[path]; ok {
				// 如果存在的话， 检查修改时间是否大于 上次的修改时间

				if thismtime > mtime {
					// 如果大于的话， 就要遍历里面的文件夹， 修改当前的mtime
					fo.LoopDir(middir)
					fo.Dir[path] = thismtime
				}
			} else {
				// 增加文件夹， 并在远程建立远程文件夹
				golog.Info("make dir ", path)
				fo.MakeDir(middir)

				fo.LoopDir(middir)
				fo.Dir[path] = thismtime
			}
		} else {

			golog.Info(path)
			now := time.Now().UnixNano()
			if _, ok := fo.File[path]; !ok {
				// 如果文件不存在， 并且是已完成的， 传输文件， 并且
				// 判断已完成的
				if now > v.ModTime().UnixNano()+2 {
					golog.Infof("copy file %s", middir)
					fo.CopyFile(middir, v.ModTime().UnixNano())
				}
			}

		}
	}
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
}

func (fo *Info) CopyFile(path string, mtime int64) {
	golog.Info(path)
	remotefile := filepath.Join(fo.Dst.Dst, path)
	localfile := filepath.Join(fo.Local, path)
	if fo.Dst.Islocal {
		fo, _ := os.Open(localfile)
		defer fo.Close()
		fi, err := os.OpenFile(remotefile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			golog.Error(err)
			log.Fatal(err)
		}
		defer fi.Close()
		_, err = io.Copy(fi, fo)
		if err != nil {
			golog.Error(err)
			log.Fatal(err)
		}
	} else {
		fo.Dst.copyfile(path)
	}
	fo.File[localfile] = mtime

}

func (r *Remote) copyfile(path string) {
	remotefile := filepath.Join(r.Dst, path)
	remotefile = strings.ReplaceAll(remotefile, "\\", "/")
	localfile := filepath.Join(r.Src, path)
	command := fmt.Sprintf(`scp -P %d %s %s@%s:%s`, r.Port, localfile, r.User, r.Host, remotefile)
	r.shell(command)
}

func (fo *Info) cleanfile() {
	for {

		for k, _ := range fo.File {
			if _, err := os.Stat(k); os.IsNotExist(err) {
				golog.Infof("file : %s is delete", k)
				delete(fo.File, k)
			}
		}
	}

}

func (fo *Info) cleandir() {
	for {

		for k, _ := range fo.Dir {
			if _, err := os.Stat(k); os.IsNotExist(err) {

				golog.Infof("dir : %s is delete", k)
				delete(fo.Dir, k)
			}
		}
	}

}

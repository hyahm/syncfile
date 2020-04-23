package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	"github.com/hyahm/goconfig"
	"github.com/hyahm/golog"
	"github.com/hyahm/syncfile/app"
)

// 检测时间间隔， 加大应该会减少cpu资源
const INTERVAL = 1 * time.Second

var exitChan chan os.Signal

func exitHandle(fo *app.Info) {

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
	var r *app.Remote
	if !islocal {
		host := goconfig.ReadString("remote.host")
		if host == "" {
			log.Fatal("if islocal is false, host must be need")
		}
		r = &app.Remote{
			Islocal: islocal,
			Dst:     dst,
			Src:     src,
			Host:    host,
			Port:    goconfig.ReadInt("remote.port", 22),
			User:    goconfig.ReadString("remote.user", "root"),
			Owner:   goconfig.ReadString("remote.owner", "root"),
		}
	} else {
		r = &app.Remote{
			Islocal: islocal,
			Dst:     dst,
			Src:     src,
			User:    goconfig.ReadString("remote.user", "root"),
			Owner:   goconfig.ReadString("remote.owner", "root"),
		}
	}

	golog.Info(src)
	interval := INTERVAL
	if goconfig.ReadUint64("server.interval", 0) > 0 {
		interval = time.Duration(goconfig.ReadUint64("server.interval", 0)) * time.Second
	}
	fo := app.NewInfo(src, r, interval)
	golog.Info(fo.Local)
	checkPath(src)
	checkPath(dst)

	exitChan = make(chan os.Signal)
	signal.Notify(exitChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	go exitHandle(fo)

	if goconfig.ReadBool("server.load", true) {
		fo.Load()
	}

	go fo.Cleanfile()
	go fo.Cleandir()
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

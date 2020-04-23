package app

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyahm/goconfig"
	"github.com/hyahm/golog"
)

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
			if len(fo.Include) > 0 {
				for _, is := range fo.Include {
					if strings.Contains(newname, is) {
						now := time.Now().UnixNano()
						if _, ok := fo.File[path]; !ok {
							// 如果文件不存在， 并且是已完成的， 传输文件， 并且
							// 判断已完成的
							if now > v.ModTime().UnixNano()+2 {
								golog.Infof("copy file %s", middir)
								fo.CopyFile(middir, v.ModTime().UnixNano())
							}
						}
						continue
					}
				}
			} else {
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

func (fo *Info) Cleanfile() {
	for {

		for k, _ := range fo.File {
			if _, err := os.Stat(k); os.IsNotExist(err) {
				golog.Infof("file : %s is delete", k)
				delete(fo.File, k)
			}
		}
		time.Sleep(fo.Interval)
	}

}

func (fo *Info) Cleandir() {
	for {

		for k, _ := range fo.Dir {
			if _, err := os.Stat(k); os.IsNotExist(err) {

				golog.Infof("dir : %s is delete", k)
				delete(fo.Dir, k)
			}
		}
		time.Sleep(fo.Interval)
	}

}

func NewInfo(local string, remote *Remote, interval time.Duration) *Info {

	golog.Info(local)

	includes := make([]string, 0)

	ib := goconfig.ReadBytes("server.include")
	err := json.Unmarshal(ib, &includes)
	if err != nil {
		golog.Error(err)
		log.Fatal(err)
	}

	return &Info{
		Interval: interval,
		Local:    local,
		Dst:      remote,
		Dir:      make(map[string]int64),
		File:     make(map[string]int64),
		Include:  includes,
	}
}

type Info struct {
	Interval time.Duration
	Local    string
	Dst      *Remote
	Dir      map[string]int64 // 保存文件夹最后修改时间
	Root     int64            // 根文件夹时间
	File     map[string]int64 // 所有的文件 信息
	Include  []string
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

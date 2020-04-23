package app

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/hyahm/goconfig"
	"github.com/hyahm/golog"
)

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

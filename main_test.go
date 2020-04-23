package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestModifyTime(t *testing.T) {
	for i := 0; i < 10000; i++ {
		f, err := os.Stat("zy_pyr0hHad_180105.mp4")
		if err != nil {
			t.Error(f)
			os.Exit(2)
		}
		fmt.Println(f.ModTime().UnixNano() / 1000)
		time.Sleep(2 * time.Millisecond)
	}

}

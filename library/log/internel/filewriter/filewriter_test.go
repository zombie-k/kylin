package filewriter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileLogWriter_daemon(t *testing.T) {
}

func TestNewFileLogWriter(t *testing.T) {
	fpath := "test.log"
	flw, err := NewLogFileWriter(fpath, SetRotateMinutely(true), SetRotateInterval(time.Second*10), SetChanSize(1024*8))
	if err != nil {
		t.Fatal(err)
	}
	defer flw.Close()

	for i := 0; i < 1000; i++ {
		_, err = flw.Write([]byte("Hello World!\n"))
		time.Sleep(time.Second)
		if err != nil {
			t.Fatal(err)
		}
	}
	//time.Sleep(time.Second * 5)
}

func TestOpenFile(t *testing.T) {
	fPath := "/Users/user/test/test.log"
	fName := filepath.Base(fPath)
	if fName == "" {
		fmt.Println("filename can't empty")
	}
	dir := filepath.Dir(fPath)
	fi, err := os.Stat(dir)
	if err == nil && !fi.IsDir() {
		fmt.Printf("%s already exists and not a directory\n", dir)
	}
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			fmt.Errorf("create dir %s error: %s", dir, err.Error())
		}
	}
	fmt.Println(fi, err)
	fmt.Println("filename:", fName, "dir:", dir)

	file, err := os.OpenFile(fPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	fi, err = file.Stat()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fi, err)
}

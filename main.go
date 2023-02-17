package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var scanner = bufio.NewScanner(os.Stdin)

func main() {
	fi := new(FileInst)
	if len(os.Args) > 1 {
		fi.OriginalPath = os.Args[1]
		fi.Directory = filepath.Dir(fi.OriginalPath)
		fi.FileName = filepath.Base(fi.OriginalPath)
	} else {
		fmt.Println("请指定文件路径，按回车确认：")
		scanner.Scan()
		fi.OriginalPath = scanner.Text()
		fi.Directory = filepath.Dir(fi.OriginalPath)
		fi.FileName = filepath.Base(fi.OriginalPath)
	}

	if err := fi.Init(); err != nil {
		fmt.Println(err)
		exitGracefully(2)
	}

	err, data := fi.Unmarshal()
	if err != nil {
		fmt.Println(err)
		exitGracefully(1)
	}

	if err := fi.Convert(data); err != nil {
		fmt.Println(err)
		exitGracefully(1)
	}

	exitGracefully(0)
}

func exitGracefully(code int) {
	switch runtime.GOOS {
	case "darwin":
		os.Exit(code)
	case "windows":
		fmt.Println("按下任意键退出……")
		reader := bufio.NewReader(os.Stdin)
		_, _, _ = reader.ReadRune()
		os.Exit(code)
	default:
		os.Exit(code)
	}
}

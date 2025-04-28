// main.go
package main

import (
	"Project/component/File_Utils"
	"Project/component/Public_Data"
	"fmt"
	"os"
)

func main() {
	var err error
	Public_Data.Dir, err = File_Utils.GetWorkpath()
	if err != nil {
		fmt.Println("获取当前工作目录失败:", err)
		return
	}
	fmt.Println("当前工作目录:", Public_Data.Dir)

	if err := File_Utils.CreateDirectories(Public_Data.Dir); err != nil {
		fmt.Println(err)
		return
	}

	Public_Data.Windir = File_Utils.InputPath()
	fmt.Println("输入路径:", Public_Data.Windir)  // 直接使用Windows路径，不进行转换

	if _, err := os.Stat(Public_Data.Windir); os.IsNotExist(err) {
		fmt.Printf("路径不存在: %s\n", Public_Data.Windir)
		return
	}

	if err := File_Utils.CopyMatchingSLXFiles(Public_Data.Windir); err != nil {
		fmt.Println("复制文件时发生错误:", err)
		return
	}

	if err := File_Utils.ProcessSLXFiles(); err != nil {
		fmt.Println("处理文件时发生错误:", err)
		return
	}
}
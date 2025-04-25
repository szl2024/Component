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
	// 使用包名前缀调用函数
	Public_Data.Dir, err = File_Utils.GetWorkpath()
	if err != nil {
		fmt.Println("获取当前工作目录失败:", err)
		return
	}
	fmt.Println("当前工作目录:", Public_Data.Dir)

	// 使用修改后的函数名
	if err := File_Utils.CreateDirectories(Public_Data.Dir); err != nil {
		fmt.Println(err)
		return
	}

	Public_Data.Windir = File_Utils.InputPath()
	Public_Data.Wsldir, err = File_Utils.ChangePath(Public_Data.Windir)
	if err != nil {
		fmt.Println("路径转换失败:", err)
		return
	}
	fmt.Println("转换后路径:", Public_Data.Wsldir)

	if _, err := os.Stat(Public_Data.Wsldir); os.IsNotExist(err) {
		fmt.Printf("路径不存在: %s\n", Public_Data.Wsldir)
		return
	}

	if err := File_Utils.CopyMatchingSLXFiles(Public_Data.Wsldir); err != nil {
		fmt.Println("复制文件时发生错误:", err)
		return
	}

	if err := File_Utils.ProcessSLXFiles(); err != nil {
		fmt.Println("处理文件时发生错误:", err)
		return
	}

}
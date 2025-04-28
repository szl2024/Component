// File_Utils.go
package File_Utils

import (
	"Project/component/Public_Data" // 正确导入 Public_Data 包
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"archive/zip"
	"Project/component/Root_System_Analysis" 
)

// GetWorkpath 获取工作目录
func GetWorkpath() (string, error) {
	var err error
	Public_Data.Dir, err = os.Getwd() // 直接赋值给 Public_Data.Dir
	if err != nil {
		return "", err
	}
	return Public_Data.Dir, nil
}

// CreateDirectories 创建工作目录
func CreateDirectories(workdir string) error {
	// 使用 Public_Data 中定义的共享变量
	Public_Data.BuildDir = filepath.Join(workdir, "build")
	Public_Data.OutputDir = filepath.Join(workdir, "output")

	// 删除已存在的目录
	if _, err := os.Stat(Public_Data.BuildDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(Public_Data.BuildDir); err != nil {
			return fmt.Errorf("删除 build 目录失败: %v", err)
		}
		fmt.Println("删除 build 目录成功")
	}

	if _, err := os.Stat(Public_Data.OutputDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(Public_Data.OutputDir); err != nil {
			return fmt.Errorf("删除 output 目录失败: %v", err)
		}
		fmt.Println("删除 output 目录成功")
	}

	// 创建新的 build 和 output 目录
	if err := os.Mkdir(Public_Data.BuildDir, 0755); err != nil {
		return fmt.Errorf("创建 build 目录失败: %v", err)
	}
	fmt.Println("创建 build 目录成功:", Public_Data.BuildDir)

	if err := os.Mkdir(Public_Data.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建 output 目录失败: %v", err)
	}
	fmt.Println("创建 output 目录成功:", Public_Data.OutputDir)

	return nil
}

// InputPath 输入路径
func InputPath() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入一个目录路径: ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// ChangePath 路径转换


// CopyMatchingSLXFiles 复制匹配的 SLX 文件
func CopyMatchingSLXFiles(sourceDir string) error {
	subDirs, err := getSubDirectories(sourceDir)
	if err != nil {
		return err
	}

	targetDir := filepath.Join(Public_Data.Dir, "build")

	// 遍历子目录并复制文件
	for _, dirName := range subDirs {
		fileName := dirName + ".slx"
		srcPath := filepath.Join(sourceDir, dirName, fileName)
		dstPath := filepath.Join(targetDir, fileName)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("文件不存在: %s\n", srcPath)
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			fmt.Printf("复制失败 %s: %v\n", fileName, err)
			continue
		}
		fmt.Printf("成功复制到: %s\n", dstPath)

		txtFileName := strings.TrimSuffix(fileName, ".slx") + ".txt"
		txtFilePath := filepath.Join(Public_Data.OutputDir, txtFileName)

		if err := createEmptyTxtFile(txtFilePath); err != nil {
			fmt.Printf("创建txt文件失败: %v\n", err)
			continue
		}
		fmt.Printf("成功创建txt文件: %s\n", txtFilePath)
	}
	return nil
}

// createEmptyTxtFile 创建空的 txt 文件
func createEmptyTxtFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// getSubDirectories 获取子目录
func getSubDirectories(path string) ([]string, error) {
	var dirs []string
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

// unzip 解压 zip 文件
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}

		outFile.Close()
		rc.Close()
	}
	return nil
}

// ProcessSLXFiles 处理 SLX 文件
func ProcessSLXFiles() error {
	files, err := ioutil.ReadDir(Public_Data.BuildDir)
	if err != nil {
		return fmt.Errorf("读取build目录失败: %v", err)
	}

	// 遍历并处理文件
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".slx") {
			continue
		}

		baseName := strings.TrimSuffix(file.Name(), ".slx")
		fmt.Printf("\n=== 开始处理模型: %s ===\n", baseName)

		oldPath := filepath.Join(Public_Data.BuildDir, file.Name())
		newPath := filepath.Join(Public_Data.BuildDir, baseName+".zip")

		if err := os.Rename(oldPath, newPath); err != nil {
			fmt.Printf("重命名失败: %v\n", err)
			continue
		}

		extractDir := filepath.Join(Public_Data.BuildDir, baseName)
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
			continue
		}

		if err := unzip(newPath, extractDir); err != nil {
			fmt.Printf("解压失败: %v\n", err)
			continue
		}

		if err := os.Remove(newPath); err != nil {
			fmt.Printf("删除zip文件失败: %v\n", err)
		}
		systemXMLPath := filepath.Join(extractDir, "simulink", "systems", "system_root.xml")
		
		if _, err := os.Stat(systemXMLPath); err == nil {
			fmt.Printf("[分析] 开始分析主系统文件: %s\n", systemXMLPath)
			if err := Root_System_Analysis.AnalyzeRootXML(systemXMLPath, baseName); err != nil {
				fmt.Printf("[分析失败] %v\n", err)
			}
		} else {
			fmt.Printf("[警告] 未找到系统文件: %s\n", systemXMLPath)
		}
		fmt.Printf("=== 完成处理模型: %s ===\n", baseName)
	}
	return nil
}
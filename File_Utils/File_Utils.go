// File_Utils.go
package File_Utils

import (
	"Project/component/Public_Data" 
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

//  Get working directory
func GetWorkpath() (string, error) {
	var err error
	Public_Data.Dir, err = os.Getwd() // Directly assign to Public_Sata.Dir
	if err != nil {
		return "", err
	}
	return Public_Data.Dir, nil
}

// Create a working directory
func CreateDirectories(workdir string) error {
	// Using shared variables defined in Public_Sata
	Public_Data.BuildDir = filepath.Join(workdir, "build")
	Public_Data.OutputDir = filepath.Join(workdir, "output")

	// Delete existing directories
	if _, err := os.Stat(Public_Data.BuildDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(Public_Data.BuildDir); err != nil {
			return fmt.Errorf("Failed to delete build directory: %v", err)
		}
		fmt.Println("Successfully deleted build directory")
	}

	if _, err := os.Stat(Public_Data.OutputDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(Public_Data.OutputDir); err != nil {
			return fmt.Errorf("Failed to delete output directory: %v", err)
		}
		fmt.Println("Successfully deleted output directory")
	}

	// Create new build and output directories
	if err := os.Mkdir(Public_Data.BuildDir, 0755); err != nil {
		return fmt.Errorf("Failed to create build directory: %v", err)
	}
	fmt.Println("Successfully created build directory:", Public_Data.BuildDir)

	if err := os.Mkdir(Public_Data.OutputDir, 0755); err != nil {
		return fmt.Errorf("Failed to create output directory: %v", err)
	}
	fmt.Println("Successfully created output directory:", Public_Data.OutputDir)

	return nil
}

// enter path 
func InputPath() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter a directory path: ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}


// Copy matching SLX files
func CopyMatchingSLXFiles(sourceDir string) error {
	subDirs, err := getSubDirectories(sourceDir)
	if err != nil {
		return err
	}

	targetDir := filepath.Join(Public_Data.Dir, "build")

	// Traverse subdirectories and copy files
	for _, dirName := range subDirs {
		fileName := dirName + ".slx"
		srcPath := filepath.Join(sourceDir, dirName, fileName)
		dstPath := filepath.Join(targetDir, fileName)

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("File does not exist: %s\n", srcPath)
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			fmt.Printf("copy failed %s: %v\n", fileName, err)
			continue
		}
		fmt.Printf("Successfully copied to: %s\n", dstPath)

		txtFileName := strings.TrimSuffix(fileName, ".slx") + ".txt"
		txtFilePath := filepath.Join(Public_Data.OutputDir, txtFileName)

		if err := createEmptyTxtFile(txtFilePath); err != nil {
			fmt.Printf("Failed to create txt file: %v\n", err)
			continue
		}
		fmt.Printf("Successfully created txt file: %s\n", txtFilePath)
	}
	return nil
}

// Create an empty txt file
func createEmptyTxtFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("An error occurred while closing the file: %v\n", cerr)
		}
	}()
	
	return nil
}

// copies files
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

// Get subdirectories
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


// Extract zip file
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

// Processing SLX files
func ProcessSLXFiles() error {
	files, err := ioutil.ReadDir(Public_Data.BuildDir)
	if err != nil {
		return fmt.Errorf("Failed to read build directory: %v", err)
	}

	// Traverse and process files
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".slx") {
			continue
		}

		baseName := strings.TrimSuffix(file.Name(), ".slx")
		fmt.Printf("\n=== Start processing the model: %s ===\n", baseName)

		oldPath := filepath.Join(Public_Data.BuildDir, file.Name())
		newPath := filepath.Join(Public_Data.BuildDir, baseName+".zip")

		if err := os.Rename(oldPath, newPath); err != nil {
			fmt.Printf("Rename failed: %v\n", err)
			continue
		}

		extractDir := filepath.Join(Public_Data.BuildDir, baseName)
		if err := os.MkdirAll(extractDir, 0755); err != nil {
			fmt.Printf("Failed to create directory: %v\n", err)
			continue
		}

		if err := unzip(newPath, extractDir); err != nil {
			fmt.Printf("Decompression failed: %v\n", err)
			continue
		}

		if err := os.Remove(newPath); err != nil {
			fmt.Printf("Failed to delete zip file: %v\n", err)
		}
		systemXMLPath := filepath.Join(extractDir, "simulink", "systems", "system_root.xml")
		
		if _, err := os.Stat(systemXMLPath); err == nil {
			fmt.Printf("[Analysis] Start analyzing the main system files: %s\n", systemXMLPath)
			if err := Root_System_Analysis.AnalyzeRootXML(systemXMLPath, baseName); err != nil {
				fmt.Printf("[Analysis failed] %v\n", err)
			}
		} else {
			fmt.Printf("[Warning] System files not found: %s\n", systemXMLPath)
		}
		fmt.Printf("=== Complete the processing model: %s ===\n", baseName)
	}
	return nil
}
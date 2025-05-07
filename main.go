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
		fmt.Println("Failed to retrieve the current working directory:", err)
		return
	}
	fmt.Println("Current working directory:", Public_Data.Dir)

	if err := File_Utils.CreateDirectories(Public_Data.Dir); err != nil {
		fmt.Println(err)
		return
	}

	Public_Data.Windir = File_Utils.InputPath()
	fmt.Println("Enter path:", Public_Data.Windir)  

	if _, err := os.Stat(Public_Data.Windir); os.IsNotExist(err) {
		fmt.Printf("Path does not exist: %s\n", Public_Data.Windir)
		return
	}

	if err := File_Utils.CopyMatchingSLXFiles(Public_Data.Windir); err != nil {
		fmt.Println("An error occurred while copying the file:", err)
		return
	}

	if err := File_Utils.ProcessSLXFiles(); err != nil {
		fmt.Println("An error occurred while processing the file:", err)
		return
	}
}
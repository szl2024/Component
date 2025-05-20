package LDI_Generator

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type LDI struct {
	XMLName  xml.Name   `xml:"ldi"`
	Elements []*Element `xml:"element"`
}

type Element struct {
	XMLName  xml.Name  `xml:"element"`
	Name     string    `xml:"name,attr"`
	Uses     []Uses    `xml:"uses,omitempty"`
	Property *Property `xml:"property,omitempty"`
}

type Uses struct {
	Provider string `xml:"provider,attr"`
	Strength int    `xml:"strength,attr"`
}

type Property struct {
	Name  string `xml:"name,attr"`
	Value int    `xml:",chardata"`
}

var subsystemConnectionRegex = regexp.MustCompile(`(📦|🏷️)\s*([\w\-\(\)\s]+)\s+\(SID:\s*\d+\)\s+→\s+(📦|🏷️)\s*([\w\-\(\)\s]+)\s+\(SID:\s*\d+\)`)
var systemRegex = regexp.MustCompile(`System:\s+([\w\-]+)\s+\(SID:`)
var systemBlockRegex = regexp.MustCompile(`[📦🏷️]+System:\s+([\w\-]+)\s+\(SID:`)
var m1Regex = regexp.MustCompile(`M1:\s+(\d+)`)

func GenerateLDIFilesFromTxt(txtDir string, ldiDir string) error {
	files, err := ioutil.ReadDir(txtDir)
	if err != nil {
		return fmt.Errorf("Failed to read txt directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".txt" {
			continue
		}

		inputPath := filepath.Join(txtDir, file.Name())
		outputPath := filepath.Join(ldiDir, strings.TrimSuffix(file.Name(), ".txt")+".ldi.xml")
		processFile(inputPath, outputPath)
	}
	return nil
}

func processFile(inputPath, outputPath string) {
	f, err := os.Open(inputPath)
	if err != nil {
		fmt.Println("Failed to open input:", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	unitIndent := detectIndentUnit(inputPath)
	fullName := make(map[string]string)
	parentMap := make(map[string]string)
	order := []string{}
	systemSet := make(map[string]bool)

	stack := []string{}
	stackLevels := []int{}

	// 第一次扫描：收集系统层级结构
	for scanner.Scan() {
		line := scanner.Text()
		indentSpaces := len(line) - len(strings.TrimLeft(line, " "))
		level := indentSpaces / unitIndent

		if matches := subsystemConnectionRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			continue
		}

		if match := systemRegex.FindStringSubmatch(line); len(match) > 1 {
			shortName := match[1]
			full := shortName

			for len(stackLevels) > 0 && stackLevels[len(stackLevels)-1] >= level {
				stack = stack[:len(stack)-1]
				stackLevels = stackLevels[:len(stackLevels)-1]
			}

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				full = fullName[parent] + "." + shortName
				parentMap[shortName] = parent
			}

			fullName[shortName] = full
			order = append(order, shortName)

			stack = append(stack, shortName)
			stackLevels = append(stackLevels, level)

			systemSet[shortName] = true
		}
	}

	// 第二次扫描：提取所有系统对应 M1 值
	f.Seek(0, 0)
	scanner = bufio.NewScanner(f)
	currentSystem := ""
	m1Map := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()

		if match := systemBlockRegex.FindStringSubmatch(line); len(match) > 1 {
			currentSystem = match[1]
		}

		if m1Match := m1Regex.FindStringSubmatch(line); len(m1Match) > 1 && currentSystem != "" {
			var m1Value int
			fmt.Sscanf(m1Match[1], "%d", &m1Value)
			m1Map[currentSystem] = m1Value
		}
	}

	// 初始化 Elements
	elements := make(map[string]*Element)
	for _, short := range order {
		full := fullName[short]
		elem := &Element{Name: full}
		if m1, ok := m1Map[short]; ok {
			elem.Property = &Property{
				Name:  "coverage.m1",
				Value: m1,
			}
		}
		elements[full] = elem
	}

	// 第三次扫描：处理连接和Strength
	f.Seek(0, 0)
	scanner = bufio.NewScanner(f)
	usesCount := make(map[string]map[string]int)

	for scanner.Scan() {
		line := scanner.Text()

		matches := subsystemConnectionRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			srcShort := strings.TrimSpace(m[2])
			dstShort := strings.TrimSpace(m[4])

			if strings.Contains(dstShort, "(Block)") {
				continue
			}
			if _, exists := fullName[dstShort]; !exists {
				continue
			}

			srcFull := fullName[srcShort]
			dstFull := fullName[dstShort]
			if srcFull == "" {
				srcFull = srcShort
			}
			if dstFull == "" {
				dstFull = dstShort
			}

			systemSet[srcShort] = true
			systemSet[dstShort] = true

			if usesCount[srcFull] == nil {
				usesCount[srcFull] = make(map[string]int)
			}
			usesCount[srcFull][dstFull]++
		}
	}

	// 更新 uses 信息
	for src, targets := range usesCount {
		if elements[src] == nil {
			elements[src] = &Element{Name: src}
		}
		for dst, count := range targets {
			elements[src].Uses = append(elements[src].Uses, Uses{
				Provider: dst,
				Strength: count,
			})
		}
	}

	// 补充所有系统（即使没有 uses）
	for sys := range systemSet {
		full := fullName[sys]
		if full == "" {
			full = sys
		}
		if elements[full] == nil {
			elem := &Element{Name: full}
			if m1, ok := m1Map[sys]; ok {
				elem.Property = &Property{
					Name:  "coverage.m1",
					Value: m1,
				}
			}
			elements[full] = elem
		}
	}

	// 输出 XML
	outputList := []*Element{}
	written := make(map[string]bool)

	for _, short := range order {
		full := fullName[short]
		if elem := elements[full]; elem != nil && !written[full] {
			outputList = append(outputList, elem)
			written[full] = true
		}
	}
	for full, elem := range elements {
		if !written[full] {
			outputList = append(outputList, elem)
			written[full] = true
		}
	}

	outputLDI := LDI{Elements: outputList}
	xmlData, err := xml.MarshalIndent(outputLDI, "", "  ")
	if err != nil {
		fmt.Println("Marshal error:", err)
		return
	}

	xmlData = append([]byte(xml.Header), xmlData...)
	err = ioutil.WriteFile(outputPath, xmlData, 0644)
	if err != nil {
		fmt.Println("Write output error:", err)
		return
	}

	fmt.Println("✅ Generated:", outputPath)
}

func detectIndentUnit(filePath string) int {
	f, err := os.Open(filePath)
	if err != nil {
		return 2
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, " ") {
			count := len(line) - len(strings.TrimLeft(line, " "))
			if count > 0 {
				return count
			}
		}
	}
	return 2
}

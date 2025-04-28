package Root_System_Analysis

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"Project/component/Public_Data" 
	"Project/component/Subsystem_Analysis" 
	"path/filepath"
)

// AnalyzeRootXML 分析根 XML 文件
func AnalyzeRootXML(xmlPath string, modelName string) error {
	file, err := os.Open(xmlPath)
	if err != nil {
		return fmt.Errorf("读取XML文件失败: %v", err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	var (
		inSubsystem      bool
		hasValidPorts    bool
		currentSystemRef string
		blockName        string
		blockSID         int
		portCountsIn     int
		portCountsOut    int
		portsParam       string
	)

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Block" {
				// 重置状态
				inSubsystem = false
				hasValidPorts = false
				currentSystemRef = ""
				blockName = ""
				blockSID = 0
				portCountsIn = 0
				portCountsOut = 0
				portsParam = ""

				// 收集基础属性
				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "Name":
						blockName = attr.Value
					case "SID":
						if sid, err := strconv.Atoi(attr.Value); err == nil {
							blockSID = sid
						}
					case "BlockType":
						if attr.Value == "SubSystem" {
							inSubsystem = true
						}
					}
				}
			}

			if inSubsystem {
				switch se.Name.Local {
				case "PortCounts":
					hasValidPorts = false
					for _, attr := range se.Attr {
						switch attr.Name.Local {
						case "in":
							if in, err := strconv.Atoi(attr.Value); err == nil {
								portCountsIn = in
								hasValidPorts = true
							}
						case "out":
							if out, err := strconv.Atoi(attr.Value); err == nil {
								portCountsOut = out
								hasValidPorts = true
							}
						}
					}
				case "P":
					for _, attr := range se.Attr {
						if attr.Name.Local == "Name" && attr.Value == "Ports" {
							portsParam = "processing"
						}
					}
				case "System":
					for _, attr := range se.Attr {
						if attr.Name.Local == "Ref" {
							currentSystemRef = attr.Value
						}
					}
				}
			}

		case xml.CharData:
			if inSubsystem && portsParam == "processing" {
				portsStr := strings.TrimSpace(string(se))
				if in, out, err := parsePortsParam(portsStr); err == nil {
					portCountsIn = in
					portCountsOut = out
					hasValidPorts = true
				}
				portsParam = ""
			}

		case xml.EndElement:
			if se.Name.Local == "Block" && inSubsystem {
				if hasValidPorts && currentSystemRef != "" {
					// 创建有效系统实例
					currentSystem := Public_Data.NewSystemFromBlock(
						blockName,
						blockSID,
						portCountsIn,
						portCountsOut,
					)
					// fmt.Printf("  创建有效系统 | SID:%d | Name:%-20s | 输入:%-2d 输出:%-2d\n",
					// 	currentSystem.SID, currentSystem.Name, currentSystem.Inputs, currentSystem.Outputs)

					// 生成路径并递归分析
					newPath := generateNewPath(xmlPath, currentSystemRef)
					//fmt.Printf("  生成路径: %s\n", newPath)
					modelName := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(xmlPath))))
					Subsystem_Analysis.AnalyzeSubSystemXML(newPath, currentSystem, modelName)

				}
				inSubsystem = false
			}
		}
	}
	return nil
}

// parsePortsParam 解析Ports参数
func parsePortsParam(portsStr string) (int, int, error) {
	trimmed := strings.Trim(strings.TrimSpace(portsStr), "[]")
	if trimmed == "" {
		return 0, 0, fmt.Errorf("invalid ports format")
	}

	parts := strings.Split(trimmed, ",")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid ports format")
	}

	in, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	out, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("parse ports failed")
	}

	return in, out, nil
}

// generateNewPath 路径生成函数
func generateNewPath(originalPath, systemRef string) string {
	return filepath.Join(
		filepath.Dir(originalPath),
		systemRef+".xml",
	)
}
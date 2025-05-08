package Subsystem_Analysis

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"Project/component/Public_Data"
	"Project/component/Port_Analysis"
	"bufio"
	"Project/component/Connect_Analysis"
)

const portNameWidth = 35
const blockNameWidth = 35
const blockInfoWidth = 40 // 控制括号内内容的对齐长度

type (
	// 解析上下文状态
	parserState struct {
		currentBlock     blockState
		currentLine      lineState
		currentSystem    *Public_Data.System
		elementStack     []xml.StartElement
		currentPName     string
		currentPContent  string
		hasPortAttr      bool 
		portsFromList []int
	}
	
	blockState struct {
		Type        string
		Name        string
		SID         int
		IsAtomic    bool
		SystemRef   string
		PortCounts  portCounts
	}

	lineState struct {
		Src      int
		Dst      int
		Branches []int  // 分支连接目标
	}
	
	portCounts struct {
		In      int
		Out     int
		Trigger int
	}
)

func AnalyzeSubSystemXML(xmlPath string, rootSystem *Public_Data.System, modelName string) error {
	queue := []struct {
		path string
		sys  *Public_Data.System
	}{{xmlPath, rootSystem}}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if err := processSystemFile(item.path, item.sys, &queue); err != nil {
			return err
		}
	}
	
	outputPath := filepath.Join(Public_Data.OutputDir, modelName+".txt")
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📦 系统架构分析结果如下：\n"))
	printSystemInfoToWriter(rootSystem, rootSystem, "", &builder)

	// 创建输出文件
	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("❌ 创建文件失败: %v\n", err)
		return err
	}
	defer func() {
    	if cerr := f.Close(); cerr != nil {
        	fmt.Printf("❌ 关闭文件失败: %v\n", cerr)
    	}
	}()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	_, err = writer.WriteString(builder.String())
	if err != nil {
		fmt.Printf("❌ 写入模型结构失败: %v\n", err)
		return err
	}
	fmt.Println("\n=== 系统统计信息 ===")
	return nil
}

// 处理单个系统文件
func processSystemFile(path string, system *Public_Data.System, queue *[]struct{path string; sys *Public_Data.System}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	state := &parserState{
		currentSystem: system,
	}

	decoder := xml.NewDecoder(file)
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch elem := token.(type) {
		case xml.StartElement:
			state.handleStartElement(elem)
		case xml.EndElement:
			state.handleEndElement(elem, queue, path)
		case xml.CharData:
			state.handleCharData(elem)
		}
	}
	return nil
}

// 处理开始标签
func (s *parserState) handleStartElement(elem xml.StartElement) {
	s.elementStack = append(s.elementStack, elem)
	switch elem.Name.Local {
	case "Block":
		s.initBlockState(elem)
	case "Line":
		s.currentLine = lineState{}
	case "PortCounts":
		s.parsePortCounts(elem)
	case "P":
		s.initPState(elem)
	case "System":
		s.parseSystemRef(elem)
	}
}

// 处理结束标签
func (s *parserState) handleEndElement(elem xml.EndElement, queue *[]struct{path string; sys *Public_Data.System}, currentPath string) {
	if len(s.elementStack) > 0 {
		s.elementStack = s.elementStack[:len(s.elementStack)-1]
	}
	switch elem.Name.Local {
	case "Block":
		s.processBlock(queue, currentPath)
	case "Line":
		s.processLine()
	case "P":
		s.processPContent()
	}
}

// 处理字符数据
func (s *parserState) handleCharData(data xml.CharData) {
	s.currentPContent = strings.TrimSpace(string(data))
}

// 初始化 Block 状态
func (s *parserState) initBlockState(elem xml.StartElement) {
	s.currentBlock = blockState{}
	s.hasPortAttr = false
	s.portsFromList = nil 
	for _, attr := range elem.Attr {
		switch attr.Name.Local {
		case "Name":
			s.currentBlock.Name = attr.Value
		case "SID":
			sidStr := attr.Value
			parts := strings.Split(sidStr, "::")
			lastPart := parts[len(parts)-1]
			s.currentBlock.SID, _ = strconv.Atoi(lastPart)
		case "BlockType":
			s.currentBlock.Type = attr.Value
		}
	}
}

// 解析端口数量
func (s *parserState) parsePortCounts(elem xml.StartElement) {
	for _, attr := range elem.Attr {
		switch attr.Name.Local {
		case "in":
			s.currentBlock.PortCounts.In, _ = strconv.Atoi(attr.Value)
		case "out":
			s.currentBlock.PortCounts.Out, _ = strconv.Atoi(attr.Value)
		case "trigger":
			s.currentBlock.PortCounts.Trigger, _ = strconv.Atoi(attr.Value)
		}
	}
}

// 初始化 P 标签状态
func (s *parserState) initPState(elem xml.StartElement) {
	s.currentPContent = ""
	for _, attr := range elem.Attr {
		if attr.Name.Local == "Name" {
			s.currentPName = attr.Value
		}
	}
}

// 解析子系统引用
func (s *parserState) parseSystemRef(elem xml.StartElement) {
	for _, attr := range elem.Attr {
		if attr.Name.Local == "Ref" {
			s.currentBlock.SystemRef = attr.Value
		}
	}
}

// 处理 Block 结束标签
func (s *parserState) processBlock(queue *[]struct{ path string; sys *Public_Data.System }, currentPath string) {
	switch s.currentBlock.Type {
	case "Inport", "Outport":
		Port_Analysis.PortAnalysis(
			s.currentBlock.Name,
			s.currentBlock.SID,
			s.currentBlock.Type,
			s.hasPortAttr, 
			s.currentSystem,
		)
	case "SubSystem":
		// 创建子系统
		subSystem := Public_Data.NewSystemFromBlock(
			s.currentBlock.Name,
			s.currentBlock.SID,
			s.currentBlock.PortCounts.In,
			s.currentBlock.PortCounts.Out,
			true,   
		)
		// 设置类型字段用于区分系统和类
		if s.currentBlock.IsAtomic {
			subSystem.Type = "system" // 如果是原子子系统，标记为 system
		} else {
			subSystem.Type = "class"  // 否则标记为 class
		}
		// 若存在系统引用，递归读取对应的 XML 文件
		if s.currentBlock.SystemRef != "" {
			newPath := generateNewPath(currentPath, s.currentBlock.SystemRef)
			*queue = append(*queue, struct {
				path string
				sys  *Public_Data.System
			}{
				path: newPath,
				sys:  subSystem,
			})
		}
		// 添加子系统到当前系统的子系统列表
		s.currentSystem.System = append(s.currentSystem.System, subSystem)
	default:
		// 常规功能块，例如 Terminator、Gain 等
		s.currentSystem.Block = append(s.currentSystem.Block, &Public_Data.Block{
			Name: strings.TrimSpace(s.currentBlock.Name),
			SID:  s.currentBlock.SID,
			Type: s.currentBlock.Type,
		})
	}
}

// 分析连接信息（例如端口连接到哪个模块，或模块输出连接到哪个端口）
func (s *parserState) processLine() {
	if s.currentLine.Src != 0 {
		// 主目标连接
		if s.currentLine.Dst != 0 {
			Connect_Analysis.ConnectAnalysis(s.currentLine.Src, s.currentLine.Dst, s.currentSystem)
		}
		// 所有分支目标连接
		for _, branchDst := range s.currentLine.Branches {
			Connect_Analysis.ConnectAnalysis(s.currentLine.Src, branchDst, s.currentSystem)
		}
	}
}

// 处理 P 标签内容
func (s *parserState) processPContent() {
	switch s.currentPName {
	case "Src":
		s.currentLine.Src = extractID(s.currentPContent)
	case "Dst":
		dstID := extractID(s.currentPContent)
		if len(s.elementStack) >= 1 && s.elementStack[len(s.elementStack)-1].Name.Local == "Branch" {
			s.currentLine.Branches = append(s.currentLine.Branches, dstID)
		} else {
			s.currentLine.Dst = dstID
		}
	case "TreatAsAtomicUnit":
		s.currentBlock.IsAtomic = (s.currentPContent == "on")
	case "Port":
		s.hasPortAttr = true
	case "Ports":
		fields := strings.Trim(s.currentPContent, "[]")
		parts := strings.Split(fields, ",")
		s.portsFromList = nil
		for _, p := range parts {
			n, _ := strconv.Atoi(strings.TrimSpace(p))
			s.portsFromList = append(s.portsFromList, n)
		}
	}
}

// 路径拼接工具函数
func generateNewPath(originalPath, systemRef string) string {
	return filepath.Join(
		filepath.Dir(originalPath),
		systemRef+".xml",
	)
}

// ID 提取工具函数
func extractID(content string) int {
	parts := strings.Split(content, "#")
	if len(parts) == 0 {
		return 0
	}
	sidPart := parts[0]
	sidSubParts := strings.Split(sidPart, "::")
	lastSID := sidSubParts[len(sidSubParts)-1]
	id, _ := strconv.Atoi(lastSID)
	return id
}

// 输出系统信息至 writer
func printSystemInfoToWriter(system *Public_Data.System, currentSystem *Public_Data.System, indent string, writer *strings.Builder) {
	systemPrefix := ""
	if currentSystem.Type == "class" {
		systemPrefix = "🏷️"
	} else if currentSystem.Type == "system" {
		systemPrefix = "📦"
	} else {
		systemPrefix = "📦🏷️"
	}
	writer.WriteString(fmt.Sprintf("%sSystem: %s (SID: %d)\n", systemPrefix, currentSystem.Name, currentSystem.SID))

	portCount := len(currentSystem.Port)
	classCount := 0
	classInputsSum := 0
	classOutputsSum := 0

	for _, subSys := range currentSystem.System {
		if subSys.Type == "class" {
			classCount++
			classInputsSum += subSys.Inputs
			classOutputsSum += subSys.Outputs
		}
	}

	inPortCount := 0
	outPortCount := 0
	for _, port := range currentSystem.Port {
		if port.IO == "In" {
			inPortCount++
		} else if port.IO == "Out" {
			outPortCount++
		}
	}

	writer.WriteString(fmt.Sprintf("%s  ├─📊 类数量: %d\n", indent, classCount))
	writer.WriteString(fmt.Sprintf("%s  ├─📊 接口汇总: %d (输入: %d 输出: %d)\n", indent, portCount, inPortCount, outPortCount))
	writer.WriteString(fmt.Sprintf("%s  ├─📊 类端口总数: %d (输入: %d 输出: %d)\n", indent, classInputsSum+classOutputsSum, classInputsSum, classOutputsSum))
	writer.WriteString(fmt.Sprintf("%s  ├─📊 M1指标: %d\n", indent, classCount*portCount*(classInputsSum+classOutputsSum)))
	// ✅ 输出当前系统发起的子系统连接（组件连接）
	if len(currentSystem.ComponentConnections) > 0 {
		writer.WriteString(fmt.Sprintf("%s  ├─🧩 Subsystem connection:\n", indent))
		for _, conn := range currentSystem.ComponentConnections {
			srcName := findBlockNameBySID(system, conn.SrcPortSID)
			dstName := findBlockNameBySID(system, conn.DstBlockSID)
	
			// 判断目标组件类型
			dstType := findBlockTypeBySID(system, conn.DstBlockSID)
			dstIcon := "📦"
			if dstType == "class" {
				dstIcon = "🏷️"
			}
	
			writer.WriteString(fmt.Sprintf("%s  │   └─📦 %s (SID: %d) → %s %s (SID: %d)\n",
				indent, srcName, conn.SrcPortSID, dstIcon, dstName, conn.DstBlockSID))
		}
	}
	for _, port := range currentSystem.Port {
		writer.WriteString(fmt.Sprintf("%s  ├─🔌 接口: %-*s (SID: %4d, 类型: %-4s, IO: %-3s)\n",
			indent, portNameWidth, port.Name, port.SID, port.PortType, port.IO))

		if len(port.Connection) > 0 {
			var targets []string
			for _, conn := range port.Connection {
				var targetSID int
				var targetName string
				if conn.Direction == "out" {
					targetSID = conn.DstBlockSID
					targetName = findBlockNameBySID(system, conn.DstBlockSID)
					targets = append(targets, fmt.Sprintf("[→] %s (SID: %d)", targetName, targetSID))
				} else {
					targetSID = conn.SrcPortSID
					targetName = findBlockNameBySID(system, conn.SrcPortSID)
					targets = append(targets, fmt.Sprintf("[←] %s (SID: %d)", targetName, targetSID))
				}
			}
			if len(targets) > 1 {
				writer.WriteString(fmt.Sprintf("%s  │   └─🔗 [%d 目标] %s\n", indent, len(targets), strings.Join(targets, ", ")))
			} else {
				writer.WriteString(fmt.Sprintf("%s  │   └─🔗 %s\n", indent, targets[0]))
			}
		}
	}

	for i, subSys := range currentSystem.System {
		last := (i == len(currentSystem.System)-1)
		prefix := "  └─"
		if !last {
			prefix = "  ├─"
		}
		writer.WriteString(fmt.Sprintf("%s%s", indent, prefix))
		printSystemInfoToWriter(system, subSys, indent+"    ", writer)
	}
}

// 根据 SID 查找模块或接口名称
func findBlockNameBySID(system *Public_Data.System, sid int) string {
	for _, port := range system.Port {
		if port.SID == sid {
			return port.Name
		}
	}
	for _, block := range system.Block {
		if block.SID == sid {
			cleanName := strings.ReplaceAll(block.Name, "\n", " ")
			cleanName = strings.ReplaceAll(cleanName, "\r", "")
			return fmt.Sprintf("(模块) %s", strings.TrimSpace(cleanName))
		}
	}
	if system.SID == sid {
		return system.Name
	}
	for _, sub := range system.System {
		name := findBlockNameBySID(sub, sid)
		if name != "" && !strings.HasPrefix(name, "Unknown") {
			return name
		}
	}
	return fmt.Sprintf("未知 (SID: %d)", sid)
}
func findBlockTypeBySID(system *Public_Data.System, sid int) string {
	if system.SID == sid {
		return system.Type
	}
	for _, sub := range system.System {
		if sub.SID == sid {
			return sub.Type
		}
		if t := findBlockTypeBySID(sub, sid); t != "" {
			return t
		}
	}
	return ""
}


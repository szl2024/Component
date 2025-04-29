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
const blockInfoWidth = 40 //Control the alignment of the length within parentheses

type (
	// Analyze the contextual state
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
		Branches []int  // Branch connection target
	}
	

	portCounts struct {
		In      int
		Out     int
		Trigger int
	}
)


func AnalyzeSubSystemXML(xmlPath string, rootSystem *Public_Data.System, modelName string) error{
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
	builder.WriteString(fmt.Sprintf("📦 The results of the system architecture analysis are as follows：\n"))
	printSystemInfoToWriter(rootSystem, rootSystem, "", &builder)

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		return err
	}
	defer func() {
    	if cerr := f.Close(); cerr != nil {
        	fmt.Printf("❌ Failed to close file: %v\n", cerr)
    	}
	}()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	_, err = writer.WriteString(builder.String())
	if err != nil {
		fmt.Printf("❌ Failed to write model structure: %v\n", err)
		return err
	}
	fmt.Println("\n=== System statistics ===")
	return nil
}

// Processing individual system files
func processSystemFile(path string, system *Public_Data.System, queue *[]struct{path string; sys *Public_Data.System}) error {
	
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Fail to open file: %w", err)
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

// Processing Start Label
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

// Processing End Label
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

// Process character data

func (s *parserState) handleCharData(data xml.CharData) {
	s.currentPContent = strings.TrimSpace(string(data))
}

// Initialize block status
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


// Analyze the number of ports
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

// Initialize P tag status
func (s *parserState) initPState(elem xml.StartElement) {
	s.currentPContent = ""
	for _, attr := range elem.Attr {
		if attr.Name.Local == "Name" {
			s.currentPName = attr.Value
		}
	}
}

// Analyze system references
func (s *parserState) parseSystemRef(elem xml.StartElement) {
	for _, attr := range elem.Attr {
		if attr.Name.Local == "Ref" {
			s.currentBlock.SystemRef = attr.Value
		}
	}
}

// End of processing block
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
		// Create subsystem
		subSystem := Public_Data.NewSystemFromBlock(
			s.currentBlock.Name,
			s.currentBlock.SID,
			s.currentBlock.PortCounts.In,
			s.currentBlock.PortCounts.Out,
			true,   
		)
		

		// Set the Type field to distinguish between System and Class
		if s.currentBlock.IsAtomic {
			subSystem.Type = "system" // If it is an atomic subsystem, set it as system
		} else {
			subSystem.Type = "class" // Otherwise, set as class
		}

		// If there is a system reference, recursively read its XML file
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

		// Add SubSystem to the System list of the parent system
		s.currentSystem.System = append(s.currentSystem.System, subSystem)

	default:
		// Common functional blocks, such as Terminator, Gain, etc
		s.currentSystem.Block = append(s.currentSystem.Block, &Public_Data.Block{
			Name: strings.TrimSpace(s.currentBlock.Name),
			SID:  s.currentBlock.SID,
			Type: s.currentBlock.Type,
		})
	}
}


// Analyze the information of the Link 
// (such as which blocks the Port port is connected to, or which block's output is connected to the Port port)
func (s *parserState) processLine() {
	if s.currentLine.Src != 0 {
		// Main target connection
		if s.currentLine.Dst != 0 {
			Connect_Analysis.ConnectAnalysis(s.currentLine.Src, s.currentLine.Dst, s.currentSystem)
		}
	
		// All branch target connections
		for _, branchDst := range s.currentLine.Branches {
			Connect_Analysis.ConnectAnalysis(s.currentLine.Src, branchDst, s.currentSystem)
		}
	}
	
	
}

// Process the content of P tags
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


// Path generation tool function
func generateNewPath(originalPath, systemRef string) string {
	return filepath.Join(
		filepath.Dir(originalPath),
		systemRef+".xml",
	)
}

// ID extraction tool function
func extractID(content string) int {
	parts := strings.Split(content, "#")
	if len(parts) == 0 {
		return 0
	}

	// Split the SID part again
	sidPart := parts[0]
	sidSubParts := strings.Split(sidPart, "::")
	lastSID := sidSubParts[len(sidSubParts)-1]

	id, _ := strconv.Atoi(lastSID)
	return id
}


func printSystemInfoToWriter(system *Public_Data.System, currentSystem *Public_Data.System, indent string, writer *strings.Builder) {
	// Determine the output symbol based on the Type field
 	systemPrefix := ""
 	if currentSystem.Type == "class" {
		systemPrefix = "🏷️"
 	} else if currentSystem.Type == "system" {
	 	systemPrefix = "📦"
 	} else {
	 	systemPrefix = "📦🏷️"
 	}

 	// Write system and statistical information
	writer.WriteString(fmt.Sprintf("%sSystem: %s (SID: %d)\n", systemPrefix, currentSystem.Name, currentSystem.SID))


 	// Count the number of ports
 	portCount := len(currentSystem.Port)
 	// Count the number of child nodes with type 'class'
 	classCount := 0
 	// Count the total sum of inputs and outputs of type "class" in the child nodes
 	classInputsSum := 0
 	classOutputsSum := 0

 	for _, subSys := range currentSystem.System {
		if subSys.Type == "class" {
			classCount++
			classInputsSum += subSys.Inputs
		 	classOutputsSum += subSys.Outputs
	 	}
 	}

 	// Count the IO types of ports
 	inPortCount := 0
 	outPortCount := 0
 	for _, port := range currentSystem.Port {
		if port.IO == "In" {
			inPortCount++
	 	} else if port.IO == "Out" {
			outPortCount++
	 	}
 	}	


    // Output statistical information
	writer.WriteString(fmt.Sprintf("%s  ├─📊 nClass: %d\n", indent, classCount))
    writer.WriteString(fmt.Sprintf("%s  ├─📊 portAsr: %d (In: %d Out: %d )\n", indent, portCount,inPortCount, outPortCount))
    writer.WriteString(fmt.Sprintf("%s  ├─📊 portSim: %d (In: %d Out: %d )\n", indent, classInputsSum + classOutputsSum, classInputsSum, classOutputsSum))
	writer.WriteString(fmt.Sprintf("%s  ├─📊 M1: %d \n", indent, classCount * portCount * ( classInputsSum + classOutputsSum )))
  
	// Output port information
	for _, port := range currentSystem.Port {
		writer.WriteString(fmt.Sprintf("%s  ├─🔌 Port: %-*s (SID: %4d, Type: %-4s, IO: %-3s)\n",
			indent, portNameWidth, port.Name, port.SID, port.PortType, port.IO))

		// Output connection information
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
				writer.WriteString(fmt.Sprintf("%s  │   └─🔗 [%d targets]  %s\n", indent, len(targets), strings.Join(targets, ", ")))
			} else {
				writer.WriteString(fmt.Sprintf("%s  │   └─🔗  %s\n", indent, targets[0]))
			}
		}
	}



	// Recursive output subsystem information
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


func findBlockNameBySID(system *Public_Data.System, sid int) string {
	for _, port := range system.Port {
		if port.SID == sid {
			return port.Name
		}
	}
	
	for _, block := range system.Block {
        if block.SID == sid {
            // Clean name by replacing newlines
            cleanName := strings.ReplaceAll(block.Name, "\n", " ")
            cleanName = strings.ReplaceAll(cleanName, "\r", "")
            return fmt.Sprintf("(Block) %s", strings.TrimSpace(cleanName))
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
	return fmt.Sprintf("Unknown (SID: %d)", sid)
}


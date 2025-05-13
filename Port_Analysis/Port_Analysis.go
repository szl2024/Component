// Port_Analysis.go
package Port_Analysis

import (
	"Project/component/Public_Data"
)

func PortAnalysis(name string, sid string, blockType string, hasPortAttribute bool, currentSystem *Public_Data.System) {
	portType := "C-S"
	if hasPortAttribute {
		portType = "S-R"
	}

	newPort := &Public_Data.Port{
		Name:     name,
		SID:      sid,
		PortType: portType,
		IO:       "In",
	}

	if blockType == "Outport" {
		newPort.IO = "Out"
	}

	currentSystem.Port = append(currentSystem.Port, newPort)
}

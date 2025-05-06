// Port_Analysis.go
package Port_Analysis

import (
	"Project/component/Public_Data"
)

// 端口分析函数：将 Inport/Outport 信息加入当前系统的端口列表
func PortAnalysis(name string, sid int, blockType string, hasPortAttribute bool, currentSystem *Public_Data.System) {
	portType := "C-S" // 默认为 Controller-to-Sensor（控制器→传感器）
	if hasPortAttribute {
		portType = "S-R" // 若标记有 Port 属性，则设为 Sensor-to-Receiver（传感器→接收器）
	}

	newPort := &Public_Data.Port{
		Name:     name,
		SID:      sid,
		PortType: portType,
		IO:       "In", // 默认 IO 为输入
	}

	if blockType == "Outport" {
		newPort.IO = "Out" // 若为输出端口，修改 IO 字段为输出
	}

	currentSystem.Port = append(currentSystem.Port, newPort) // 添加到当前系统的端口列表中
}

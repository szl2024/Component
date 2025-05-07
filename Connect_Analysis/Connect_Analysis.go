package Connect_Analysis

import "Project/component/Public_Data"

func ConnectAnalysis(srcID int, dstID int, currentSystem *Public_Data.System) {
	// 原有逻辑：分析端口连接
	for _, port := range currentSystem.Port {
		if port.SID == srcID {
			port.Connection = append(port.Connection, &Public_Data.Connection{
				SrcPortSID:  srcID,
				DstBlockSID: dstID,
				Direction:   "out",
			})
		} else if port.SID == dstID {
			port.Connection = append(port.Connection, &Public_Data.Connection{
				SrcPortSID:  srcID,
				DstBlockSID: dstID,
				Direction:   "in",
			})
		}
		if port.SID == srcID || port.SID == dstID {
			break
		}
	}

	// ✅ 新增逻辑：识别是否是子系统（组件）之间的连接
	var srcSys, dstSys *Public_Data.System
	for _, sub := range currentSystem.System {
		if sub.SID == srcID {
			srcSys = sub
		} else if sub.SID == dstID {
			dstSys = sub
		}
	}

	// ✅ 若是子系统连接，记录进 ComponentConnections
// ✅ 若是子系统连接，记录到源组件（srcSys）自身
if srcSys != nil && dstSys != nil {
	srcSys.ComponentConnections = append(srcSys.ComponentConnections, &Public_Data.Connection{
		SrcPortSID:  srcID,
		DstBlockSID: dstID,
		Direction:   "component",
	})
}
//归处理所有子系统
	for _, sub := range currentSystem.System {
		ConnectAnalysis(srcID, dstID, sub)
	}
}

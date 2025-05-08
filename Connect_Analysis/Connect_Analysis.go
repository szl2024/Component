package Connect_Analysis

import "Project/component/Public_Data"

func ConnectAnalysis(srcID int, dstID int, currentSystem *Public_Data.System) {
	// 分析 Port 端口连接
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
	}

	// ✅ 查找是否是子系统之间连接
	var srcSys, dstSys *Public_Data.System
	for _, sub := range currentSystem.System {
		if sub.SID == srcID {
			srcSys = sub
		} else if sub.SID == dstID {
			dstSys = sub
		}
	}

	// ✅ 如果是组件连接，记录到“源组件”srcSys中（注意不是 currentSystem）
	if srcSys != nil && dstSys != nil {
		srcSys.ComponentConnections = append(srcSys.ComponentConnections, &Public_Data.Connection{
			SrcPortSID:  srcID,
			DstBlockSID: dstID,
			Direction:   "component",
		})
	}

	// ✅ 递归处理所有子系统
	for _, sub := range currentSystem.System {
		ConnectAnalysis(srcID, dstID, sub)
	}
}

package Connect_Analysis

import "Project/component/Public_Data"

func ConnectAnalysis(srcID string, dstID string, currentSystem *Public_Data.System) {
	// Analyze Port Port Connections
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

	// Check if it is a connection between subsystems
	var srcSys, dstSys *Public_Data.System
	for _, sub := range currentSystem.System {
		if sub.SID == srcID {
			srcSys = sub
		} else if sub.SID == dstID {
			dstSys = sub
		}
	}

	// If it is a component connection, record it in the "source component" srcSys (note not the current system)
	if srcSys != nil && dstSys != nil {
		srcSys.ComponentConnections = append(srcSys.ComponentConnections, &Public_Data.Connection{
			SrcPortSID:  srcID,
			DstBlockSID: dstID,
			Direction:   "component",
		})
	}

	// Recursive processing of all subsystems
	for _, sub := range currentSystem.System {
		ConnectAnalysis(srcID, dstID, sub)
	}
}

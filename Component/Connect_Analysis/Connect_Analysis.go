package Connect_Analysis

import "Project/component/Public_Data"

func ConnectAnalysis(srcID int, dstID int, currentSystem *Public_Data.System) {
    // A. 如果当前系统下某个 Port 是 srcID，就在它上挂 "out" 连接
    for _, port := range currentSystem.Port {
        if port.SID == srcID {
            conn := &Public_Data.Connection{
                SrcPortSID:  srcID,
                DstBlockSID: dstID,
                Direction:   "out",
            }
            port.Connection = append(port.Connection, conn)
            break
        }
    }

    // B. 如果某个 Port 是 dstID，就在它上挂 "in" 连接
    for _, port := range currentSystem.Port {
        if port.SID == dstID {
            conn := &Public_Data.Connection{
                SrcPortSID:  srcID,
                DstBlockSID: dstID,
                Direction:   "in",
            }
            port.Connection = append(port.Connection, conn)
            break
        }
    }

    // C. 递归处理子系统
    for _, sub := range currentSystem.System {
        ConnectAnalysis(srcID, dstID, sub)
    }
}

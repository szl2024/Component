package Connect_Analysis

import "Project/component/Public_Data"

func ConnectAnalysis(srcID int, dstID int, currentSystem *Public_Data.System) {
    // A. 如果当前系统中某个端口的 SID 等于 srcID，则在其上挂载一个 “out” 方向的连接
    for _, port := range currentSystem.Port {
        if port.SID == srcID {
            port.Connection = append(port.Connection, &Public_Data.Connection{
                SrcPortSID:  srcID,
                DstBlockSID: dstID,
                Direction:   "out", // 出方向：该端口连接到了其他模块
            })
        } else if port.SID == dstID {
            port.Connection = append(port.Connection, &Public_Data.Connection{
                SrcPortSID:  srcID,
                DstBlockSID: dstID,
                Direction:   "in", // 入方向：其他模块连接到了该端口
            })
        }
        if port.SID == srcID || port.SID == dstID {
            break
        }
    }

    // C. 递归处理所有子系统
    for _, sub := range currentSystem.System {
        ConnectAnalysis(srcID, dstID, sub)
    }
}

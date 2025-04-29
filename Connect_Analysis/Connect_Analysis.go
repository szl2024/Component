package Connect_Analysis

import "Project/component/Public_Data"

func ConnectAnalysis(srcID int, dstID int, currentSystem *Public_Data.System) {
    // A. If a port in the current system is srcID, hang an "out" connection on it
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

    //C. Recursive processing subsystem
    for _, sub := range currentSystem.System {
        ConnectAnalysis(srcID, dstID, sub)
    }
}

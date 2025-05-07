// Public_Data.go
package Public_Data


var (
	Dir       string
	BuildDir  string
	OutputDir string
	Windir    string 
	Systems   []*System
)

type Connection struct {
	SrcPortSID  int
	DstBlockSID int
	Direction   string //'out' means I am connecting with others, 'in' means others are connecting with me
}
type Block struct {
	Name string
	SID  int
	Type string
}


type Port struct {
    Name       string
    SID        int
    PortType   string
    IO         string
    Connection []*Connection
}

type System struct {
	Name    string
	SID     int
	Inputs  int
	Outputs int
	Type    string
	Port    []*Port
	System  []*System
	Block   []*Block

	ComponentConnections []*Connection // ✅ 新增：记录子系统之间的连接
}

func NewSystemFromBlock(name string, sid int, inputs int, outputs int, register bool) *System {
    sys := &System{
        Name:    name,
        SID:     sid,
        Inputs:  inputs,
        Outputs: outputs,
    }
    if register {
        Systems = append(Systems, sys)
    }
    return sys
}

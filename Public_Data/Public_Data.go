// Public_Data.go
package Public_Data


var (
	Dir       string
	BuildDir  string
	OutputDir string
	Windir    string
	LdiDir    string
	TxtDir    string
	Systems   []*System
)

type Connection struct {
	SrcPortSID  string
	DstBlockSID string
	Direction   string //'out' means I am connecting with others, 'in' means others are connecting with me
}
type Block struct {
	Name string
	SID  string
	Type string
}


type Port struct {
    Name       string
    SID        string
    PortType   string
    IO         string
    Connection []*Connection
}

type System struct {
	Name    string
	SID     string
	Inputs  int
	Outputs int
	Type    string
	Port    []*Port
	System  []*System
	Block   []*Block

	ComponentConnections []*Connection 
}


func NewSystemFromBlock(name string, sid string, inputs int, outputs int, register bool) *System {
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

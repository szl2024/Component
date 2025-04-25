// Public_Data.go
package Public_Data

// 导出变量（首字母大写）
var (
	Dir       string
	BuildDir  string
	OutputDir string
	Windir    string
	Wsldir    string
	Classin   int
	Classout  int
    Systems []*System
)

type Connection struct {
	SrcPortSID  int
	DstBlockSID int
	Direction   string // "out" 表示我是连接别人, "in" 表示别人连我
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
    Name     string
    SID      int
    Inputs   int
    Outputs  int
    Type    string
    Port     []*Port
    System   []*System // 子节点列表（多个子节点）

    Block []*Block // ✅ 加入这个新列表
}

func NewSystemFromBlock(name string, sid int, inputs int, outputs int) *System {
    sys := &System{
        Name:    name,
        SID:     sid,
        Inputs:  inputs,
        Outputs: outputs,
    }
    Systems = append(Systems, sys)
    return sys
}
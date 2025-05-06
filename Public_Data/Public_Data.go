// Public_Data.go
package Public_Data

var (
	Dir       string      // 当前工作目录
	BuildDir  string      // 构建目录
	OutputDir string      // 输出目录
	Windir    string      // 用户输入路径
	Systems   []*System   // 所有注册的系统
)

// 连接信息结构体
type Connection struct {
	SrcPortSID  int    // 源端口的 SID
	DstBlockSID int    // 目标模块的 SID
	Direction   string // 'out' 表示我连接其他对象；'in' 表示其他对象连接我
}

// 功能块结构体
type Block struct {
	Name string // 模块名称
	SID  int    // 模块 SID
	Type string // 模块类型
}

// 接口结构体
type Port struct {
	Name       string         // 接口名称
	SID        int            // 接口 SID
	PortType   string         // 接口类型
	IO         string         // 输入/输出标识（In/Out）
	Connection []*Connection  // 连接信息
}

// 系统结构体
type System struct {
	Name    string     // 系统名称
	SID     int        // 系统 SID
	Inputs  int        // 输入数量
	Outputs int        // 输出数量
	Type    string     // 类型（system/class）
	Port    []*Port    // 接口列表
	System  []*System  // 子系统列表
	Block   []*Block   // 功能块列表
}

// 根据模块创建新系统（可选择是否注册进全局系统列表）
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

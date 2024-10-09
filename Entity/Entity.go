package Entity

import "time"

// License 授权信息列表 包括：授权起始时间、授权到期时间、允许节点数量、MAC地址列表、主板ID
type License struct {
	StartTime         string      `json:"start_time"`          // 开始时间，格式为：YYYY-MM-ddTHH:mm:SS
	EndTime           string      `json:"end_time"`            // 到期时间，格式为：YYYY-MM-ddTHH:mm:SS
	ClientTimeZone    string      `json:"client_time_zone"`    // 客户端时区
	LicenseCreateTime string      `json:"license_create_time"` // License创建时间
	AllowNodes        int         `json:"allow_nodes"`         // 允许接入的计算节点数
	UseNodes          int         `json:"use_nodes"`           // 已接入计算节点数
	MacAddr           string      `json:"mac_addr"`            // 授权的管理节点MAC地址
	MotherBoardID     string      `json:"mother_board_id"`     // 授权的管理节点主板编号
	PermanentAuth     bool        `json:"permanent_auth"`      // 永久授权
	CustomerTag       string      `json:"customer_tag"`        // 客户标记
	ModelRoute        string      `json:"model_route"`         // 模块路由Prefix
	CheckCode         string      `json:"check_code"`          // 校验码
	LastCheckTime     *time.Time  `json:"last_check_time"`     // 最后一次校验时间
	CheckStatus       bool        `json:"check_status"`        // 校验状态
	NodeList          []*NodeInfo `json:"node_list"`           // 节点列表
}

// NodeInfo 节点信息，记录仪授权的节点的基础信息
type NodeInfo struct {
	NodeIP            string `json:"node_ip"`              // 节点IP
	NodeName          string `json:"node_name"`            // 节点名
	NodeMac           string `json:"node_mac"`             // 节点MAC地址
	NodeMotherBoardID string `json:"node_mother_board_id"` // 节点主板ID
}

// NetCard 网卡详细信息
type NetCard struct {
	ID   int
	Name string
	MAC  string
	IP   string
}

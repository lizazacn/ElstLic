package Node

import (
	"github.com/lizazacn/ElstLic/Entity"
	"github.com/lizazacn/ElstLic/Utils"
	"os"
	"runtime"
)

type Node struct {
	MgrNetCard string `json:"mgr_net_card"` // 管理网卡名
}

func (n *Node) GetNodeInfo() (*Entity.NodeInfo, error) {
	var nodeInfo = new(Entity.NodeInfo)
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	nodeInfo.NodeName = hostname
	// 根据系统类型生成主板ID
	switch runtime.GOOS {
	case "linux":
		nodeInfo.NodeMotherBoardID = Utils.GetLinuxMotherBoardID()
	case "windows":
		nodeInfo.NodeMotherBoardID = Utils.GetWinMotherBoardID()
	default:
		nodeInfo.NodeMotherBoardID = ""
	}
	// 根据管理网卡名获取网卡基本信息
	netCard, err := Utils.GetAllNetCardInfoByName(n.MgrNetCard)
	if err != nil {
		return nil, err
	}
	nodeInfo.NodeIP = netCard.IP
	nodeInfo.NodeMac = netCard.MAC
	return nodeInfo, nil
}

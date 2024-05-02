package Utils

import (
	"encoding/json"
	"github.com/lizazacn/ElstLic/Entity"
	"github.com/lizazacn/ElstLic/Utils/GM"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

// CheckData 校验数据
func CheckData(lic *Entity.License) (bool, error) {
	var oldCheckCode = lic.CheckCode
	lic.CheckCode = ""
	licByte, err := json.Marshal(lic)
	if err != nil {
		return false, err
	}
	checkCode := GM.SM3SUM(string(licByte))
	if oldCheckCode == checkCode {
		return true, nil
	}
	return false, nil
}

// GetGMCipherAndKey 获取GM密文和SM4Key
func GetGMCipherAndKey(ciphertext []byte, offset, step int) ([]byte, []byte) {
	var key = make([]byte, 0)
	var idxList = make([]int, 0)
	for i := 0; i < 16; i++ {
		var idx = i*step + offset
		key = append(key, ciphertext[idx])
		idxList = append(idxList, idx)
	}
	for idx := len(idxList) - 1; idx >= 0; idx-- {
		index := idxList[idx]
		ciphertext = append(ciphertext[:index], ciphertext[index+1:]...)
	}
	return ciphertext, key
}

// AddKeyToGMCipher 添加SM4Key到GM密文
func AddKeyToGMCipher(ciphertext, key []byte, offset, step int) []byte {
	for i := range key {
		var idx = i*step + offset
		var tmp = make([]byte, idx)
		copy(tmp, ciphertext[:idx])
		tmp = append(tmp, key[i])
		ciphertext = append(tmp, ciphertext[idx:]...)
	}
	return ciphertext
}

// GetLinuxMotherBoardUUID 获取Linux主板UUID
func GetLinuxMotherBoardUUID() string {
	cmd := exec.Command("dmidecode", "-t", "system")
	result, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	var uuid string
	compile, err := regexp.Compile(`Serial Number:(.+)\n`)
	if err != nil {
		return ""
	}
	find := compile.Find(result)
	uuid = string(find)
	uuid = strings.ReplaceAll(uuid, "Serial Number:", "")
	uuid = strings.ReplaceAll(uuid, " ", "")
	uuid = strings.ReplaceAll(uuid, "\n", "")
	return uuid
}

// GetWinMotherBoardUUID 获取Windows主板ID
func GetWinMotherBoardUUID() string {
	cmd := exec.Command("wmic", "baseboard", "get", "SerialNumber")
	result, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	var uuid string
	uuid = string(result)
	uuid = strings.ReplaceAll(uuid, "SerialNumber", "")
	uuid = strings.ReplaceAll(uuid, " ", "")
	uuid = strings.ReplaceAll(uuid, "\n", "")
	uuid = strings.ReplaceAll(uuid, "\r", "")
	return uuid
}

// GetAllNetCardInfo 获取全部网卡信息
func GetAllNetCardInfo() []Entity.NetCard {
	var result = make([]Entity.NetCard, 0)
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for idx, eth := range interfaces {
		result = append(result, Entity.NetCard{
			ID:   idx + 1,
			Name: eth.Name,
			MAC:  eth.HardwareAddr.String(),
		})
	}
	return result
}

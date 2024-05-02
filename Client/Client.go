package Client

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/lizazacn/ElstLic/Entity"
	"github.com/lizazacn/ElstLic/Utils"
	"github.com/lizazacn/ElstLic/Utils/GM"
	"github.com/manifoldco/promptui"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"
)

type Client struct {
	Offset  int
	Step    int
	DevInfo string
}

// CreateNodeInfoFile 创建节点信息文件
func (c *Client) CreateNodeInfoFile() error {
	if c.Offset == 0 {
		c.Offset = 1
	}
	if c.Step == 0 {
		c.Step = 1
	}
	licData, err := c.createLicData()
	if err != nil {
		return err
	}
	return c.encryptDataToFile(licData)
}

// createLicData 初始化Lic证书数据
func (c *Client) createLicData() (*Entity.License, error) {
	var lic = new(Entity.License)
	lic.StartTime = time.Now().Format("2006-01-02T15:04:05")

	// 根据系统类型生成主板ID
	switch runtime.GOOS {
	case "linux":
		lic.MotherBoardUUID = Utils.GetLinuxMotherBoardUUID()
	case "windows":
		lic.MotherBoardUUID = Utils.GetWinMotherBoardUUID()
	default:
		lic.MotherBoardUUID = ""
	}

	if lic.MotherBoardUUID == "" {
		return nil, errors.New("未获取到主板ID")
	}
	var netCards = Utils.GetAllNetCardInfo()
	var template = &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "{{ \">\" | green }} {{ .ID | cyan }} {{ .Name | cyan }} ({{ .MAC | red }})",
		Inactive: "{{ .ID | cyan }} {{ .Name | cyan }} ({{ .MAC | red }})",
		Selected: "{{ .Name | cyan }} ({{ .MAC | red }})",
	}
	prompt := promptui.Select{
		Label:     "请选择需要授权的网卡：",
		Items:     netCards,
		Templates: template,
		Size:      5,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	lic.MacAddr = netCards[idx].MAC
	return lic, err
}

// EncryptDataToFile 加密数据到文件
func (c *Client) encryptDataToFile(lic *Entity.License) error {
	var path = "./"
	prompt := promptui.Prompt{
		Label:   "请输入node.info文件保存路径",
		Default: path,
	}

	result, err := prompt.Run()
	if err != nil {
		return err
	}
	if result != "" {
		path = result
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/node.info", path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	licByte, err := json.Marshal(lic)
	if err != nil {
		return err
	}
	lic.CheckCode = GM.SM3SUM(string(licByte))

	licByte, err = json.Marshal(lic)
	if err != nil {
		return err
	}
	var key = lic.CheckCode[:16]
	encrypt, err := GM.SM4Encrypt(licByte, []byte(key), []byte(key))
	if err != nil {
		return err
	}
	encrypt = Utils.AddKeyToGMCipher(encrypt, []byte(key), c.Offset, c.Step)
	_, err = file.Write(encrypt)
	if err != nil {
		return err
	}
	return nil
}

// DecryptDataFromFile 解密License文件
func (c *Client) DecryptDataFromFile(path ...string) (*Entity.License, error) {
	if c.Offset == 0 {
		c.Offset = 1
	}
	if c.Step == 0 {
		c.Step = 1
	}
	var licPath string
	if len(path) >= 1 && path[0] != "" {
		licPath = path[0]
	} else {
		inPutFile := flag.String("l", "", "请指定license.lic文件路径")
		flag.Parse()
		licPath = *inPutFile
	}
	if licPath == "" {
		prompt := promptui.Prompt{
			Label:   "请输入license.lic文件路径",
			Default: "./license.lic",
		}

		result, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		if result != "" {
			licPath = result
		}
	}
	ciphertext, err := os.ReadFile(licPath)
	if err != nil {
		return nil, err
	}
	cipher, key := Utils.GetGMCipherAndKey(ciphertext, c.Offset, c.Step)
	decrypt, err := GM.SM4Decrypt(cipher, key, key)
	if err != nil {
		return nil, err
	}
	var lic = new(Entity.License)
	err = json.Unmarshal(decrypt, lic)
	if err != nil {
		return nil, err
	}
	stat, err := Utils.CheckData(lic)
	if err != nil {
		return nil, err
	}
	if !stat {
		return nil, errors.New(fmt.Sprintf("数据疑似被篡改，请联系:%s！", c.DevInfo))
	}
	return lic, nil
}

// EnableLicCheck 启动Lic证书校验
func (c *Client) EnableLicCheck() {
	go c.licCheck(time.Now())
}

// licCheck Lic证书校验
func (c *Client) licCheck(lastRunTime time.Time) {
	for true {
		rand.Seed(time.Now().UnixNano())
		var randomInt = rand.Intn(240)
		var sleepTime = randomInt * 6
		time.Sleep(time.Duration(sleepTime) * time.Minute)
		var now = time.Now()
		if now.Sub(lastRunTime).Minutes()-float64(sleepTime) >= 30 {
			log.Println("请勿随意修改系统时间，否则会影响License授权！")
			os.Exit(0)
		}
		lastRunTime = now
	}
}

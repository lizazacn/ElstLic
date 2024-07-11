package Client

import (
	"encoding/json"
	"errors"
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
	licPath string
}

type Lic func() bool

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
		lic.MotherBoardID = Utils.GetLinuxMotherBoardID()
	case "windows":
		lic.MotherBoardID = Utils.GetWinMotherBoardID()
	default:
		lic.MotherBoardID = ""
	}

	if lic.MotherBoardID == "" {
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

// encryptDataToFile 加密数据到文件
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

// encryptDataToLicFile 加密数据到认证文件
func (c *Client) encryptDataToLicFile(lic *Entity.License, licPath string) error {
	file, err := os.OpenFile(licPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
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
func (c *Client) EnableLicCheck(lic Lic) {
	go c.licCheck(time.Now(), lic)
}

// EnableDefaultLicCheck 启动默认Lic证书校验机制
func (c *Client) EnableDefaultLicCheck(licPath string) {
	c.licPath = licPath
	go c.licCheck(time.Now(), c.DefaultLic)
}

// licCheck Lic证书校验
func (c *Client) licCheck(lastRunTime time.Time, lic Lic) {
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
		status := lic()
		if !status {
			os.Exit(0)
		}
		lastRunTime = now
	}
}

// DefaultLic 默认证书校验规则
func (c *Client) DefaultLic() bool {
	license, err := c.DecryptDataFromFile(c.licPath)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 获取主板ID
	var motherBoardID string
	switch runtime.GOOS {
	case "linux":
		motherBoardID = Utils.GetLinuxMotherBoardID()
	case "windows":
		motherBoardID = Utils.GetWinMotherBoardID()
	default:
		log.Println("未识别到主板ID")
		return false
	}

	// 验证主板ID是否一致
	if motherBoardID != license.MotherBoardID {
		log.Println("主板ID比对异常，请检查是否使用了正确的license文件")
		return false
	}

	// 验证系统license是否过期
	var now = time.Now()
	startAt, err := time.ParseInLocation("2006-01-02T15:04:05", license.StartTime, time.Local)
	if err != nil {
		log.Println("解析时间异常，疑似数据被篡改！")
		return false
	}

	if now.Before(startAt) {
		log.Println("证书不在有效期！")
		return false
	}
	endAt, err := time.ParseInLocation("2006-01-02T15:04:05", license.EndTime, time.Local)
	if err != nil {
		log.Println("解析时间异常，疑似数据被篡改！")
		return false
	}
	//endAt = now.AddDate(100, 0, 0)
	if now.After(endAt) {
		log.Println("证书已过期，请联系销售人员重新获取授权！")
		return false
	}
	return true
}

// RegisterNodeToLicense 注册新节点
func (c *Client) RegisterNodeToLicense(info *Entity.NodeInfo, licPath string) error {
	license, err := c.DecryptDataFromFile(licPath)
	if err != nil {
		return err
	}
	if license.UseNodes >= license.AllowNodes {
		return errors.New("超出允许的节点范围，请联系产品供应商扩容许可")
	}

	if license.NodeList == nil {
		license.NodeList = make([]*Entity.NodeInfo, 0)
	}
	license.NodeList = append(license.NodeList, info)
	return c.encryptDataToLicFile(license, licPath)
}

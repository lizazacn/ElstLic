package Server

import (
	"ElstLic/Entity"
	"ElstLic/Utils"
	"ElstLic/Utils/GM"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Server struct {
	Offset  int
	Step    int
	DevInfo string
}

// CreateLicFile 创建license授权文件
func (s *Server) CreateLicFile() error {
	if s.Offset == 0 {
		s.Offset = 1
	}
	if s.Step == 0 {
		s.Step = 1
	}
	inPutFile := flag.String("i", "", "请指定node.info文件路径")
	flag.Parse()
	if *inPutFile == "" {
		prompt := promptui.Prompt{
			Label:   "请输入node.info文件路径",
			Default: "./node.info",
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		if result != "" {
			*inPutFile = result
		}
	}
	ciphertext, err := os.ReadFile(*inPutFile)
	if err != nil {
		return err
	}
	cipher, key := Utils.GetGMCipherAndKey(ciphertext, s.Offset, s.Step)
	decrypt, err := GM.SM4Decrypt(cipher, key, key)
	if err != nil {
		return err
	}
	var lic = new(Entity.License)
	err = json.Unmarshal(decrypt, lic)
	if err != nil {
		return err
	}
	stat, err := Utils.CheckData(lic)
	if err != nil {
		return err
	}
	if !stat {
		return errors.New(fmt.Sprintf("数据疑似被篡改，请联系:%s！", s.DevInfo))
	}
	// 输入必要字段
	err = s.inputLicData(lic)
	if err != nil {
		return err
	}
	// 回显License信息
	fmt.Println("###############授权信息###############")
	licJson, err := json.MarshalIndent(lic, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(licJson))
	// 加密数据到文件
	err = s.encryptDataToFile(lic)
	if err != nil {
		return err
	}
	return nil
}

// inputLicData 填充Lic必要数据
func (s *Server) inputLicData(lic *Entity.License) error {
	var nowTime = time.Now()
	start, err := time.ParseInLocation("2006-01-02T15:04:05", lic.StartTime, time.Local)
	if err != nil {
		fmt.Printf("解析开始时间字段异常: %v", err.Error())
		start = nowTime
		lic.StartTime = nowTime.Format("2006-01-02T15:04:05")
	}
	if start.AddDate(0, 0, 1).Before(nowTime) {
		return errors.New("node.info文件已超出48小时有效期！")
	}
	if start.After(nowTime) {
		start = nowTime
		lic.StartTime = nowTime.Format("2006-01-02T15:04:05")
	}
	lic.LicenseCreateTime = nowTime.Format("2006-01-02T15:04:05")

reInNodes:
	// 设置最大节点数
	prompt := promptui.Prompt{
		Label:   "请输入允许接入的最大节点数",
		Default: "3",
	}
	result, err := prompt.Run()
	if err != nil {
		return err
	}
	atoi, err := strconv.Atoi(result)
	if err != nil {
		fmt.Println("输入格式异常，请输入数字类型数据！")
		goto reInNodes
	}
	if atoi <= 3 {
		atoi = 3
	}
	lic.AllowNodes = atoi

	// 设置是否永久授权（100年）
	promptSelect := promptui.Select{
		Label: "选择是否永久授权",
		Items: []string{"yes", "no"},
		Size:  2,
	}

	_, result, err = promptSelect.Run()
	if err != nil {
		return err
	}
	// 设置结束时间(100年)
	if result == "yes" {
		end := start.AddDate(100, 0, 0)
		lic.EndTime = end.Format("2006-01-02T15:04:05")
	}

	if result != "yes" {
	reInDate:
		// 设置过期时间
		prompt = promptui.Prompt{
			Label:   "设置过期时间",
			Default: start.AddDate(0, 30, 0).Format("2006-01-02T15:04:05"),
		}
		result, err = prompt.Run()
		if err != nil {
			return err
		}
		// 验证输入格式是否正确
		dateTimeComp, err := regexp.Compile(`\d\d\d\d-\d\d-\d\dT\d\d:\d\d:\d\d`)
		if err != nil {
			return err
		}
		if !dateTimeComp.Match([]byte(result)) {
			fmt.Println("输入时间格式不正确，请重新输入！")
			goto reInDate
		}
		lic.EndTime = result
	}

	// 设置客户标记
	prompt = promptui.Prompt{
		Label:   "设置客户标记",
		Default: "",
	}
	result, err = prompt.Run()
	if err != nil {
		return err
	}
	lic.CustomerTag = result
	if result == "" {
		lic.CustomerTag = lic.MacAddr
	}
	return nil
}

// EncryptDataToFile 加密数据到文件
func (s *Server) encryptDataToFile(licData *Entity.License) error {
	var path = "./"
	prompt := promptui.Prompt{
		Label:   "请输入license.lic文件保存路径",
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

	file, err := os.OpenFile(fmt.Sprintf("%s/license.lic", path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	licByte, err := json.Marshal(licData)
	if err != nil {
		return err
	}
	licData.CheckCode = GM.SM3SUM(string(licByte))

	licByte, err = json.Marshal(licData)
	if err != nil {
		return err
	}
	var key = licData.CheckCode[:16]
	encrypt, err := GM.SM4Encrypt(licByte, []byte(key), []byte(key))
	if err != nil {
		return err
	}
	encrypt = Utils.AddKeyToGMCipher(encrypt, []byte(key), s.Offset, s.Step)
	_, err = file.Write(encrypt)
	if err != nil {
		return err
	}
	return nil
}

package main

import (
	"fmt"
	"github.com/lizazacn/ElstLic/Client"
	"github.com/lizazacn/ElstLic/Server"
	"github.com/manifoldco/promptui"
)

func main() {
	var client = Client.Client{
		Offset:  3,
		Step:    3,
		DevInfo: "www.elst.dev",
	}
	var server = Server.Server{
		Offset:  3,
		Step:    3,
		DevInfo: "www.elst.dev",
	}

	// 选择执行的操作
	promptSelect := promptui.Select{
		Label: "选择执行的操作（Client：生成节点信息文件；Server：生成license文件）：",
		Items: []string{"Client", "Server"},
		Size:  2,
	}
	_, result, err := promptSelect.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	if result == "Client" {
		err = client.CreateNodeInfoFile()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	if result == "Server" {
		err = server.CreateLicFile()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("############生成授权数据完成############")
		fmt.Println("###############授权信息###############")
		license, err := client.DecryptDataFromFile()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(*license)
	}
}

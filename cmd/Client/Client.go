package main

import (
	"bufio"
	"cache/cachepb/cachepb"
	"cache/service"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var (
	addr   string
	client *service.Client
)

func main() {
	createClient()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("[" + addr + "]:")
		input, err := reader.ReadString('\n') // 读取整行输入
		if err != nil {
			showError(err)
			continue
		}
		input = strings.TrimSpace(input) // 去除首尾空格
		if input == "exit" {
			os.Exit(0)
		} else {
			commandExplainer(input)
		}
	}
}

func createClient() {
	var input string
	for {
		fmt.Println("[Zcache]:enter the target server's addr,like 127.0.0.1:8888")
		_, err := fmt.Scanln(&input)
		if err != nil {
			showError(err)
			continue
		}
		// 检测输入的地址是否有效
		parts := strings.SplitN(input, ":", 2)
		if len(parts) != 2 {
			fmt.Println("[Error]: addr format error")
			continue
		}
		// 检测连通性
		conn, err := net.DialTimeout("tcp", input, 3*time.Second)
		if err != nil {
			fmt.Printf("[Error]: cant't connect to the server %v，error: %v\n", input, err)
			continue
		}
		conn.Close()
		fmt.Println("success")
		addr = input
		client = service.NewClient("http://" + input)
		break
	}
}

// 解析命令
func commandExplainer(input string) {
	words := strings.Fields(input)
	inputLen := len(words)
	if inputLen == 0 {
		return
	}
	if inputLen == 1 {
		//当指令只有一条时，先进行解析，当没有对应的指令时再进行用法的解释
		if explainOneCommand(words[0]) {
			return
		}
		explainUsage(words[0])
		return
	}
	switch words[0] {
	case "createGroup":
		if inputLen == 3 {
			client.CreateGroup(words[1])
			fmt.Println("success")
		} else {
			showError(errors.New("unexpected command,use createGroup to see the usage"))
		}
	case "set":
		in := cachepb.SetRequest{}
		if inputLen == 4 {
			in.Group = words[1]
			in.Key = words[2]
			in.Value = []byte(words[3])
		} else if inputLen == 3 {
			in.Group = "default"
			in.Key = words[1]
			in.Value = []byte(words[2])
		} else {
			showError(errors.New("unexpected command,use set to see the usage"))
			return
		}
		out := cachepb.Response{}
		if err := client.Set(&in, &out); err != nil {
			showError(err)
		} else {
			fmt.Println(string(out.Value))
		}
	case "get":
		in := cachepb.GetRequest{}
		if inputLen == 3 {
			in.Group = words[1]
			in.Key = words[2]
		} else if inputLen == 2 {
			in.Group = "default"
			in.Key = words[1]
		}
		out := cachepb.Response{}
		if err := client.Get(&in, &out); err != nil {
			showError(err)
		} else {
			fmt.Println(string(out.Value))
		}
	case "getKeys":
		if inputLen != 2 {
			showError(errors.New("unexpected command,use getKeys to see the usage"))
			return
		}
		out := cachepb.GroupKeyList{}
		if err := client.GetGroupKeyList(words[1], &out); err != nil {
			showError(err)
			return
		}
		for _, v := range out.Key {
			fmt.Println(v)
		}
	default:
		showError(errors.New("unknown command"))
	}
}

// 解析一条指令
func explainOneCommand(command string) bool {
	switch command {
	case "getGroups":
		out := cachepb.GroupList{}
		if err := client.GetGroupList(&out); err != nil {
			showError(err)
		}
		for _, v := range out.GroupName {
			fmt.Println(v)
		}
	case "getKeys":
		out := cachepb.GroupKeyList{}
		if err := client.GetGroupKeyList("default", &out); err != nil {
			showError(err)
			return true
		}
		for _, v := range out.Key {
			fmt.Println(v)
		}
	default:
		return false
	}
	return true
}

// 解析帮助
func explainUsage(word string) {
	switch word {
	case "createGroup":
		fmt.Println("createGroup -GroupName -MaxBytes")
	case "set":
		fmt.Println("set -GroupName(default='default') -Key -Value")
	case "get":
		fmt.Println("get -GroupName(default='default') -Key")
	default:
		showError(errors.New("unknown command"))
	}
}
func showError(err error) {
	fmt.Println("[Error]:", err.Error())
}

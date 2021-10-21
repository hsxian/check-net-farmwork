package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	registry "golang.org/x/sys/windows/registry"
)

type FileEx struct {
	FileName      string `mapstructure:"FileName" json:"FileName" ini:"FileName"`
	RunningTip    string `mapstructure:"RunningTip" json:"RunningTip" ini:"RunningTip"`
	IsWaitRunning bool   `mapstructure:"IsWaitRunning" json:"IsWaitRunning" ini:"IsWaitRunning"`
	AbsPath       string
	AbsDir        string
}
type Config struct {
	MinVersion          string   `mapstructure:"MinVersion" json:"MinVersion" ini:"MinVersion"`
	DotNetFramworkFiles []FileEx `mapstructure:"DotNetFramworkFiles" json:"DotNetFramworkFiles" ini:"DotNetFramworkFiles"`
	StartExexFiles      []FileEx `mapstructure:"StartExexFiles" json:"StartExexFiles" ini:"StartExexFiles"`
}

var (
	CONFIG = new(Config)
)

// reg query "HKLM\Software\Microsoft\NET Framework Setup\NDP" /s /v version | findstr /i version | sort /+26 /r
func main() {
	InitConfigFromJson()
	path := "Software\\Microsoft\\NET Framework Setup\\NDP"
	subkeys := ReadReadAllSubKeyNames(path)
	versions := []string{}
	for _, subkey := range subkeys {
		v, err := ReadValue(subkey, "version")
		if err == nil {
			versions = append(versions, v)
		}
	}
	minVersion := CONFIG.MinVersion
	a := VersionOrdinal(minVersion)
	versions = RemoveDuplicatesAndEmpty(versions)
	isQualified := false
	minVersionTip := ".Net Farmwork运行时检测：" + minVersion + "(最低要求)"
	for _, v := range versions {
		b := VersionOrdinal(v)
		switch {
		case a > b:
			fmt.Println(minVersionTip, ">", v+"(本机已安装)")
		case a < b:
			isQualified = true
			fmt.Println(minVersionTip, "<", v+"(本机已安装)")
		case a == b:
			fmt.Println(minVersionTip, "=", v+"(本机已安装)")
		}
	}
	if isQualified {
		for _, v := range CONFIG.StartExexFiles {
			RunExec(v)
		}
	} else {
		fmt.Printf("缺少.net framwork环境，最小依赖版本：%s\n", minVersion)
		for _, v := range CONFIG.DotNetFramworkFiles {
			RunExec(v)
		}
	}
}

func InitConfigFromJson() {
	// 打开文件
	file, _ := os.Open("config-check-runtime-environment.json")
	// 关闭文件
	defer file.Close()
	//NewDecoder创建一个从file读取并解码json对象的*Decoder，解码器有自己的缓冲，并可能超前读取部分json数据。
	decoder := json.NewDecoder(file)
	//Decode从输入流读取下一个json编码值并保存在v指向的值里
	err := decoder.Decode(&CONFIG)
	if err != nil {
		panic(err)
	}
	// fmt.Println(CONFIG)
}
func VersionOrdinal(version string) string {
	// ISO/IEC 14651:2011
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			panic("VersionOrdinal: invalid version")
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}
func RemoveDuplicatesAndEmpty(a []string) (ret []string) {
	a_len := len(a)
	for i := 0; i < a_len; i++ {
		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}
func ReadReadAllSubKeyNames(path string) []string {
	ret := []string{}
	subkeys := ReadReadSubKeyNames(path)
	if len(subkeys) != 0 {
		for _, v := range subkeys {
			subPath := path + "\\" + v
			ret = append(ret, subPath)
			paths := ReadReadAllSubKeyNames(subPath)
			if (len(paths)) > 0 {
				ret = append(ret, paths...)
			}
		}
	}
	return ret
}
func ReadReadSubKeyNames(path string) []string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.READ)

	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	params, err := k.ReadSubKeyNames(0)
	if err != nil {
		log.Printf("Can't ReadSubKeyNames  %#v", err)
	}
	return params
}
func ReadValue(path string, name string) (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.READ)

	if err != nil {
		log.Fatal(err)
		return "", err
	}
	defer k.Close()

	v, _, err := k.GetStringValue(name)
	if err != nil {
		// log.Println(err)
		return "", err
	}
	return v, nil
}
func ReadValueNames(path string) []string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.READ)

	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	params, err := k.ReadValueNames(0)
	if err != nil {
		log.Printf("Can't ReadValueNames  %#v", err)
	}
	for _, param := range params {
		v, _, _ := k.GetStringValue(param)
		fmt.Printf("%s :%s %s\n", path, param, v)
	}
	return params
}

// func test() {
// 	cmd2 := exec.Command("reg", `query "HKLM\Software\Microsoft\NET Framework Setup\NDP" /s /v version | findstr /i version | sort /+26 /r`)
// 	buf, err := cmd2.Output()
// 	fmt.Printf("output: %s\n", buf)
// 	fmt.Printf("err: %v", err)
// }
func RunExec(fileEx FileEx) {
	if Exists(&fileEx) {
		cmd := exec.Command(fileEx.AbsPath)
		cmd.Dir = fileEx.AbsDir
		fmt.Println(fileEx.RunningTip)
		if fileEx.IsWaitRunning {
			cmd.Run()
		} else {
			cmd.Start()
		}
	}
}

// 判断所给路径文件/文件夹是否存在
func Exists(fileEx *FileEx) bool {
	_, err := os.Stat(fileEx.FileName) //os.Stat获取文件信息
	file := fileEx.FileName
	fileEx.AbsPath, _ = filepath.Abs(file)
	fileEx.AbsDir, _ = filepath.Abs(filepath.Dir(file))
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

//管道模式
func RunCMDPipe(m string) {
	cmd1 := exec.Command("reg", "query \"HKLM\\Software\\Microsoft\\NET Framework Setup\\NDP\" /s /v version | findstr /i version | sort /+26 /r")
	var outputbuf1 bytes.Buffer
	cmd1.Stdout = &outputbuf1 //设置输入
	if err := cmd1.Start(); err != nil {
		fmt.Println(err)
		return
	}
	if err := cmd1.Wait(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s", outputbuf1.Bytes())

}

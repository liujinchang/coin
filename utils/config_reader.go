package utils

import (
	"bufio"
	"io"
	"os"
	"strings"
)

/*
 * 读取key=value类型的配置文件
 * 		当前只开发读取属性文件，后期有时间可以增加读取yml文件、json文件、xml文件
 */
var config map[string]string

func initConfig(path string) map[string]string {
	config := make(map[string]string)
	if FileExists(path) {
		f, err := os.Open(path)
		defer f.Close()
		ErrorLog(err)

		r := bufio.NewReader(f)
		for {
			b, _, err := r.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			s := strings.TrimSpace(string(b))
			//当以#为开头时，为注释行，不去处理
			if strings.Index(s, "#") == 0 {
				continue
			}
			index := strings.Index(s, "=")
			if index < 0 {
				continue
			}
			key := strings.TrimSpace(s[:index])
			if len(key) == 0 {
				continue
			}
			value := strings.TrimSpace(s[index+1:])
			if len(value) == 0 {
				continue
			}
			config[key] = value
		}
	}
	return config
}
func GetConfig(file string, key string) string{
	if config == nil {
		config = initConfig(file)
	}
	value := config[key]
	switch key {
		case "known_nodes" :
			if value == "" {
				value = "localhost:3000"
			}
	}
	return value
}

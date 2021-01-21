package utils

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)
/*
 * 读取key=value类型的配置文件
 * 		当前只开发读取属性文件，后期有时间可以增加读取yml文件、json文件、xml文件
 */
func ReadConfig(path string) (map[string]string, error) {
	if FileExists(path) {
		config := make(map[string]string)
		f, err := os.Open(path)
		ErrorLog(err)
		defer f.Close()

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
		return config, nil
	} else {
		return nil, errors.New("Config file is not exist!")
	}
}

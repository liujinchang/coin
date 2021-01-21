package config

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"utils"
)

const (
	//根目录
	Root = ".bitcoin"
	//配置文件
	ConfigFile = "bitcoin.conf"
	//上限值
	UpperLimitTransactionInMemory = "upper_limit_transaction_in_memory"
	//下限值
	LowerLimitTransactionInMemory = "lower_limit_transaction_in_memory"
	//已知结点
	KnownNodes = "known_nodes"
)
var conf map[string]string
/*
 * 获取配置项的值
 */
func GetConfig(key string) string{
	if conf == nil {
		InitConfigs()
	}
	return conf[key]
}
/*
 * 从数据库配置里的文件
 */
func InitConfigs() {
	if conf != nil {
		return
	}
	config, err := utils.ReadConfig(Root+"/config/"+ConfigFile)
	defaultConfig := getDeaultConfig()
	//当file不存在时, 使用系统默认的配置
	if err != nil {
		log.Println("Config file is not exist,Use default config!")
		conf = defaultConfig
	} else {
		if ok, err := validConfig(config); !ok {
			/*
			 * 校验配置未通过
			 */
			if err != nil {
				utils.ErrorLog(err)
			} else {
				utils.ErrorLog(errors.New("Valid config is not pass!"))
			}
		} else {
			conf = config
			/*
			 * 校验配置通过
			 */
			if conf[UpperLimitTransactionInMemory] == "" {
				conf[UpperLimitTransactionInMemory] = defaultConfig[UpperLimitTransactionInMemory]
			}
			if conf[LowerLimitTransactionInMemory] == "" {
				conf[LowerLimitTransactionInMemory] = defaultConfig[LowerLimitTransactionInMemory]
			}
			if conf[KnownNodes] == "" {
				conf[KnownNodes] = defaultConfig[KnownNodes]
			}
		}
		log.Println("config message start")
		for k := range conf {
			log.Printf("%s = %s", k, conf[k])
		}
		log.Println("config message end")
	}
}
/*
 * 获取系统配置项的默认值，当从配置文件中读不到相关的配置时使用
 */
func getDeaultConfig() map[string]string {
	defaultConfig := make(map[string]string)
	defaultConfig[KnownNodes] = "localhost:3000"
	defaultConfig[UpperLimitTransactionInMemory] = "15"
	defaultConfig[LowerLimitTransactionInMemory] = "10"
	return defaultConfig
}
/*
 * 校验配置文件里的配置
 */
func validConfig(config map[string]string) (bool, error) {
	upperLimitTransactionInMemory := config[UpperLimitTransactionInMemory]
	lowerLimitTransactionInMemory := config[LowerLimitTransactionInMemory]
	if upperLimitTransactionInMemory != "" {
		if ok := isPositiveInteger(upperLimitTransactionInMemory); !ok {
			return ok, errors.New("upper_limit_transaction_in_memory is error!")
		}
	}
	if lowerLimitTransactionInMemory != "" {
		if ok := isPositiveInteger(lowerLimitTransactionInMemory); !ok {
			return ok, errors.New("lower_limit_transaction_in_memory is error!")
		}
	}
	upper, _ := strconv.Atoi(upperLimitTransactionInMemory)
	lower, _ := strconv.Atoi(lowerLimitTransactionInMemory)
	if upper < lower {
		return false, errors.New("lower_limit_transaction_in_memory is bigger than upper_limit_transaction_in_memory")
	}
	return true, nil
}
/*
 *　判断字符串是不是正整数
 */
func isPositiveInteger(str string) bool {
	reg := regexp.MustCompile(`[1-9]\d*`)
	return reg.MatchString(str)
}
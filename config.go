package main

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"reflect"
	//"regexp"
	"strconv"
	"sync"
)

var m sync.Mutex
var ConfigFile string
var File_Extension string
var Count_config_file []string

type GetConfigParamAsString func(string, *string, string) string
type GetConfigParamAsInt64 func(string, *string, string) int64
type GetConfigParamAsFloat64 func(string, *string, string) float64
type Config struct {
	GetConfigParamAsString  GetConfigParamAsString
	GetConfigParamAsInt64   GetConfigParamAsInt64
	GetConfigParamAsFloat64 GetConfigParamAsFloat64
}
type KeyFunc struct {
	Key          string
	CallBackFunc func()
}

var Pair []KeyFunc

const bitSize = 64 // Don't think about it to much. It's just 64 bits.
var MapConfig map[string]interface{}
var MapJson map[string]interface{}
var OldConfig = make(map[string]interface{})

func Decode(value interface{}) {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &MapJson)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func Common(key string, deep_key *string, default_val string) interface{} {
	var old_value interface{}
	value, found := MapConfig[key]
	if !found {
		old_value, found = OldConfig[key]
		if !found {
			return default_val
		}
		return old_value
	}
	if deep_key == nil {
		return value
	}

	Decode(value)

	deep_value, found := MapJson[*deep_key]
	if !found {
		Decode(old_value)
		deep_value_old, found := MapJson[*deep_key]
		if !found {

			return default_val
		}
		return deep_value_old
	}

	return deep_value
}

func MisMatchedKey(old map[string]interface{}, updated map[string]interface{}) []string {
	var keys []string
	for k, _ := range old {
		eq := reflect.DeepEqual(old[k], updated[k])
		if eq {

		} else {
			keys = append(keys, k)
		}
	}
	return keys
}

// func VisitFile(path string, info os.FileInfo, err error) error {

// 	if err != nil {

// 		fmt.Println(err)
// 		return nil
// 	}

// 	if info.IsDir() || filepath.Ext(path) != File_Extension {

// 		return nil
// 	}

// 	reg, err2 := regexp.Compile(ConfigFile)

// 	if err2 != nil {

// 		return err2
// 	}

// 	if reg.MatchString(info.Name()) {
// 		Count_config_file = append(Count_config_file, path)
// 		if len(Count_config_file) != 1 {
// 			panic("to many config files with same name")
// 		}
// 	}
//

// 	return nil
// }

func CallFuncIfExists(key []string) bool {
	fmt.Println(key)
	for _, K := range key {
		for _, v := range Pair {
			if v.Key == K {
				v.CallBackFunc()

			}
		}
	}
	return true
}

func InitConfig(file_path string, pair []KeyFunc) (*Config, error) {

	//ConfigFile = file_name_with_extension
	// File_Extension = filepath.Ext(file_name_with_extension)
	// err := filepath.Walk(".", VisitFile)

	// if err != nil {

	// }
	_, err := os.Open(file_path)
	if err == nil {
		viper.SetConfigFile(file_path)
	} else {
		filename := filepath.Base(file_path)
		file := fmt.Sprintf("%s%s", "./", filename)
		viper.SetConfigFile(file)
	}

	Pair = pair
	var config = new(Config)
	viper.AutomaticEnv()

	if MapConfig == nil {
		if err := viper.ReadInConfig(); err != nil {
			return config, err
		}
		err := viper.Unmarshal(&MapConfig)
		if err != nil {
			return config, err
		}
	}

	viper.OnConfigChange(func(e fsnotify.Event) {

		fmt.Println("Config file changed:", e.Name)

		b, err := json.Marshal(MapConfig)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(b, &OldConfig)
		if err != nil {
			fmt.Println("error:", err)
		}

		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("read failed")
		}

		err = viper.Unmarshal(&MapConfig)
		if err != nil {
			fmt.Println("unmarshal failed")
		}

		fmt.Println("config updated & Checking for any call_back func")

		key := MisMatchedKey(OldConfig, MapConfig)
		if key != nil {
			ok := CallFuncIfExists(key)
			if ok {
				fmt.Println("ok")
			}
		}

	})
	viper.WatchConfig()

	config = &Config{

		GetConfigParamAsString: func(key string, deep_key *string, default_val string) string {
			val := Common(key, deep_key, default_val)
			return fmt.Sprintf("%v", val)
		},
		GetConfigParamAsInt64: func(key string, deep_key *string, default_val string) int64 {
			val := Common(key, deep_key, default_val)
			Num, err := strconv.ParseInt(fmt.Sprintf("%v", val), 10, bitSize)
			if err != nil {
				fmt.Println("error:", err)
				Num, err := strconv.ParseInt(default_val, 10, bitSize)
				if err != nil {
					fmt.Println("error:", err)
				}
				return Num
			}
			return Num
		},
		GetConfigParamAsFloat64: func(key string, deep_key *string, default_val string) float64 {
			val := Common(key, deep_key, default_val)
			Num, err := strconv.ParseFloat(fmt.Sprintf("%v", val), bitSize)
			if err != nil {
				fmt.Println("error:", err)
				Num, err := strconv.ParseFloat(default_val, bitSize)
				if err != nil {
					fmt.Println("error:", err)
				}
				return Num
			}
			return Num
		},
	}

	return config, nil

}

func Prints1() {
	fmt.Println("db updated")
}
func Prints2() {
	fmt.Println("app updated")
}
func Prints3() {
	fmt.Println("key updated")
}
func main() {
	e := echo.New()
	var pair []KeyFunc
	pair = []KeyFunc{
		{Key: "db", CallBackFunc: Prints1},
		{Key: "app", CallBackFunc: Prints2},
		{Key: "v", CallBackFunc: Prints3},
	}
	config, err := InitConfig("./g/e/config.dev.json", pair)

	if err != nil {
		fmt.Println(err.Error())
	}
	p := "port"
	// val := config.GetConfigParamAsString("app", nil, "1010")
	// fmt.Println(val)
	val2 := config.GetConfigParamAsInt64("app", &p, "1010")
	fmt.Println(val2)
	e.Start(":8080")
}

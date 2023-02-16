package main

import (
	"encoding/json"
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"math/bits"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func (fi *FileInst) Init() error {
	info, err := os.Stat(fi.OriginalPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("文件 %s 不存在", fi.OriginalPath)
		}
	} else if info.IsDir() {
		return fmt.Errorf("%s 指向路径而非文件", fi.OriginalPath)
	} else {
		switch filepath.Ext(fi.FileName) {
		case ".json":
			fi.FileType = "JSON"
		case ".yaml":
			fi.FileType = "YAML"
		default:
			fmt.Print("文件的后缀名不是.json或.yaml，请指定文件类型\nj = JSON，y = yaml (j/y)\n")
			scanner.Scan()
			switch scanner.Text() {
			case "j":
				fi.FileType = "JSON"
			case "y":
				fi.FileType = "YAML"
			default:
				return fmt.Errorf("无法识别的输入 %s", scanner.Text())
			}
		}
	}
	return nil
}

func (fi *FileInst) Unmarshal() (error, map[string]interface{}) {
	var container = make(map[string]interface{})
	data, err := os.ReadFile(fi.OriginalPath)
	if err != nil {
		fmt.Printf("读取文件 %s 时出现错误：%s", fi.FileName, err)
	}
	switch fi.FileType {
	case "JSON":
		if err := json.Unmarshal(data, &container); err != nil {
			return fmt.Errorf("解析JSON时出现错误：%s", err), nil
		}
	case "YAML":
		if err := yaml.Unmarshal(data, &container); err != nil {
			return fmt.Errorf("解析YAML时出现错误：%s", err), nil
		}
	}
	return nil, container
}

func (fi *FileInst) Convert(data map[string]interface{}) error {
	newName := strings.TrimSuffix(fi.FileName, filepath.Ext(fi.FileName))
	if fi.FileType == "JSON" {
		newName += ".yaml"
	} else if fi.FileType == "YAML" {
		newName += ".json"
	}

	file, err := os.Create(fi.Directory + "/" + newName)
	if err != nil {
		return err
	}

	var keys []string

	switch fi.FileType {
	case "JSON":
		hasKeys := false
		for k, v := range data {
			if !strings.HasPrefix(k, "_") {
				keys = append(keys, k)
				continue
			} else {
				delete(data, k)
				val, ok := v.([]interface{})
				if ok {
					if k == "_title" {
						data["name"] = val[0]
					} else if k == "_author" {
						data["author"] = val[0]
					} else if k == "_version" {
						reg := regexp.MustCompile(`[\!\$\%\^\&\*\(\)\_\+\|\~\\\-\=\{\}\[\]\:\"\;\'\<\>\?\,\.\/]`)
						vad := reg.ReplaceAllString(val[0].(string), "")
						pi, err := strconv.ParseInt(vad, 10, bits.UintSize)
						if err != nil {
							return fmt.Errorf("转换日期时出现错误：%s", err)
						}
						data["version"] = pi
					} else if k == "_brief" {
						data["desc"] = val[0]
					} else if k == "_keys" {
						hasKeys = true
						data["includes"] = val
					} else if k == "_updateDate" {
						//
					} else {
						data[k] = val
					}
				}
				reg := regexp.MustCompile(`\{%(.*)\}`)
				for _, va := range data {
					val, ok := va.([]interface{})
					if ok {
						for k2, va2 := range val {
							val[k2] = reg.ReplaceAllString(va2.(string), `{$$$1}`)
						}
					}
					va = val
				}
			}
		}
		if !hasKeys {
			data["includes"] = keys
		}

		marshalledData, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("转换为YAML时出现错误：%s", err)
		}

		if _, err := file.Write(marshalledData); err != nil {
			return fmt.Errorf("写入文件时出现错误：%s", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("试图退出文件时出现错误：%s", err)
		}
		fmt.Printf("文件已经写入 %s\n", fi.Directory+"/"+newName)
	case "YAML":
		for k, v := range data {
			delete(data, k)
			if k == "includes" {
				data["_keys"] = v
			} else if k == "name" {
				v = []interface{}{v}
				data["_title"] = v
			} else if k == "author" || k == "version" {
				v = []interface{}{v}
				k2 := "_" + k
				data[k2] = v
			} else if k == "desc" {
				v = []interface{}{v}
				data["_brief"] = v
			} else if k == "command" {
				//
			} else {
				data[k] = v
			}
			reg := regexp.MustCompile(`\{\$(.*)\}`)
			for _, va := range data {
				val, ok := va.([]interface{})
				if ok {
					for k2, va2 := range val {
						switch va2.(type) {
						case int:
							val[k2] = reg.ReplaceAllString(strconv.Itoa(va2.(int)), `{%$1}`)
						default:
							val[k2] = reg.ReplaceAllString(va2.(string), `{%$1}`)
						}
					}
				}
				va = val
			}
		}

		marshalledData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("转换为JSON时出现错误：%s", err)
		}

		if _, err := file.Write(marshalledData); err != nil {
			return fmt.Errorf("写入文件时出现错误：%s", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("试图退出文件时出现错误：%s", err)
		}
		fmt.Printf("文件已经写入 %s\n", fi.Directory+"/"+newName)
	}
	return nil
}

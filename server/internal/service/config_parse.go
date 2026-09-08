package service

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"kenaito-conf/server/internal/model"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigParseService struct{}

func NewConfigParseService() *ConfigParseService {
	return &ConfigParseService{}
}

// ParseConfigToKVMap 解析配置内容，语法错误时返回 error 而不是静默返回空结果。
func (s *ConfigParseService) ParseConfigToKVMap(config *model.AppConfig) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	switch config.Type {
	case 1: // file
		name := strings.ToLower(config.Name)
		switch {
		case strings.HasSuffix(name, ".xml"):
			if err := s.parseXML(result, config.Contents); err != nil {
				return nil, fmt.Errorf("解析 XML 配置 %q 失败: %w", config.Name, err)
			}
		case strings.HasSuffix(name, ".yml"), strings.HasSuffix(name, ".yaml"):
			if err := s.parseYAML(result, config.Contents); err != nil {
				return nil, fmt.Errorf("解析 YAML 配置 %q 失败: %w", config.Name, err)
			}
		default:
			s.parseProperties(result, config.Contents)
		}
	case 2: // key-value
		result[config.Name] = config.Contents
	case 3: // JSON
		if err := json.Unmarshal([]byte(config.Contents), &result); err != nil {
			return nil, fmt.Errorf("解析 JSON 配置 %q 失败: %w", config.Name, err)
		}
	default:
		return nil, fmt.Errorf("不支持的配置类型: %d", config.Type)
	}

	// resolve placeholders
	s.resolvePlaceholders(result)
	return result, nil
}

// ValidateConfig 在保存配置前校验语法。
func (s *ConfigParseService) ValidateConfig(config *model.AppConfig) error {
	_, err := s.ParseConfigToKVMap(config)
	return err
}

func (s *ConfigParseService) parseProperties(result map[string]interface{}, content string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key != "" {
			result[key] = value
		}
	}
}

// parseXML 使用标准库解析 <entry key="...">value</entry> 结构，
// 支持嵌套标签、CDATA 和实体转义，语法错误会返回 error。
func (s *ConfigParseService) parseXML(result map[string]interface{}, content string) error {
	decoder := xml.NewDecoder(strings.NewReader(content))

	var (
		currentKey   string
		currentValue strings.Builder
		depth        int
	)

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if depth == 0 && t.Name.Local == "entry" {
				currentKey = ""
				for _, attr := range t.Attr {
					if attr.Name.Local == "key" {
						currentKey = attr.Value
						break
					}
				}
				currentValue.Reset()
				depth = 1
			} else if depth > 0 {
				depth++
			}
		case xml.CharData:
			if depth > 0 {
				currentValue.Write(t)
			}
		case xml.EndElement:
			if depth > 0 {
				depth--
				if depth == 0 {
					value := strings.TrimSpace(currentValue.String())
					if currentKey != "" && value != "" {
						result[currentKey] = value
					}
				}
			}
		}
	}

	if depth != 0 {
		return errors.New("XML 标签未闭合")
	}
	return nil
}

func (s *ConfigParseService) parseYAML(result map[string]interface{}, content string) error {
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return err
	}
	s.flattenYAML("", result, data)
	return nil
}

func (s *ConfigParseService) flattenYAML(prefix string, result map[string]interface{}, data map[string]interface{}) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if subMap, ok := v.(map[string]interface{}); ok {
			s.flattenYAML(key, result, subMap)
		} else {
			result[key] = v
		}
	}
}

func (s *ConfigParseService) resolvePlaceholders(result map[string]interface{}) {
	for key, val := range result {
		strVal, ok := val.(string)
		if !ok {
			continue
		}
		if strings.Contains(strVal, "${") {
			result[key] = s.replacePlaceholder(result, strVal)
		}
	}
}

func (s *ConfigParseService) replacePlaceholder(data map[string]interface{}, value string) string {
	result := value
	maxIterations := 100
	processedPlaceholders := make(map[string]bool)

	for i := 0; i < maxIterations; i++ {
		start := strings.Index(result, "${")
		if start < 0 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end < 0 {
			break
		}
		end += start

		placeholder := result[start+2 : end]
		defaultVal := ""
		refKey := placeholder

		// check for default value
		if idx := strings.Index(placeholder, ":"); idx >= 0 {
			refKey = placeholder[:idx]
			defaultVal = placeholder[idx+1:]
		}

		// 检测循环引用
		if processedPlaceholders[refKey] {
			// 已经处理过这个占位符，说明存在循环引用，保留原样
			break
		}

		var replacement string
		if v, ok := data[refKey]; ok {
			replacement = fmt.Sprintf("%v", v)
		} else if defaultVal != "" {
			replacement = defaultVal
		} else {
			replacement = "${" + placeholder + "}"
		}

		// 标记已处理
		processedPlaceholders[refKey] = true

		result = result[:start] + replacement + result[end+1:]
	}
	return result
}

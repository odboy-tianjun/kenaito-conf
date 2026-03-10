package service

import (
	"encoding/json"
	"fmt"
	"kenaito-conf/server/internal/model"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigParseService struct{}

func NewConfigParseService() *ConfigParseService {
	return &ConfigParseService{}
}

// ParseConfigToKVMap parses config content based on type
func (s *ConfigParseService) ParseConfigToKVMap(config *model.AppConfig) map[string]interface{} {
	result := make(map[string]interface{})

	switch config.Type {
	case 1: // file
		if strings.HasSuffix(strings.ToLower(config.Name), ".xml") {
			s.parseXML(result, config.Contents)
		} else if strings.HasSuffix(strings.ToLower(config.Name), ".yml") ||
			strings.HasSuffix(strings.ToLower(config.Name), ".yaml") {
			s.parseYAML(result, config.Contents)
		} else {
			s.parseProperties(result, config.Contents)
		}
	case 2: // key-value
		result[config.Name] = config.Contents
	case 3: // JSON
		json.Unmarshal([]byte(config.Contents), &result)
	}

	// resolve placeholders
	s.resolvePlaceholders(result)
	return result
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

func (s *ConfigParseService) parseXML(result map[string]interface{}, content string) {
	// simple XML parsing for <entry key="...">value</entry> format
	// This is a simplified version - for production use a proper XML parser
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "<entry") {
			continue
		}
		// extract key
		keyStart := strings.Index(line, `key="`)
		if keyStart < 0 {
			continue
		}
		keyStart += 5
		keyEnd := strings.Index(line[keyStart:], `"`)
		if keyEnd < 0 {
			continue
		}
		key := line[keyStart : keyStart+keyEnd]

		// extract value
		valStart := strings.Index(line, ">")
		if valStart < 0 {
			continue
		}
		valEnd := strings.Index(line, "</entry>")
		if valEnd < 0 {
			continue
		}
		value := strings.TrimSpace(line[valStart+1 : valEnd])
		if key != "" && value != "" {
			result[key] = value
		}
	}
}

func (s *ConfigParseService) parseYAML(result map[string]interface{}, content string) {
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return
	}
	s.flattenYAML("", result, data)
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

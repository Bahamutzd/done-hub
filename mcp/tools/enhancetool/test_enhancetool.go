package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// 模拟ToolCall结构用于测试
type ToolCall struct {
	Type       string                 `json:"type"`
	Function   FunctionCall          `json:"function"`
	ID         string                 `json:"id,omitempty"`
	Parameters map[string]interface{} `json:"parameters"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func main() {
	// 测试数据
	testToolCalls := []ToolCall{
		{
			Type: "function",
			Function: FunctionCall{
				Name: "calculator",
				Arguments: `{"operation": "add", "x": 10, "y": 20}`,
			},
		},
		{
			Type: "function",
			Function: FunctionCall{
				Name: "calculator",
				Arguments: `{"operation": "multiply", "x": "5", "y": "3"}`, // 字符串数字
			},
		},
	}

	// 测试JSON序列化
	result, err := json.MarshalIndent(testToolCalls, "", "  ")
	if err != nil {
		log.Fatalf("JSON序列化失败: %v", err)
	}

	fmt.Printf("测试数据:\n%s\n", string(result))
	fmt.Println("enhancetool工具创建成功！")
}
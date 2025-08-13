// Package enhancetool 提供了工具调用参数的容错处理功能
// Package enhancetool provides error tolerance handling for tool call parameters
package enhancetool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
)

const NAME = "enhancetool"

// EnhanceTool 实现了工具调用参数容错处理工具
type EnhanceTool struct{}

// ToolCall 表示一个工具调用
type ToolCall struct {
	Type       string                 `json:"type"`
	Function   FunctionCall          `json:"function"`
	ID         string                 `json:"id,omitempty"`
	Parameters map[string]interface{} `json:"parameters"`
}

// FunctionCall 表示函数调用信息
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCallParam 表示enhancetool的参数结构
type ToolCallParam struct {
	ToolCalls []ToolCall `json:"tool_calls" description:"需要处理的工具调用数组" required:"true"`
	StrictMode bool      `json:"strict_mode" description:"是否启用严格模式，默认为false" required:"false"`
}

// ParameterSchema 表示参数的schema信息
type ParameterSchema struct {
	Type        string      `json:"type"`
	Description string     `json:"description"`
	Required    []string    `json:"required"`
	Enum        []string    `json:"enum,omitempty"`
	Properties  map[string]interface{} `json:"properties"`
}

// GetTool 返回enhancetool的定义
func (e *EnhanceTool) GetTool() *protocol.Tool {
	// 创建一个新的enhancetool
	enhanceTool, _ := protocol.NewTool(
		NAME,
		"对LLM返回的工具调用参数增加一层容错处理，验证和修正参数格式、类型和缺失值",
		ToolCallParam{},
	)

	return enhanceTool
}

// HandleRequest 处理enhancetool的请求
func (e *EnhanceTool) HandleRequest(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	param := ToolCallParam{}

	if err := protocol.VerifyAndUnmarshal(req.RawArguments, &param); err != nil {
		return nil, fmt.Errorf("参数解析失败: %v", err)
	}

	if len(param.ToolCalls) == 0 {
		return nil, fmt.Errorf("tool_calls参数不能为空")
	}

	// 处理每个工具调用
	enhancedCalls := make([]ToolCall, 0, len(param.ToolCalls))
	var errorMessages []string

	for i, toolCall := range param.ToolCalls {
		enhancedCall, err := e.processToolCall(toolCall, param.StrictMode)
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("工具调用[%d]处理失败: %v", i, err))
			if param.StrictMode {
				continue // 严格模式下跳过错误调用
			}
			// 非严格模式下，使用原始调用
			enhancedCall = toolCall
		}
		enhancedCalls = append(enhancedCalls, enhancedCall)
	}

	// 构建响应内容
	content := e.buildResponseContent(enhancedCalls, errorMessages)

	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}, nil
}

// processToolCall 处理单个工具调用
func (e *EnhanceTool) processToolCall(toolCall ToolCall, strictMode bool) (ToolCall, error) {
	enhancedCall := toolCall

	// 解析函数参数
	if toolCall.Function.Arguments != "" {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			if strictMode {
				return toolCall, fmt.Errorf("参数JSON解析失败: %v", err)
			}
			// 非严格模式下尝试修复常见JSON错误
			args, err = e.tryFixJSON(toolCall.Function.Arguments)
			if err != nil {
				return toolCall, fmt.Errorf("参数JSON解析失败且无法修复: %v", err)
			}
		}

		// 验证和修复参数
		enhancedArgs, err := e.validateAndFixParameters(args, toolCall.Function.Name)
		if err != nil {
			if strictMode {
				return toolCall, fmt.Errorf("参数验证失败: %v", err)
			}
			// 非严格模式下使用修复后的参数
			enhancedArgs = args
		}

		// 重新序列化参数
		argsJSON, err := json.Marshal(enhancedArgs)
		if err != nil {
			return toolCall, fmt.Errorf("参数重新序列化失败: %v", err)
		}
		enhancedCall.Function.Arguments = string(argsJSON)
		enhancedCall.Parameters = enhancedArgs
	}

	return enhancedCall, nil
}

// tryFixJSON 尝试修复常见JSON错误
func (e *EnhanceTool) tryFixJSON(jsonStr string) (map[string]interface{}, error) {
	// 尝试修复常见的JSON格式问题
	fixedStr := strings.TrimSpace(jsonStr)
	
	// 移除可能的前缀/后缀
	fixedStr = strings.TrimPrefix(fixedStr, "```json")
	fixedStr = strings.TrimSuffix(fixedStr, "```")
	fixedStr = strings.TrimSpace(fixedStr)

	// 尝试解析
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(fixedStr), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// validateAndFixParameters 验证和修复参数
func (e *EnhanceTool) validateAndFixParameters(args map[string]interface{}, functionName string) (map[string]interface{}, error) {
	fixedArgs := make(map[string]interface{})
	
	for key, value := range args {
		// 类型转换和修复
		fixedValue, err := e.fixParameterType(value)
		if err != nil {
			return nil, fmt.Errorf("参数'%s'类型修复失败: %v", key, err)
		}
		fixedArgs[key] = fixedValue
	}

	// 根据函数名进行特定验证和修复
	switch functionName {
	case "calculator":
		return e.validateCalculatorParams(fixedArgs)
	case "available_model":
		return e.validateAvailableModelParams(fixedArgs)
	// 可以添加更多函数的特定验证规则
	default:
		return fixedArgs, nil
	}
}

// fixParameterType 修复参数类型
func (e *EnhanceTool) fixParameterType(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case string:
		// 尝试将字符串转换为适当的类型
		return e.tryConvertString(v)
	case float64:
		// JSON数字默认解析为float64
		if v == float64(int(v)) {
			return int(v), nil // 如果是整数，转换为int
		}
		return v, nil
	case int, int32, int64:
		return v, nil
	case bool:
		return v, nil
	case map[string]interface{}, []interface{}:
		return v, nil
	default:
		// 对于其他类型，转换为字符串
		return fmt.Sprintf("%v", v), nil
	}
}

// tryConvertString 尝试将字符串转换为更合适的类型
func (e *EnhanceTool) tryConvertString(str string) (interface{}, error) {
	str = strings.TrimSpace(str)
	
	// 尝试解析为布尔值
	if strings.ToLower(str) == "true" {
		return true, nil
	}
	if strings.ToLower(str) == "false" {
		return false, nil
	}
	
	// 尝试解析为数字
	if num, err := strconv.ParseInt(str, 10, 64); err == nil {
		return num, nil
	}
	if num, err := strconv.ParseFloat(str, 64); err == nil {
		return num, nil
	}
	
	// 返回原始字符串
	return str, nil
}

// validateCalculatorParams 验证计算器参数
func (e *EnhanceTool) validateCalculatorParams(args map[string]interface{}) (map[string]interface{}, error) {
	required := []string{"operation", "x", "y"}
	for _, key := range required {
		if _, exists := args[key]; !exists {
			return nil, fmt.Errorf("缺少必需参数: %s", key)
		}
	}

	// 验证operation参数
	if op, ok := args["operation"].(string); ok {
		validOps := []string{"add", "subtract", "multiply", "divide"}
		valid := false
		for _, validOp := range validOps {
			if op == validOp {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("无效的operation参数: %s", op)
		}
	}

	// 验证数字参数
	for _, key := range []string{"x", "y"} {
		if _, ok := args[key].(float64); !ok {
			if _, ok := args[key].(int); !ok {
				return nil, fmt.Errorf("参数'%s'必须是数字", key)
			}
		}
	}

	return args, nil
}

// validateAvailableModelParams 验证可用模型参数
func (e *EnhanceTool) validateAvailableModelParams(args map[string]interface{}) (map[string]interface{}, error) {
	// 这里可以添加特定于available_model工具的验证逻辑
	// 目前只进行基本验证
	return args, nil
}

// buildResponseContent 构建响应内容
func (e *EnhanceTool) buildResponseContent(enhancedCalls []ToolCall, errors []string) string {
	// 生成处理后的工具调用JSON
	resultJSON, err := json.MarshalIndent(enhancedCalls, "", "  ")
	if err != nil {
		resultJSON = []byte("[]")
	}

	content := fmt.Sprintf("处理后的工具调用:\n%s", string(resultJSON))

	if len(errors) > 0 {
		content += "\n\n处理过程中的错误:\n"
		for _, err := range errors {
			content += fmt.Sprintf("- %s\n", err)
		}
	}

	return content
}
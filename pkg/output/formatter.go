package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Format 输出格式类型
type Format string

const (
	FormatText     Format = "text"     // 纯文本格式
	FormatJSON     Format = "json"     // JSON 格式
	FormatYAML     Format = "yaml"     // YAML 格式
	FormatTemplate Format = "template" // 自定义模板格式
	FormatTable    Format = "table"    // 表格格式
)

// Formatter 输出格式化器
type Formatter struct {
	format         Format
	writer         io.Writer
	templateString string
	quiet          bool
}

// NewFormatter 创建新的格式化器
func NewFormatter(format Format, writer io.Writer) *Formatter {
	if writer == nil {
		writer = os.Stdout
	}
	return &Formatter{
		format: format,
		writer: writer,
		quiet:  false,
	}
}

// SetTemplate 设置自定义模板
func (f *Formatter) SetTemplate(tmpl string) {
	f.templateString = tmpl
}

// SetQuiet 设置静默模式
func (f *Formatter) SetQuiet(quiet bool) {
	f.quiet = quiet
}

// Print 打印普通消息
func (f *Formatter) Print(message string) {
	if f.quiet {
		return
	}
	fmt.Fprintln(f.writer, message)
}

// Printf 格式化打印
func (f *Formatter) Printf(format string, args ...interface{}) {
	if f.quiet {
		return
	}
	fmt.Fprintf(f.writer, format, args...)
}

// Success 打印成功消息
func (f *Formatter) Success(message string) {
	if f.quiet {
		return
	}
	fmt.Fprintf(f.writer, "✓ %s\n", message)
}

// Error 打印错误消息
func (f *Formatter) Error(message string) {
	fmt.Fprintf(f.writer, "✗ %s\n", message)
}

// Warning 打印警告消息
func (f *Formatter) Warning(message string) {
	if f.quiet {
		return
	}
	fmt.Fprintf(f.writer, "⚠ %s\n", message)
}

// Info 打印信息消息
func (f *Formatter) Info(message string) {
	if f.quiet {
		return
	}
	fmt.Fprintf(f.writer, "ℹ %s\n", message)
}

// Output 输出结构化数据
func (f *Formatter) Output(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.outputJSON(data)
	case FormatYAML:
		return f.outputYAML(data)
	case FormatTemplate:
		return f.outputTemplate(data)
	case FormatTable:
		return f.outputTable(data)
	case FormatText:
		fallthrough
	default:
		return f.outputText(data)
	}
}

// outputJSON 输出 JSON 格式
func (f *Formatter) outputJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputYAML 输出 YAML 格式
func (f *Formatter) outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

// outputTemplate 使用自定义模板输出
func (f *Formatter) outputTemplate(data interface{}) error {
	if f.templateString == "" {
		return fmt.Errorf("template not set")
	}

	tmpl, err := template.New("output").Parse(f.templateString)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(f.writer, data)
}

// outputText 输出纯文本格式
func (f *Formatter) outputText(data interface{}) error {
	fmt.Fprintln(f.writer, data)
	return nil
}

// outputTable 输出表格格式
func (f *Formatter) outputTable(data interface{}) error {
	w := tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// 尝试将数据转换为可迭代的格式
	switch v := data.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil
		}
		// 打印表头和数据
		return f.printTableFromSlice(w, v)
	case map[string]interface{}:
		// 打印键值对
		return f.printTableFromMap(w, v)
	default:
		// 回退到文本格式
		return f.outputText(data)
	}
}

// printTableFromSlice 从切片打印表格
func (f *Formatter) printTableFromSlice(w *tabwriter.Writer, data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// 假设第一个元素是 map，获取表头
	if firstItem, ok := data[0].(map[string]interface{}); ok {
		// 打印表头
		for key := range firstItem {
			fmt.Fprintf(w, "%s\t", key)
		}
		fmt.Fprintln(w)

		// 打印数据行
		for _, item := range data {
			if row, ok := item.(map[string]interface{}); ok {
				for key := range firstItem {
					fmt.Fprintf(w, "%v\t", row[key])
				}
				fmt.Fprintln(w)
			}
		}
	}

	return nil
}

// printTableFromMap 从 map 打印表格
func (f *Formatter) printTableFromMap(w *tabwriter.Writer, data map[string]interface{}) error {
	fmt.Fprintln(w, "KEY\tVALUE")
	for key, value := range data {
		fmt.Fprintf(w, "%s\t%v\n", key, value)
	}
	return nil
}

// Summary 输出总结信息
type Summary struct {
	Total     int      `json:"total" yaml:"total"`
	Success   int      `json:"success" yaml:"success"`
	Failed    int      `json:"failed" yaml:"failed"`
	Skipped   int      `json:"skipped" yaml:"skipped"`
	Errors    []string `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings  []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	StartTime string   `json:"start_time" yaml:"start_time"`
	EndTime   string   `json:"end_time" yaml:"end_time"`
	Duration  string   `json:"duration" yaml:"duration"`
}

// PrintSummary 打印总结信息
func (f *Formatter) PrintSummary(summary *Summary) error {
	if f.format == FormatJSON || f.format == FormatYAML {
		return f.Output(summary)
	}

	// 文本格式输出
	fmt.Fprintln(f.writer, "\n===== Summary =====")
	fmt.Fprintf(f.writer, "Total:    %d\n", summary.Total)
	fmt.Fprintf(f.writer, "Success:  %d\n", summary.Success)
	fmt.Fprintf(f.writer, "Failed:   %d\n", summary.Failed)
	fmt.Fprintf(f.writer, "Skipped:  %d\n", summary.Skipped)
	fmt.Fprintf(f.writer, "Duration: %s\n", summary.Duration)

	if len(summary.Errors) > 0 {
		fmt.Fprintln(f.writer, "\nErrors:")
		for _, err := range summary.Errors {
			fmt.Fprintf(f.writer, "  - %s\n", err)
		}
	}

	if len(summary.Warnings) > 0 {
		fmt.Fprintln(f.writer, "\nWarnings:")
		for _, warn := range summary.Warnings {
			fmt.Fprintf(f.writer, "  - %s\n", warn)
		}
	}

	fmt.Fprintln(f.writer, "===================")
	return nil
}

// Package color 提供终端 ANSI 彩色输出的便捷封装。
// 支持常见前景/背景色、亮色、文本样式、8bit(256色)与24bit(RGB)。
// 使用示例：
//
//	fmt.Println(color.Red("错误"), color.Bold(color.Yellow("警告")))
//	fmt.Println(color.RGB(255, 0, 128, "真彩色"))
//	fmt.Println(color.Index(202, "8bit橙色"), color.BgIndex(27, "8bit海蓝"))
//	fmt.Println(color.Wrap("组合", color.FgBlack, color.BoldAttr, color.UnderlineAttr))
package color

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
)

// 是否启用颜色（原子布尔，便于并发环境动态开关）
var enabled atomic.Bool

func init() {
	enabled.Store(true)
}

// Enable 开启彩色输出（默认已开启）
func Enable() { enabled.Store(true) }

// Disable 关闭彩色输出（所有方法将返回原文，不带颜色）
func Disable() { enabled.Store(false) }

// ----------- ANSI 基础码 -----------
const (
	esc   = "\x1b["
	reset = esc + "0m"
)

// Reset 返回 ANSI 重置码（一般无需手动用，本包会自动补）
func Reset() string { return reset }

// ----------- 属性/样式 枚举（便于组合） -----------
// Attr 表示一个 ANSI SGR 属性（选择图形渲染属性）
// 例如 31=红色前景, 1=粗体, 4=下划线 等。
type Attr int

// 文本样式
const (
	ResetAttr     Attr = 0 // 重置所有属性
	BoldAttr      Attr = 1 // 粗体/高亮
	DimAttr       Attr = 2 // 微亮/变淡
	ItalicAttr    Attr = 3 // 斜体（有些终端不支持）
	UnderlineAttr Attr = 4 // 下划线
	BlinkAttr     Attr = 5 // 闪烁（大多终端禁用）
	InverseAttr   Attr = 7 // 反色（前景与背景互换）
	HiddenAttr    Attr = 8 // 隐藏（常用于密码输入）
	StrikeAttr    Attr = 9 // 删除线
)

// 标准前景色
const (
	FgBlack Attr = 30 + iota
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

// 亮色前景
const (
	FgBrightBlack Attr = 90 + iota
	FgBrightRed
	FgBrightGreen
	FgBrightYellow
	FgBrightBlue
	FgBrightMagenta
	FgBrightCyan
	FgBrightWhite
)

// 标准背景色
const (
	BgBlack Attr = 40 + iota
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// 亮色背景
const (
	BgBrightBlack Attr = 100 + iota
	BgBrightRed
	BgBrightGreen
	BgBrightYellow
	BgBrightBlue
	BgBrightMagenta
	BgBrightCyan
	BgBrightWhite
)

// SprintAttr 把多个 Attr 组装成形如 "\x1b[31;1m" 的前缀。
func SprintAttr(attrs ...Attr) string {
	if len(attrs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(esc)
	for i, a := range attrs {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(strconv.Itoa(int(a)))
	}
	b.WriteByte('m')
	return b.String()
}

// Wrap 使用若干 Attr 包裹文本，自动在末尾追加 Reset。
func Wrap(s string, attrs ...Attr) string {
	if !enabled.Load() || len(attrs) == 0 || s == "" {
		return s
	}
	return SprintAttr(attrs...) + s + reset
}

// ----------- 8bit(256色) 与 24bit 真彩 -----------

// Index 返回 8bit 前景色（0-255），示例：Index(202, "橙色")
func Index(idx int, s string) string {
	if !enabled.Load() || s == "" {
		return s
	}
	if idx < 0 {
		idx = 0
	} else if idx > 255 {
		idx = 255
	}
	return fmt.Sprintf("%s38;5;%dm%s%s", esc, idx, s, reset)
}

// BgIndex 返回 8bit 背景色（0-255）
func BgIndex(idx int, s string) string {
	if !enabled.Load() || s == "" {
		return s
	}
	if idx < 0 {
		idx = 0
	} else if idx > 255 {
		idx = 255
	}
	return fmt.Sprintf("%s48;5;%dm%s%s", esc, idx, s, reset)
}

// RGB 使用 24bit 真彩前景色，r/g/b 范围 0-255
func RGB(r, g, b int, s string) string {
	if !enabled.Load() || s == "" {
		return s
	}
	r, g, b = clamp255(r), clamp255(g), clamp255(b)
	return fmt.Sprintf("%s38;2;%d;%d;%dm%s%s", esc, r, g, b, s, reset)
}

// BgRGB 使用 24bit 真彩背景色
func BgRGB(r, g, b int, s string) string {
	if !enabled.Load() || s == "" {
		return s
	}
	r, g, b = clamp255(r), clamp255(g), clamp255(b)
	return fmt.Sprintf("%s48;2;%d;%d;%dm%s%s", esc, r, g, b, s, reset)
}

func clamp255(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// ----------- 便捷前景色函数 -----------

func Black(s string) string         { return Wrap(s, FgBlack) }
func Red(s string) string           { return Wrap(s, FgRed) }
func Green(s string) string         { return Wrap(s, FgGreen) }
func Yellow(s string) string        { return Wrap(s, FgYellow) }
func Blue(s string) string          { return Wrap(s, FgBlue) }
func Magenta(s string) string       { return Wrap(s, FgMagenta) }
func Cyan(s string) string          { return Wrap(s, FgCyan) }
func White(s string) string         { return Wrap(s, FgWhite) }
func BrightBlack(s string) string   { return Wrap(s, FgBrightBlack) }
func BrightRed(s string) string     { return Wrap(s, FgBrightRed) }
func BrightGreen(s string) string   { return Wrap(s, FgBrightGreen) }
func BrightYellow(s string) string  { return Wrap(s, FgBrightYellow) }
func BrightBlue(s string) string    { return Wrap(s, FgBrightBlue) }
func BrightMagenta(s string) string { return Wrap(s, FgBrightMagenta) }
func BrightCyan(s string) string    { return Wrap(s, FgBrightCyan) }
func BrightWhite(s string) string   { return Wrap(s, FgBrightWhite) }

// ----------- 便捷背景色函数 -----------

func BgBlackText(s string) string         { return Wrap(s, BgBlack) }
func BgRedText(s string) string           { return Wrap(s, BgRed) }
func BgGreenText(s string) string         { return Wrap(s, BgGreen) }
func BgYellowText(s string) string        { return Wrap(s, BgYellow) }
func BgBlueText(s string) string          { return Wrap(s, BgBlue) }
func BgMagentaText(s string) string       { return Wrap(s, BgMagenta) }
func BgCyanText(s string) string          { return Wrap(s, BgCyan) }
func BgWhiteText(s string) string         { return Wrap(s, BgWhite) }
func BgBrightBlackText(s string) string   { return Wrap(s, BgBrightBlack) }
func BgBrightRedText(s string) string     { return Wrap(s, BgBrightRed) }
func BgBrightGreenText(s string) string   { return Wrap(s, BgBrightGreen) }
func BgBrightYellowText(s string) string  { return Wrap(s, BgBrightYellow) }
func BgBrightBlueText(s string) string    { return Wrap(s, BgBrightBlue) }
func BgBrightMagentaText(s string) string { return Wrap(s, BgBrightMagenta) }
func BgBrightCyanText(s string) string    { return Wrap(s, BgBrightCyan) }
func BgBrightWhiteText(s string) string   { return Wrap(s, BgBrightWhite) }

// ----------- 便捷样式函数 -----------

func Bold(s string) string      { return Wrap(s, BoldAttr) }
func Dim(s string) string       { return Wrap(s, DimAttr) }
func Italic(s string) string    { return Wrap(s, ItalicAttr) }
func Underline(s string) string { return Wrap(s, UnderlineAttr) }
func Blink(s string) string     { return Wrap(s, BlinkAttr) }
func Inverse(s string) string   { return Wrap(s, InverseAttr) }
func Hidden(s string) string    { return Wrap(s, HiddenAttr) }
func Strike(s string) string    { return Wrap(s, StrikeAttr) }

// Mix 组合常见前景色 + 样式，示例：Mix("警告", FgYellow, BoldAttr)
func Mix(s string, attrs ...Attr) string { return Wrap(s, attrs...) }

// ----------- ANSI 清理 -----------

// Strip 移除字符串里的 ANSI 转义（用于日志/落盘）
var reANSI = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func Strip(s string) string {
	return reANSI.ReplaceAllString(s, "")
}

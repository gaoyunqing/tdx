package protocol

import "strings"

// 通达信盘后逐股统计数据文件(随 zhb.zip 下发, GBK 文本, `|` 分隔, 每行一只股票, 全市场)。
const (
	FileTdxStat  = "tdxstat.cfg"  // 个股综合统计指标(35 字段)
	FileTdxStat2 = "tdxstat2.cfg" // 个股资金流向 + 板块归属(21 字段)
)

// TdxStat 个股综合统计指标(来自 tdxstat.cfg, 35 字段)。
//
// 仅 Market/Code/Date 字段语义确定; 其余比率/市值/阶段涨跌/股本等字段为通达信内部格式,
// 无官方文档, 不强行命名, 完整原始字段保留在 Fields(下标与文件列一一对应, 0 基)。
// 经验推断(未核验): [2]换手率% [3]涨跌幅% [6..10]量比/市盈/市净等 [11,14]流通/总市值
// [17..21]阶段涨跌幅 [24,34]股本。需要精确语义请按实盘核验后再用 Fields 取值。
type TdxStat struct {
	Market uint8    // [0] 市场 0=深 1=沪 2=京
	Code   string   // [1] 证券代码
	Date   string   // [4] 数据日期 YYYYMMDD
	Fields []string // 全部 35 个原始字段
}

// TdxStat2 个股资金流向 + 板块归属(来自 tdxstat2.cfg, 21 字段)。
//
// BlockIndex([13]) 为该股所属/领涨板块指数代码(880xxx 行业概念 / 881xxx 地域), 已核验,
// 可能为空(部分京市/无归属个股)。其余资金分档/价格字段语义为推断, 原始值保留在 Fields。
type TdxStat2 struct {
	Market     uint8    // [0] 市场
	Code       string   // [1] 证券代码
	Date       string   // [2] 数据日期 YYYYMMDD
	BlockIndex string   // [13] 板块指数代码(id), 可能为空
	Fields     []string // 全部 21 个原始字段
}

// splitStatLines 按行分割 GBK 文本并解码, 去空行。
func splitStatLines(data []byte) []string {
	return strings.Split(string(UTF8ToGBK(data)), "\n")
}

func field(f []string, i int) string {
	if i < len(f) {
		return f[i]
	}
	return ""
}

// ParseTdxStat 解析 tdxstat.cfg → 个股综合统计指标。
func ParseTdxStat(data []byte) []*TdxStat {
	lines := splitStatLines(data)
	out := make([]*TdxStat, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimRight(ln, "\r")
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		f := strings.Split(ln, "|")
		if len(f) < 5 || f[1] == "" {
			continue
		}
		out = append(out, &TdxStat{
			Market: uint8(Uint16FromStr(f[0])),
			Code:   f[1],
			Date:   field(f, 4),
			Fields: f,
		})
	}
	return out
}

// ParseTdxStat2 解析 tdxstat2.cfg → 个股资金流向 + 板块归属。
func ParseTdxStat2(data []byte) []*TdxStat2 {
	lines := splitStatLines(data)
	out := make([]*TdxStat2, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimRight(ln, "\r")
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		f := strings.Split(ln, "|")
		if len(f) < 14 || f[1] == "" {
			continue
		}
		out = append(out, &TdxStat2{
			Market:     uint8(Uint16FromStr(f[0])),
			Code:       f[1],
			Date:       field(f, 2),
			BlockIndex: field(f, 13),
			Fields:     f,
		})
	}
	return out
}

// StockBlockIndex 从 tdxstat2 数据提取 证券代码→所属板块指数代码(id) 映射, 跳过无归属个股。
// 这是 block_*.dat(板块→成分股) 的反向映射, 且免名称匹配。
func StockBlockIndex(stats []*TdxStat2) map[string]string {
	m := make(map[string]string, len(stats))
	for _, s := range stats {
		if s.BlockIndex != "" {
			m[s.Code] = s.BlockIndex
		}
	}
	return m
}

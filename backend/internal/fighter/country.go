package fighter

import "strings"

var countryToZH = map[string]string{
	"brazil":        "巴西",
	"china":         "中国",
	"england":       "英格兰",
	"france":        "法国",
	"georgia":       "格鲁吉亚",
	"ireland":       "爱尔兰",
	"kazakhstan":    "哈萨克斯坦",
	"kyrgyzstan":    "吉尔吉斯斯坦",
	"mexico":        "墨西哥",
	"myanmar":       "缅甸",
	"netherlands":   "荷兰",
	"new zealand":   "新西兰",
	"nigeria":       "尼日利亚",
	"poland":        "波兰",
	"russia":        "俄罗斯",
	"south korea":   "韩国",
	"spain":         "西班牙",
	"sweden":        "瑞典",
	"thailand":      "泰国",
	"usa":           "美国",
	"united states": "美国",
}

// TranslateCountryToZH maps common UFC country names/abbreviations to Chinese.
func TranslateCountryToZH(country string) string {
	text := strings.TrimSpace(country)
	if text == "" {
		return ""
	}
	key := strings.ToLower(text)
	if value, ok := countryToZH[key]; ok {
		return value
	}
	return ""
}

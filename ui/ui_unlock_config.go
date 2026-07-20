package ui

// unlockRegionCodes maps Select option index to the region code string passed to MediaTest.
var unlockRegionCodes = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21",
}

// unlockRegionLabelsZH are the display labels shown in the Chinese UI, parallel to unlockRegionCodes.
var unlockRegionLabelsZH = []string{
	"0: 跨国平台",
	"1: 跨国+台湾",
	"2: 跨国+香港",
	"3: 跨国+日本",
	"4: 跨国+韩国",
	"5: 跨国+北美",
	"6: 跨国+南美",
	"7: 跨国+欧洲",
	"8: 跨国+非洲",
	"9: 跨国+大洋洲",
	"10: 仅台湾",
	"11: 仅香港",
	"12: 仅日本",
	"13: 仅韩国",
	"14: 仅北美",
	"15: 仅南美",
	"16: 仅欧洲",
	"17: 仅非洲",
	"18: 仅大洋洲",
	"19: 仅体育",
	"20: 全部平台",
	"21: 仅 AI 平台",
}

// unlockRegionLabelsEN are the display labels shown in the English UI, parallel to unlockRegionCodes.
var unlockRegionLabelsEN = []string{
	"0: International",
	"1: Intl+Taiwan",
	"2: Intl+HongKong",
	"3: Intl+Japan",
	"4: Intl+Korea",
	"5: Intl+N.America",
	"6: Intl+S.America",
	"7: Intl+Europe",
	"8: Intl+Africa",
	"9: Intl+Oceania",
	"10: Taiwan only",
	"11: HongKong only",
	"12: Japan only",
	"13: Korea only",
	"14: N.America only",
	"15: S.America only",
	"16: Europe only",
	"17: Africa only",
	"18: Oceania only",
	"19: Sports only",
	"20: All platforms",
	"21: AI only",
}

// unlockRegionLabelsForLang returns the label slice for the given UI language.
func unlockRegionLabelsForLang(lang string) []string {
	if lang == langEN {
		return unlockRegionLabelsEN
	}
	return unlockRegionLabelsZH
}

// unlockRegionLabelToCode converts a displayed label to its numeric code string.
// Returns "0" if no match is found.
func unlockRegionLabelToCode(label, lang string) string {
	labels := unlockRegionLabelsForLang(lang)
	for i, l := range labels {
		if l == label {
			return unlockRegionCodes[i]
		}
	}
	return "0"
}

// unlockRegionCodeToLabel converts a numeric code string to the displayed label for the given language.
// Returns the first label if no match is found.
func unlockRegionCodeToLabel(code, lang string) string {
	labels := unlockRegionLabelsForLang(lang)
	for i, c := range unlockRegionCodes {
		if c == code {
			if i < len(labels) {
				return labels[i]
			}
		}
	}
	return labels[0]
}

// unlockIpVersionOptions returns the options for the IP-version selector.
// These strings are used as both the Select display value and the code passed to MediaTest.
var unlockIpVersionOptions = []string{"auto", "ipv4", "ipv6"}

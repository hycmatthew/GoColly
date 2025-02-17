package main

import (
	"regexp"
	"strings"
)

func GetLastSegment(url string) string {
	// Find the last occurrence of '/'
	lastSlashIndex := strings.LastIndex(url, "/")
	if lastSlashIndex == -1 || lastSlashIndex == len(url)-1 {
		return "" // No segment found or URL ends with '/'
	}
	return url[lastSlashIndex+1:] // Return the substring after the last '/'
}

func ExtractTypeFromURL(url string) string {
	// Find the position of "browse/"
	browseIndex := strings.Index(url, "browse/")
	if browseIndex == -1 {
		return ""
	}
	// Get the substring after "browse/"
	substr := url[browseIndex+len("browse/"):]
	// Find the last "/"
	lastSlashIndex := strings.LastIndex(substr, "/")
	if lastSlashIndex == -1 {
		return substr // Return the whole substring if there's no "/"
	}
	// Return the substring before the last "/"
	return substr[:lastSlashIndex]
}

func RemoveBrandsFromName(brand, name string) string {
	pattern := "(?i)" + regexp.QuoteMeta(brand)
	re := regexp.MustCompile(pattern)

	// 计数器控制替换次数
	count := 0
	return re.ReplaceAllStringFunc(name, func(matched string) string {
		if count < 1 {
			count++
			return ""
		}
		return matched // 后续匹配保留原内容
	})
}

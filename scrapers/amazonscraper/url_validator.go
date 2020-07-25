package amazonscraper

import (
	"regexp"
	"strings"
)

func ExtractManyAmazonProductLinkFromString(s string) []string {
	lines := strings.Split(s, "\n")
	result := []string{}

	for _, line := range lines {
		link, success := ExtractOneAmazonProductLinkFromString(line)
		if !success {
			continue
		}
		result = append(result, link)
	}

	return result
}

func ExtractOneAmazonProductLinkFromString(s string) (link string, found bool) {
	regex, _ := regexp.Compile(`(http[s]?:\/\/)?(www\.)?amazon\..*\/.*(dp|gp)\/\S*`)
	link = regex.FindString(s)
	found = len(link) > 0
	return link, found
}

func IsAmazonProductLink(s string) bool {
	_, result := ExtractOneAmazonProductLinkFromString(s)
	return result
}

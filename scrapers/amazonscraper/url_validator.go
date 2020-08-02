package amazonscraper

import (
	"regexp"
	"strings"
)

func ExtractManyProductLinkFromString(s string) []string {
	lines := strings.Split(s, "\n")
	result := []string{}

	for _, line := range lines {
		link, success := ExtractOneProductLinkFromString(line)
		if !success {
			continue
		}
		result = append(result, link)
	}

	return result
}

func ExtractOneProductLinkFromString(s string) (link string, found bool) {
	regex, _ := regexp.Compile(`(http[s]?:\/\/)?(www\.)?amazon\..*\/.*(dp|gp)\/\S*`)
	link = regex.FindString(s)
	found = len(link) > 0
	return link, found
}

func IsProductLink(s string) bool {
	_, result := ExtractOneProductLinkFromString(s)
	return result
}

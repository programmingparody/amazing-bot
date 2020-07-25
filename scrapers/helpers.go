package scrapers

import (
	"strconv"
	"unicode"
)

//Takes all numbers from a string and returns a slice of floats
func NumbersFromString(input string) (numbers []float64) {
	resultString := ""

	tryToAppendString := func(resultString *string) {
		if len(*resultString) > 0 {
			newNumber, error := strconv.ParseFloat(*resultString, 64)
			if error == nil {
				numbers = append(numbers, newNumber)
			}
			*resultString = ""
		}
	}

	for _, c := range input {
		if unicode.IsDigit(c) || c == '.' || c == ',' {
			if c == ',' {
				continue
			}
			resultString += string(c)
		} else {
			tryToAppendString(&resultString)
		}
	}

	tryToAppendString(&resultString)
	return numbers
}

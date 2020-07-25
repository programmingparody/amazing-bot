package scrapers

import "testing"

func test(t *testing.T, testString string, expectedResult []float64) {
	result := NumbersFromString(testString)

	if len(result) != len(expectedResult) {
		t.Errorf("%v %v", expectedResult, result)
	}

	for index, result := range result {
		expected := expectedResult[index]
		if result != expected {
			t.Errorf("[%v]: %v %v", index, expected, result)
		}
	}
}
func TestNumbersFromString(t *testing.T) {
	test(t, "foaskfoa 12.30 fiasfj 22222 fkla .1", []float64{12.30, 22222, 0.1})
	test(t, "   2.30    \n57 ", []float64{2.30, 57})
	test(t, "1,653 ratings", []float64{1653})
}

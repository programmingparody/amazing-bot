package amazonscraper

import "testing"

func TestIsAmazonProductLink(t *testing.T) {
	testTable := []struct {
		input    string
		expected bool
	}{
		{
			input:    "",
			expected: false,
		},
		{
			input:    "https://www.amazon.com/Acer-R240HY-bidx-23-8-Inch-Widescreen/dp/B0148NNKTC/ref=as_li_ss_tl?ie=UTF8&linkCode=sl1&linkId=9392be61161004ae4cbc7e45d400b758&language=en_US",
			expected: true,
		},
		{
			input:    "https://www.amazon.com/Acer-R240HY-bidx-23-8-Inch-Widescreen/dp/B0148NNKTC/ref=as_li_ss_tl?ie=UTF8&linkCode=sl1",
			expected: true,
		},
		{
			input:    "https://www.amazon.com/gp/product/B07MPCSHQD",
			expected: true,
		},
		{
			input:    "https://www.amazon.ca/HP-13-AQ1001CA-Laptop-i5-1035G1-7YZ81UA/dp/B0899J7B28/ref=br_msw_pdt-4/140-4680546-1561768?_encoding=UTF8&smid=A3DWYIK6Y9EEQB&pf_rd_m=A3DWYIK6Y9EEQB&pf_rd_s=&pf_rd_r=5VTCJ7WH87G2D7XQ1N9V&pf_rd_t=36701&pf_rd_p=7341f510-3e19-4c9e-a178-4ce130685e05&pf_rd_i=desktop",
			expected: true,
		},
		{
			input:    "https://www.amazon.com",
			expected: false,
		},
	}

	for _, test := range testTable {
		result := IsAmazonProductLink(test.input)
		if result != test.expected {
			t.Errorf("Input: %v Expected: %v Result: %v", test.input, test.expected, result)
		}
	}
}

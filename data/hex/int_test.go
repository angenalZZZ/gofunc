package hex

import "testing"

var testIntTable = []struct {
	name string
	i    int64
}{
	{"1 digits", 9},
	{"2 digits", 19},
	{"3 digits", 255},
	{"4 digits", 1234},
	{"5 digits", 65535},
	{"6 digits", 112245},
	{"7 digits", 1008080},
	{"8 digits", 16777215},
	{"9 digits", 133677660},
	{"10 digits", 1537208076},
	{"11 digits", 18639687201},
	{"12 digits", 198876102091},
	{"13 digits", 1938392494373},
	{"14 digits", 10918913829245},
	{"15 digits", 109382742849943},
	{"16 digits", 1102939248284945},
}

func TestFormatInt(t *testing.T) {
	for _, tt := range testIntTable {
		t.Run(tt.name, func(t *testing.T) {
			i, s := tt.i, FormatInt(tt.i)
			if iS, err := ParseInt(s); err != nil {
				t.Error(err)
			} else if iS != i {
				t.Fatal("convert failure")
			} else {
				t.Log(iS, "=>", s)
			}
		})
	}
}

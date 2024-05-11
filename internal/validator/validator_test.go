package validator

import (
	"testing"
)

func TestPermissivePhoneNumberRegex(t *testing.T) {
	var re = PermissivePhoneNumberRX

	// Define test cases
	testCases := []struct {
		name   string
		number string
		valid  bool
	}{
		{"Valid US Number", "123-456-7890", true},
		{"Valid US Number with Parentheses", "(123) 456-7890", true},
		{"Valid with Leading Plus", "+123 456 7890", true},
		{"Invalid - Too Long", "1234-456-7890", false},
		{"Invalid - Letters Included", "123-456-ABCD", false},
		{"Invalid - Special Characters", "(123)-456-7890#", false},
		{"Valid Full International", "+(123) 456-7890", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if re.MatchString(tc.number) != tc.valid {
				t.Errorf("Test %s failed. Expected %t, got %t", tc.name, tc.valid, !tc.valid)
			}
		})
	}
}

func TestE164PhoneNumber(t *testing.T) {
	var re = E164PhoneNumber

	// Define test cases
	testCases := []struct {
		name   string
		number string
		valid  bool
	}{
		{"No plus", "1234567890", false},
		{"Valid Number", "+1234567890", true},
		{"Maximum Digits", "+123456789012345", true},
		{"Invalid - Starts with Zero", "+0234567890", false},
		{"Invalid - Missing Plus", "1234567890", false},
		{"Invalid - Too Long", "+1234567890123456", false},
		{"Invalid - Contains Letters", "+1234abcd567890", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if re.MatchString(tc.number) != tc.valid {
				t.Errorf("Failed %s. Expected %t, got %t", tc.name, tc.valid, !tc.valid)
			}
		})
	}
}

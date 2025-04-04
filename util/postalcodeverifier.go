package util

func VerifyPostalCode(postalCode string) bool {
	if len(postalCode) != 6 {
		return false
	}

	// Check odd positions (1,3,5) are letters
	for i := 0; i < 6; i += 2 {
		if postalCode[i] < 'A' || postalCode[i] > 'Z' {
			return false
		}
	}

	// Check even positions (2,4,6) are numbers
	for i := 1; i < 6; i += 2 {
		if postalCode[i] < '0' || postalCode[i] > '9' {
			return false
		}
	}

	return true
}

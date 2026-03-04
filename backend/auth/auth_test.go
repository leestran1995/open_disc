package auth

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestCheckPasswordStrengthValidPassword(t *testing.T) {
	validPassword := "Pa$$w0rd"
	
	want := CheckPasswordResult{
		HasUppercase:  true,
		HasLowercase:  true,
		HasNumber:     true,
		HasSpecial:    true,
		HasEightChars: true,
	}

	if got := CheckPasswordStrength(validPassword); !cmp.Equal(got, want) {
		t.Errorf("CheckPasswordStrength(%v) = %v, want %v", validPassword, got, want)
	}
	
}
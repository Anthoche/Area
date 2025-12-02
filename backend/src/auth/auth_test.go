package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	const password = "supersecret"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hashed == "" || hashed == password {
		t.Fatalf("hashed password should be non-empty and different from input")
	}
	if err := CheckPassword(hashed, password); err != nil {
		t.Fatalf("CheckPassword should succeed: %v", err)
	}
	if err := CheckPassword(hashed, "wrong"); err == nil {
		t.Fatalf("CheckPassword should fail for wrong password")
	}
}

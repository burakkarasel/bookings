package forms

import (
	"net/url"
	"testing"
)

// TestForm_Valid must return true because we didn't pass any value, and it can't have an error, and it must be equal 0
func TestForm_Valid(t *testing.T) {
	postedData := url.Values{}

	form := New(postedData)

	isValid := form.Valid()

	if !isValid {
		t.Error("got invalid when should have been valid")
	}
}

// TestForm_Required tests our Required func in forms.go
func TestForm_Required(t *testing.T) {
	// created a new url.Values holder
	postedData := url.Values{}
	form := New(postedData)

	// expecting not valid
	form.Required("a", "b", "c")

	if form.Valid() {
		t.Error("form.Valid should have returned false because we have required fields but no value passed into")
	}

	// Added key-value pairs to posted data
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	// Checked with new data
	form = New(postedData)
	form.Required("a", "b", "c")

	// expecting valid
	if !form.Valid() {
		t.Error("form.Valid should have returned true because we passed value into required fields")
	}

}

// TestForm_IsEmail is a test func for IsEmail func in forms.go
func TestForm_IsEmail(t *testing.T) {
	// created a new url.Values holder
	postedData := url.Values{}
	form := New(postedData)

	// expecting error for non-existing key
	form.IsEmail("non_existing_key")

	if form.Valid() {
		t.Error("form shows valid, even if there must be an error because we didnt create a key")
	}

	// expecting error
	postedData = url.Values{}
	postedData.Add("invalid", "asd")
	form = New(postedData)

	form.IsEmail("invalid")

	if form.Valid() {
		t.Error("expected not valid but returned valid")
	}

	// expecting no error
	postedData = url.Values{}
	postedData.Add("valid", "bck@here.com")
	form = New(postedData)

	form.IsEmail("valid")

	if !form.Valid() {
		t.Error("expected valid but returned invalid")
	}
}

// TestForm_Has is a test func for Has func in forms.go
func TestForm_Has(t *testing.T) {
	// created a new url.Values holder
	postedData := url.Values{}
	form := New(postedData)

	// non-existing val
	has := form.Has("whatever")
	if has {
		t.Error("didnt pass any key-value pair, but has returned true")
	}

	postedData = url.Values{}
	postedData.Add("valid", "abc")
	postedData.Add("invalid", "")

	form = New(postedData)

	// checks valid value
	valid := form.Has("valid")

	if !valid {
		t.Error("passed key-value pair but returned false")
	}

	// checks invalid value
	invalid := form.Has("invalid")

	if invalid {
		t.Error("passed empty value expected error but didn't get one")
	}
}

// TestForm_MinLength is a test func for Minlength func in forms.go
func TestForm_MinLength(t *testing.T) {
	// created a new url.Values holder
	postedData := url.Values{}
	form := New(postedData)

	// checks non-existing key
	nonExistingKey := form.MinLength("nonExistingKey", 3)

	if nonExistingKey {
		t.Error("expected false got true for none existing key")
	}

	// to cover Get func in errors
	isErr := form.Errors.Get("nonExistingKey")

	if isErr == "" {
		t.Error("should have an error, but didnt get one")
	}

	postedData = url.Values{}
	postedData.Add("invalid", "a")
	postedData.Add("valid", "abcde")
	postedData.Add("another_valid", "abcdefghijk")

	form = New(postedData)

	// checks invalid value
	invalid := form.MinLength("invalid", 3)

	if invalid {
		t.Error("expected false got true for invalid case")
	}

	// checks valid value
	valid := form.MinLength("valid", 3)

	if !valid {
		t.Error("expected true got false for valid case")
	}

	// to cover Get func in errors.go
	isErr = form.Errors.Get("valid")

	if isErr != "" {
		t.Error("shouldn't got an error, but got one")
	}

	// checks with another length should return true

	anotherValid := form.MinLength("another_valid", 10)

	if !anotherValid {
		t.Error("expected true for len(another)valid == 11 && len(another)valid > 10 but got false")
	}
}

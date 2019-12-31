package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestCreateFirebaseAuthUserClient(t *testing.T) {
	request := struct {
		Email         string
		EmailVerified bool
		PhoneNumber   string
		Password      string
		DisplayName   string
		PhotoURL      string
		Disabled      bool
	}{
		Email:         "tanopwan@gmail.com",
		EmailVerified: true,
		PhoneNumber:   "+66803363515",
		Password:      "123456",
		DisplayName:   "Tanopwan",
		PhotoURL:      "https://lh3.googleusercontent.com/a-/AAuE7mBzDSJN1bPOTnvaF80rODwAjiZ1jpA8UznvelFTAQ=k-s64",
		Disabled:      false,
	}
	jsonStr, err := json.Marshal(request)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	res, err := http.Post("http://localhost:8080/api/auth/firebase-auth/register", "application/jon", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Error(res.Status)
		t.FailNow()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Log(string(body))
}

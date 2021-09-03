package handlers

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"
)

func TestStringSliceSort(t *testing.T) {
	words := []string{"tester", "dev", "tester", "sre", "tester", "sre", "dev", "ac", "ab"}
	sort.Slice(words, func(i, j int) bool {
		// don't use: words[i] == words[j]
		return words[i] > words[j]
	})
	fmt.Println(words)
}

func TestStructToMap(t *testing.T) {
	user := tableUser{
		ID:     "id01",
		Name:   "name-01",
		Role:   "tester",
		Skills: []string{"manual", "auto"},
	}
	fmt.Printf("user: %+v\n", user)

	b, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}

	userMap := make(map[string]interface{})
	if err = json.Unmarshal(b, &userMap); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("user map: %v\n", userMap)
}

func TestUsersGroupByRole(t *testing.T) {
	users, err := buildMockTableUsers()
	if err != nil {
		t.Fatal(err)
	}
	spanUsers, err := addDefaultSpanValues(users)
	if err != nil {
		t.Fatal(err)
	}
	usersSpanByRole(spanUsers)

	for _, user := range spanUsers {
		fmt.Printf("%+v\n", user)
	}
}

package ddb

import (
	"reflect"
	"slices"
	"testing"
)

func TestFilterOwnedBookIDsKeepsOwnedBooksOnly(t *testing.T) {
	resp := AvailableUserContent{
		Status: "success",
		Licenses: []LicenseBlock{
			{
				EntityTypeID: EntityTypeIDBooks,
				Entities: []LicenseEntity{
					{ID: 1, IsOwned: true},   // PHB
					{ID: 2, IsOwned: true},   // MM
					{ID: 5, IsOwned: false},  // XGE: not owned
					{ID: 11, IsOwned: true},  // TCE
				},
			},
			{
				EntityTypeID: "953599357", // dice sets — must be ignored
				Entities: []LicenseEntity{
					{ID: 999, IsOwned: true},
				},
			},
		},
	}

	got := filterOwnedBookIDs(resp)
	slices.Sort(got)
	want := []int{1, 2, 11}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterOwnedBookIDs = %v, want %v", got, want)
	}
}

func TestFilterOwnedBookIDsEmptyWhenNothingOwned(t *testing.T) {
	resp := AvailableUserContent{
		Status: "success",
		Licenses: []LicenseBlock{
			{
				EntityTypeID: EntityTypeIDBooks,
				Entities: []LicenseEntity{
					{ID: 1, IsOwned: false},
				},
			},
		},
	}
	if got := filterOwnedBookIDs(resp); len(got) != 0 {
		t.Fatalf("filterOwnedBookIDs = %v, want empty", got)
	}
}

func TestBadSourceIDsContainsAdventureMuncherList(t *testing.T) {
	// Sanity check: the IDs we ported from ddb-adventure-muncher must remain
	// in the BadSourceIDs map. If a future commit drops one, this test fails
	// loudly so the change is intentional.
	want := []int{31, 53, 42, 29, 4, 26, 30}
	for _, id := range want {
		if _, ok := BadSourceIDs[id]; !ok {
			t.Errorf("BadSourceIDs missing id %d", id)
		}
	}
}

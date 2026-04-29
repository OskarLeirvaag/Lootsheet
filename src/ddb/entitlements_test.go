package ddb

import (
	"encoding/json"
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
				EntityTypeID: 953599357, // dice sets — must be ignored
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

// TestAvailableUserContentDecodesRealResponseShape pins the actual JSON shape
// observed from /mobile/api/v6/available-user-content (verified via
// tmp/check_ddb.go on 2026-04-29): Licenses are nested under `data` and
// EntityTypeID is a JSON number, not a string. Regression guard.
func TestAvailableUserContentDecodesRealResponseShape(t *testing.T) {
	body := []byte(`{
		"status": "success",
		"data": {
			"Licenses": [
				{
					"EntityTypeID": 496802664,
					"Entities": [
						{"id": 1, "isOwned": true},
						{"id": 2, "isOwned": true},
						{"id": 5, "isOwned": false}
					]
				},
				{
					"EntityTypeID": 701257905,
					"Entities": [{"id": 99, "isOwned": true}]
				}
			]
		}
	}`)

	var resp AvailableUserContent
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	books := resp.Books()
	if len(books) != 2 {
		t.Fatalf("Books() = %d blocks, want 2", len(books))
	}
	if books[0].EntityTypeID != EntityTypeIDBooks {
		t.Errorf("first block EntityTypeID = %d, want %d", books[0].EntityTypeID, EntityTypeIDBooks)
	}

	owned := filterOwnedBookIDs(resp)
	slices.Sort(owned)
	want := []int{1, 2}
	if !reflect.DeepEqual(owned, want) {
		t.Fatalf("filterOwnedBookIDs = %v, want %v", owned, want)
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

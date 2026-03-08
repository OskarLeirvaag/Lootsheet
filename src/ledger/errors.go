package ledger

import (
	"errors"
	"fmt"
)

// ErrImmutableEntry is returned when an operation attempts to modify or delete
// a journal entry that has been posted or reversed. Corrections must use
// reversal, adjustment, or reclassification.
var ErrImmutableEntry = fmt.Errorf("posted or reversed journal entries are immutable; use reversal or adjustment to correct")

// ErrEntryNotReversible is returned when a reversal is attempted on an entry
// that is not in 'posted' status (e.g., it is a draft or already reversed).
var ErrEntryNotReversible = fmt.Errorf("only posted journal entries can be reversed")

// ErrAccountHasPostings is returned when an attempt is made to delete an
// account that has journal line postings referencing it.
var ErrAccountHasPostings = errors.New("account has journal postings and cannot be deleted")

// ErrAccountNotFound is returned when an account lookup by code finds no match.
var ErrAccountNotFound = errors.New("account not found")

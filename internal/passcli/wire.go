package passcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// Upstream spells the same concept three different ways depending on where it
// appears, because each Rust type derives Serialize differently. Keeping three
// constants rather than one prevents "helpfully" normalising them.
// See docs/schema.md §3.1.
const (
	// itemTypeSSHKey is the JSON value of item_type: snake_case, from a
	// #[serde(rename_all = "snake_case")] enum.
	itemTypeSSHKey = "ssh_key"

	// filterTypeSSHKey is the --filter-type flag value: kebab-case, parsed by
	// a hand-written FromStr.
	filterTypeSSHKey = "ssh-key"

	// stateActive and stateTrashed are JSON values of state: PascalCase, from
	// a plain derive with no rename attribute.
	stateActive  = "Active"
	stateTrashed = "Trashed"
)

// civilTimeLayout matches jiff::civil::DateTime, which serialises with no zone
// and no offset. time.RFC3339 does not parse it. See docs/schema.md §3.2.
const civilTimeLayout = "2006-01-02T15:04:05"

// civilTime decodes upstream's zoneless timestamps. The value is wall-clock
// only and carries no offset, so it is never converted to an absolute instant.
type civilTime struct {
	time.Time
}

// UnmarshalJSON accepts the civil layout and tolerates null or an empty string,
// both of which yield the zero time rather than an error: a missing timestamp
// should not drop an otherwise valid item from the list.
func (c *civilTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("timestamp is not a string: %w", err)
	}
	if raw == "" {
		return nil
	}

	parsed, err := time.Parse(civilTimeLayout, raw)
	if err != nil {
		return fmt.Errorf("parse timestamp %q as %q: %w", raw, civilTimeLayout, err)
	}

	c.Time = parsed

	return nil
}

// vaultListResponse is the envelope of `vault list --output json`.
type vaultListResponse struct {
	Vaults []vaultEntry `json:"vaults"`
}

type vaultEntry struct {
	Name    string `json:"name"`
	VaultID string `json:"vault_id"`
	ShareID string `json:"share_id"`
}

// itemListResponse is the envelope of `item list --output json`. Upstream
// always wraps the array; it is never a bare list. An "internal" build adds a
// sibling "folders" key, which decodes to nothing here by design.
type itemListResponse struct {
	Items []itemSummary `json:"items"`
}

// itemSummary mirrors the redacted summary form of `item list`. protui never
// passes --show-secrets, so no field here can carry key material.
type itemSummary struct {
	ID         string    `json:"id"`
	ShareID    string    `json:"share_id"`
	VaultID    string    `json:"vault_id"`
	State      string    `json:"state"`
	Flags      []string  `json:"flags"`
	CreateTime civilTime `json:"create_time"`
	ModifyTime civilTime `json:"modify_time"`
	FolderID   *string   `json:"folder_id"`
	Title      string    `json:"title"`
	ItemType   string    `json:"item_type"`
}

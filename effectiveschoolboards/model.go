package effectiveschoolboards

import (
	"cmp"
	"slices"
	"strings"
)

// Model is the model for the effective school boards survey questions.
type Model struct {
	Items           []Item
	CategoryTierMap map[string]map[int]int
}

// Item is a single item in the model.
type Item struct {
	ID               string         // This is the item ID.
	Text             string         // This is the text of the item.
	Tier             int            // This is the tier of the item.
	CategoryScoreMap map[string]int // This maps a category to its score for the item.
	RequiredItemIDs  []string       // This is the list of item IDs that must be present in order for this item to be answered.  Each of these will have a "+" or "-" prefix to indicate if it must be answered affirmatively or negatively.
}

// MeetsRequirements returns true if the list of IDs given is sufficient to meet the item's requirements.
func (i Item) MeetsRequirements(matchedIDs []string) bool {
	for _, requiredItemID := range i.RequiredItemIDs {
		if !slices.Contains(matchedIDs, requiredItemID) {
			return false
		}
	}
	return true
}

// AllItems returns a list of items that can be answered.
//
// If provided, the list of matched IDs are assumed to have been answered affirmatively or negatively, based on the "+" or "-" prefix.
func (m *Model) AllItems(matchedIDs []string) []Item {
	nextItems := []Item{}
	for _, item := range m.Items {
		if !item.MeetsRequirements(matchedIDs) {
			continue
		}
		nextItems = append(nextItems, item)
	}

	slices.SortFunc(nextItems, func(a, b Item) int {
		diff := cmp.Compare(a.Tier, b.Tier)
		if diff != 0 {
			return diff
		}
		return strings.Compare(a.Text, b.Text)
	})

	return nextItems
}

// NextItems returns a list of items that can be answered next.
//
// If provided, the list of matched IDs are assumed to have been answered affirmatively or negatively, based on the "+" or "-" prefix.
func (m *Model) NextItems(matchedIDs []string) []Item {
	items := m.AllItems(matchedIDs)

	nextItems := []Item{}
	for _, item := range items {
		if slices.Contains(matchedIDs, "+"+item.ID) || slices.Contains(matchedIDs, "-"+item.ID) {
			continue
		}

		nextItems = append(nextItems, item)
	}

	return nextItems
}

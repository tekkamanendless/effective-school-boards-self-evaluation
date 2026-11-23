package effectiveschoolboards

import (
	"cmp"
	"slices"
	"strings"
)

// Model is the model for the effective school boards survey questions.
type Model struct {
	Items                  []Item
	Categories             []Category
	CategoryTierToIndexMap map[string]map[int]int
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
func (i Item) MeetsRequirements(answers []string) bool {
	for _, requiredItemID := range i.RequiredItemIDs {
		if !slices.Contains(answers, requiredItemID) {
			return false
		}
	}
	return true
}

type Category struct {
	Name  string         // This is the name of the category.
	Tiers []CategoryTier // This is the list of tiers for the category.
}

type CategoryTier struct {
	Tier            int      // This is the tier of the category (the number of points that it is worth).
	RequiredItemIDs []string // This is the list of item IDs that must be present in order for this item to be answered.  Each of these will have a "+" or "-" prefix to indicate if it must be answered affirmatively or negatively.
}

// MeetsRequirements returns true if the list of IDs given is sufficient to meet the category tier's requirements.
func (i CategoryTier) MeetsRequirements(answers []string) bool {
	for _, requiredItemID := range i.RequiredItemIDs {
		if !slices.Contains(answers, requiredItemID) {
			return false
		}
	}
	return true
}

// AllItems returns a list of items that can be answered.
//
// If provided, the list of matched IDs are assumed to have been answered affirmatively or negatively, based on the "+" or "-" prefix.
func (m *Model) AllItems(answers []string) []Item {
	nextItems := []Item{}
	for _, item := range m.Items {
		if !item.MeetsRequirements(answers) {
			continue
		}
		nextItems = append(nextItems, item)
	}

	slices.SortFunc(nextItems, func(a, b Item) int {
		if a.Tier == 0 && b.Tier != 0 {
			return -1
		}
		if a.Tier != 0 && b.Tier == 0 {
			return 1
		}

		if !(a.Tier == 0 && b.Tier == 0) {
			var aPrimaryCategory string
			var bPrimaryCategory string
			for _, category := range m.Categories {
				if _, ok := a.CategoryScoreMap[category.Name]; ok {
					aPrimaryCategory = category.Name
				}
				if _, ok := b.CategoryScoreMap[category.Name]; ok {
					bPrimaryCategory = category.Name
				}
			}

			if aPrimaryCategory != bPrimaryCategory {
				return strings.Compare(aPrimaryCategory, bPrimaryCategory)
			}
		}

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
func (m *Model) NextItems(answers []string) []Item {
	items := m.AllItems(answers)

	nextItems := []Item{}
	for _, item := range items {
		if slices.Contains(answers, "+"+item.ID) || slices.Contains(answers, "-"+item.ID) {
			continue
		}

		nextItems = append(nextItems, item)
	}

	return nextItems
}

type Score struct {
	Total          int            // This is the total score.
	CategoryScores map[string]int // This maps a category to its score.
}

// CalculateScore calculates the score based on the given answers.
func (m *Model) CalculateScore(answers []string) Score {
	score := Score{
		CategoryScores: map[string]int{},
	}

	for _, category := range m.Categories {
		for _, categoryTier := range category.Tiers {
			if categoryTier.MeetsRequirements(answers) {
				score.CategoryScores[category.Name] = categoryTier.Tier
			}
		}
		score.Total += score.CategoryScores[category.Name]
	}

	return score
}

package effectiveschoolboards

import (
	"cmp"
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"
)

func LoadModel() (*Model, error) {
	model := &Model{
		Items:                  []Item{},
		CategoryTierToIndexMap: map[string]map[int]int{},
	}

	file, err := FS.Open("model.csv")
	if err != nil {
		return nil, fmt.Errorf("could not open model: %w", err)
	}

	csvReader := csv.NewReader(file)
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read model: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("model is empty")
	}

	header := rows[0]
	rows = rows[1:]

	headerNameToIndexMap := map[string]int{}
	for index, name := range header {
		name = strings.ToLower(strings.TrimSpace(name))
		headerNameToIndexMap[name] = index
	}

	categoryNames := []string{
		"Focus Mindset",
		"Clarify Priorities 1",
		"Clarify Priorities 2",
		"Monitor Progress",
		"Align Resources",
		"Communicate Results",
	}

	for _, row := range rows {
		item := Item{
			ID:               row[headerNameToIndexMap["id"]],
			RequiredItemIDs:  []string{},
			Text:             row[headerNameToIndexMap["item"]],
			CategoryScoreMap: map[string]int{},
		}

		{
			value := row[headerNameToIndexMap["tier"]]
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("could not convert %s to int: %w", value, err)
			}
			item.Tier = intValue
		}

		for _, categoryName := range categoryNames {
			value := row[headerNameToIndexMap[strings.ToLower(strings.TrimSpace(categoryName))]]
			if value != "" {
				intValue, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("could not convert %s to int: %w", value, err)
				}
				item.CategoryScoreMap[categoryName] = intValue
			}
		}

		model.Items = append(model.Items, item)
	}

	for _, categoryName := range categoryNames {
		model.CategoryTierToIndexMap[categoryName] = map[int]int{}

		categoryScoreMap := map[int]bool{}
		for _, item := range model.Items {
			if score, ok := item.CategoryScoreMap[categoryName]; ok {
				categoryScoreMap[score] = true
			}
		}
		categoryScores := slices.Collect(maps.Keys(categoryScoreMap))
		slices.Sort(categoryScores)

		for tierIndex, score := range categoryScores {
			model.CategoryTierToIndexMap[categoryName][score] = tierIndex
		}
	}

	findItemsInCategoryBelowTier := func(categoryName string, tier int) []Item {
		items := []Item{}
		for _, item := range model.Items {
			if _, ok := item.CategoryScoreMap[categoryName]; !ok {
				continue
			}
			if item.CategoryScoreMap[categoryName] < tier {
				items = append(items, item)
			}
		}
		return items
	}

	for itemIndex, item := range model.Items {
		var requiredItemIDs []string
		for categoryName, score := range item.CategoryScoreMap {
			requiredItems := findItemsInCategoryBelowTier(categoryName, score)
			for _, requiredItem := range requiredItems {
				if requiredItem.CategoryScoreMap[categoryName] == 0 {
					requiredItemIDs = append(requiredItemIDs, "-"+requiredItem.ID)
				} else {
					requiredItemIDs = append(requiredItemIDs, "+"+requiredItem.ID)
				}
			}
		}
		slices.Sort(requiredItemIDs)
		requiredItemIDs = slices.Compact(requiredItemIDs)
		model.Items[itemIndex].RequiredItemIDs = requiredItemIDs
	}

	for categoryName, scoreToTierMap := range model.CategoryTierToIndexMap {
		category := Category{
			Name:  categoryName,
			Tiers: []CategoryTier{},
		}

		for score := range scoreToTierMap {
			var requiredItemIDs []string
			for _, item := range model.Items {
				if _, ok := item.CategoryScoreMap[categoryName]; !ok {
					continue
				}
				if item.CategoryScoreMap[categoryName] <= score {
					if item.CategoryScoreMap[categoryName] == 0 {
						requiredItemIDs = append(requiredItemIDs, "-"+item.ID)
					} else {
						requiredItemIDs = append(requiredItemIDs, "+"+item.ID)
					}
				}
			}
			slices.Sort(requiredItemIDs)
			requiredItemIDs = slices.Compact(requiredItemIDs)

			categoryTier := CategoryTier{
				Tier:            score,
				RequiredItemIDs: requiredItemIDs,
			}
			category.Tiers = append(category.Tiers, categoryTier)
		}
		slices.SortFunc(category.Tiers, func(a, b CategoryTier) int {
			return cmp.Compare(a.Tier, b.Tier)
		})

		model.Categories = append(model.Categories, category)
	}
	slices.SortFunc(model.Categories, func(a, b Category) int {
		return cmp.Compare(a.Name, b.Name)
	})

	slog.InfoContext(context.TODO(), "Model loaded.", "items", len(model.Items))
	for _, item := range model.Items {
		slog.InfoContext(context.TODO(), "Item.", "id", item.ID, "text", item.Text, "requiredItemIDs", item.RequiredItemIDs)
	}
	for _, category := range model.Categories {
		slog.InfoContext(context.TODO(), "Category.", "name", category.Name)
		for _, categoryTier := range category.Tiers {
			slog.InfoContext(context.TODO(), "CategoryTier.", "tier", categoryTier.Tier, "requiredItemIDs", categoryTier.RequiredItemIDs)
		}
	}

	return model, nil
}

package effectiveschoolboards

import (
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
		Items:           []Item{},
		CategoryTierMap: map[string]map[int]int{},
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

	categories := []string{
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

		for _, category := range categories {
			value := row[headerNameToIndexMap[strings.ToLower(strings.TrimSpace(category))]]
			if value != "" {
				intValue, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("could not convert %s to int: %w", value, err)
				}
				item.CategoryScoreMap[category] = intValue
			}
		}

		model.Items = append(model.Items, item)
	}

	for _, category := range categories {
		model.CategoryTierMap[category] = map[int]int{}

		categoryScoreMap := map[int]bool{}
		for _, item := range model.Items {
			if score, ok := item.CategoryScoreMap[category]; ok {
				categoryScoreMap[score] = true
			}
		}
		categoryScores := slices.Collect(maps.Keys(categoryScoreMap))
		slices.Sort(categoryScores)

		for tierIndex, score := range categoryScores {
			model.CategoryTierMap[category][score] = tierIndex
		}
	}

	findItemsInCategoryBelowTier := func(category string, tier int) []Item {
		items := []Item{}
		for _, item := range model.Items {
			if _, ok := item.CategoryScoreMap[category]; !ok {
				continue
			}
			if item.CategoryScoreMap[category] < tier {
				items = append(items, item)
			}
		}
		return items
	}

	for itemIndex, item := range model.Items {
		var requiredItemIDs []string
		for category, score := range item.CategoryScoreMap {
			requiredItems := findItemsInCategoryBelowTier(category, score)
			for _, requiredItem := range requiredItems {
				if requiredItem.CategoryScoreMap[category] == 0 {
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

	slog.InfoContext(context.TODO(), "Model loaded.", "items", len(model.Items))
	for _, item := range model.Items {
		slog.InfoContext(context.TODO(), "Item.", "id", item.ID, "text", item.Text, "requiredItemIDs", item.RequiredItemIDs)
	}

	return model, nil
}

package types

import "Hyper/models"

type ContentTagItem = models.ContentTagItem

type ContentTagBindingRequest struct {
	TagIDs []int64 `json:"tag_ids"`
}

func BuildContentTagItems(categories []models.AdminCategory) []ContentTagItem {
	items := make([]ContentTagItem, 0, len(categories))
	for _, category := range categories {
		items = append(items, ContentTagItem{ID: category.ID, Name: category.Name, Image: category.Image, Value: category.Value, Sort: category.Sort})
	}
	return items
}

func ContentTagIDs(categories []models.AdminCategory) []int64 {
	ids := make([]int64, 0, len(categories))
	for _, category := range categories {
		ids = append(ids, category.ID)
	}
	return ids
}

func ContentTagNames(categories []models.AdminCategory) []string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		names = append(names, category.Name)
	}
	return names
}

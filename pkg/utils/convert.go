package utils

import (
	"Hyper/models"
	"Hyper/types"
	"encoding/json"
)

func ConvertNoteModelToDTO(note *models.Note) (*types.Notes, error) {
	if note == nil {
		return nil, nil
	}

	dto := &types.Notes{
		ID:          int64(note.ID),
		UserID:      int64(note.UserID),
		Title:       note.Title,
		Content:     note.Content,
		Type:        note.Type,
		Status:      note.Status,
		VisibleConf: note.VisibleConf,
		CreatedAt:   note.CreatedAt,
		UpdatedAt:   note.UpdatedAt,
	}

	// topic_ids
	if err := unmarshalJSON(note.TopicIDs, &dto.TopicIDs); err != nil {
		dto.TopicIDs = make([]int64, 0)
	}

	// location
	if err := unmarshalJSON(note.Location, &dto.Location); err != nil {
		dto.Location = types.Location{}
	}

	// media_data
	if err := unmarshalJSON(note.MediaData, &dto.MediaData); err != nil {
		dto.MediaData = make([]types.NoteMedia, 0)
	}

	return dto, nil
}

func unmarshalJSON[T any](src string, dst *T) error {
	if src == "" {
		return nil
	}
	return json.Unmarshal([]byte(src), dst)
}

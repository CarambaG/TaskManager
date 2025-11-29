package tests

import (
	"TaskManager/internal/models"
	"testing"
)

func TestTaskValidation(t *testing.T) {
	validDueDate := "2025-12-15"

	tests := []struct {
		name    string
		task    models.Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: models.Task{
				ID:          "00000000-0000-0000-0000-000000000000",
				UserID:      "00000000-0000-0000-0000-000000000000",
				Title:       "Test Task",
				Description: "Test Description",
				Priority:    "high",
				DueDate:     &validDueDate,
			},
			wantErr: false,
		},
		{
			name: "invalid uuid in ID",
			task: models.Task{
				ID:          "00000000-0000-0000-0000-00000000000Z",
				UserID:      "00000000-0000-0000-0000-000000000000",
				Title:       "Test Title",
				Description: "Test Description",
				Priority:    "high",
				DueDate:     nil,
			},
			wantErr: true,
		},
		{
			name: "invalid UserID",
			task: models.Task{
				ID:          "00000000-0000-0000-0000-000000000000",
				UserID:      "00000000-0000-0000-0000-00000000000Z",
				Title:       "Test Title",
				Description: "Test Description",
				Priority:    "high",
				DueDate:     nil,
			},
			wantErr: true,
		},
		{
			name: "empty title",
			task: models.Task{
				ID:          "00000000-0000-0000-0000-000000000000",
				UserID:      "00000000-0000-0000-0000-000000000000",
				Title:       "",
				Description: "Test Description",
				Priority:    "high",
				DueDate:     nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

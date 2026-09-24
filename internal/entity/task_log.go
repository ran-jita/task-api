package entity

import "time"

type TaskLog struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	ChangedBy   string    `json:"changed_by"`
	OldAssignee *string   `json:"old_assignee,omitempty"`
	NewAssignee string    `json:"new_assignee"`
	CreatedAt   time.Time `json:"created_at"`
}

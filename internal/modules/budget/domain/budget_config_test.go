package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewBudgetConfig_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		projectID  string
		mode       BudgetMode
		amountMln  float64
	}{
		{
			name:       "limited mode with positive amount",
			projectID:  "proj-1",
			mode:       BudgetModeLimited,
			amountMln:  5000.0,
		},
		{
			name:       "unlimited mode with zero amount",
			projectID:  "proj-2",
			mode:       BudgetModeUnlimited,
			amountMln:  0,
		},
		{
			name:       "unlimited mode ignores amount",
			projectID:  "proj-3",
			mode:       BudgetModeUnlimited,
			amountMln:  1000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBudgetConfig(tt.projectID, tt.mode, tt.amountMln)
			if err != nil {
				t.Fatalf("NewBudgetConfig() unexpected error: %v", err)
			}

			if config.ProjectID() != tt.projectID {
				t.Errorf("ProjectID() = %s, want %s", config.ProjectID(), tt.projectID)
			}
			if config.BudgetMode() != tt.mode {
				t.Errorf("BudgetMode() = %s, want %s", config.BudgetMode(), tt.mode)
			}
			if tt.mode == BudgetModeLimited && config.BudgetAmountMln() != tt.amountMln {
				t.Errorf("BudgetAmountMln() = %f, want %f", config.BudgetAmountMln(), tt.amountMln)
			}
			if config.CreatedAt().IsZero() {
				t.Error("CreatedAt() should not be zero")
			}
			if config.UpdatedAt().IsZero() {
				t.Error("UpdatedAt() should not be zero")
			}
		})
	}
}

func TestNewBudgetConfig_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		projectID string
		mode      BudgetMode
		amountMln float64
		wantErr   error
	}{
		{
			name:      "empty project ID",
			projectID: "",
			mode:      BudgetModeLimited,
			amountMln: 5000,
			wantErr:   ErrProjectIDRequired,
		},
		{
			name:      "invalid mode",
			projectID: "proj-1",
			mode:      BudgetMode("invalid"),
			amountMln: 5000,
			wantErr:   ErrInvalidBudgetMode,
		},
		{
			name:      "limited mode with zero amount",
			projectID: "proj-1",
			mode:      BudgetModeLimited,
			amountMln: 0,
			wantErr:   ErrInvalidBudgetAmount,
		},
		{
			name:      "limited mode with negative amount",
			projectID: "proj-1",
			mode:      BudgetModeLimited,
			amountMln: -100,
			wantErr:   ErrInvalidBudgetAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBudgetConfig(tt.projectID, tt.mode, tt.amountMln)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewBudgetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBudgetConfig_SetBudget(t *testing.T) {
	t.Parallel()

	config, err := NewBudgetConfig("proj-1", BudgetModeLimited, 5000)
	if err != nil {
		t.Fatalf("NewBudgetConfig() unexpected error: %v", err)
	}

	// Меняем на unlimited
	if err := config.SetBudget(BudgetModeUnlimited, 0); err != nil {
		t.Fatalf("SetBudget() unexpected error: %v", err)
	}
	if config.BudgetMode() != BudgetModeUnlimited {
		t.Errorf("BudgetMode() = %s, want unlimited", config.BudgetMode())
	}

	// Меняем обратно на limited с новым лимитом
	if err := config.SetBudget(BudgetModeLimited, 8000); err != nil {
		t.Fatalf("SetBudget() unexpected error: %v", err)
	}
	if config.BudgetMode() != BudgetModeLimited {
		t.Errorf("BudgetMode() = %s, want limited", config.BudgetMode())
	}
	if config.BudgetAmountMln() != 8000 {
		t.Errorf("BudgetAmountMln() = %f, want 8000", config.BudgetAmountMln())
	}

	// Ошибка: invalid mode
	if err := config.SetBudget("invalid", 0); !errors.Is(err, ErrInvalidBudgetMode) {
		t.Errorf("SetBudget() error = %v, want ErrInvalidBudgetMode", err)
	}

	// Ошибка: limited с нулевой суммой
	if err := config.SetBudget(BudgetModeLimited, 0); !errors.Is(err, ErrInvalidBudgetAmount) {
		t.Errorf("SetBudget() error = %v, want ErrInvalidBudgetAmount", err)
	}
}

func TestBudgetConfig_RemainingBudget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mode       BudgetMode
		amountMln  float64
		costMln    float64
		want       float64
		wantErr    error
	}{
		{
			name:      "limited with room",
			mode:      BudgetModeLimited,
			amountMln: 5000,
			costMln:   3000,
			want:      2000,
			wantErr:   nil,
		},
		{
			name:      "limited exactly at limit",
			mode:      BudgetModeLimited,
			amountMln: 5000,
			costMln:   5000,
			want:      0,
			wantErr:   nil,
		},
		{
			name:      "limited over budget",
			mode:      BudgetModeLimited,
			amountMln: 5000,
			costMln:   7000,
			want:      0,
			wantErr:   nil,
		},
		{
			name:      "unlimited mode",
			mode:      BudgetModeUnlimited,
			amountMln: 0,
			costMln:   5000,
			want:      0,
			wantErr:   ErrBudgetNotConfigured,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewBudgetConfig("proj-1", tt.mode, tt.amountMln)
			if err != nil {
				t.Fatalf("NewBudgetConfig() unexpected error: %v", err)
			}

			got, err := config.RemainingBudget(tt.costMln)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("RemainingBudget() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("RemainingBudget() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestRestoreBudgetConfig(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	config := RestoreBudgetConfig("proj-1", BudgetModeLimited, 5000, now, now)

	if config.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %s, want proj-1", config.ProjectID())
	}
	if config.BudgetMode() != BudgetModeLimited {
		t.Errorf("BudgetMode() = %s, want limited", config.BudgetMode())
	}
	if config.CreatedAt().IsZero() {
		t.Error("CreatedAt() should not be zero")
	}
}

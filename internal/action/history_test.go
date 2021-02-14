package action

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/service"
)

func TestHistoryActionParams_ValidateAndFill(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name           string
		args           args
		expectedParams *HistoryActionParams
		wantErr        bool
	}{
		{
			name:           "no args",
			args:           args{[]string{}},
			expectedParams: &HistoryActionParams{limit: 10},
			wantErr:        false,
		},
		{
			name:           "handle all limit",
			args:           args{args: []string{"all"}},
			expectedParams: &HistoryActionParams{limit: 0},
			wantErr:        false,
		},
		{
			name:           "handle some limit",
			args:           args{args: []string{"3"}},
			expectedParams: &HistoryActionParams{limit: 3},
			wantErr:        false,
		},
		{
			name:           "non-numeric limit",
			args:           args{args: []string{"kek"}},
			expectedParams: &HistoryActionParams{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &HistoryActionParams{}
			err := p.ValidateAndFill(tt.args.args)
			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.expectedParams, p)
		})
	}
}

func TestHistoryAction_Run(t *testing.T) {
	type fields struct {
		svc *service.MigrationService
	}
	type args struct {
		params interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "invalid action params type passed",
			fields:  fields{},
			args:    args{params: struct{}{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &DownAction{
				svc: tt.fields.svc,
			}
			if err := a.Run(tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

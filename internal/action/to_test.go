package action

import (
	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/service"
	"testing"
)

func TestToActionParams_ValidateAndFill(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name           string
		args           args
		expectedParams *ToActionParams
		wantErr        bool
	}{
		{
			name: "not enough args err",
			args: args{
				args: []string{},
			},
			expectedParams: &ToActionParams{},
			wantErr:        true,
		},
		{
			name: "invalid version pattern",
			args: args{
				args: []string{"m200101_000000_+`"},
			},
			expectedParams: &ToActionParams{},
			wantErr:        true,
		},
		{
			name: "success",
			args: args{
				args: []string{"m200101_000000_test"},
			},
			expectedParams: &ToActionParams{
				version: "m200101_000000_test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ToActionParams{}
			err := p.ValidateAndFill(tt.args.args)
			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.expectedParams, p)
		})
	}
}

func TestToAction_Run(t *testing.T) {
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
			a := &ToAction{
				svc: tt.fields.svc,
			}
			if err := a.Run(tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

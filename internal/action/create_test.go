package action

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/migration"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
)

func TestCreateActionParams_ValidateAndFill(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name           string
		expectedParams *CreateActionParams
		args           args
		wantErr        error
	}{
		{
			name:           "not enough args",
			expectedParams: &CreateActionParams{},
			args:           args{args: []string{}},
			wantErr:        errorsInternal.ErrNotEnoughArgs,
		},
		{
			name:           "empty name",
			expectedParams: &CreateActionParams{},
			args:           args{args: []string{""}},
			wantErr:        ErrEmptyName,
		},
		{
			name:           "invalid name pattern (common check)",
			expectedParams: &CreateActionParams{},
			args:           args{args: []string{"create+test_table"}},
			wantErr:        ErrInvalidName,
		},
		{
			name: "default filetype assign",
			expectedParams: &CreateActionParams{
				name:  "create_some_table",
				mType: migration.TypeGo,
			},
			args:    args{args: []string{"create_some_table"}},
			wantErr: nil,
		},
		{
			name:           "unknown migration type",
			expectedParams: &CreateActionParams{},
			args:           args{args: []string{"create_some_table", "kek"}},
			wantErr:        ErrUnknownMigrationType,
		},
		{
			name: "success validate .go",
			expectedParams: &CreateActionParams{
				name:  "create_some_table",
				mType: migration.TypeGo,
			},
			args:    args{args: []string{"create_some_table", "go"}},
			wantErr: nil,
		},
		{
			name: "success validate .sql",
			expectedParams: &CreateActionParams{
				name:  "create_some_table",
				mType: migration.TypeSQL,
			},
			args:    args{args: []string{"create_some_table", "sql"}},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &CreateActionParams{}
			if err := p.ValidateAndFill(tt.args.args); err != nil && err != tt.wantErr {
				require.Equal(t, tt.wantErr, err)
			}

			require.Equal(t, tt.expectedParams, p)
		})
	}
}

func TestCreateAction_Run(t *testing.T) {
	type fields struct {
		migrationsPath string
	}
	type args struct {
		params interface{}
	}
	// todo: check for all errs
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name:    "invalid action params type passed",
			fields:  fields{},
			args:    args{params: struct{}{}},
			wantErr: errorsInternal.ErrInvalidActionParamsType,
		},
		{
			name:   "success case",
			fields: fields{migrationsPath: dir},
			args: args{params: &CreateActionParams{
				name:  "some_name",
				mType: migration.TypeGo,
			}},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &CreateAction{
				migrationsPath: tt.fields.migrationsPath,
			}
			if err := a.Run(tt.args.params); err != nil && err != tt.wantErr {
				require.Error(t, tt.wantErr, err)
			}
		})
	}
}

package oapi_builder

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var logger = slog.Default()

func TestNew(t *testing.T) {
	type args struct {
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want *OapiBuilder
	}{
		{
			name: "test",
			args: args{
				logger: logger,
			},
			want: &OapiBuilder{
				logger: logger,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.logger)

			require.Equal(t, tt.want, got)

		})
	}
}

type mainInfoSchemer struct {
	scheme string
}

func (m *mainInfoSchemer) Scheme() string {
	return m.scheme
}

func TestOapiBuilder_AddMainInfo(t *testing.T) {
	mainInfoScheme := `
	openapi: 3.0.0
	info:
		title: Test
		version: 1.0.0
	`

	yamlBuilder := strings.Builder{}
	yamlBuilder.WriteString(mainInfoScheme)

	type fields struct {
		oapiDoc     string
		yamlBuilder strings.Builder
		logger      *slog.Logger
	}
	type args struct {
		schemer Schemer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *OapiBuilder
	}{
		{
			name: "test",
			fields: fields{
				logger: logger,
			},
			args: args{
				schemer: &mainInfoSchemer{scheme: mainInfoScheme},
			},
			want: &OapiBuilder{
				logger:      logger,
				yamlBuilder: yamlBuilder,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := &OapiBuilder{
				oapiDoc:     tt.fields.oapiDoc,
				yamlBuilder: tt.fields.yamlBuilder,
				logger:      tt.fields.logger,
			}

			got := ob.AddMainInfo(tt.args.schemer)

			s := strings.NewReplacer("\n", "", "\t", "", " ", "")
			require.Contains(t, s.Replace(tt.want.String()), s.Replace(got.String()))
		})
	}
}

package oapi_builder

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var pathsGetByIdScheme = `
/admin/user/{id}:
  get:
    tags:
      - admin
      - user
    summary: Get user details
    operationId: getUserById
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
          format: uuid
    responses:
      200:
        description: Success
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
`

var pathsDeleteByIdScheme = `
/admin/user/{id}:
  delete:
    tags:
      - admin
      - user
    summary: Delete user
    operationId: deleteUser
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
          format: uuid
    responses:
      200:
        description: Success
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Status'
`

var pathsGetListScheme = `
/admin/user:
  get:
    tags:
      - admin
      - user
    summary: Get list of users
    operationId: getUsers
    responses:
      200:
        description: Success
        content:
          application/json:
            schema:
              type: array
              items:
                $ref: '#/components/schemas/User'
`

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

func TestOapiBuilder_AddPaths(t *testing.T) {
	mainInfoScheme := `
openapi: 3.0.0
info:
	title: Test
	version: 1.0.0
`

	type args struct {
		schemers []Schemer
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",

			args: args{
				schemers: []Schemer{
					&mainInfoSchemer{scheme: pathsGetByIdScheme},
					&mainInfoSchemer{scheme: pathsGetListScheme},
					&mainInfoSchemer{scheme: pathsDeleteByIdScheme},
				},
			},
			want: `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /admin/user/{id}:
    get:
      tags:
        - admin
        - user
      summary: Get user details
      operationId: getUserById
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
    delete:
      tags:
        - admin
        - user
      summary: Delete user
      operationId: deleteUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Status'
  /admin/user:
    get:
      tags:
        - admin
        - user
      summary: Get list of users
      operationId: getUsers
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := New(logger)
			ob.AddMainInfo(&mainInfoSchemer{scheme: mainInfoScheme})
			ob.AddPaths(tt.args.schemers)

			require.Equal(t, tt.want, ob.String())
		})
	}
}

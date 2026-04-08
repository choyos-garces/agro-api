package schema

import "testing"

func TestValidateGeoJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty string is allowed",
			input:   "",
			wantErr: false,
		},
		{
			name:    "valid point geometry",
			input:   `{"type":"Point","coordinates":[102.0,0.5]}`,
			wantErr: false,
		},
		{
			name:    "valid geometry collection",
			input:   `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[102.0,0.5]},{"type":"LineString","coordinates":[[100.0,0.0],[101.0,1.0]]}]}`,
			wantErr: false,
		},
		{
			name:    "valid feature",
			input:   `{"type":"Feature","geometry":{"type":"Point","coordinates":[102.0,0.5]},"properties":{"name":"test"}}`,
			wantErr: false,
		},
		{
			name:    "valid feature collection",
			input:   `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[102.0,0.5]},"properties":{"name":"test"}}]}`,
			wantErr: false,
		},
		{
			name:    "malformed json",
			input:   `{"type":"Point","coordinates":[102.0,0.5]`,
			wantErr: true,
		},
		{
			name:    "missing type",
			input:   `{"coordinates":[102.0,0.5]}`,
			wantErr: true,
		},
		{
			name:    "unsupported top level type",
			input:   `{"type":"NotGeoJSON","coordinates":[102.0,0.5]}`,
			wantErr: true,
		},
		{
			name:    "invalid point coordinates",
			input:   `{"type":"Point","coordinates":"bad"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateGeoJSON(tc.input)
			if tc.wantErr && err == nil {
				t.Fatalf("expected an error, got nil")
			}

			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

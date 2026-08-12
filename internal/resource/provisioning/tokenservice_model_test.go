package provisioning_test

import (
	"testing"

	provisioning "github.com/gamefabric/terraform-provider-gamefabric/internal/resource/provisioning"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFirstPlatformKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want types.String
	}{
		{
			name: "empty string",
			in:   "",
			want: types.StringNull(),
		},
		{
			name: "single platform",
			in:   `{"pc":["key1","key2"]}`,
			want: types.StringValue("key1"),
		},
		{
			name: "first platform in original key order",
			in:   `{"xbox":["key-xbox"],"pc":["key-pc"]}`,
			want: types.StringValue("key-xbox"),
		},
		{
			name: "empty object",
			in:   `{}`,
			want: types.StringNull(),
		},
		{
			name: "empty key list",
			in:   `{"pc":[]}`,
			want: types.StringNull(),
		},
		{
			name: "malformed json",
			in:   `not json`,
			want: types.StringNull(),
		},
		{
			name: "truncated json after first key (unclosed object)",
			in:   `{"pc":["key1"]`,
			want: types.StringNull(),
		},
		{
			name: "not an object",
			in:   `["pc"]`,
			want: types.StringNull(),
		},
		{
			name: "value not a list",
			in:   `{"pc":"key1"}`,
			want: types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := provisioning.FirstPlatformKey(tt.in)
			if !got.Equal(tt.want) {
				t.Errorf("FirstPlatformKey(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

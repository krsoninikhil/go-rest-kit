package filestorage

import "testing"

func TestIsOverMaxBytes(t *testing.T) {
	tests := []struct {
		name       string
		maxBytes   int64
		headerSize int64
		actualSize int64
		want       bool
	}{
		{
			name:       "returns false when max disabled",
			maxBytes:   0,
			headerSize: 10,
			actualSize: 10,
			want:       false,
		},
		{
			name:       "returns true when header exceeds limit",
			maxBytes:   50,
			headerSize: 51,
			actualSize: 0,
			want:       true,
		},
		{
			name:       "returns true when actual exceeds limit",
			maxBytes:   50,
			headerSize: 0,
			actualSize: 51,
			want:       true,
		},
		{
			name:       "returns false when both sizes within limit",
			maxBytes:   50,
			headerSize: 50,
			actualSize: 50,
			want:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isOverMaxBytes(tc.maxBytes, tc.headerSize, tc.actualSize)
			if got != tc.want {
				t.Fatalf("isOverMaxBytes(%d, %d, %d) = %v, want %v", tc.maxBytes, tc.headerSize, tc.actualSize, got, tc.want)
			}
		})
	}
}

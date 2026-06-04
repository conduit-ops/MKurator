package mqadmin

import "testing"

func TestNormalizeQueueType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   QueueType
		want QueueType
	}{
		{in: "", want: QueueTypeLocal},
		{in: QueueTypeAlias, want: QueueTypeAlias},
		{in: QueueTypeRemote, want: QueueTypeRemote},
	}
	for _, tt := range tests {
		if got := NormalizeQueueType(tt.in); got != tt.want {
			t.Fatalf("NormalizeQueueType(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

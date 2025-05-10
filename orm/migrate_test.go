package orm

import "testing"

func Test_removeSQLComment(t *testing.T) {

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "Case1",
			arg:  "-- SEQUENCE\nCREATE SEQUENCE IF",
			want: "CREATE SEQUENCE IF",
		},
		{
			name: "Case2",
			arg:  "\n-- SEQUENCE\n-- SEQUENCE\n-- SEQUENCE\n\nCREATE SEQUENCE IF",
			want: "CREATE SEQUENCE IF",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeSQLComment(tt.arg); got != tt.want {
				t.Errorf("removeSQLComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

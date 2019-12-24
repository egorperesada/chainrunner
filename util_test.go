package chainrunner

import (
	"reflect"
	"testing"
)

func Test_deleteFromSlice(t *testing.T) {
	type args struct {
		data     []interface{}
		toDelete []int
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			"inRow",
			args{
				data:     []interface{}{1, 2, 3, 4, 5},
				toDelete: []int{0, 1, 2},
			},
			[]interface{}{4, 5},
		},
		{
			"discordantly",
			args{
				data:     []interface{}{1, 2, 3, 4, 5, 6, 7},
				toDelete: []int{0, 3, 6},
			},
			[]interface{}{2, 3, 5, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deleteFromSlice(tt.args.data, tt.args.toDelete); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deleteFromSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

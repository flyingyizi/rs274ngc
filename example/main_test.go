package main

import (
	"reflect"
	"testing"

	. "github.com/flyingyizi/rs274ngc"
	"github.com/flyingyizi/rs274ngc/inc"
)

func TestCNC_interpret_from_file(t *testing.T) {
	type args struct {
		do_next      int
		block_delete ON_OFF
		print_stack  ON_OFF
	}
	var cnc = CNC{}
	cnc.init()
	cnc.open("")

	tests := []struct {
		name string
		c    *CNC
		args args
		want inc.STATUS
	}{
		// TODO: Add test cases.
		{name: "simple",
			args: args{do_next: 1, block_delete: OFF, print_stack: OFF},
			c:    &cnc,
			want: inc.RS274NGC_EXIT},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.interpret_from_file(tt.args.do_next, tt.args.block_delete, tt.args.print_stack); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CNC.interpret_from_file() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCNC_open(t *testing.T) {
	type args struct {
		filename string
	}
	var cnc = CNC{}
	cnc.init()

	tests := []struct {
		name string
		c    *CNC
		args args
		want inc.STATUS
	}{
		// TODO: Add test cases.
		{name: "simple", c: &cnc, args: args{""}, want: inc.RS274NGC_OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.open(tt.args.filename); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CNC.open() = %v, want %v", got, tt.want)
			}
		})
	}
}

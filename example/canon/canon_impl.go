package canon

import (
	"github.com/flyingyizi/rs274ngc/inc"
)

var (
	/* Dummy world model */

	_active_plane                             = inc.CANON_PLANE_XY
	_active_slot                              = 1
	_feed_rate          float64               = 0.0
	_flood                                    = 0
	_length_unit_factor float64               = 1.0 /* 1 for MM 25.4 for inch */
	_length_unit_type   inc.CANON_UNITS       = inc.CANON_UNITS_MM
	_line_number                              = 1
	_mist                                     = 0
	_motion_mode        inc.CANON_MOTION_MODE = inc.CANON_CONTINUOUS
	/*Not static.Driver writes*/
	_parameter_file_name string

	_probe_position_a float64 = 0.0 /*AA*/
	_probe_position_b float64 = 0.0 /*BB*/
	_probe_position_c float64 = 0.0 /*CC*/
	_probe_position_x float64 = 0.0
	_probe_position_y float64 = 0.0
	_probe_position_z float64 = 0.0

	_program_origin_a   float64 = 0.0 /*AA*/
	_program_origin_b   float64 = 0.0 /*BB*/
	_program_origin_c   float64 = 0.0 /*CC*/
	_program_origin_x   float64 = 0.0
	_program_origin_y   float64 = 0.0
	_program_origin_z   float64 = 0.0
	_program_position_a float64 = 0.0 /*AA*/
	_program_position_b float64 = 0.0 /*BB*/
	_program_position_c float64 = 0.0 /*CC*/

	_program_position_x float64 = 0.0
	_program_position_y float64 = 0.0
	_program_position_z float64 = 0.0
	_spindle_speed      float64
	_spindle_turning    inc.CANON_DIRECTION
	_tool_max           = 68                                     /*Not static. Driver reads  */
	_tools              [inc.CANON_TOOL_MAX]inc.CANON_TOOL_TABLE /*Not static. Driver writes */
	_traverse_rate      float64
)

type Canon_t struct {
}

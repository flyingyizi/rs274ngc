package canon

import (
	"math"

	"github.com/flyingyizi/rs274ngc/inc"
)

/************************************************************************/

/* Canonical "Do it" functions

   This is a set of dummy definitions for the canonical machining functions
   given in canon.hh. These functions just print themselves and, if necessary,
   update the dummy world model. On each output line is printed:
   1. an output line number (sequential, starting with 1).
   2. an input line number read from the input (or ... if not provided).
   3. a printed representation of the function call which was made.

   If an interpreter which makes these calls is compiled with this set of
   definitions, it can be used as a translator by redirecting output from
   stdout to a file.

*/

/* Representation */

func (canon Canon_t) SET_ORIGIN_OFFSETS(
	x, y, z float64,
	a, /*AA*/
	b, /*BB*/
	c float64) { /*CC*/

	myFprintf("%5d ", _line_number)
	_line_number++
	//TODO //TODO print_nc_line_number()
	myFprintf("SET_ORIGIN_OFFSETS(%.4f, %.4f, %.4f, %.4f, %.4f, %.4f)\n", x, y, z, a, b, c)
	_program_position_x = _program_position_x + _program_origin_x - x
	_program_position_y = _program_position_y + _program_origin_y - y
	_program_position_z = _program_position_z + _program_origin_z - z
	_program_position_a = _program_position_a + _program_origin_a - a
	_program_position_b = _program_position_b + _program_origin_b - b
	_program_position_c = _program_position_c + _program_origin_c - c

	_program_origin_x = x
	_program_origin_y = y
	_program_origin_z = z

	_program_origin_a = a /*AA*/
	_program_origin_b = b /*BB*/
	_program_origin_c = c /*CC*/
}

func (c Canon_t) USE_LENGTH_UNITS(in_unit inc.CANON_UNITS) {
	if in_unit == inc.CANON_UNITS_INCHES {
		myFprintf("USE_LENGTH_UNITS(CANON_UNITS_INCHES)\n")
		if _length_unit_type == inc.CANON_UNITS_MM {
			_length_unit_type = inc.CANON_UNITS_INCHES
			_length_unit_factor = 25.4
			_program_origin_x = (_program_origin_x / 25.4)
			_program_origin_y = (_program_origin_y / 25.4)
			_program_origin_z = (_program_origin_z / 25.4)
			_program_position_x = (_program_position_x / 25.4)
			_program_position_y = (_program_position_y / 25.4)
			_program_position_z = (_program_position_z / 25.4)
		}
	} else if in_unit == inc.CANON_UNITS_MM {
		myFprintf("USE_LENGTH_UNITS(CANON_UNITS_MM)\n")
		if _length_unit_type == inc.CANON_UNITS_INCHES {
			_length_unit_type = inc.CANON_UNITS_MM
			_length_unit_factor = 1.0
			_program_origin_x = (_program_origin_x * 25.4)
			_program_origin_y = (_program_origin_y * 25.4)
			_program_origin_z = (_program_origin_z * 25.4)
			_program_position_x = (_program_position_x * 25.4)
			_program_position_y = (_program_position_y * 25.4)
			_program_position_z = (_program_position_z * 25.4)
		}
	} else {
		myFprintf("USE_LENGTH_UNITS(UNKNOWN)\n")
	}
}

/* Free Space Motion */
func (c Canon_t) SET_TRAVERSE_RATE(rate float64) {
	myFprintf("SET_TRAVERSE_RATE(%.4f)\n", rate)
	_traverse_rate = rate
}

func (canon Canon_t) STRAIGHT_TRAVERSE(x, y, z, a, b, c float64) { /*CC*/

	myFprintf("%5d ", _line_number)
	_line_number++
	//TODO //TODO print_nc_line_number()
	myFprintf("STRAIGHT_TRAVERSE(%.4f, %.4f, %.4f, %.4f, %.4f, %.4f)\n", x, y, z, a, b, c) /*CC*/

	_program_position_x = x
	_program_position_y = y
	_program_position_z = z
	_program_position_a = a /*AA*/
	_program_position_b = b /*BB*/
	_program_position_c = c /*CC*/
}

/* Machining Attributes */
func (c Canon_t) SET_FEED_RATE(rate float64) {
	myFprintf("SET_FEED_RATE(%.4f)\n", rate)
	_feed_rate = rate
}

func (c Canon_t) SET_FEED_REFERENCE(reference inc.CANON_FEED_REFERENCE) {
	myFprintf("SET_FEED_REFERENCE(%s)\n", inc.If(reference == inc.CANON_WORKPIECE, "CANON_WORKPIECE", "CANON_XYZ").(string))
}

func (c Canon_t) SET_MOTION_CONTROL_MODE(mode inc.CANON_MOTION_MODE) {
	if mode == inc.CANON_EXACT_STOP {
		myFprintf("SET_MOTION_CONTROL_MODE(CANON_EXACT_STOP)\n")
		_motion_mode = inc.CANON_EXACT_STOP
	} else if mode == inc.CANON_EXACT_PATH {
		myFprintf("SET_MOTION_CONTROL_MODE(CANON_EXACT_PATH)\n")
		_motion_mode = inc.CANON_EXACT_PATH
	} else if mode == inc.CANON_CONTINUOUS {
		myFprintf("SET_MOTION_CONTROL_MODE(CANON_CONTINUOUS)\n")
		_motion_mode = inc.CANON_CONTINUOUS
	} else {
		myFprintf("SET_MOTION_CONTROL_MODE(UNKNOWN)\n")
	}
}

func (c Canon_t) SELECT_PLANE(in_plane inc.CANON_PLANE) {
	//TODO
	//myFprintf("SELECT_PLANE(CANON_PLANE_%s)\n",
	//	((in_plane == CANON_PLANE_XY) ? "XY" :
	//(in_plane == CANON_PLANE_YZ) ? "YZ" :
	//inc.If(in_plane == CANON_PLANE_XZ, "XZ", "UNKNOWN"));
	//_active_plane = in_plane;
}

func (c Canon_t) SET_CUTTER_RADIUS_COMPENSATION(radius float64) {
	myFprintf("SET_CUTTER_RADIUS_COMPENSATION(%.4f)\n", radius)
}

func (c Canon_t) START_CUTTER_RADIUS_COMPENSATION(side inc.CANON_SIDE) {
	myFprintf("START_CUTTER_RADIUS_COMPENSATION(%s)\n",
		inc.If(side == inc.CANON_SIDE_LEFT, "LEFT", inc.If(side == inc.CANON_SIDE_RIGHT, "RIGHT", "UNKNOWN").(string)).(string))
}

func (c Canon_t) STOP_CUTTER_RADIUS_COMPENSATION() {
	myFprintf("STOP_CUTTER_RADIUS_COMPENSATION()\n")
}

func (c Canon_t) START_SPEED_FEED_SYNCH() {
	myFprintf("START_SPEED_FEED_SYNCH()\n")
}

func (c Canon_t) STOP_SPEED_FEED_SYNCH() {
	myFprintf("STOP_SPEED_FEED_SYNCH()\n")
}

/* Machining Functions */

func (canon Canon_t) ARC_FEED(
	first_end, second_end, first_axis, second_axis float64, rotation int,
	axis_end_point, a, b, c float64) { /*CC*/

	myFprintf("%5d ", _line_number)
	_line_number++
	//TODO print_nc_line_number()
	myFprintf("ARC_FEED(%.4f, %.4f, %.4f, %.4f, %d, %.4f, %.4f, %.4f, %.4f)\n",
		first_end, second_end, first_axis, second_axis, rotation, axis_end_point, a, b, c) /*CC*/

	if _active_plane == inc.CANON_PLANE_XY {
		_program_position_x = first_end
		_program_position_y = second_end
		_program_position_z = axis_end_point
	} else if _active_plane == inc.CANON_PLANE_YZ {
		_program_position_x = axis_end_point
		_program_position_y = first_end
		_program_position_z = second_end
	} else { /* if (_active_plane == CANON_PLANE_XZ) */
		_program_position_x = second_end
		_program_position_y = axis_end_point
		_program_position_z = first_end
	}

	_program_position_a = a /*AA*/

	_program_position_b = b /*BB*/

	_program_position_c = c /*CC*/

}

func (canon Canon_t) STRAIGHT_FEED(
	x, y, z, a, b, c float64) { /*CC*/

	myFprintf("%5d ", _line_number)
	_line_number++
	//TODO print_nc_line_number()
	myFprintf("STRAIGHT_FEED(%.4f, %.4f, %.4f, %.4f, %.4f, %.4f)\n", x, y, z, a, b, c)

	_program_position_x = x
	_program_position_y = y
	_program_position_z = z
	_program_position_a = a /*AA*/
	_program_position_b = b /*BB*/
	_program_position_c = c /*CC*/
}

/* This models backing the probe off 0.01 inch or 0.254 mm from the probe
   point towards the previous location after the probing, if the probe
   point is not the same as the previous point -- which it should not be. */

func (canon Canon_t) STRAIGHT_PROBE(
	x, y, z, a, b, c float64) { /*CC*/

	var distance, dx, dy, dz, backoff float64

	dx = (_program_position_x - x)
	dy = (_program_position_y - y)
	dz = (_program_position_z - z)

	distance = math.Sqrt((dx * dx) + (dy * dy) + (dz * dz))

	myFprintf("%5d ", _line_number)
	_line_number++
	//TODO print_nc_line_number()
	myFprintf("STRAIGHT_PROBE(%.4f, %.4f, %.4f, %.4f, %.4f, %.4f)\n", x, y, z, a, b, c)

	_probe_position_x = x
	_probe_position_y = y
	_probe_position_z = z
	_probe_position_a = a /*AA*/
	_probe_position_b = b /*BB*/
	_probe_position_c = c /*CC*/
	if distance == 0 {
		_program_position_x = _program_position_x
		_program_position_y = _program_position_y
		_program_position_z = _program_position_z
	} else {
		//TODO
		backoff = inc.If(_length_unit_type == inc.CANON_UNITS_MM, 0.254, 0.01).(float64)
		_program_position_x = (x + (backoff * (dx / distance)))
		_program_position_y = (y + (backoff * (dy / distance)))
		_program_position_z = (z + (backoff * (dz / distance)))
	}
	_program_position_a = a /*AA*/
	_program_position_b = b /*BB*/
	_program_position_c = c /*CC*/
}

/*
	   func (c Canon_t) PARAMETRIC_2D_CURVE_FEED(FunctionPtr f1, FunctionPtr f2,
						 double start_parameter_value,
	   double end_parameter_value) {}

	   func (c Canon_t) PARAMETRIC_3D_CURVE_FEED(FunctionPtr xfcn, FunctionPtr yfcn,
	   FunctionPtr zfcn, double start_parameter_value,
	   double end_parameter_value) {}
*/

func (c Canon_t) DWELL(seconds float64) {
	myFprintf("DWELL(%.4f)\n", seconds)
}

/* Spindle Functions */
func (c Canon_t) SPINDLE_RETRACT_TRAVERSE() {
	myFprintf("SPINDLE_RETRACT_TRAVERSE()\n")
}

func (c Canon_t) START_SPINDLE_CLOCKWISE() {
	myFprintf("START_SPINDLE_CLOCKWISE()\n")
	_spindle_turning = inc.If(_spindle_speed == 0, inc.CANON_STOPPED, inc.CANON_CLOCKWISE).(inc.CANON_DIRECTION)
}

func (c Canon_t) START_SPINDLE_COUNTERCLOCKWISE() {
	myFprintf("START_SPINDLE_COUNTERCLOCKWISE()\n")
	_spindle_turning = inc.If(_spindle_speed == 0, inc.CANON_STOPPED, inc.CANON_COUNTERCLOCKWISE).(inc.CANON_DIRECTION)
}

func (c Canon_t) SET_SPINDLE_SPEED(rpm float64) {
	myFprintf("SET_SPINDLE_SPEED(%.4f)\n", rpm)
	_spindle_speed = rpm
}

func (c Canon_t) STOP_SPINDLE_TURNING() {
	myFprintf("STOP_SPINDLE_TURNING()\n")
	_spindle_turning = inc.CANON_STOPPED
}

func (c Canon_t) SPINDLE_RETRACT() {
	myFprintf("SPINDLE_RETRACT()\n")
}

func (c Canon_t) ORIENT_SPINDLE(orientation float64, direction inc.CANON_DIRECTION) {
	myFprintf("ORIENT_SPINDLE(%.4f, %s)\n",
		orientation,
		inc.If(direction == inc.CANON_CLOCKWISE, "CANON_CLOCKWISE", "CANON_COUNTERCLOCKWISE").(string))
}

func (c Canon_t) USE_NO_SPINDLE_FORCE() {
	myFprintf("USE_NO_SPINDLE_FORCE()\n")
}

/* Tool Functions */

func (c Canon_t) USE_TOOL_LENGTH_OFFSET(length float64) {
	myFprintf("USE_TOOL_LENGTH_OFFSET(%.4f)\n", length)
}

func (c Canon_t) CHANGE_TOOL(slot int) {
	myFprintf("CHANGE_TOOL(%d)\n", slot)
	_active_slot = slot
}

func (c Canon_t) SELECT_TOOL(slot int) {
	myFprintf("SELECT_TOOL(%d)\n", slot)
}

/* Misc Functions */

func (c Canon_t) CLAMP_AXIS(axis inc.CANON_AXIS) {
	myFprintf("CLAMP_AXIS(%s)\n",
		inc.If(axis == inc.CANON_AXIS_X, "CANON_AXIS_X",
			inc.If(axis == inc.CANON_AXIS_Y, "CANON_AXIS_Y",
				inc.If(axis == inc.CANON_AXIS_Z, "CANON_AXIS_Z",
					inc.If(axis == inc.CANON_AXIS_A, "CANON_AXIS_A",
						inc.If(axis == inc.CANON_AXIS_C, "CANON_AXIS_C", "UNKNOWN").(string)).(string)).(string)).(string)).(string))
}

func (c Canon_t) COMMENT(s string) {
	myFprintf("COMMENT(\"%s\")\n", s)
}

func (c Canon_t) DISABLE_FEED_OVERRIDE() {
	myFprintf("DISABLE_FEED_OVERRIDE()\n")
}

func (c Canon_t) DISABLE_SPEED_OVERRIDE() {
	myFprintf("DISABLE_SPEED_OVERRIDE()\n")
}

func (c Canon_t) ENABLE_FEED_OVERRIDE() {
	myFprintf("ENABLE_FEED_OVERRIDE()\n")
}

func (c Canon_t) ENABLE_SPEED_OVERRIDE() {
	myFprintf("ENABLE_SPEED_OVERRIDE()\n")
}

func (c Canon_t) FLOOD_OFF() {
	myFprintf("FLOOD_OFF()\n")
	_flood = 0
}

func (c Canon_t) FLOOD_ON() {
	myFprintf("FLOOD_ON()\n")
	_flood = 1
}

func (c Canon_t) INIT_CANON() {
}

func (c Canon_t) MESSAGE(s []byte) {
	myFprintf("MESSAGE(\"%s\")\n", s)
}

func (c Canon_t) MIST_OFF() {
	myFprintf("MIST_OFF()\n")
	_mist = 0
}

func (c Canon_t) MIST_ON() {
	myFprintf("MIST_ON()\n")
	_mist = 1
}

func (c Canon_t) PALLET_SHUTTLE() {
	myFprintf("PALLET_SHUTTLE()\n")
}

func (c Canon_t) TURN_PROBE_OFF() {
	myFprintf("TURN_PROBE_OFF()\n")
}

func (c Canon_t) TURN_PROBE_ON() {
	myFprintf("TURN_PROBE_ON()\n")
}

func (c Canon_t) UNCLAMP_AXIS(axis inc.CANON_AXIS) {
	//TODO
	myFprintf("UNCLAMP_AXIS(%s)\n",
		inc.If(axis == inc.CANON_AXIS_X, "CANON_AXIS_X",
			inc.If(axis == inc.CANON_AXIS_Y, "CANON_AXIS_Y",
				inc.If(axis == inc.CANON_AXIS_Z, "CANON_AXIS_Z",
					inc.If(axis == inc.CANON_AXIS_A, "CANON_AXIS_A",
						inc.If(axis == inc.CANON_AXIS_B, "CANON_AXIS_B",
							inc.If(axis == inc.CANON_AXIS_C, "CANON_AXIS_C", "UNKNOWN").(string)).(string)).(string)).(string)).(string)).(string))
}

/* Program Functions */

func (c Canon_t) PROGRAM_STOP() {
	myFprintf("PROGRAM_STOP()\n")
}

func (c Canon_t) OPTIONAL_PROGRAM_STOP() {
	myFprintf("OPTIONAL_PROGRAM_STOP()\n")
}

func (c Canon_t) PROGRAM_END() {
	myFprintf("PROGRAM_END()\n")
}

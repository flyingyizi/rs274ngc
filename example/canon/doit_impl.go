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

	//TODO fmt.Println(_outfile, "%5d ", _line_number)
	_line_number++
	//TODO //TODO print_nc_line_number()
	//TODO fmt.Println(_outfile, "SET_ORIGIN_OFFSETS(%.4f, %.4f, %.4f, %.4f, %.4f, %.4f)\n", x, y, z, a, b, c)
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
		//TODO fmt.Println("USE_LENGTH_UNITS(CANON_UNITS_INCHES)\n")
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
		//TODO fmt.Println("USE_LENGTH_UNITS(CANON_UNITS_MM)\n")
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
		//TODO fmt.Println("USE_LENGTH_UNITS(UNKNOWN)\n")
	}
}

/* Free Space Motion */
func (c Canon_t) SET_TRAVERSE_RATE(rate float64) {
	//TODO fmt.Println("SET_TRAVERSE_RATE(%.4f)\n", rate)
	_traverse_rate = rate
}

func (canon Canon_t) STRAIGHT_TRAVERSE(x, y, z, a, b, c float64) { /*CC*/

	//TODO fmt.Println(_outfile, "%5d ", _line_number)
	_line_number++
	//TODO //TODO print_nc_line_number()
	//TODO fmt.Println(_outfile, "STRAIGHT_TRAVERSE(%.4f, %.4f, %.4f, %.4f, %.4f, %.4f)\n", x, y, z,a, 	b, c) /*CC*/

	_program_position_x = x
	_program_position_y = y
	_program_position_z = z
	_program_position_a = a /*AA*/
	_program_position_b = b /*BB*/
	_program_position_c = c /*CC*/
}

/* Machining Attributes */
func (c Canon_t) SET_FEED_RATE(rate float64) {
	//TODO fmt.Println("SET_FEED_RATE(%.4f)\n", rate)
	_feed_rate = rate
}

func (c Canon_t) SET_FEED_REFERENCE(reference inc.CANON_FEED_REFERENCE) {
	//TODO fmt.Println("SET_FEED_REFERENCE(%s)\n",	inc.If(reference == CANON_WORKPIECE, "CANON_WORKPIECE", "CANON_XYZ").(string))
}

func (c Canon_t) SET_MOTION_CONTROL_MODE(mode inc.CANON_MOTION_MODE) {
	if mode == inc.CANON_EXACT_STOP {
		//TODO fmt.Println("SET_MOTION_CONTROL_MODE(CANON_EXACT_STOP)\n")
		_motion_mode = inc.CANON_EXACT_STOP
	} else if mode == inc.CANON_EXACT_PATH {
		//TODO fmt.Println("SET_MOTION_CONTROL_MODE(CANON_EXACT_PATH)\n")
		_motion_mode = inc.CANON_EXACT_PATH
	} else if mode == inc.CANON_CONTINUOUS {
		//TODO fmt.Println("SET_MOTION_CONTROL_MODE(CANON_CONTINUOUS)\n")
		_motion_mode = inc.CANON_CONTINUOUS
	} else {
		//TODO fmt.Println("SET_MOTION_CONTROL_MODE(UNKNOWN)\n")
	}
}

func (c Canon_t) SELECT_PLANE(in_plane inc.CANON_PLANE) {
	//TODO
	////TODO fmt.Println("SELECT_PLANE(CANON_PLANE_%s)\n",
	//	((in_plane == CANON_PLANE_XY) ? "XY" :
	//(in_plane == CANON_PLANE_YZ) ? "YZ" :
	//inc.If(in_plane == CANON_PLANE_XZ, "XZ", "UNKNOWN"));
	//_active_plane = in_plane;
}

func (c Canon_t) SET_CUTTER_RADIUS_COMPENSATION(radius float64) {
	//TODO fmt.Println("SET_CUTTER_RADIUS_COMPENSATION(%.4f)\n", radius)
}

func (c Canon_t) START_CUTTER_RADIUS_COMPENSATION(side int) {
	//TODO fmt.Println("START_CUTTER_RADIUS_COMPENSATION(%s)\n",		inc.If(side == CANON_SIDE_LEFT, "LEFT",			inc.If(side == CANON_SIDE_RIGHT, "RIGHT", "UNKNOWN").(string)).(string))
}

func (c Canon_t) STOP_CUTTER_RADIUS_COMPENSATION() {
	//TODO fmt.Println("STOP_CUTTER_RADIUS_COMPENSATION()\n")
}

func (c Canon_t) START_SPEED_FEED_SYNCH() {
	//TODO fmt.Println("START_SPEED_FEED_SYNCH()\n")
}

func (c Canon_t) STOP_SPEED_FEED_SYNCH() {
	//TODO fmt.Println("STOP_SPEED_FEED_SYNCH()\n")
}

/* Machining Functions */

func (canon Canon_t) ARC_FEED(
	first_end, second_end, first_axis, second_axis float64, rotation int,
	axis_end_point, a, b, c float64) { /*CC*/

	//TODO fmt.Println(_outfile, "%5d ", _line_number)
	_line_number++
	//TODO print_nc_line_number()
	//TODO fmt.Println(_outfile, "ARC_FEED(%.4f, %.4f, %.4f, %.4f, %d, %.4f, %.4f, %.4f, %.4f)\n", first_end, second_end, first_axis, second_axis,		rotation, axis_end_point, a, b, c) /*CC*/

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

	//TODO fmt.Println(_outfile, "%5d ", _line_number)
	_line_number++
	//TODO print_nc_line_number()
	//TODO fmt.Println(_outfile, "STRAIGHT_FEED(%.4f, %.4f, %.4f, %.4f, %.4f, %.4f)\n", x, y, z, a, b, c)

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

	var distance, dx, dy, dz /*,backoff*/ float64

	dx = (_program_position_x - x)
	dy = (_program_position_y - y)
	dz = (_program_position_z - z)

	distance = math.Sqrt((dx * dx) + (dy * dy) + (dz * dz))

	//TODO fmt.Println(_outfile, "%5d ", _line_number)
	_line_number++
	//TODO print_nc_line_number()
	//TODO fmt.Println(_outfile, "STRAIGHT_PROBE(%.4f, %.4f, %.4f, %.4f, %.4f, %.4f)\n", x, y, z, a, b, c)

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
		//backoff = ((_length_unit_type == CANON_UNITS_MM) ? 0.254 : 0.01);
		//_program_position_x = (x + (backoff * (dx / distance)));
		//_program_position_y = (y + (backoff * (dy / distance)));
		//_program_position_z = (z + (backoff * (dz / distance)));
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
	//TODO fmt.Println("DWELL(%.4f)\n", seconds)
}

/* Spindle Functions */
func (c Canon_t) SPINDLE_RETRACT_TRAVERSE() {
	//TODO fmt.Println("SPINDLE_RETRACT_TRAVERSE()\n")
}

func (c Canon_t) START_SPINDLE_CLOCKWISE() {
	//TODO fmt.Println("START_SPINDLE_CLOCKWISE()\n")
	//TODO _spindle_turning = ((_spindle_speed == 0) ? CANON_STOPPED :		CANON_CLOCKWISE);
}

func (c Canon_t) START_SPINDLE_COUNTERCLOCKWISE() {
	//TODO fmt.Println("START_SPINDLE_COUNTERCLOCKWISE()\n")
	//TDO _spindle_turning = ((_spindle_speed == 0) ? CANON_STOPPED :		CANON_COUNTERCLOCKWISE);
}

func (c Canon_t) SET_SPINDLE_SPEED(rpm float64) {
	//TODO fmt.Println("SET_SPINDLE_SPEED(%.4f)\n", rpm)
	_spindle_speed = rpm
}

func (c Canon_t) STOP_SPINDLE_TURNING() {
	//TODO fmt.Println("STOP_SPINDLE_TURNING()\n")
	_spindle_turning = inc.CANON_STOPPED
}

func (c Canon_t) SPINDLE_RETRACT() {
	//TODO fmt.Println("SPINDLE_RETRACT()\n")
}

func (c Canon_t) ORIENT_SPINDLE(orientation float64, direction inc.CANON_DIRECTION) {
	//TODO PRINT2("ORIENT_SPINDLE(%.4f, %s)\n", orientation,			(direction == CANON_CLOCKWISE) ? "CANON_CLOCKWISE" :	"CANON_COUNTERCLOCKWISE");
}

func (c Canon_t) USE_NO_SPINDLE_FORCE() {
	//TODO fmt.Println("USE_NO_SPINDLE_FORCE()\n")
}

/* Tool Functions */

func (c Canon_t) USE_TOOL_LENGTH_OFFSET(length float64) {
	//TODO fmt.Println("USE_TOOL_LENGTH_OFFSET(%.4f)\n", length)
}

func (c Canon_t) CHANGE_TOOL(slot int) {
	//TODO fmt.Println("CHANGE_TOOL(%d)\n", slot)
	_active_slot = slot
}

func (c Canon_t) SELECT_TOOL(slot int) {
	//TODO fmt.Println("SELECT_TOOL(%d)\n", slot)
}

/* Misc Functions */

func (c Canon_t) CLAMP_AXIS(axis inc.CANON_AXIS) {
	//TODO
	////TODO fmt.Println("CLAMP_AXIS(%s)\n",
	//	(axis == CANON_AXIS_X) ? "CANON_AXIS_X" :
	//(axis == CANON_AXIS_Y) ? "CANON_AXIS_Y" :
	//(axis == CANON_AXIS_Z) ? "CANON_AXIS_Z" :
	//(axis == CANON_AXIS_A) ? "CANON_AXIS_A" :
	//(axis == CANON_AXIS_C) ? "CANON_AXIS_C" : "UNKNOWN");
}

func (c Canon_t) COMMENT(s string) {
	//TODO fmt.Println("COMMENT(\"%s\")\n", s)
}

func (c Canon_t) DISABLE_FEED_OVERRIDE() {
	//TODO fmt.Println("DISABLE_FEED_OVERRIDE()\n")
}

func (c Canon_t) DISABLE_SPEED_OVERRIDE() {
	//TODO fmt.Println("DISABLE_SPEED_OVERRIDE()\n")
}

func (c Canon_t) ENABLE_FEED_OVERRIDE() {
	//TODO fmt.Println("ENABLE_FEED_OVERRIDE()\n")
}

func (c Canon_t) ENABLE_SPEED_OVERRIDE() {
	//TODO fmt.Println("ENABLE_SPEED_OVERRIDE()\n")
}

func (c Canon_t) FLOOD_OFF() {
	//TODO fmt.Println("FLOOD_OFF()\n")
	_flood = 0
}

func (c Canon_t) FLOOD_ON() {
	//TODO fmt.Println("FLOOD_ON()\n")
	_flood = 1
}

func (c Canon_t) INIT_CANON() {
}

func (c Canon_t) MESSAGE(s []byte) {
	//TODO fmt.Println("MESSAGE(\"%s\")\n", s)
}

func (c Canon_t) MIST_OFF() {
	//TODO fmt.Println("MIST_OFF()\n")
	_mist = 0
}

func (c Canon_t) MIST_ON() {
	//TODO fmt.Println("MIST_ON()\n")
	_mist = 1
}

func (c Canon_t) PALLET_SHUTTLE() {
	//TODO fmt.Println("PALLET_SHUTTLE()\n")
}

func (c Canon_t) TURN_PROBE_OFF() {
	//TODO fmt.Println("TURN_PROBE_OFF()\n")
}

func (c Canon_t) TURN_PROBE_ON() {
	//TODO fmt.Println("TURN_PROBE_ON()\n")
}

func (c Canon_t) UNCLAMP_AXIS(axis inc.CANON_AXIS) {
	//TODO
	////TODO fmt.Println("UNCLAMP_AXIS(%s)\n",
	//	(axis == CANON_AXIS_X) ? "CANON_AXIS_X" :
	//(axis == CANON_AXIS_Y) ? "CANON_AXIS_Y" :
	//(axis == CANON_AXIS_Z) ? "CANON_AXIS_Z" :
	//(axis == CANON_AXIS_A) ? "CANON_AXIS_A" :
	//(axis == CANON_AXIS_B) ? "CANON_AXIS_B" :
	//(axis == CANON_AXIS_C) ? "CANON_AXIS_C" : "UNKNOWN");
}

/* Program Functions */

func (c Canon_t) PROGRAM_STOP() {
	//TODO fmt.Println("PROGRAM_STOP()\n")
}

func (c Canon_t) OPTIONAL_PROGRAM_STOP() {
	//TODO fmt.Println("OPTIONAL_PROGRAM_STOP()\n")
}

func (c Canon_t) PROGRAM_END() { //TODO fmt.Println("PROGRAM_END()\n")
}

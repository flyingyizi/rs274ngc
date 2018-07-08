package canon

import "github.com/flyingyizi/rs274ngc/inc"

/*************************************************************************/

/* Canonical "Give me information" functions

   In general, returned values are valid only if any canonical do it commands
   that may have been called for have been executed to completion. If a function
   returns a valid value regardless of execution, that is noted in the comments
   below.

*/

/* The interpreter is not using this function
   // Returns the system angular unit factor, in units / degree
   extern double GET_EXTERNAL_ANGLE_UNIT_FACTOR()
   {
   return 1;
   }
*/

/* Returns the system feed rate */
func (c Canon_t) GET_EXTERNAL_FEED_RATE() float64 {
	return _feed_rate
}

/* Returns the system flood coolant setting zero = off, non-zero = on */
func (c Canon_t) GET_EXTERNAL_FLOOD() int {
	return _flood
}

/* Returns the system length unit factor, in units per mm */
func (c Canon_t) GET_EXTERNAL_LENGTH_UNIT_FACTOR() float64 {
	return 1 / _length_unit_factor
}

/* Returns the system length unit type */
func (c Canon_t) GET_EXTERNAL_LENGTH_UNIT_TYPE() inc.CANON_UNITS {
	return _length_unit_type
}

/* Returns the system mist coolant setting zero = off, non-zero = on */
func (c Canon_t) GET_EXTERNAL_MIST() int {
	return _mist
}

// Returns the current motion control mode
func (c Canon_t) GET_EXTERNAL_MOTION_CONTROL_MODE() inc.CANON_MOTION_MODE {
	return _motion_mode
}

/* The interpreter is not using these six GET_EXTERNAL_ORIGIN functions

#ifdef AA
// returns the current a-axis origin offset
double GET_EXTERNAL_ORIGIN_A()
{
return _program_origin_a;
}
#endif

#ifdef BB
// returns the current b-axis origin offset
double GET_EXTERNAL_ORIGIN_B()
{
return _program_origin_b;
}
#endif

#ifdef CC
// returns the current c-axis origin offset
double GET_EXTERNAL_ORIGIN_C()
{
return _program_origin_c;
}
#endif

// returns the current x-axis origin offset
double GET_EXTERNAL_ORIGIN_X()
{
return _program_origin_x;
}

// returns the current y-axis origin offset
double GET_EXTERNAL_ORIGIN_Y()
{
return _program_origin_y;
}

// returns the current z-axis origin offset
double GET_EXTERNAL_ORIGIN_Z()
{
return _program_origin_z;
}

*/

func (c Canon_t) GET_EXTERNAL_PARAMETER_FILE_NAME(
	max_size int) string { /* maximum number of characters to copy */
	return _parameter_file_name
}

func (c Canon_t) GET_EXTERNAL_PLANE() inc.CANON_PLANE {
	return _active_plane
}

/* returns the current a-axis position */
func (c Canon_t) GET_EXTERNAL_POSITION_A() float64 {
	return _program_position_a
}

/* returns the current b-axis position */
func (c Canon_t) GET_EXTERNAL_POSITION_B() float64 {
	return _program_position_b
}

/* returns the current c-axis position */
func (c Canon_t) GET_EXTERNAL_POSITION_C() float64 {
	return _program_position_c
}

/* returns the current x-axis position */
func (c Canon_t) GET_EXTERNAL_POSITION_X() float64 {
	return _program_position_x
}

/* returns the current y-axis position */
func (c Canon_t) GET_EXTERNAL_POSITION_Y() float64 {
	return _program_position_y
}

/* returns the current z-axis position */
func (c Canon_t) GET_EXTERNAL_POSITION_Z() float64 {
	return _program_position_z
}

/* returns the a-axis position at the last probe trip. This is only valid
   once the probe command has executed to completion. */
func (c Canon_t) GET_EXTERNAL_PROBE_POSITION_A() float64 {
	return _probe_position_a
}

/* returns the b-axis position at the last probe trip. This is only valid
   once the probe command has executed to completion. */
func (c Canon_t) GET_EXTERNAL_PROBE_POSITION_B() float64 {
	return _probe_position_b
}

/* returns the c-axis position at the last probe trip. This is only valid
   once the probe command has executed to completion. */
func (c Canon_t) GET_EXTERNAL_PROBE_POSITION_C() float64 {
	return _probe_position_c
}

/* returns the x-axis position at the last probe trip. This is only valid
   once the probe command has executed to completion. */
func (c Canon_t) GET_EXTERNAL_PROBE_POSITION_X() float64 {
	return _probe_position_x
}

/* returns the y-axis position at the last probe trip. This is only valid
   once the probe command has executed to completion. */
func (c Canon_t) GET_EXTERNAL_PROBE_POSITION_Y() float64 {
	return _probe_position_y
}

/* returns the z-axis position at the last probe trip. This is only valid
   once the probe command has executed to completion. */
func (c Canon_t) GET_EXTERNAL_PROBE_POSITION_Z() float64 {
	return _probe_position_z
}

/* Returns the value for any analog non-contact probing. */
/* This is a dummy of a dummy, returning a useless value. */
/* It is not expected this will ever be called. */
func (c Canon_t) GET_EXTERNAL_PROBE_VALUE() float64 {
	return 1.0
}

/* Returns zero if queue is not empty, non-zero if the queue is empty */
/* In the stand-alone interpreter, there is no queue, so it is always empty */
func (c Canon_t) GET_EXTERNAL_QUEUE_EMPTY() int {
	return 1
}

/* Returns the system value for spindle speed in rpm */
func (c Canon_t) GET_EXTERNAL_SPEED() float64 {
	return _spindle_speed
}

/* Returns the system value for direction of spindle turning */
func (c Canon_t) GET_EXTERNAL_SPINDLE() inc.CANON_DIRECTION {
	return _spindle_turning
}

/* Returns the system value for the carousel slot in which the tool
   currently in the spindle belongs. Return value zero means there is no
   tool in the spindle. */
func (c Canon_t) GET_EXTERNAL_TOOL_SLOT() int {
	return _active_slot
}

/* Returns maximum number of tools */
func (c Canon_t) GET_EXTERNAL_TOOL_MAX() int {
	return _tool_max
}

/* Returns the CANON_TOOL_TABLE structure associated with the tool
   in the given pocket */
func (c Canon_t) GET_EXTERNAL_TOOL_TABLE(pocket int) inc.CANON_TOOL_TABLE {
	return _tools[pocket]
}

/* Returns the system traverse rate */
func (c Canon_t) GET_EXTERNAL_TRAVERSE_RATE() float64 {
	return _traverse_rate
}

package rs274ngc

import "github.com/flyingyizi/rs274ngc/inc"

type CANON_UNITS int

//	  type Plane int
//	  const (
//		  _ Plane = iota
//		  XY
//		  YZ
//		  XZ
//	  )

type CANON_PLANE int

const (
	_ CANON_PLANE = iota
	CANON_PLANE_XY
	CANON_PLANE_YZ
	CANON_PLANE_XZ
)

const (
	_ CANON_UNITS = iota
	CANON_UNITS_INCHES
	CANON_UNITS_MM
	CANON_UNITS_CM
)

//type Motion int
//const (
//	_ Motion = iota
//	Exact_Stop
//	Exact_Path
//	Continuous
//)

type CANON_MOTION_MODE int

const (
	_ CANON_MOTION_MODE = iota
	CANON_EXACT_STOP
	CANON_EXACT_PATH
	CANON_CONTINUOUS
)

//type Direction int
//const (
//	_ Direction = iota
//	Stop
//	Clockwise
//	CounterClockwise
//)
type CANON_DIRECTION int

const (
	_ CANON_DIRECTION = iota
	CANON_STOPPED
	CANON_CLOCKWISE
	CANON_COUNTERCLOCKWISE
)

//type Tool struct {
//	id       int
//	length   float64
//	diameter float64
//	//Tool();
//}
type CANON_TOOL_TABLE struct {
	id       int
	length   float64
	diameter float64
}

type cannon_i interface {
	//******Miscellaneous Functions

	COMMENT(s []byte)
	DISABLE_FEED_OVERRIDE()
	DISABLE_SPEED_OVERRIDE()
	ENABLE_FEED_OVERRIDE()
	ENABLE_SPEED_OVERRIDE()
	FLOOD_OFF()
	FLOOD_ON()
	INIT_CANON()
	MESSAGE([]byte)
	MIST_OFF()
	MIST_ON()
	PALLET_SHUTTLE()
	//******Miscellaneous Functions  END

	//******Machining 	Attributes
	SELECT_PLANE(plane CANON_PLANE)
	SET_FEED_RATE(rate float64)
	SET_FEED_REFERENCE(reference inc.CANON_FEED_REFERENCE)
	SET_MOTION_CONTROL_MODE(mode CANON_MOTION_MODE)
	START_SPEED_FEED_SYNCH()
	STOP_SPEED_FEED_SYNCH()
	//******Machining 	Attributes  END

	//******Spindle Functions
	ORIENT_SPINDLE(orientation float64, direction CANON_DIRECTION)
	SET_SPINDLE_SPEED(r float64)
	START_SPINDLE_CLOCKWISE()
	START_SPINDLE_COUNTERCLOCKWISE()
	STOP_SPINDLE_TURNING()
	//******Spindle Functions END

	//******Program 	Functions
	OPTIONAL_PROGRAM_STOP()
	PROGRAM_END()
	PROGRAM_STOP()
	//******Program 	Functions END

	//******Tool 	Functions
	CHANGE_TOOL(slot int)
	SELECT_TOOL(i int)
	USE_TOOL_LENGTH_OFFSET(offset float64)
	//******Tool 	Functions END

	//******Machining 	Functions
	ARC_FEED(first_end, second_end, first_axis,
		second_axis float64, rotation int, axis_end_point, a, b, c float64)
	DWELL(seconds float64)
	STRAIGHT_FEED(x, y, z, a, b, c float64)
	//******Machining 	Functions END

	//******Probe 	Functions
	STRAIGHT_PROBE(x, y, z, a, b, c float64)
	//******Probe 	Functions END

	//******Free Space	Motion
	STRAIGHT_TRAVERSE(x, y, z, a, b, c float64)

	USE_LENGTH_UNITS(in_unit CANON_UNITS)
	SET_ORIGIN_OFFSETS(x, y, z, a, b, c float64)

	TURN_PROBE_OFF()
	TURN_PROBE_ON()

	//D.6 World-give-information Functions
	//This section describes the world-give-information functions. These functions get information for
	//the Interpreter. They are arranged alphabetically. All function names start with “GET_EXTERNAL_”.
	//Return the system angular unit factor, in units / degree. The Interpreter is not currently using
	//this function.
	GET_EXTERNAL_ANGLE_UNIT_FACTOR() float64

	//Return the system feed rate.
	GET_EXTERNAL_FEED_RATE() float64

	//Return the system value for flood coolant, zero = off, non-zero = on.
	GET_EXTERNAL_FLOOD() int

	//Return the system length unit factor, in units / mm. The Interpreter is not currently using this
	//function.
	GET_EXTERNAL_LENGTH_UNIT_FACTOR() float64

	//Return the system length unit type.
	GET_EXTERNAL_LENGTH_UNIT_TYPE() CANON_UNITS

	//Return the system value for mist coolant, zero = off, non-zero = on.
	GET_EXTERNAL_MIST() int

	//Return the current path control mode.
	GET_EXTERNAL_MOTION_CONTROL_MODE() CANON_MOTION_MODE

	//The Interpreter is not using these six GET_EXTERNAL_ORIGIN functions, each of which
	//returns the current value of the origin offset for the axis it names.
	//double GET_EXTERNAL_ORIGIN_A();
	//double GET_EXTERNAL_ORIGIN_B();
	//double GET_EXTERNAL_ORIGIN_C();
	//double GET_EXTERNAL_ORIGIN_X();
	//double GET_EXTERNAL_ORIGIN_Y();
	//double GET_EXTERNAL_ORIGIN_Z();

	//Return nothing but copy the name of the parameter file into the filename array, stopping at
	//max_size if the name is longer. An empty string may be placed in filename.
	GET_EXTERNAL_PARAMETER_FILE_NAME(filename []byte, max_size int)

	//Return the currently active plane.
	GET_EXTERNAL_PLANE() CANON_PLANE

	//Each of the six functions above returns the current position for the axis it names.
	GET_EXTERNAL_POSITION_A() float64
	GET_EXTERNAL_POSITION_B() float64
	GET_EXTERNAL_POSITION_C() float64
	GET_EXTERNAL_POSITION_X() float64
	GET_EXTERNAL_POSITION_Y() float64
	GET_EXTERNAL_POSITION_Z() float64

	//Each of the six functions above returns the position at the last probe trip for the axis it names.
	GET_EXTERNAL_PROBE_VALUE() float64
	//Return the value for any analog non-contact probing.
	GET_EXTERNAL_PROBE_POSITION_A() float64
	GET_EXTERNAL_PROBE_POSITION_B() float64
	GET_EXTERNAL_PROBE_POSITION_C() float64
	GET_EXTERNAL_PROBE_POSITION_X() float64
	GET_EXTERNAL_PROBE_POSITION_Y() float64
	GET_EXTERNAL_PROBE_POSITION_Z() float64

	//Return the system value for the spindle speed setting in revolutions per minute (rpm). The
	//actual spindle speed may differ from this.
	GET_EXTERNAL_SPEED() float64
	//Return the system value for direction of spindle turning.
	GET_EXTERNAL_SPINDLE() CANON_DIRECTION

	//Return the current tool length offset.
	GET_EXTERNAL_TOOL_LENGTH_OFFSET() float64
	//returns number of slots in carousel.
	GET_EXTERNAL_TOOL_MAX() int
	//Returns the system value for the carousel slot in which the tool currently in the spindle
	//belongs. Return value zero means there is no tool in the spindle.
	GET_EXTERNAL_TOOL_SLOT() int
	//Returns the CANON_TOOL_TABLE structure associated with the tool in the given pocket. A
	//CANON_TOOL_TABLE structure has three data elements: id (an int), length (a double), and
	//diameter (a double).
	GET_EXTERNAL_TOOL_TABLE(pocket int) CANON_TOOL_TABLE
	//Returns the system traverse rate.
	GET_EXTERNAL_TRAVERSE_RATE() float64

	// Returns zero if queue is not empty, non-zero if the queue is empty
	// This always returns a valid value
	GET_EXTERNAL_QUEUE_EMPTY() int
}

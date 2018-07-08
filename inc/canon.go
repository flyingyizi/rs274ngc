package inc

import (
	"math"
)

/************************************************************************

      Copyright 2008 Mark Pictor

  This file is part of RS274NGC.

  RS274NGC is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  RS274NGC is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with RS274NGC.  If not, see <http://www.gnu.org/licenses/>.

  This software is based on software that was produced by the National
  Institute of Standards and Technology (NIST).

  ************************************************************************/

/* canon.hh

This is the header file that all applications that use the
canonical commands for three- to six-axis machining should include.

Three mutually orthogonal (in a right-handed system) X, Y, and Z axes
are always present. In addition, there may be zero to three rotational
axes: A (parallel to the X-axis), B (parallel to the Y-axis), and C
(parallel to the Z-axis).

In the functions that use rotational axes, the axis value is that of a
wrapped linear axis, in degrees.

It is assumed in these activities that the spindle tip is always at
some location called the "current location," and the controller always
knows where that is. It is also assumed that there is always a
"selected plane" which must be the XY-plane, the YZ-plane, or the
ZX-plane of the machine.

*/

//type Position struct {
//	x float64
//	y float64
//	z float64
//	a float64
//	b float64
//	c float64
//
//	//Position();
//	//Position(double x, double y, double z, double a, double b, double c);
//	//Position(double x, double y, double z);
//	//
//	//Position operator+(const Position& p) const;
//	//Position operator-(const Position& p) const;
//}

type CANON_POSITION struct {
	X float64
	Y float64
	Z float64
	A float64
	B float64
	C float64
}

//Length length of path between start and end points
func (start *CANON_POSITION) Length(end *CANON_POSITION) float64 {
	/* straight line */
	if (start.X != end.X) || (start.Y != end.Y) || (start.Z != end.Z) || ((end.A == start.A) && (end.B == start.B) && (end.C == start.C)) {
		return math.Sqrt(math.Pow(end.X-start.X, 2) + math.Pow(end.Y-start.Y, 2) + math.Pow(end.Z-start.Z, 2))
	} else {
		return math.Sqrt(math.Pow(end.A-start.A, 2) + math.Pow(end.B-start.B, 2) + math.Pow(end.C-start.C, 2))
	}
}

//type SpeedFeedMode int
//const (
//	_ SpeedFeedMode = iota
//	Synched
//	Independent
//)

type CANON_SPEED_FEED_MODE int

const (
	_ CANON_SPEED_FEED_MODE = iota
	CANON_SYNCHED
	CANON_INDEPENDENT
)

//type FeedReference int
//const (
//	_ FeedReference = iota
//	Workpiece
//	XYZ
//)
type CANON_FEED_REFERENCE int

const (
	_ CANON_FEED_REFERENCE = iota
	CANON_WORKPIECE
	CANON_XYZ
)

//type Side int
//
//const (
//	_ Side = iota
//	Right
//	Left
//	Off
//)

type CANON_SIDE int

const (
	_ CANON_SIDE = iota
	CANON_SIDE_RIGHT
	CANON_SIDE_LEFT
	CANON_SIDE_OFF
)

//type Axis int
//
//const (
//	_ Axis = iota
//	X
//	Y
//	Z
//	A
//	B
//	C
//)

type CANON_AXIS int

const (
	_ CANON_AXIS = iota
	CANON_AXIS_X
	CANON_AXIS_Y
	CANON_AXIS_Z
	CANON_AXIS_A
	CANON_AXIS_B
	CANON_AXIS_C
)

/* Tools are numbered 1..CANON_TOOL_MAX, with tool 0 meaning no tool. */
const (
	CANON_TOOL_MAX       = 128 // max size of carousel handled
	CANON_TOOL_ENTRY_LEN = 256 // how long each file line can be

)

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
	Length   float64
	Diameter float64
}

type Canon_i interface {
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
	SET_FEED_REFERENCE(reference CANON_FEED_REFERENCE)
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

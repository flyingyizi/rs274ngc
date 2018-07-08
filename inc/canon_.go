package inc

import "math"

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

/* Initialization */

/* reads world model data into the canonical interface */
func INIT_CANON() {
}

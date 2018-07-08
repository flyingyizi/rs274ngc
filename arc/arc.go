package arc

import "math"

//import "github.com/flyingyizi/rs274ngc/inc"
import "github.com/flyingyizi/rs274ngc/inc"

/*
 * arc.cpp
 *
 *  Created on: 2013-08-27
 *      Author: nicholas
 */

//If simulate  ?: operator
func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

/***********************************************************************/

/* arc_data_comp_ijk

Returned Value: int
If any of the following errors occur, this returns the error code shown.
Otherwise, it returns RS274NGC_OK.
1. The two calculable values of the radius differ by more than
tolerance: NCE_RADIUS_TO_END_OF_ARC_DIFFERS_FROM_RADIUS_TO_START
2. move is not G_2 or G_3: NCE_BUG_CODE_NOT_G2_OR_G3

Side effects:
This finds and sets the values of center_x, center_y, and turn.

Called by: convert_arc_comp1

This finds the center coordinates and number of full or partial turns
counterclockwise of a helical or circular arc in ijk-format in the XY
plane. The center is computed easily from the current point and center
offsets, which are given. It is checked that the end point lies one
tool radius from the arc.

*/

func Arc_data_comp_ijk( /* ARGUMENTS                               */
	move inc.GCodes, /* either G_2 (cw arc) or G_3 (ccw arc)             */
	side inc.CANON_SIDE, /* either RIGHT or LEFT                             */
	tool_radius, /* radius of the tool                               */
	current_x, /* first coordinate of current point                */
	current_y, /* second coordinate of current point               */
	end_x, /* first coordinate of arc end point                */
	end_y, /* second coordinate of arc end point               */
	i_number, /* first coordinate offset of center from current   */
	j_number float64, /* second coordinate offset of center from current  */
	center_x, /* pointer to first coordinate of center of arc     */
	center_y *float64, /* pointer to second coordinate of center of arc    */
	turn *int, /* pointer to number of full or partial circles CCW */
	tolerance float64) inc.STATUS { /* tolerance of differing radii                     */

	*center_x = (current_x + i_number)
	*center_y = (current_y + j_number)
	arc_radius := math.Hypot(i_number, j_number)
	radius2 := math.Hypot((*center_x - end_x), (*center_y - end_y))
	radius2 =
		inc.If(((side == inc.CANON_SIDE_LEFT) && (move == 30)) ||
			((side == inc.CANON_SIDE_RIGHT) && (move == 20)),
			(radius2 - tool_radius), (radius2 + tool_radius)).(float64)
	if math.Abs(arc_radius-radius2) > tolerance {
		return inc.NCE_RADIUS_TO_END_OF_ARC_DIFFERS_FROM_RADIUS_TO_START
	}

	/* This catches an arc too small for the tool, also */
	if move == inc.G_2 {
		*turn = -1
	} else if move == inc.G_3 {
		*turn = 1
	} else {
		return inc.NCE_BUG_CODE_NOT_G2_OR_G3
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* arc_data_comp_r

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The arc radius is too small to reach the end point:
   NCE_RADIUS_TOO_SMALL_TO_REACH_END_POINT
   2. The arc radius is not greater than the tool_radius, but should be:
   NCE_TOOL_RADIUS_NOT_LESS_THAN_ARC_RADIUS_WITH_COMP
   3. An imaginary value for offset would be found, which should never
   happen if the theory is correct: NCE_BUG_IN_TOOL_RADIUS_COMP

   Side effects:
   This finds and sets the values of center_x, center_y, and turn.

   Called by: convert_arc_comp1

   This finds the center coordinates and number of full or partial turns
   counterclockwise of a helical or circular arc (call it arc1) in
   r-format in the XY plane.  Arc2 is constructed so that it is tangent
   to a circle whose radius is tool_radius and whose center is at the
   point (current_x, current_y) and passes through the point (end_x,
   end_y). Arc1 has the same center as arc2. The radius of arc1 is one
   tool radius larger or smaller than the radius of arc2.

   If the value of the big_radius argument is negative, that means [NCMS,
   page 21] that an arc larger than a semicircle is to be made.
   Otherwise, an arc of a semicircle or less is made.

   The algorithm implemented here is to construct a line L from the
   current point to the end point, and a perpendicular to it from the
   center of the arc which intersects L at point P. Since the distance
   from the end point to the center and the distance from the current
   point to the center are known, two equations for the length of the
   perpendicular can be written. The right sides of the equations can be
   set equal to one another and the resulting equation solved for the
   length of the line from the current point to P. Then the location of
   P, the length of the perpendicular, the angle of the perpendicular,
   and the location of the center, can be found in turn.

   This needs to be better documented, with figures. There are eight
   possible arcs, since there are three binary possibilities: (1) tool
   inside or outside arc, (2) clockwise or counterclockwise (3) two
   positions for each arc (of the given radius) tangent to the tool
   outline and through the end point. All eight are calculated below,
   since theta, radius2, and turn may each have two values.

   To see two positions for each arc, imagine the arc is a hoop, the
   tool is a cylindrical pin, and the arc may rotate around the end point.
   The rotation covers all possible positions of the arc. It is easy to
   see the hoop is constrained by the pin at two different angles, whether
   the pin is inside or outside the hoop.

*/

func Arc_data_comp_r( /* ARGUMENTS                                 */
	move inc.GCodes, /* either G_2 (cw arc) or G_3 (ccw arc)             */
	side inc.CANON_SIDE, /* either RIGHT or LEFT                             */
	tool_radius, /* radius of the tool                               */
	current_x, /* first coordinate of current point                */
	current_y, /* second coordinate of current point               */
	end_x, /* first coordinate of arc end point                */
	end_y, /* second coordinate of arc end point               */
	big_radius float64, /* radius of arc                                    */
	center_x, /* pointer to first coordinate of center of arc     */
	center_y *float64, /* pointer to second coordinate of center of arc    */
	turn *int) inc.STATUS { /* pointer to number of full or partial circles CCW */

	//static char name[] = "arc_data_comp_r";
	var (
		abs_radius, /* absolute value of big_radius          */
		alpha, /* direction of line from current to end */
		distance, /* length of line L from current to end  */
		mid_length, /* length from current point to point P  */
		offset, /* length of line from P to center       */
		radius2, /* distance from center to current point */
		mid_x, /* x-value of point P                    */
		mid_y, /* y-value of point P                    */
		theta float64 /* direction of line from P to center    */
	)
	abs_radius = math.Abs(big_radius)
	if (abs_radius <= tool_radius) && (((side == inc.CANON_SIDE_LEFT) && (move == inc.G_3)) ||
		((side == inc.CANON_SIDE_RIGHT) && (move == inc.G_2))) {
		return inc.NCE_TOOL_RADIUS_NOT_LESS_THAN_ARC_RADIUS_WITH_COMP
	}

	distance = math.Hypot((end_x - current_x), (end_y - current_y))
	alpha = math.Atan2((end_y - current_y), (end_x - current_x))
	theta = inc.If(((move == inc.G_3) && (big_radius > 0)) ||
		((move == inc.G_2) && (big_radius < 0)),
		(alpha + inc.PI2), (alpha - inc.PI2)).(float64)
	radius2 = inc.If(((side == inc.CANON_SIDE_LEFT) && (move == inc.G_3)) ||
		((side == inc.CANON_SIDE_RIGHT) && (move == inc.G_2)),
		(abs_radius - tool_radius), (abs_radius + tool_radius)).(float64)

	if distance > (radius2 + abs_radius) {
		return inc.NCE_RADIUS_TOO_SMALL_TO_REACH_END_POINT
	}

	mid_length = (((radius2 * radius2) + (distance * distance) -
		(abs_radius * abs_radius)) / (2.0 * distance))
	mid_x = (current_x + (mid_length * math.Cos(alpha)))
	mid_y = (current_y + (mid_length * math.Sin(alpha)))

	if (radius2 * radius2) <= (mid_length * mid_length) {
		return inc.NCE_BUG_IN_TOOL_RADIUS_COMP
	}

	offset = math.Sqrt((radius2 * radius2) - (mid_length * mid_length))
	*center_x = mid_x + (offset * math.Cos(theta))
	*center_y = mid_y + (offset * math.Sin(theta))
	*turn = inc.If(move == inc.G_2, -1, 1).(int)

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* arc_data_ijk

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The two calculable values of the radius differ by more than
   tolerance: NCE_RADIUS_TO_END_OF_ARC_DIFFERS_FROM_RADIUS_TO_START
   2. The move code is not G_2 or G_3: NCE_BUG_CODE_NOT_G2_OR_G3
   3. Either of the two calculable values of the radius is zero:
   NCE_ZERO_RADIUS_ARC

   Side effects:
   This finds and sets the values of center_x, center_y, and turn.

   Called by:
   convert_arc2
   convert_arc_comp2

   This finds the center coordinates and number of full or partial turns
   counterclockwise of a helical or circular arc in ijk-format. This
   function is used by convert_arc2 for all three planes, so "x" and
   "y" really mean "first_coordinate" and "second_coordinate" wherever
   they are used here as suffixes of variable names. The i and j prefixes
   are handled similarly.

*/

func Arc_data_ijk( /* ARGUMENTS                                       */
	move inc.GCodes, /* either G_2 (cw arc) or G_3 (ccw arc)            */
	current_x, /* first coordinate of current point               */
	current_y, /* second coordinate of current point              */
	end_x, /* first coordinate of arc end point               */
	end_y, /* second coordinate of arc end point              */
	i_number, /* first coordinate offset of center from current  */
	j_number float64, /* second coordinate offset of center from current */
	center_x, /* pointer to first coordinate of center of arc    */
	center_y *float64, /* pointer to second coordinate of center of arc   */
	turn *int, /* pointer to no. of full or partial circles CCW   */
	tolerance float64) inc.STATUS { /* tolerance of differing radii                    */

	//static char name[] = "arc_data_ijk";
	var (
		radius, /* radius to current point */
		radius2 float64 /* radius to end point     */
	)
	*center_x = (current_x + i_number)
	*center_y = (current_y + j_number)
	radius = math.Hypot((*center_x - current_x), (*center_y - current_y))
	radius2 = math.Hypot((*center_x - end_x), (*center_y - end_y))
	if (radius == 0.0) || (radius2 == 0.0) {
		return inc.NCE_ZERO_RADIUS_ARC
	}
	if math.Abs(radius-radius2) > tolerance {
		return inc.NCE_RADIUS_TO_END_OF_ARC_DIFFERS_FROM_RADIUS_TO_START
	}

	if move == inc.G_2 {
		*turn = -1

	} else if move == inc.G_3 {
		*turn = 1

	} else {
		return inc.NCE_BUG_CODE_NOT_G2_OR_G3
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* arc_data_r

   Returned Value: int
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. The radius is too small to reach the end point:
   NCE_ARC_RADIUS_TOO_SMALL_TO_REACH_END_POINT
   2. The current point is the same as the end point of the arc
   (so that it is not possible to locate the center of the circle):
   NCE_CURRENT_POINT_SAME_AS_END_POINT_OF_ARC

   Side effects:
   This finds and sets the values of center_x, center_y, and turn.

   Called by:
   convert_arc2
   convert_arc_comp2

   This finds the center coordinates and number of full or partial turns
   counterclockwise of a helical or circular arc in the r format. This
   function is used by convert_arc2 for all three planes, so "x" and
   "y" really mean "first_coordinate" and "second_coordinate" wherever
   they are used here as suffixes of variable names.

   If the value of the radius argument is negative, that means [NCMS,
   page 21] that an arc larger than a semicircle is to be made.
   Otherwise, an arc of a semicircle or less is made.

   The algorithm used here is based on finding the midpoint M of the line
   L between the current point and the end point of the arc. The center
   of the arc lies on a line through M perpendicular to L.

*/

func Arc_data_r( /* ARGUMENTS                                     */
	move inc.GCodes, /* either G_2 (cw arc) or G_3 (ccw arc)          */
	current_x, /* first coordinate of current point             */
	current_y, /* second coordinate of current point            */
	end_x, /* first coordinate of arc end point             */
	end_y, /* second coordinate of arc end point            */
	radius float64, /* radius of arc                                 */
	center_x, /* pointer to first coordinate of center of arc  */
	center_y *float64, /* pointer to second coordinate of center of arc */
	turn *int) inc.STATUS { /* pointer to no. of full or partial circles CCW */

	//static char name[] = "arc_data_r";
	//double abs_radius;                        /* absolute value of given radius */
	//double half_length;                       /* distance from M to end point   */
	//double mid_x;                             /* first coordinate of M          */
	//double mid_y;                             /* second coordinate of M         */
	var offset float64 /* distance from M to center      */
	var theta float64  /* angle of line from M to center */
	var turn2 float64  /* absolute value of half of turn */

	if (end_x == current_x) && (end_y == current_y) {
		return inc.NCE_CURRENT_POINT_SAME_AS_END_POINT_OF_ARC
	}

	abs_radius := math.Abs(radius)
	mid_x := (end_x + current_x) / 2.0
	mid_y := (end_y + current_y) / 2.0
	half_length := math.Hypot((mid_x - end_x), (mid_y - end_y))

	if (half_length / abs_radius) > (1 + inc.TINY) {
		return inc.NCE_ARC_RADIUS_TOO_SMALL_TO_REACH_END_POINT
	}

	if (half_length / abs_radius) > (1 - inc.TINY) {
		half_length = abs_radius /* allow a small error for semicircle */
	}
	/* check needed before calling asin   */
	if ((move == inc.G_2) && (radius > 0)) ||
		((move == inc.G_3) && (radius < 0)) {
		theta = math.Atan2((end_y-current_y), (end_x-current_x)) - inc.PI2
	} else {
		theta = math.Atan2((end_y-current_y), (end_x-current_x)) + inc.PI2
	}

	turn2 = math.Asin(half_length / abs_radius)
	offset = abs_radius * math.Cos(turn2)

	*center_x = mid_x + (offset * math.Cos(theta))
	*center_y = mid_y + (offset * math.Sin(theta))
	*turn = inc.If(move == inc.G_2, -1, 1).(int)

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* find_arc_length

Returned Value: double (length of path between start and end points)

Side effects: none

Called by:
inverse_time_rate_arc
inverse_time_rate_arc2
inverse_time_rate_as

This calculates the length of the path that will be made relative to
the XYZ axes for a motion in which the X,Y,Z, motion is a circular or
helical arc with its axis parallel to the Z-axis. If tool length
compensation is on, this is the path of the tool tip; if off, the
length of the path of the spindle tip. Any rotary axis motion is
ignored.

If the arc is helical, it is coincident with the hypotenuse of a right
triangle wrapped around a cylinder. If the triangle is unwrapped, its
base is [the radius of the cylinder times the number of radians in the
helix] and its height is [z2 - z1], and the path length can be found
by the Pythagorean theorem.

This is written as though it is only for arcs whose axis is parallel to
the Z-axis, but it will serve also for arcs whose axis is parallel
to the X-axis or Y-axis, with suitable permutation of the arguments.

This works correctly when turn is zero (find_turn returns 0 in that
case).

*/

func Find_arc_length( /* ARGUMENTS                          */
	x1, /* X-coordinate of start point        */
	y1, /* Y-coordinate of start point        */
	z1, /* Z-coordinate of start point        */
	center_x, /* X-coordinate of arc center         */
	center_y float64, /* Y-coordinate of arc center         */
	turn int, /* no. of full or partial circles CCW */
	x2, /* X-coordinate of end point          */
	y2, /* Y-coordinate of end point          */
	z2 float64) float64 { /* Z-coordinate of end point          */

	radius := math.Hypot((center_x - x1), (center_y - y1))

	/* amount of turn of arc in radians */
	theta := find_turn(x1, y1, center_x, center_y, turn, x2, y2)
	if z2 == z1 {
		return (radius * math.Abs(theta))
	} else {
		return math.Hypot((radius * theta), (z2 - z1))
	}
}

/****************************************************************************/

/* find_straight_length

   Returned Value: double (length of path between start and end points)

   Side effects: none

   Called by:
   inverse_time_rate_straight
   inverse_time_rate_as

   This calculates a number to use in feed rate calculations when inverse
   time feed mode is used, for a motion in which X,Y,Z,A,B, and C each change
   linearly or not at all from their initial value to their end value.

   This is used when the feed_reference mode is CANON_XYZ, which is
   always in rs274NGC.

   If any of the X, Y, or Z axes move or the A-axis, B-axis, and C-axis
   do not move, this is the length of the path relative to the XYZ axes
   from the first point to the second, and any rotary axis motion is
   ignored. The length is the simple Euclidean distance.

   The formula for the Euclidean distance "length" of a move involving
   only the A, B and C axes is based on a conversation with Jim Frohardt at
   Boeing, who says that the Fanuc controller on their 5-axis machine
   interprets the feed rate this way. Note that if only one rotary axis
   moves, this formula returns the absolute value of that axis move,
   which is what is desired.

*/

func Find_straight_length( /* ARGUMENTS   */
	x2, /* X-coordinate of end point    */
	y2, /* Y-coordinate of end point    */
	z2 float64, /* Z-coordinate of end point    */
	AA_2, /* A-coordinate of end point    */ /*AA*/
	BB_2, /* B-coordinate of end point    */ /*BB*/
	CC_2 float64, /* C-coordinate of end point    */ /*CC*/
	x1, /* X-coordinate of start point  */
	y1, /* Y-coordinate of start point  */
	z1 float64, /* Z-coordinate of start point  */
	AA_1, /* A-coordinate of start point  */ /*AA*/
	BB_1, /* B-coordinate of start point  */ /*BB*/
	CC_1 float64) float64 { /* C-coordinate of start point  */ /*CC*/

	if (x1 != x2) || (y1 != y2) || (z1 != z2) ||
		((AA_2 == AA_1) && (BB_2 == BB_1) && (CC_2 == CC_1)) { /* straight line */
		return math.Sqrt(math.Pow((x2-x1), 2) + math.Pow((y2-y1), 2) + math.Pow((z2-z1), 2))
	} else {

		return math.Sqrt(math.Pow((AA_2-AA_1), 2) + math.Pow((BB_2-BB_1), 2) + math.Pow((CC_2-CC_1), 2))
	}
}

/****************************************************************************/

/* find_turn

Returned Value: double (angle in radians between two radii of a circle)

Side effects: none

Called by: find_arc_length

All angles are in radians.

*/

func find_turn( /* ARGUMENTS                          */
	x1, /* X-coordinate of start point        */
	y1, /* Y-coordinate of start point        */
	center_x, /* X-coordinate of arc center         */
	center_y float64, /* Y-coordinate of arc center         */
	turn int, /* no. of full or partial circles CCW */
	x2, /* X-coordinate of end point          */
	y2 float64) float64 { /* Y-coordinate of end point          */

	var (
		alpha, /* angle of first radius                      */
		beta, /* angle of second radius                     */
		theta float64 /* amount of turn of arc CCW - negative if CW */
	)
	if turn == 0 {
		return 0.0
	}
	alpha = math.Atan2((y1 - center_y), (x1 - center_x))
	beta = math.Atan2((y2 - center_y), (x2 - center_x))
	if turn > 0 {
		if beta <= alpha {
			beta = (beta + inc.TWO_PI)
		}
		theta = ((beta - alpha) + ((float64)(turn-1) * inc.TWO_PI))
	} else { /* turn < 0 */
		if alpha <= beta {
			alpha = (alpha + inc.TWO_PI)
		}
		theta = ((beta - alpha) + ((float64)(turn+1) * inc.TWO_PI))
	}
	return theta
}

package rs274ngc

import (
	"math"

	"github.com/flyingyizi/rs274ngc/arc"
	"github.com/flyingyizi/rs274ngc/inc"
)

/****************************************************************************/

/* convert_straight

   Returned Value: int
   If convert_straight_comp1 or convert_straight_comp2 is called
   and returns an error code, this returns that code.
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. x, y, z, a, b, and c are all missing from the block:
   NCE_ALL_AXES_MISSING_WITH_G0_OR_G1
   2. The value of move is not G_0 or G_1:
   NCE_BUG_CODE_NOT_G0_OR_G1
   3. A straight feed (g1) move is called with feed rate set to 0:
   NCE_CANNOT_DO_G1_WITH_ZERO_FEED_RATE
   4. A straight feed (g1) move is called with inverse time feed in effect
   but no f word (feed time) is provided:
   NCE_F_WORD_MISSING_WITH_INVERSE_TIME_G1_MOVE
   5. A move is called with G53 and cutter radius compensation on:
   NCE_CANNOT_USE_G53_WITH_CUTTER_RADIUS_COMP

   Side effects:
   This executes a STRAIGHT_FEED command at cutting feed rate
   (if move is G_1) or a STRAIGHT_TRAVERSE command (if move is G_0).
   It also updates the setting of the position of the tool point to the
   end point of the move. If cutter radius compensation is on, it may
   also generate an arc before the straight move. Also, in INVERSE_TIME
   feed mode, SET_FEED_RATE will be called the feed rate setting changed.

   Called by: convert_motion.

   The approach to operating in incremental distance mode (g91) is to
   put the the absolute position values into the block before using the
   block to generate a move.

   In inverse time feed mode, a lower bound of 0.1 is placed on the feed
   rate so that the feed rate is never set to zero. If the destination
   point is the same as the current point, the feed rate would be
   calculated as zero otherwise.

*/

func (cnc *rs274ngc_t) convert_straight( /* ARGUMENTS                                */
	move inc.GCodes) inc.STATUS { /* either G_0 or G_1                        */

	//static char name[] = "convert_straight";
	var (
		end_x, end_y, end_z, AA_end, BB_end, CC_end float64
	)

	var status inc.STATUS

	if move == inc.G_1 {
		if cnc._setup.feed_mode == inc.UNITS_PER_MINUTE {
			if cnc._setup.feed_rate == 0.0 {
				return inc.NCE_CANNOT_DO_G1_WITH_ZERO_FEED_RATE
			}

		} else if cnc._setup.feed_mode == inc.INVERSE_TIME {
			if cnc._setup.block1.f_number == -1.0 {
				return inc.NCE_F_WORD_MISSING_WITH_INVERSE_TIME_G1_MOVE
			}
		}
	}

	cnc._setup.motion_mode = move
	cnc.find_ends(&end_x, &end_y, &end_z, &AA_end, &BB_end, &CC_end)
	/* NOT "== ON" */
	if (cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF) &&
		(cnc._setup.cutter_comp_radius > 0.0) { /* radius always is >= 0 */

		if cnc._setup.block1.g_modes[0] == inc.G_53 {
			return inc.NCE_CANNOT_USE_G53_WITH_CUTTER_RADIUS_COMP
		}

		if cnc._setup.program_x == inc.UNKNOWN {
			status =
				cnc.convert_straight_comp1(move, end_x, end_y,
					end_z, AA_end, BB_end, CC_end)
			if status != inc.RS274NGC_OK {
				return status
			}
			//CHP(status)
		} else {
			status =
				cnc.convert_straight_comp2(move, end_x, end_y,
					end_z, AA_end, BB_end, CC_end)
			//CHP(status)
			if status != inc.RS274NGC_OK {
				return status
			}
		}
	} else if move == inc.G_0 {
		cnc.canon.STRAIGHT_TRAVERSE(end_x, end_y, end_z, AA_end, BB_end, CC_end)
		cnc._setup.current.X = end_x
		cnc._setup.current.Y = end_y
	} else if move == inc.G_1 {
		if cnc._setup.feed_mode == inc.INVERSE_TIME {
			cnc.inverse_time_rate_straight(end_x, end_y, end_z, AA_end, BB_end, CC_end)
		}
		cnc.canon.STRAIGHT_FEED(end_x, end_y, end_z, AA_end, BB_end, CC_end)
		cnc._setup.current.X = end_x
		cnc._setup.current.Y = end_y
	} else {
		return inc.NCE_BUG_CODE_NOT_G0_OR_G1
	}

	cnc._setup.current.Z = end_z
	cnc._setup.current.A = AA_end /*AA*/
	cnc._setup.current.B = BB_end /*BB*/
	cnc._setup.current.C = CC_end /*CC*/
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_straight_comp1

   Returned Value: int
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. The side is not RIGHT or LEFT:
   NCE_BUG_SIDE_NOT_RIGHT_OR_LEFT
   2. The destination tangent point is not more than a tool radius
   away (indicating gouging): NCE_CUTTER_GOUGING_WITH_CUTTER_RADIUS_COMP
   3. The value of move is not G_0 or G_1
   NCE_BUG_CODE_NOT_G0_OR_G1

   Side effects:
   This executes a STRAIGHT_MOVE command at cutting feed rate
   or a STRAIGHT_TRAVERSE command.
   It also updates the setting of the position of the tool point
   to the end point of the move and updates the programmed point.
   If INVERSE_TIME feed rate mode is in effect, it resets the feed rate.

   Called by: convert_straight.

   This is called if cutter radius compensation is on and cnc._setup.program_x
   is UNKNOWN, indicating that this is the first move after cutter radius
   compensation is turned on.

   The algorithm used here for determining the path is to draw a straight
   line from the destination point which is tangent to a circle whose
   center is at the current point and whose radius is the radius of the
   cutter. The destination point of the cutter tip is then found as the
   center of a circle of the same radius tangent to the tangent line at
   the destination point.

*/

func (cnc *rs274ngc_t) convert_straight_comp1( /* ARGUMENTS                       */
	move inc.GCodes, /* either G_0 or G_1                         */
	px, /* X coordinate of end point                 */
	py, /* Y coordinate of end point                 */
	end_z, /* Z coordinate of end point                 */
	AA_end, /* A coordinate of end point           */ /*AA*/
	BB_end, /* B coordinate of end point           */ /*BB*/
	CC_end float64) inc.STATUS { /* C coordinate of end point           */ /*CC*/

	//static char name[] = "convert_straight_comp1";
	side := cnc._setup.cutter_comp_side
	cx := cnc._setup.current.X /* first current point x then end point x */
	cy := cnc._setup.current.Y /* first current point y then end point y */

	/* always will be positive */
	radius := cnc._setup.cutter_comp_radius
	distance := math.Hypot((px - cx), (py - cy))

	if (side != inc.CANON_SIDE_LEFT) && (side != inc.CANON_SIDE_RIGHT) {
		return inc.NCE_BUG_SIDE_NOT_RIGHT_OR_LEFT
	}
	if distance <= radius {
		return inc.NCE_CUTTER_GOUGING_WITH_CUTTER_RADIUS_COMP
	}

	theta := math.Acos(radius / distance)
	alpha := inc.If(side == inc.CANON_SIDE_LEFT, (math.Atan2((cy-py), (cx-px)) - theta), (math.Atan2((cy-py), (cx-px)) + theta)).(float64)
	cx = (px + (radius * math.Cos(alpha))) /* reset to end location */
	cy = (py + (radius * math.Sin(alpha)))
	if move == inc.G_0 {
		cnc.canon.STRAIGHT_TRAVERSE(cx, cy, end_z, AA_end, BB_end, CC_end)
	} else if move == inc.G_1 {
		if cnc._setup.feed_mode == inc.INVERSE_TIME {
			cnc.inverse_time_rate_straight(cx, cy, end_z, AA_end, BB_end, CC_end)
		}
		cnc.canon.STRAIGHT_FEED(cx, cy, end_z, AA_end, BB_end, CC_end)
	} else {
		return inc.NCE_BUG_CODE_NOT_G0_OR_G1
	}

	cnc._setup.current.X = cx
	cnc._setup.current.Y = cy
	cnc._setup.program_x = px
	cnc._setup.program_y = py
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_straight_comp2

   Returned Value: int
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. The compensation side is not RIGHT or LEFT:
   NCE_BUG_SIDE_NOT_RIGHT_OR_LEFT
   2. A concave corner is found:
   NCE_CONCAVE_CORNER_WITH_CUTTER_RADIUS_COMP

   Side effects:
   This executes a STRAIGHT_FEED command at cutting feed rate
   or a STRAIGHT_TRAVERSE command.
   It also generates an ARC_FEED to go around a corner, if necessary.
   It also updates the setting of the position of the tool point to
   the end point of the move and updates the programmed point.
   If INVERSE_TIME feed mode is in effect, it also calls SET_FEED_RATE
   and resets the feed rate in the machine model.

   Called by: convert_straight.

   This is called if cutter radius compensation is on and
   cnc._setup.program_x is not UNKNOWN, indicating that this is not the
   first move after cutter radius compensation is turned on.

   The algorithm used here is:
   1. Determine the direction of the last motion. This is done by finding
   the direction of the line from the last programmed point to the
   current tool tip location. This line is a radius of the tool and is
   perpendicular to the direction of motion since the cutter is tangent
   to that direction.
   2. Determine the direction of the programmed motion.
   3. If there is a convex corner, insert an arc to go around the corner.
   4. Find the destination point for the tool tip. The tool will be
   tangent to the line from the last programmed point to the present
   programmed point at the present programmed point.
   5. Go in a straight line from the current tool tip location to the
   destination tool tip location.

   This uses an angle tolerance of TOLERANCE_CONCAVE_CORNER (0.01 radian)
   to determine if:
   1) an illegal concave corner exists (tool will not fit into corner),
   2) no arc is required to go around the corner (i.e. the current line
   is in the same direction as the end of the previous move), or
   3) an arc is required to go around a convex corner and start off in
   a new direction.

   If a rotary axis is moved in this block and an extra arc is required
   to go around a sharp corner, all the rotary axis motion occurs on the
   arc.  An alternative might be to distribute the rotary axis motion
   over the arc and the straight move in proportion to their lengths.

   If the Z-axis is moved in this block and an extra arc is required to
   go around a sharp corner, all the Z-axis motion occurs on the straight
   line and none on the extra arc.  An alternative might be to distribute
   the Z-axis motion over the extra arc and the straight line in
   proportion to their lengths.

   This handles inverse time feed rates by computing the length of the
   compensated path.

   This handles the case of there being no XY motion.

   This handles G0 moves. Where an arc is inserted to round a corner in a
   G1 move, no arc is inserted for a G0 move; a STRAIGHT_TRAVERSE is made
   from the current point to the end point. The end point for a G0
   move is the same as the end point for a G1 move, however.

*/

func (cnc *rs274ngc_t) convert_straight_comp2( /* ARGUMENTS                       */
	move inc.GCodes, /* either G_0 or G_1                         */
	px, /* X coordinate of programmed end point      */
	py, /* Y coordinate of programmed end point      */
	end_z, /* Z coordinate of end point                 */
	AA_end, /* A coordinate of end point           */ /*AA*/
	BB_end, /* B coordinate of end point           */ /*BB*/
	CC_end float64) inc.STATUS { /* C coordinate of end point           */ /*CC*/

	/* radians, testing corners */
	small := inc.TOLERANCE_CONCAVE_CORNER
	var (
		theta,
		alpha,
		beta,
		end_x, /* x-coordinate of actual end point */
		end_y, /* y-coordinate of actual end point */
		gamma,
		mid_x, /* x-coordinate of end of added arc, if needed */
		mid_y float64 /* y-coordinate of end of added arc, if needed */
	)

	start_x := cnc._setup.program_x /* programmed beginning point */
	start_y := cnc._setup.program_y
	if (py == start_y) && (px == start_x) { /* no XY motion */
		end_x = cnc._setup.current.X
		end_y = cnc._setup.current.Y
		if move == inc.G_0 {
			cnc.canon.STRAIGHT_TRAVERSE(end_x, end_y, end_z, AA_end, BB_end, CC_end)

		} else if move == inc.G_1 {
			if cnc._setup.feed_mode == inc.INVERSE_TIME {
				cnc.inverse_time_rate_straight(end_x, end_y, end_z, AA_end, BB_end, CC_end)
			}
			cnc.canon.STRAIGHT_FEED(end_x, end_y, end_z, AA_end, BB_end, CC_end)
		} else {
			return inc.NCE_BUG_CODE_NOT_G0_OR_G1
		}
	} else {
		side := cnc._setup.cutter_comp_side
		/* will always be positive */
		radius := cnc._setup.cutter_comp_radius
		theta = math.Atan2(cnc._setup.current.Y-start_y,
			cnc._setup.current.X-start_x)
		alpha = math.Atan2(py-start_y, px-start_x)

		if side == inc.CANON_SIDE_LEFT {
			if theta < alpha {
				theta = (theta + inc.TWO_PI)
			}
			beta = ((theta - alpha) - inc.PI2)
			gamma = inc.PI2
		} else if side == inc.CANON_SIDE_RIGHT {
			if alpha < theta {
				alpha = (alpha + inc.TWO_PI)
			}
			beta = ((alpha - theta) - inc.PI2)
			gamma = -inc.PI2
		} else {
			return inc.NCE_BUG_SIDE_NOT_RIGHT_OR_LEFT
		}

		end_x = (px + (radius * math.Cos(alpha+gamma)))
		end_y = (py + (radius * math.Sin(alpha+gamma)))
		mid_x = (start_x + (radius * math.Cos(alpha+gamma)))
		mid_y = (start_y + (radius * math.Sin(alpha+gamma)))

		if (beta < -small) || (beta > (inc.PI + small)) {
			return inc.NCE_CONCAVE_CORNER_WITH_CUTTER_RADIUS_COMP
		}

		if move == inc.G_0 {
			cnc.canon.STRAIGHT_TRAVERSE(end_x, end_y, end_z, AA_end, BB_end, CC_end)
		} else if move == inc.G_1 {
			if beta > small { /* ARC NEEDED */
				if cnc._setup.feed_mode == inc.INVERSE_TIME {
					if side == inc.CANON_SIDE_LEFT {
						cnc.inverse_time_rate_as(start_x, start_y,
							-1, mid_x, mid_y, end_x, end_y, end_z, AA_end, BB_end, CC_end)
					} else {
						cnc.inverse_time_rate_as(start_x, start_y,
							1, mid_x, mid_y, end_x, end_y, end_z, AA_end, BB_end, CC_end)
					}
				}
				if side == inc.CANON_SIDE_LEFT {
					cnc.canon.ARC_FEED(mid_x, mid_y, start_x, start_y, -1,
						cnc._setup.current.Z, AA_end, BB_end, CC_end)
				} else {
					cnc.canon.ARC_FEED(mid_x, mid_y, start_x, start_y, 1,
						cnc._setup.current.Z, AA_end, BB_end, CC_end)
				}

				cnc.canon.STRAIGHT_FEED(end_x, end_y, end_z, AA_end, BB_end, CC_end)
			} else {
				if cnc._setup.feed_mode == inc.INVERSE_TIME {
					cnc.inverse_time_rate_straight(end_x, end_y, end_z, AA_end, BB_end, CC_end)
				}
				cnc.canon.STRAIGHT_FEED(end_x, end_y, end_z, AA_end, BB_end, CC_end)
			}
		} else {
			return inc.NCE_BUG_CODE_NOT_G0_OR_G1
		}
	}

	cnc._setup.current.X = end_x
	cnc._setup.current.Y = end_y
	cnc._setup.program_x = px
	cnc._setup.program_y = py
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* inverse_time_rate_straight

   Returned Value: int (RS274NGC_OK)

   Side effects: a call is made to SET_FEED_RATE and _setup.feed_rate is set.

   Called by:
   convert_straight
   convert_straight_comp1
   convert_straight_comp2

   This finds the feed rate needed by an inverse time straight move. Most
   of the work here is in finding the length of the line.

*/

func (cnc *rs274ngc_t) inverse_time_rate_straight( /* ARGUMENTS                    */
	end_x, /* x coordinate of end point of straight line */
	end_y, /* y coordinate of end point of straight line */
	end_z, /* z coordinate of end point of straight line */
	AA_end, /* A coordinate of end point of straight line */ /*AA*/
	BB_end, /* B coordinate of end point of straight line */ /*BB*/
	CC_end float64) inc.STATUS { /* C coordinate of end point of straight line */ /*CC*/

	//static char name[] = "inverse_time_rate_straight";

	length := arc.Find_straight_length(end_x, end_y, end_z, AA_end, BB_end, CC_end, cnc._setup.current.X,
		cnc._setup.current.Y, cnc._setup.current.Z, cnc._setup.current.A,
		cnc._setup.current.B, cnc._setup.current.C)

	rate := math.Max(0.1, (length * cnc._setup.block1.f_number))
	cnc.canon.SET_FEED_RATE(rate)
	cnc._setup.feed_rate = rate

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* inverse_time_rate_as

   Returned Value: int (RS274NGC_OK)

   Side effects: a call is made to SET_FEED_RATE and _setup.feed_rate is set.

   Called by: convert_straight_comp2

   This finds the feed rate needed by an inverse time move in
   convert_straight_comp2. The move consists of an extra arc and a straight
   line. Most of the work here is in finding the lengths of the arc and
   the line.

   All rotary motion is assumed to occur on the arc, as done by
   convert_straight_comp2.

   All z motion is assumed to occur on the line, as done by
   convert_straight_comp2.

*/

func (cnc *rs274ngc_t) inverse_time_rate_as( /* ARGUMENTS */
	start_x, /* x coord of last program point, extra arc center x */
	start_y float64, /* y coord of last program point, extra arc center y */
	turn int, /* turn of extra arc                                 */
	mid_x, /* x coord of end point of extra arc                 */
	mid_y, /* y coord of end point of extra arc                 */
	end_x, /* x coord of end point of straight line             */
	end_y, /* y coord of end point of straight line             */
	end_z float64, /* z coord of end point of straight line             */
	AA_end, /* A coord of end point of straight line       */ /*AA*/
	BB_end, /* B coord of end point of straight line       */ /*BB*/
	CC_end float64) inc.STATUS { /* C coord of end point of straight line       */ /*CC*/

	length := (arc.Find_arc_length(cnc._setup.current.X, cnc._setup.current.Y,
		cnc._setup.current.Z, start_x, start_y,
		turn, mid_x, mid_y, cnc._setup.current.Z) +
		arc.Find_straight_length(end_x, end_y,
			end_z, AA_end, BB_end, CC_end, mid_x, mid_y,
			cnc._setup.current.Z, AA_end, BB_end, CC_end))
	rate := math.Max(0.1, (length * cnc._setup.block1.f_number))
	cnc.canon.SET_FEED_RATE(rate)
	cnc._setup.feed_rate = rate

	return inc.RS274NGC_OK
}

package rs274ngc

import (
	"math"

	"github.com/flyingyizi/rs274ngc/arc"
	"github.com/flyingyizi/rs274ngc/inc"
)

/****************************************************************************/

/* convert_arc

   Returned Value: int
   If one of the following functions returns an error code,
   this returns that error code.
   convert_arc_comp1
   convert_arc_comp2
   convert_arc2
   If any of the following errors occur, this returns the error code shown.
   Otherwise, this returns RS274NGC_OK.
   1. The block has neither an r value nor any i,j,k values:
   NCE_R_I_J_K_WORDS_ALL_MISSING_FOR_ARC
   2. The block has both an r value and one or more i,j,k values:
   NCE_MIXED_RADIUS_IJK_FORMAT_FOR_ARC
   3. In the ijk format the XY-plane is selected and
   the block has a k value: NCE_K_WORD_GIVEN_FOR_ARC_IN_XY_PLANE
   4. In the ijk format the YZ-plane is selected and
   the block has an i value: NCE_I_WORD_GIVEN_FOR_ARC_IN_YZ_PLANE
   5. In the ijk format the XZ-plane is selected and
   the block has a j value: NCE_J_WORD_GIVEN_FOR_ARC_IN_XZ_PLANE
   6. In either format any of the following occurs.
   a. The XY-plane is selected and the block has no x or y value:
   NCE_X_AND_Y_WORDS_MISSING_FOR_ARC_IN_XY_PLANE
   b. The YZ-plane is selected and the block has no y or z value:
   NCE_Y_AND_Z_WORDS_MISSING_FOR_ARC_IN_YZ_PLANE
   c. The ZX-plane is selected and the block has no z or x value:
   NCE_X_AND_Z_WORDS_MISSING_FOR_ARC_IN_XZ_PLANE
   7. The selected plane is an unknown plane:
   NCE_BUG_PLANE_NOT_XY_YZ__OR_XZ
   8. The feed rate mode is UNITS_PER_MINUTE and feed rate is zero:
   NCE_CANNOT_MAKE_ARC_WITH_ZERO_FEED_RATE
   9. The feed rate mode is INVERSE_TIME and the block has no f word:
   NCE_F_WORD_MISSING_WITH_INVERSE_TIME_ARC_MOVE

   Side effects:
   This generates and executes an arc command at feed rate
   (and, possibly a second arc command). It also updates the setting
   of the position of the tool point to the end point of the move.

   Called by: convert_motion.

   This converts a helical or circular arc.  The function calls:
   convert_arc2 (when cutter radius compensation is off) or
   convert_arc_comp1 (when cutter comp is on and this is the first move) or
   convert_arc_comp2 (when cutter comp is on and this is not the first move).

   If the ijk format is used, at least one of the offsets in the current
   plane must be given in the block; it is common but not required to
   give both offsets. The offsets are always incremental [NCMS, page 21].

*/

func (cnc *rs274ngc_t) convert_arc( /* ARGUMENTS                                */
	move inc.GCodes) inc.STATUS { /* either G_2 (cw arc) or G_3 (ccw arc)     */

	//static char name[] = "convert_arc";
	var (
		end_x, end_y, end_z, AA_end, BB_end, CC_end float64
		status                                      inc.STATUS
	)

	/* flag set ON if any of i,j,k present in NC code  */
	ijk_flag :=
		inc.If(((cnc._setup.block1.i_flag == ON || cnc._setup.block1.j_flag == ON) || cnc._setup.block1.k_flag == ON),
			ON, OFF).(ON_OFF)

	/* flag set ON if this is first move after comp ON */
	first := (cnc._setup.program_x == inc.UNKNOWN)

	if (cnc._setup.block1.r_flag != ON) && (ijk_flag != ON) {
		return inc.NCE_R_I_J_K_WORDS_ALL_MISSING_FOR_ARC
	}

	if (cnc._setup.block1.r_flag == ON) && (ijk_flag == ON) {
		return inc.NCE_MIXED_RADIUS_IJK_FORMAT_FOR_ARC
	}

	if cnc._setup.feed_mode == inc.UNITS_PER_MINUTE {
		if cnc._setup.feed_rate == 0.0 {
			return inc.NCE_CANNOT_MAKE_ARC_WITH_ZERO_FEED_RATE
		}

	} else if cnc._setup.feed_mode == inc.INVERSE_TIME {
		if cnc._setup.block1.f_number == -1.0 {
			return inc.NCE_F_WORD_MISSING_WITH_INVERSE_TIME_ARC_MOVE
		}
	}
	if ijk_flag {
		if cnc._setup.plane == inc.CANON_PLANE_XY {
			if cnc._setup.block1.k_flag {
				return inc.NCE_K_WORD_GIVEN_FOR_ARC_IN_XY_PLANE
			}
			if cnc._setup.block1.i_flag == OFF { /* i or j flag on to get here */
				cnc._setup.block1.i_number = 0.0
			} else if cnc._setup.block1.j_flag == OFF {
				cnc._setup.block1.j_number = 0.0
			}
		} else if cnc._setup.plane == inc.CANON_PLANE_YZ {
			if cnc._setup.block1.i_flag {
				return inc.NCE_I_WORD_GIVEN_FOR_ARC_IN_YZ_PLANE
			}
			if cnc._setup.block1.j_flag == OFF { /* j or k flag on to get here */
				cnc._setup.block1.j_number = 0.0
			} else if cnc._setup.block1.k_flag == OFF {
				cnc._setup.block1.k_number = 0.0

			}
		} else if cnc._setup.plane == inc.CANON_PLANE_XZ {
			if cnc._setup.block1.j_flag {
				return inc.NCE_J_WORD_GIVEN_FOR_ARC_IN_XZ_PLANE
			}
			if cnc._setup.block1.i_flag == OFF { /* i or k flag on to get here */
				cnc._setup.block1.i_number = 0.0
			} else if cnc._setup.block1.k_flag == OFF {
				cnc._setup.block1.k_number = 0.0
			}
		} else {
			return inc.NCE_BUG_PLANE_NOT_XY_YZ_OR_XZ
		}
	} else {

	} /* r format arc; no other checks needed specific to this format */

	if cnc._setup.plane == inc.CANON_PLANE_XY { /* checks for both formats */
		if (cnc._setup.block1.x_flag == OFF) && (cnc._setup.block1.y_flag == OFF) {
			return inc.NCE_X_AND_Y_WORDS_MISSING_FOR_ARC_IN_XY_PLANE
		}

	} else if cnc._setup.plane == inc.CANON_PLANE_YZ {
		if (cnc._setup.block1.y_flag == OFF) && (cnc._setup.block1.z_flag == OFF) {
			return inc.NCE_Y_AND_Z_WORDS_MISSING_FOR_ARC_IN_YZ_PLANE
		}

	} else if cnc._setup.plane == inc.CANON_PLANE_XZ {
		if (cnc._setup.block1.x_flag == OFF) && (cnc._setup.block1.z_flag == OFF) {
			return inc.NCE_X_AND_Z_WORDS_MISSING_FOR_ARC_IN_XZ_PLANE
		}
	}

	cnc.find_ends(&end_x, &end_y, &end_z, &AA_end, &BB_end, &CC_end)
	cnc._setup.motion_mode = move

	if cnc._setup.plane == inc.CANON_PLANE_XY {
		if (cnc._setup.cutter_comp_side == inc.CANON_SIDE_OFF) ||
			(cnc._setup.cutter_comp_radius == 0.0) {
			status =
				cnc.convert_arc2(move,
					&(cnc._setup.current.X), &(cnc._setup.current.Y),
					&(cnc._setup.current.Z), end_x, end_y,
					end_z, AA_end, BB_end, CC_end, cnc._setup.block1.i_number,
					cnc._setup.block1.j_number)
			//CHP(status)
			if status != inc.RS274NGC_OK {
				return status
			}

		} else if first {
			status =
				cnc.convert_arc_comp1(move, end_x, end_y,
					end_z, AA_end, BB_end, CC_end)
			//CHP(status)
			if status != inc.RS274NGC_OK {
				return status
			}

		} else {
			status =
				cnc.convert_arc_comp2(move, end_x, end_y,
					end_z, AA_end, BB_end, CC_end)
			//CHP(status)
			if status != inc.RS274NGC_OK {
				return status
			}

		}
	} else if cnc._setup.plane == inc.CANON_PLANE_XZ {
		status =
			cnc.convert_arc2(move,
				&(cnc._setup.current.Z), &(cnc._setup.current.X),
				&(cnc._setup.current.Y), end_z, end_x,
				end_y, AA_end, BB_end, CC_end, cnc._setup.block1.k_number,
				cnc._setup.block1.i_number)
		//CHP(status)
		if status != inc.RS274NGC_OK {
			return status
		}

	} else if cnc._setup.plane == inc.CANON_PLANE_YZ {
		status =
			cnc.convert_arc2(move,
				&(cnc._setup.current.Y), &(cnc._setup.current.Z),
				&(cnc._setup.current.X), end_y, end_z,
				end_x, AA_end, BB_end, CC_end, cnc._setup.block1.j_number, cnc._setup.block1.k_number)
		//CHP(status)
		if status != inc.RS274NGC_OK {
			return status
		}

	} else {
		return inc.NCE_BUG_PLANE_NOT_XY_YZ_OR_XZ
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_arc_comp1

   Returned Value: int
   If arc_data_comp_ijk or arc_data_comp_r returns an error code,
   this returns that code.
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   This executes an arc command at
   feed rate. It also updates the setting of the position of
   the tool point to the end point of the move.

   Called by: convert_arc.

   This function converts a helical or circular arc, generating only one
   arc. The axis must be parallel to the z-axis. This is called when
   cutter radius compensation is on and this is the first cut after the
   turning on.

   The arc which is generated is derived from a second arc which passes
   through the programmed end point and is tangent to the cutter at its
   current location. The generated arc moves the tool so that it stays
   tangent to the second arc throughout the move.

*/

func (cnc *rs274ngc_t) convert_arc_comp1( /* ARGUMENTS                                   */
	move inc.GCodes, /* either G_2 (cw arc) or G_3 (ccw arc)             */
	end_x, /* x-value at end of programmed (then actual) arc   */
	end_y, /* y-value at end of programmed (then actual) arc   */
	end_z, /* z-value at end of arc                            */
	AA_end, /* a-value at end of arc                      */ /*AA*/
	BB_end, /* b-value at end of arc                      */ /*BB*/
	CC_end float64) inc.STATUS { /* c-value at end of arc                      */ /*CC*/

	var turn int /* 1 for counterclockwise, -1 for clockwise */

	var (
		center_x, center_y float64
	)

	/* offset side - right or left              */
	side := cnc._setup.cutter_comp_side
	/* always is positive */
	tool_radius := cnc._setup.cutter_comp_radius
	/* tolerance for difference of radii        */
	tolerance := inc.If(cnc._setup.length_units == inc.CANON_UNITS_INCHES, inc.TOLERANCE_INCH, inc.TOLERANCE_MM).(float64)

	if math.Hypot((end_x-cnc._setup.current.X),
		(end_y-cnc._setup.current.Y)) <= tool_radius {
		return inc.NCE_CUTTER_GOUGING_WITH_CUTTER_RADIUS_COMP
	}

	if cnc._setup.block1.r_flag {
		arc.Arc_data_comp_r(move, side, tool_radius, cnc._setup.current.X,
			cnc._setup.current.Y, end_x, end_y, cnc._setup.block1.r_number,
			&center_x, &center_y, &turn)
	} else {
		arc.Arc_data_comp_ijk(move, side, tool_radius, cnc._setup.current.X,
			cnc._setup.current.Y, end_x, end_y,
			cnc._setup.block1.i_number, cnc._setup.block1.j_number,
			&center_x, &center_y, &turn, tolerance)
	}
	gamma :=
		inc.If(((side == inc.CANON_SIDE_LEFT) && (move == inc.G_3)) || ((side == inc.CANON_SIDE_RIGHT) && (move == inc.G_2)),
			math.Atan2((center_y-end_y), (center_x-end_x)),
			math.Atan2((end_y-center_y), (end_x-center_x))).(float64)

	cnc._setup.program_x = end_x
	cnc._setup.program_y = end_y
	/* end_x reset actual */
	end_x = (end_x + (tool_radius * math.Cos(gamma)))
	/* end_y reset actual */
	end_y = (end_y + (tool_radius * math.Sin(gamma)))

	if cnc._setup.feed_mode == inc.INVERSE_TIME {
		cnc.inverse_time_rate_arc(cnc._setup.current.X, cnc._setup.current.Y,
			cnc._setup.current.Z, center_x, center_y, turn,
			end_x, end_y, end_z)
	}

	cnc.canon.ARC_FEED(end_x, end_y, center_x, center_y, turn, end_z, AA_end, BB_end, CC_end)
	cnc._setup.current.X = end_x
	cnc._setup.current.Y = end_y
	cnc._setup.current.Z = end_z

	cnc._setup.current.A = AA_end /*AA*/
	cnc._setup.current.B = BB_end /*BB*/
	cnc._setup.current.C = CC_end /*CC*/

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_arc_comp2

   Returned Value: int
   If arc_data_ijk or arc_data_r returns an error code,
   this returns that code.
   If any of the following errors occurs, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. A concave corner is found: NCE_CONCAVE_CORNER_WITH_CUTTER_RADIUS_COMP
   2. The tool will not fit inside an arc:
   NCE_TOOL_RADIUS_NOT_LESS_THAN_ARC_RADIUS_WITH_COMP

   Side effects:
   This executes an arc command feed rate. If needed, at also generates
   an arc to go around a convex corner. It also updates the setting of
   the position of the tool point to the end point of the move. If
   inverse time feed rate mode is in effect, the feed rate is reset.

   Called by: convert_arc.

   This function converts a helical or circular arc. The axis must be
   parallel to the z-axis. This is called when cutter radius compensation
   is on and this is not the first cut after the turning on.

   If one or more rotary axes is moved in this block and an extra arc is
   required to go around a sharp corner, all the rotary axis motion
   occurs on the main arc and none on the extra arc.  An alternative
   might be to distribute the rotary axis motion over the extra arc and
   the programmed arc in proportion to their lengths.

   If the Z-axis is moved in this block and an extra arc is required to
   go around a sharp corner, all the Z-axis motion occurs on the main arc
   and none on the extra arc.  An alternative might be to distribute the
   Z-axis motion over the extra arc and the main arc in proportion to
   their lengths.

*/

func (cnc *rs274ngc_t) convert_arc_comp2( /* ARGUMENTS                                 */
	move inc.GCodes, /* either G_2 (cw arc) or G_3 (ccw arc)           */
	end_x, /* x-value at end of programmed (then actual) arc */
	end_y, /* y-value at end of programmed (then actual) arc */
	end_z, /* z-value at end of arc                          */
	AA_end, /* a-value at end of arc                    */ /*AA*/
	BB_end, /* b-value at end of arc                    */ /*BB*/
	CC_end float64) inc.STATUS { /* c-value at end of arc                    */ /*CC*/

	//static char name[] = "convert_arc_comp2";
	var (
		alpha, /* direction of tangent to start of arc */
		arc_radius,
		beta, /* angle between two tangents above */
		center_x, /* center of arc */
		center_y,
		delta, /* direction of radius from start of arc to center of arc */
		gamma, /* direction of perpendicular to arc at end */
		mid_x,
		mid_y float64

		/* angle for testing corners */
		small = inc.TOLERANCE_CONCAVE_CORNER

		turn int
	)
	/* find basic arc data: center_x, center_y, and turn */

	start_x := cnc._setup.program_x
	start_y := cnc._setup.program_y

	tolerance := inc.If(cnc._setup.length_units == inc.CANON_UNITS_INCHES,
		inc.TOLERANCE_INCH, inc.TOLERANCE_MM).(float64)

	if cnc._setup.block1.r_flag {
		arc.Arc_data_r(move, start_x, start_y, end_x, end_y, cnc._setup.block1.r_number, &center_x, &center_y, &turn)
	} else {
		arc.Arc_data_ijk(move, start_x, start_y, end_x, end_y,
			cnc._setup.block1.i_number, cnc._setup.block1.j_number, &center_x, &center_y, &turn, tolerance)
	}

	/* compute other data */
	side := cnc._setup.cutter_comp_side
	/* always is positive */
	tool_radius := cnc._setup.cutter_comp_radius
	arc_radius = math.Hypot((center_x - end_x), (center_y - end_y))
	theta := math.Atan2(cnc._setup.current.Y-start_y, cnc._setup.current.X-start_x)
	theta = inc.If(side == inc.CANON_SIDE_LEFT, (theta - inc.PI2), (theta + inc.PI2)).(float64)
	delta = math.Atan2(center_y-start_y, center_x-start_x)
	alpha = inc.If(move == inc.G_3, (delta - inc.PI2), (delta + inc.PI2)).(float64)
	beta = inc.If(side == inc.CANON_SIDE_LEFT, (theta - alpha), (alpha - theta)).(float64)
	beta = inc.If(beta > (1.5*inc.PI), (beta - inc.TWO_PI),
		inc.If(beta < -inc.PI2, (beta+inc.TWO_PI), beta).(float64)).(float64)

	if ((side == inc.CANON_SIDE_LEFT) && (move == inc.G_3)) ||
		((side == inc.CANON_SIDE_RIGHT) && (move == inc.G_2)) {
		gamma = math.Atan2((center_y - end_y), (center_x - end_x))
		if arc_radius <= tool_radius {
			return inc.NCE_TOOL_RADIUS_NOT_LESS_THAN_ARC_RADIUS_WITH_COMP
		}
	} else {
		gamma = math.Atan2((end_y - center_y), (end_x - center_x))
		delta = (delta + inc.PI)
	}

	cnc._setup.program_x = end_x
	cnc._setup.program_y = end_y
	/* end_x reset actual */
	end_x = (end_x + (tool_radius * math.Cos(gamma)))
	/* end_y reset actual */
	end_y = (end_y + (tool_radius * math.Sin(gamma)))

	/* check if extra arc needed and insert if so */

	if (beta < -small) || (beta > (inc.PI + small)) {
		return inc.NCE_CONCAVE_CORNER_WITH_CUTTER_RADIUS_COMP
	}

	if beta > small { /* two arcs needed */
		mid_x = (start_x + (tool_radius * math.Cos(delta)))
		mid_y = (start_y + (tool_radius * math.Sin(delta)))
		if cnc._setup.feed_mode == inc.INVERSE_TIME {
			if side == inc.CANON_SIDE_LEFT {
				cnc.inverse_time_rate_arc2(start_x, start_y, -1,
					mid_x, mid_y, center_x, center_y, turn,
					end_x, end_y, end_z)
			} else {
				cnc.inverse_time_rate_arc2(start_x, start_y, 1,
					mid_x, mid_y, center_x, center_y, turn,
					end_x, end_y, end_z)
			}
		}

		if side == inc.CANON_SIDE_LEFT {
			cnc.canon.ARC_FEED(mid_x, mid_y, start_x, start_y, -1,
				cnc._setup.current.Z, AA_end, BB_end, CC_end)
		} else {
			cnc.canon.ARC_FEED(mid_x, mid_y, start_x, start_y, 1, cnc._setup.current.Z, AA_end, BB_end, CC_end)
		}
		cnc.canon.ARC_FEED(end_x, end_y, center_x, center_y, turn, end_z, AA_end, BB_end, CC_end)
	} else { /* one arc needed */

		if cnc._setup.feed_mode == inc.INVERSE_TIME {
			cnc.inverse_time_rate_arc(cnc._setup.current.X, cnc._setup.current.Y,
				cnc._setup.current.Z, center_x, center_y, turn,
				end_x, end_y, end_z)
		}
		cnc.canon.ARC_FEED(end_x, end_y, center_x, center_y, turn, end_z, AA_end, BB_end, CC_end)
	}

	cnc._setup.current.X = end_x
	cnc._setup.current.Y = end_y
	cnc._setup.current.Z = end_z

	cnc._setup.current.A = AA_end /*AA*/

	cnc._setup.current.B = BB_end /*BB*/

	cnc._setup.current.C = CC_end /*CC*/

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_arc2

   Returned Value: int
   If arc_data_ijk or arc_data_r returns an error code,
   this returns that code.
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   This executes an arc command at feed rate. It also updates the
   setting of the position of the tool point to the end point of the move.
   If inverse time feed rate is in effect, it also resets the feed rate.

   Called by: convert_arc.

   This converts a helical or circular arc.

*/

func (cnc *rs274ngc_t) convert_arc2( /* ARGUMENTS                                */
	move inc.GCodes, /* either G_2 (cw arc) or G_3 (ccw arc)     */
	current1, /* pointer to current value of coordinate 1 */
	current2, /* pointer to current value of coordinate 2 */
	current3 *float64, /* pointer to current value of coordinate 3 */
	end1, /* coordinate 1 value at end of arc         */
	end2, /* coordinate 2 value at end of arc         */
	end3 float64, /* coordinate 3 value at end of arc         */
	AA_end, /* a-value at end of arc                    */ /*AA*/
	BB_end, /* b-value at end of arc                    */ /*BB*/
	CC_end, /* c-value at end of arc                    */ /*CC*/
	offset1, /* offset of center from current1           */
	offset2 float64) inc.STATUS { /* offset of center from current2           */

	var turn int /* number of full or partial turns CCW in arc */

	var center1, center2 float64

	/* tolerance for difference of radii          */
	tolerance := inc.If(cnc._setup.length_units == inc.CANON_UNITS_INCHES,
		inc.TOLERANCE_INCH, inc.TOLERANCE_MM).(float64)

	if cnc._setup.block1.r_flag {
		arc.Arc_data_r(move, *current1, *current2, end1, end2,
			cnc._setup.block1.r_number, &center1, &center2, &turn)
	} else {
		arc.Arc_data_ijk(move, *current1, *current2, end1, end2, offset1,
			offset2, &center1, &center2, &turn, tolerance)
	}
	if cnc._setup.feed_mode == inc.INVERSE_TIME {

	}
	cnc.inverse_time_rate_arc(*current1, *current2, *current3, center1, center2,
		turn, end1, end2, end3)
	cnc.canon.ARC_FEED(end1, end2, center1, center2, turn, end3, AA_end, BB_end, CC_end)
	*current1 = end1
	*current2 = end2
	*current3 = end3

	cnc._setup.current.A = AA_end /*AA*/

	cnc._setup.current.B = BB_end /*BB*/

	cnc._setup.current.C = CC_end /*CC*/

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* inverse_time_rate_arc

   Returned Value: int (RS274NGC_OK)

   Side effects: a call is made to SET_FEED_RATE and _setup.feed_rate is set.

   Called by:
   convert_arc2
   convert_arc_comp1
   convert_arc_comp2

   This finds the feed rate needed by an inverse time move. The move
   consists of an a single arc. Most of the work here is in finding the
   length of the arc.

*/

func (cnc *rs274ngc_t) inverse_time_rate_arc( /* ARGUMENTS                       */
	x1, /* x coord of start point of arc            */
	y1, /* y coord of start point of arc            */
	z1, /* z coord of start point of arc            */
	cx, /* x coord of center of arc                 */
	cy float64, /* y coord of center of arc                 */
	turn int, /* turn of arc                              */
	x2, /* x coord of end point of arc              */
	y2, /* y coord of end point of arc              */
	z2 float64) inc.STATUS { /* z coord of end point of arc              */

	length := arc.Find_arc_length(x1, y1, z1, cx, cy, turn, x2, y2, z2)
	rate := math.Max(0.1, (length * cnc._setup.block1.f_number))
	cnc.canon.SET_FEED_RATE(rate)
	cnc._setup.feed_rate = rate

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* inverse_time_rate_arc2

   Returned Value: int (RS274NGC_OK)

   Side effects: a call is made to SET_FEED_RATE and _setup.feed_rate is set.

   Called by: convert_arc_comp2

   This finds the feed rate needed by an inverse time move in
   convert_arc_comp2. The move consists of an extra arc and a main
   arc. Most of the work here is in finding the lengths of the two arcs.

   All rotary motion is assumed to occur on the extra arc, as done by
   convert_arc_comp2.

   All z motion is assumed to occur on the main arc, as done by
   convert_arc_comp2.

*/

func (cnc *rs274ngc_t) inverse_time_rate_arc2( /* ARGUMENTS */
	start_x, /* x coord of last program point, extra arc center x */
	start_y float64, /* y coord of last program point, extra arc center y */
	turn1 int, /* turn of extra arc                                 */
	mid_x, /* x coord of end point of extra arc                 */
	mid_y, /* y coord of end point of extra arc                 */
	cx, /* x coord of center of main arc                     */
	cy float64, /* y coord of center of main arc                     */
	turn2 int, /* turn of main arc                                  */
	end_x, /* x coord of end point of main arc                  */
	end_y, /* y coord of end point of main arc                  */
	end_z float64) inc.STATUS { /* z coord of end point of main arc                  */

	length := (arc.Find_arc_length(cnc._setup.current.X, cnc._setup.current.Y,
		cnc._setup.current.Z, start_x, start_y,
		turn1, mid_x, mid_y, cnc._setup.current.Z) +
		arc.Find_arc_length(mid_x, mid_y, cnc._setup.current.Z,
			cx, cy, turn2, end_x, end_y, end_z))
	rate := math.Max(0.1, (length * cnc._setup.block1.f_number))
	cnc.canon.SET_FEED_RATE(rate)
	cnc._setup.feed_rate = rate

	return inc.RS274NGC_OK
}

package rs274ngc

import (
	"github.com/flyingyizi/rs274ngc/inc"
)

/****************************************************************************/

/* convert_cycle

   Returned Value: int
   If any of the specific functions called returns an error code,
   this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The r-value is not given the first time this code is called after
   some other motion mode has been in effect:
   NCE_R_CLEARANCE_PLANE_UNSPECIFIED_IN_CYCLE
   2. The l number is zero: NCE_CANNOT_DO_ZERO_REPEATS_OF_CYCLE
   3. The currently selected plane in not XY, YZ, or XZ.
   NCE_BUG_PLANE_NOT_XY_YZ_OR_XZ

   Side effects:
   A number of moves are made to execute a canned cycle. The current
   position is reset. The values of the cycle attributes in the settings
   may be reset.

   Called by: convert_motion

   This function makes a couple checks and then calls one of three
   functions, according to which plane is currently selected.

   See the documentation of convert_cycle_xy for most of the details.

*/

func (cnc *rs274ngc_t) convert_cycle( /* ARGUMENTS                                      */
	motion inc.GCodes) inc.STATUS { /* a g-code between G_81 and G_89, a canned cycle */

	//static char name[] = "convert_cycle";

	plane := cnc._setup.plane
	if cnc._setup.block1.r_flag == OFF {
		if cnc._setup.motion_mode == motion {
			cnc._setup.block1.r_number = cnc._setup.cycle.r
		} else {
			return inc.NCE_R_CLEARANCE_PLANE_UNSPECIFIED_IN_CYCLE
		}
	}

	if cnc._setup.block1.l_number == 0 {
		return inc.NCE_CANNOT_DO_ZERO_REPEATS_OF_CYCLE
	}
	if cnc._setup.block1.l_number == -1 {
		cnc._setup.block1.l_number = 1
	}

	if plane == CANON_PLANE_XY {
		cnc.convert_cycle_xy(motion)
	} else if plane == CANON_PLANE_YZ {
		cnc.convert_cycle_yz(motion)
	} else if plane == CANON_PLANE_XZ {
		cnc.convert_cycle_zx(motion)
	} else {
		return inc.NCE_BUG_PLANE_NOT_XY_YZ_OR_XZ
	}

	cnc._setup.cycle.l = cnc._setup.block1.l_number
	cnc._setup.cycle.r = cnc._setup.block1.r_number
	cnc._setup.motion_mode = motion
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* cycle_feed

   Returned Value: int (RS274NGC_OK)

   Side effects:
   STRAIGHT_FEED is called.

   Called by:
   convert_cycle_g81
   convert_cycle_g82
   convert_cycle_g83
   convert_cycle_g84
   convert_cycle_g85
   convert_cycle_g86
   convert_cycle_g87
   convert_cycle_g88
   convert_cycle_g89

   This writes a STRAIGHT_FEED command appropriate for a cycle move with
   respect to the given plane. No rotary axis motion takes place.

*/

func (cnc *rs274ngc_t) cycle_feed( /* ARGUMENTS                  */
	plane CANON_PLANE, /* currently selected plane   */
	end1, /* first coordinate value     */
	end2, /* second coordinate value    */
	end3 float64) inc.STATUS { /* third coordinate value     */

	//static char name[] = "cycle_feed";

	if plane == CANON_PLANE_XY {
		cnc.canon.STRAIGHT_FEED(end1, end2, end3, cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)

	} else if plane == CANON_PLANE_YZ {
		cnc.canon.STRAIGHT_FEED(end3, end1, end2, cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)

	} else { /* if (plane IS CANON_PLANE_XZ) */
		cnc.canon.STRAIGHT_FEED(end2, end3, end1, cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* cycle_traverse

   Returned Value: int (RS274NGC_OK)

   Side effects:
   STRAIGHT_TRAVERSE is called.

   Called by:
   convert_cycle
   convert_cycle_g81
   convert_cycle_g82
   convert_cycle_g83
   convert_cycle_g86
   convert_cycle_g87
   convert_cycle_xy (via CYCLE_MACRO)
   convert_cycle_yz (via CYCLE_MACRO)
   convert_cycle_zx (via CYCLE_MACRO)

   This writes a STRAIGHT_TRAVERSE command appropriate for a cycle
   move with respect to the given plane. No rotary axis motion takes place.

*/

func (cnc *rs274ngc_t) cycle_traverse( /* ARGUMENTS                 */
	plane CANON_PLANE, /* currently selected plane  */
	end1, /* first coordinate value    */
	end2, /* second coordinate value   */
	end3 float64) inc.STATUS { /* third coordinate value    */

	//static char name[] = "cycle_traverse";
	if plane == CANON_PLANE_XY {
		cnc.canon.STRAIGHT_TRAVERSE(end1, end2, end3, cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)

	} else if plane == CANON_PLANE_YZ {
		cnc.canon.STRAIGHT_TRAVERSE(end3, end1, end2, cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)

	} else { /* if (plane == CANON_PLANE_XZ) */
		cnc.canon.STRAIGHT_TRAVERSE(end2, end3, end1, cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g81

   Returned Value: int (RS274NGC_OK)

   Side effects: See below

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   For the XY plane, this implements the following RS274/NGC cycle, which
   is usually drilling:
   1. Move the z-axis only at the current feed rate to the specified bottom_z.
   2. Retract the z-axis at traverse rate to clear_z.

   See [NCMS, page 99].

   CYCLE_MACRO has positioned the tool at (x, y, r, a, b, c) when this starts.

   For the XZ and YZ planes, this makes analogous motions.

*/

func (cnc *rs274ngc_t) convert_cycle_g81( /* ARGUMENTS                        */
	plane CANON_PLANE, /* selected plane                   */
	x, /* x-value where cycle is executed  */
	y, /* y-value where cycle is executed  */
	clear_z, /* z-value of clearance plane       */
	bottom_z float64) inc.STATUS { /* value of z at bottom of cycle    */

	//static char name[] = "convert_cycle_g81";

	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.cycle_traverse(plane, x, y, clear_z)

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g82

   Returned Value: int (RS274NGC_OK)

   Side effects: See below

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   For the XY plane, this implements the following RS274/NGC cycle, which
   is usually drilling:
   1. Move the z_axis only at the current feed rate to the specified z-value.
   2. Dwell for the given number of seconds.
   3. Retract the z-axis at traverse rate to the clear_z.

   CYCLE_MACRO has positioned the tool at (x, y, r, a, b, c) when this starts.

   For the XZ and YZ planes, this makes analogous motions.

*/

func (cnc *rs274ngc_t) convert_cycle_g82( /* ARGUMENTS                        */
	plane CANON_PLANE, /* selected plane                   */
	x, /* x-value where cycle is executed  */
	y, /* y-value where cycle is executed  */
	clear_z, /* z-value of clearance plane       */
	bottom_z, /* value of z at bottom of cycle    */
	dwell float64) inc.STATUS { /* dwell time                       */

	//static char name[] = "convert_cycle_g82";

	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.canon.DWELL(dwell)
	cnc.cycle_traverse(plane, x, y, clear_z)

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g83

   Returned Value: int (RS274NGC_OK)

   Side effects: See below

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   For the XY plane, this implements the following RS274/NGC cycle,
   which is usually peck drilling:
   1. Move the z-axis only at the current feed rate downward by delta or
   to the specified bottom_z, whichever is less deep.
   2. Rapid back out to the clear_z.
   3. Rapid back down to the current hole bottom, backed off a bit.
   4. Repeat steps 1, 2, and 3 until the specified bottom_z is reached.
   5. Retract the z-axis at traverse rate to clear_z.

   CYCLE_MACRO has positioned the tool at (x, y, r, a, b, c) when this starts.

   The rapid out and back in causes any long stringers (which are common
   when drilling in aluminum) to be cut off and clears chips from the
   hole.

   For the XZ and YZ planes, this makes analogous motions.

*/

const (
	G83_RAPID_DELTA = 0.010
)

/* how far above hole bottom for rapid
   return, in inches */

func (cnc *rs274ngc_t) convert_cycle_g83( /* ARGUMENTS                        */
	plane CANON_PLANE, /* selected plane                   */
	x, /* x-value where cycle is executed  */
	y, /* y-value where cycle is executed  */
	r, /* initial z-value                  */
	clear_z, /* z-value of clearance plane       */
	bottom_z, /* value of z at bottom of cycle    */
	delta float64) inc.STATUS { /* size of z-axis feed increment    */

	//        static char name[] = "convert_cycle_g83";

	rapid_delta := G83_RAPID_DELTA
	if cnc._setup.length_units == CANON_UNITS_MM {
		rapid_delta = (rapid_delta * 25.4)

	}

	current_depth := (r - delta)
	for ; current_depth > bottom_z; current_depth = (current_depth - delta) {
		cnc.cycle_feed(plane, x, y, current_depth)
		cnc.cycle_traverse(plane, x, y, clear_z)
		cnc.cycle_traverse(plane, x, y, current_depth+rapid_delta)
	}
	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.cycle_traverse(plane, x, y, clear_z)

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g84

   Returned Value: int
   If the spindle is not turning clockwise, this returns
   NCE_SPINDLE_NOT_TURNING_CLOCKWISE_IN_G84.
   Otherwise, it returns RS274NGC_OK.

   Side effects: See below

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   For the XY plane, this implements the following RS274/NGC cycle,
   which is right-hand tapping:
   1. Start speed-feed synchronization.
   2. Move the z-axis only at the current feed rate to the specified bottom_z.
   3. Stop the spindle.
   4. Start the spindle counterclockwise.
   5. Retract the z-axis at current feed rate to clear_z.
   6. If speed-feed synch was not on before the cycle started, stop it.
   7. Stop the spindle.
   8. Start the spindle clockwise.

   CYCLE_MACRO has positioned the tool at (x, y, r, a, b, c) when this starts.
   The direction argument must be clockwise.

   For the XZ and YZ planes, this makes analogous motions.

*/

func (cnc *rs274ngc_t) convert_cycle_g84( /* ARGUMENTS                           */
	plane CANON_PLANE, /* selected plane                      */
	x, /* x-value where cycle is executed     */
	y, /* y-value where cycle is executed     */
	clear_z, /* z-value of clearance plane          */
	bottom_z float64, /* value of z at bottom of cycle       */
	direction CANON_DIRECTION, /* direction spindle turning at outset */
	mode inc.CANON_SPEED_FEED_MODE) inc.STATUS { /* the speed-feed mode at outset       */

	//static char name[] = "convert_cycle_g84";

	if direction != CANON_CLOCKWISE {
		return inc.NCE_SPINDLE_NOT_TURNING_CLOCKWISE_IN_G84
	}

	cnc.canon.START_SPEED_FEED_SYNCH()
	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.canon.STOP_SPINDLE_TURNING()
	cnc.canon.START_SPINDLE_COUNTERCLOCKWISE()
	cnc.cycle_feed(plane, x, y, clear_z)
	if mode != inc.CANON_SYNCHED {
		cnc.canon.STOP_SPEED_FEED_SYNCH()

	}
	cnc.canon.STOP_SPINDLE_TURNING()
	cnc.canon.START_SPINDLE_CLOCKWISE()

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g85

   Returned Value: int (RS274NGC_OK)

   Side effects:
   A number of moves are made as described below.

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   For the XY plane, this implements the following RS274/NGC cycle,
   which is usually boring or reaming:
   1. Move the z-axis only at the current feed rate to the specified z-value.
   2. Retract the z-axis at the current feed rate to clear_z.

   CYCLE_MACRO has positioned the tool at (x, y, r, ?, ?) when this starts.

   For the XZ and YZ planes, this makes analogous motions.

*/

func (cnc *rs274ngc_t) convert_cycle_g85( /* ARGUMENTS                        */
	plane CANON_PLANE, /* selected plane                   */
	x, /* x-value where cycle is executed  */
	y, /* y-value where cycle is executed  */
	clear_z, /* z-value of clearance plane       */
	bottom_z float64) inc.STATUS { /* value of z at bottom of cycle    */

	//static char name[] = "convert_cycle_g85";

	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.cycle_feed(plane, x, y, clear_z)

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g86

   Returned Value: int
   If the spindle is not turning clockwise or counterclockwise,
   this returns NCE_SPINDLE_NOT_TURNING_IN_G86.
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   A number of moves are made as described below.

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   For the XY plane, this implements the RS274/NGC following cycle,
   which is usually boring:
   1. Move the z-axis only at the current feed rate to bottom_z.
   2. Dwell for the given number of seconds.
   3. Stop the spindle turning.
   4. Retract the z-axis at traverse rate to clear_z.
   5. Restart the spindle in the direction it was going.

   CYCLE_MACRO has positioned the tool at (x, y, r, a, b, c) when this starts.

   For the XZ and YZ planes, this makes analogous motions.

*/

func (cnc *rs274ngc_t) convert_cycle_g86( /* ARGUMENTS                           */
	plane CANON_PLANE, /* selected plane                      */
	x, /* x-value where cycle is executed     */
	y, /* y-value where cycle is executed     */
	clear_z, /* z-value of clearance plane          */
	bottom_z, /* value of z at bottom of cycle       */
	dwell float64, /* dwell time                          */
	direction CANON_DIRECTION) inc.STATUS { /* direction spindle turning at outset */

	// static char name[] = "convert_cycle_g86";

	if (direction != CANON_CLOCKWISE) &&
		(direction != CANON_COUNTERCLOCKWISE) {
		return inc.NCE_SPINDLE_NOT_TURNING_IN_G86
	}

	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.canon.DWELL(dwell)
	cnc.canon.STOP_SPINDLE_TURNING()
	cnc.cycle_traverse(plane, x, y, clear_z)
	if direction == CANON_CLOCKWISE {
		cnc.canon.START_SPINDLE_CLOCKWISE()

	} else {
		cnc.canon.START_SPINDLE_COUNTERCLOCKWISE()
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g87

   Returned Value: int
   If the spindle is not turning clockwise or counterclockwise,
   this returns NCE_SPINDLE_NOT_TURNING_IN_G87.
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   A number of moves are made as described below. This cycle is a
   modified version of [Monarch, page 5-24] since [NCMS, pages 98 - 100]
   gives no clue as to what the cycle is supposed to do. [KT] does not
   have a back boring cycle. [Fanuc, page 132] in "Canned cycle II"
   describes the G87 cycle as given here, except that the direction of
   spindle turning is always clockwise and step 7 below is omitted
   in [Fanuc].

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   For the XY plane, this implements the following RS274/NGC cycle, which
   is usually back boring.  The situation is that you have a through hole
   and you want to counterbore the bottom of hole. To do this you put an
   L-shaped tool in the spindle with a cutting surface on the UPPER side
   of its base. You stick it carefully through the hole when it is not
   spinning and is oriented so it fits through the hole, then you move it
   so the stem of the L is on the axis of the hole, start the spindle,
   and feed the tool upward to make the counterbore. Then you get the
   tool out of the hole.

   1. Move at traverse rate parallel to the XY-plane to the point
   with x-value offset_x and y-value offset_y.
   2. Stop the spindle in a specific orientation.
   3. Move the z-axis only at traverse rate downward to the bottom_z.
   4. Move at traverse rate parallel to the XY-plane to the x,y location.
   5. Start the spindle in the direction it was going before.
   6. Move the z-axis only at the given feed rate upward to the middle_z.
   7. Move the z-axis only at the given feed rate back down to bottom_z.
   8. Stop the spindle in the same orientation as before.
   9. Move at traverse rate parallel to the XY-plane to the point
   with x-value offset_x and y-value offset_y.
   10. Move the z-axis only at traverse rate to the clear z value.
   11. Move at traverse rate parallel to the XY-plane to the specified x,y
   location.
   12. Restart the spindle in the direction it was going before.

   CYCLE_MACRO has positioned the tool at (x, y, r, a, b, c) before this starts.

   It might be useful to add a check that clear_z > middle_z > bottom_z.
   Without the check, however, this can be used to counterbore a hole in
   material that can only be accessed through a hole in material above it.

   For the XZ and YZ planes, this makes analogous motions.

*/

func (cnc *rs274ngc_t) convert_cycle_g87( /* ARGUMENTS                           */
	plane CANON_PLANE, /* selected plane                      */
	x, /* x-value where cycle is executed     */
	offset_x, /* x-axis offset position              */
	y, /* y-value where cycle is executed     */
	offset_y, /* y-axis offset position              */
	r, /* z_value of r_plane                  */
	clear_z, /* z-value of clearance plane          */
	middle_z, /* z-value of top of back bore         */
	bottom_z float64, /* value of z at bottom of cycle       */
	direction CANON_DIRECTION) inc.STATUS { /* direction spindle turning at outset */

	//static char name[] = "convert_cycle_g87";

	if (direction != CANON_CLOCKWISE) &&
		(direction != CANON_COUNTERCLOCKWISE) {
		return inc.NCE_SPINDLE_NOT_TURNING_IN_G87
	}

	cnc.cycle_traverse(plane, offset_x, offset_y, r)
	cnc.canon.STOP_SPINDLE_TURNING()
	cnc.canon.ORIENT_SPINDLE(0.0, direction)
	cnc.cycle_traverse(plane, offset_x, offset_y, bottom_z)
	cnc.cycle_traverse(plane, x, y, bottom_z)
	if direction == CANON_CLOCKWISE {
		cnc.canon.START_SPINDLE_CLOCKWISE()

	} else {
		cnc.canon.START_SPINDLE_COUNTERCLOCKWISE()

	}
	cnc.cycle_feed(plane, x, y, middle_z)
	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.canon.STOP_SPINDLE_TURNING()
	cnc.canon.ORIENT_SPINDLE(0.0, direction)
	cnc.cycle_traverse(plane, offset_x, offset_y, bottom_z)
	cnc.cycle_traverse(plane, offset_x, offset_y, clear_z)
	cnc.cycle_traverse(plane, x, y, clear_z)
	if direction == CANON_CLOCKWISE {
		cnc.canon.START_SPINDLE_CLOCKWISE()

	} else {
		cnc.canon.START_SPINDLE_COUNTERCLOCKWISE()

	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g88

   Returned Value: int
   If the spindle is not turning clockwise or counterclockwise, this
   returns NCE_SPINDLE_NOT_TURNING_IN_G88.
   Otherwise, it returns RS274NGC_OK.

   Side effects: See below

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   For the XY plane, this implements the following RS274/NGC cycle,
   which is usually boring:
   1. Move the z-axis only at the current feed rate to the specified z-value.
   2. Dwell for the given number of seconds.
   3. Stop the spindle turning.
   4. Stop the program so the operator can retract the spindle manually.
   5. Restart the spindle.

   CYCLE_MACRO has positioned the tool at (x, y, r, a, b, c) when this starts.

   For the XZ and YZ planes, this makes analogous motions.

*/

func (cnc *rs274ngc_t) convert_cycle_g88( /* ARGUMENTS                           */
	plane CANON_PLANE, /* selected plane                      */
	x, /* x-value where cycle is executed     */
	y, /* y-value where cycle is executed     */
	bottom_z, /* value of z at bottom of cycle       */
	dwell float64, /* dwell time                          */
	direction CANON_DIRECTION) inc.STATUS { /* direction spindle turning at outset */

	//static char name[] = "convert_cycle_g88";

	if (direction != CANON_CLOCKWISE) &&
		(direction != CANON_COUNTERCLOCKWISE) {
		return inc.NCE_SPINDLE_NOT_TURNING_IN_G88
	}

	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.canon.DWELL(dwell)
	cnc.canon.STOP_SPINDLE_TURNING()
	cnc.canon.PROGRAM_STOP() /* operator retracts the spindle here */
	if direction == CANON_CLOCKWISE {
		cnc.canon.START_SPINDLE_CLOCKWISE()

	} else {
		cnc.canon.START_SPINDLE_COUNTERCLOCKWISE()

	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_g89

   Returned Value: int (RS274NGC_OK)

   Side effects: See below

   Called by:
   convert_cycle_xy
   convert_cycle_yz
   convert_cycle_zx

   This implements the following RS274/NGC cycle, which is intended for boring:
   1. Move the z-axis only at the current feed rate to the specified z-value.
   2. Dwell for the given number of seconds.
   3. Retract the z-axis at the current feed rate to clear_z.

   CYCLE_MACRO has positioned the tool at (x, y, r, a, b, c) when this starts.

   For the XZ and YZ planes, this makes analogous motions.

*/

func (cnc *rs274ngc_t) convert_cycle_g89( /* ARGUMENTS                        */
	plane CANON_PLANE, /* selected plane                   */
	x, /* x-value where cycle is executed  */
	y, /* y-value where cycle is executed  */
	clear_z, /* z-value of clearance plane       */
	bottom_z, /* value of z at bottom of cycle    */
	dwell float64) inc.STATUS { /* dwell time                       */

	//static char name[] = "convert_cycle_g89";

	cnc.cycle_feed(plane, x, y, bottom_z)
	cnc.canon.DWELL(dwell)
	cnc.cycle_feed(plane, x, y, clear_z)

	return inc.RS274NGC_OK
}

/* convert_cycle_xy

   Returned Value: int
   If any of the specific functions called returns an error code,
   this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The z-value is not given the first time this code is called after
   some other motion mode has been in effect:
   NCE_Z_VALUE_UNSPECIFIED_IN_XY_PLANE_CANNED_CYCLE
   2. The r clearance plane is below the bottom_z:
   NCE_R_LESS_THAN_Z_IN_CYCLE_IN_XY_PLANE
   3. the distance mode is neither absolute or incremental:
   NCE_BUG_DISTANCE_MODE_NOT_G90_OR_G91
   4. G82, G86, G88, or G89 is called when it is not already in effect,
   and no p number is in the block:
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G82
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G86
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G88
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G89
   5. G83 is called when it is not already in effect,
   and no q number is in the block: NCE_Q_WORD_MISSING_WITH_G83
   6. G87 is called when it is not already in effect,
   and any of the i number, j number, or k number is missing:
   NCE_I_WORD_MISSING_WITH_G87
   NCE_J_WORD_MISSING_WITH_G87
   NCE_K_WORD_MISSING_WITH_G87
   7. the G code is not between G_81 and G_89.
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED

   Side effects:
   A number of moves are made to execute the g-code

   Called by: convert_cycle

   The function does not require that any of x,y,z, or r be specified in
   the block, except that if the last motion mode command executed was
   not the same as this one, the r-value and z-value must be specified.

   This function is handling the repeat feature of RS274/NGC, wherein
   the L word represents the number of repeats [NCMS, page 99]. We are
   not allowing L=0, contrary to the manual. We are allowing L > 1
   in absolute distance mode to mean "do the same thing in the same
   place several times", as provided in the manual, although this seems
   abnormal.

   In incremental distance mode, x, y, and r values are treated as
   increments to the current position and z as an increment from r.  In
   absolute distance mode, x, y, r, and z are absolute. In g87, i and j
   will always be increments, regardless of the distance mode setting, as
   implied in [NCMS, page 98], but k (z-value of top of counterbore) will
   be an absolute z-value in absolute distance mode, and an increment
   (from bottom z) in incremental distance mode.

   If the r position of a cycle is above the current.Z position, this
   retracts the z-axis to the r position before moving parallel to the
   XY plane.

   In the code for this function, there is a nearly identical "for" loop
   in every case of the switch. The loop is the done with a compiler
   macro, "CYCLE_MACRO" so that the code is easy to read, automatically
   kept identical from case to case and, and much shorter than it would
   be without the macro. The loop could be put outside the switch, but
   then the switch would run every time around the loop, not just once,
   as it does here. The loop could also be placed in the called
   functions, but then it would not be clear that all the loops are the
   same, and it would be hard to keep them the same when the code is
   modified.  The macro would be very awkward as a regular function
   because it would have to be passed all of the arguments used by any of
   the specific cycles, and, if another switch in the function is to be
   avoided, it would have to passed a function pointer, but the different
   cycle functions have different arguments so the type of the pointer
   could not be declared unless the cycle functions were re-written to
   take the same arguments (in which case most of them would have several
   unused arguments).

   The motions within the CYCLE_MACRO (but outside a specific cycle) are
   a straight traverse parallel to the selected plane to the given
   position in the plane and a straight traverse of the third axis only
   (if needed) to the r position.

   The CYCLE_MACRO is defined here but is also used in convert_cycle_yz
   and convert_cycle_zx. The variables aa, bb, and cc are used in
   CYCLE_MACRO and in the other two functions just mentioned. Those
   variables represent the first axis of the selected plane, the second
   axis of the selected plane, and third axis which is perpendicular to
   the selected plane.  In this function aa represents x, bb represents
   y, and cc represents z. This usage makes it possible to have only one
   version of each of the cycle functions.  The cycle_traverse and
   cycle_feed functions help accomplish this.

   The height of the retract move at the end of each repeat of a cycle is
   determined by the setting of the retract_mode: either to the r
   position (if the retract_mode is R_PLANE) or to the original
   z-position (if that is above the r position and the retract_mode is
   not R_PLANE). This is a slight departure from [NCMS, page 98], which
   does not require checking that the original z-position is above r.

   The rotary axes may not move during a canned cycle.

*/

//   #define CYCLE_MACRO(call) for (repeat = cnc._setup.block1.l_number; \
//	repeat > 0; \
//	repeat--) \
//	{ \
//		aa = (aa + aa_increment); \
//		bb = (bb + bb_increment); \
//		cycle_traverse(plane, aa, bb, old_cc); \
//		if (old_cc != r) \
//		cycle_traverse(plane, aa, bb, r); \
//		CHP(call); \
//		old_cc = clear_cc; \
//	}

func (cnc *rs274ngc_t) convert_cycle_xy( /* ARGUMENTS                                 */
	motion inc.GCodes) inc.STATUS { /* a g-code between G_81 and G_89, a canned cycle */

	//	static char name[] = "convert_cycle_xy";
	var (
		r,
		aa,
		aa_increment,
		bb,
		bb_increment,
		cc,
		clear_cc,
		i,
		j,
		k,
		old_cc float64

		save_mode CANON_MOTION_MODE
	)
	plane := CANON_PLANE_XY

	if cnc._setup.motion_mode != motion {
		if cnc._setup.block1.z_flag == OFF {
			return inc.NCE_Z_VALUE_UNSPECIFIED_IN_XY_PLANE_CANNED_CYCLE
		}
	}
	cnc._setup.block1.z_number =
		inc.If(cnc._setup.block1.z_flag == ON, cnc._setup.block1.z_number, cnc._setup.cycle.cc).(float64)
	old_cc = cnc._setup.current.Z

	if cnc._setup.distance_mode == inc.MODE_ABSOLUTE {
		aa_increment = 0.0
		bb_increment = 0.0
		r = cnc._setup.block1.r_number
		cc = cnc._setup.block1.z_number
		aa = inc.If(cnc._setup.block1.x_flag == ON, cnc._setup.block1.x_number, cnc._setup.current.X).(float64)
		bb = inc.If(cnc._setup.block1.y_flag == ON, cnc._setup.block1.y_number, cnc._setup.current.Y).(float64)
	} else if cnc._setup.distance_mode == inc.MODE_INCREMENTAL {
		aa_increment = cnc._setup.block1.x_number
		bb_increment = cnc._setup.block1.y_number
		r = (cnc._setup.block1.r_number + old_cc)
		cc = (r + cnc._setup.block1.z_number) /* [NCMS, page 98] */
		aa = cnc._setup.current.X
		bb = cnc._setup.current.Y
	} else {
		return inc.NCE_BUG_DISTANCE_MODE_NOT_G90_OR_G91
	}
	if r < cc {
		return inc.NCE_R_LESS_THAN_Z_IN_CYCLE_IN_XY_PLANE
	}

	if old_cc < r {
		cnc.canon.STRAIGHT_TRAVERSE(cnc._setup.current.X, cnc._setup.current.Y, r,
			cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)
		old_cc = r
	}
	clear_cc = inc.If(cnc._setup.retract_mode == inc.R_PLANE, r, old_cc).(float64)

	save_mode = cnc.canon.GET_EXTERNAL_MOTION_CONTROL_MODE()
	if save_mode != CANON_EXACT_PATH {
		cnc.canon.SET_MOTION_CONTROL_MODE(CANON_EXACT_PATH)
	}

	switch motion {
	case inc.G_81:
		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g81(CANON_PLANE_XY, aa, bb, clear_cc, cc)
			old_cc = clear_cc
		}
		break
	case inc.G_82:
		if (cnc._setup.motion_mode != inc.G_82) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G82
		}
		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g82(CANON_PLANE_XY, aa, bb, clear_cc, cc, cnc._setup.block1.p_number)
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	case inc.G_83:
		if (cnc._setup.motion_mode != inc.G_83) && (cnc._setup.block1.q_number == -1.0) {
			return inc.NCE_Q_WORD_MISSING_WITH_G83
		}

		cnc._setup.block1.q_number =
			inc.If(cnc._setup.block1.q_number == -1.0, cnc._setup.cycle.q, cnc._setup.block1.q_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g83(CANON_PLANE_XY, aa, bb, r, clear_cc, cc, cnc._setup.block1.q_number)
			old_cc = clear_cc
		}

		cnc._setup.cycle.q = cnc._setup.block1.q_number
		break
	case inc.G_84:
		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g84(CANON_PLANE_XY, aa, bb, clear_cc, cc, cnc._setup.spindle_turning, cnc._setup.speed_feed_mode)
			old_cc = clear_cc
		}

		break
	case inc.G_85:
		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g85(CANON_PLANE_XY, aa, bb, clear_cc, cc)
			old_cc = clear_cc
		}

		break
	case inc.G_86:
		if (cnc._setup.motion_mode != inc.G_86) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G86
		}

		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g86(CANON_PLANE_XY, aa, bb, clear_cc, cc, cnc._setup.block1.p_number, cnc._setup.spindle_turning)
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	case inc.G_87:
		if cnc._setup.motion_mode != inc.G_87 {
			if cnc._setup.block1.i_flag == OFF {
				return inc.NCE_I_WORD_MISSING_WITH_G87
			}
			if cnc._setup.block1.j_flag == OFF {
				return inc.NCE_J_WORD_MISSING_WITH_G87
			}
			if cnc._setup.block1.k_flag == OFF {
				return inc.NCE_K_WORD_MISSING_WITH_G87
			}
		}
		i = inc.If(cnc._setup.block1.i_flag == ON, cnc._setup.block1.i_number, cnc._setup.cycle.i).(float64)
		j = inc.If(cnc._setup.block1.j_flag == ON, cnc._setup.block1.j_number, cnc._setup.cycle.j).(float64)
		k = inc.If(cnc._setup.block1.k_flag == ON, cnc._setup.block1.k_number, cnc._setup.cycle.k).(float64)
		cnc._setup.cycle.i = i
		cnc._setup.cycle.j = j
		cnc._setup.cycle.k = k
		if cnc._setup.distance_mode == inc.MODE_INCREMENTAL {
			k = (cc + k) /* k always absolute in function call below */
		}

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g87(CANON_PLANE_XY, aa, (aa + i), bb, (bb + j), r, clear_cc, k, cc, cnc._setup.spindle_turning)
			old_cc = clear_cc
		}

		break
	case inc.G_88:
		if (cnc._setup.motion_mode != inc.G_88) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G88
		}

		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g88(CANON_PLANE_XY, aa, bb, cc, cnc._setup.block1.p_number, cnc._setup.spindle_turning)
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	case inc.G_89:
		if (cnc._setup.motion_mode != inc.G_89) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G89
		}

		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g89(CANON_PLANE_XY, aa, bb, clear_cc, cc, cnc._setup.block1.p_number)
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	default:
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	cnc._setup.current.X = aa /* CYCLE_MACRO updates aa and bb */
	cnc._setup.current.Y = bb
	cnc._setup.current.Z = clear_cc
	cnc._setup.cycle.cc = cnc._setup.block1.z_number

	if save_mode != CANON_EXACT_PATH {
		cnc.canon.SET_MOTION_CONTROL_MODE(save_mode)

	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_yz

   Returned Value: int
   If any of the specific functions called returns an error code,
   this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The x-value is not given the first time this code is called after
   some other motion mode has been in effect:
   NCE_X_VALUE_UNSPECIFIED_IN_YZ_PLANE_CANNED_CYCLE
   2. The r clearance plane is below the bottom_x:
   NCE_R_LESS_THAN_X_IN_CYCLE_IN_YZ_PLANE
   3. the distance mode is neither absolute or incremental:
   NCE_BUG_DISTANCE_MODE_NOT_G90_OR_G91
   4. G82, G86, G88, or G89 is called when it is not already in effect,
   and no p number is in the block:
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G82
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G86
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G88
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G89
   5. G83 is called when it is not already in effect,
   and no q number is in the block: NCE_Q_WORD_MISSING_WITH_G83
   6. G87 is called when it is not already in effect,
   and any of the i number, j number, or k number is missing:
   NCE_I_WORD_MISSING_WITH_G87
   NCE_J_WORD_MISSING_WITH_G87
   NCE_K_WORD_MISSING_WITH_G87
   7. the G code is not between G_81 and G_89.
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED

   Side effects:
   A number of moves are made to execute a canned cycle.

   Called by: convert_cycle

   See the documentation of convert_cycle_xy. This function is entirely
   similar. In this function aa represents y, bb represents z, and cc
   represents x.

   The CYCLE_MACRO is defined just before the convert_cycle_xy function.

   Tool length offsets work only when the tool axis is parallel to the
   Z-axis, so if this function is used, tool length offsets should be
   turned off, and the NC code written to take tool length into account.

*/

func (cnc *rs274ngc_t) convert_cycle_yz( /* ARGUMENTS                                 */
	motion inc.GCodes) inc.STATUS { /* a g-code between G_81 and G_89, a canned cycle */

	//			static char name[] = "convert_cycle_yz";
	var (
		aa,
		aa_increment,
		bb,
		bb_increment,
		cc,
		clear_cc,
		i,
		j,
		k,
		old_cc,
		r float64
		//repeat    int
		save_mode CANON_MOTION_MODE
	)

	plane := CANON_PLANE_YZ

	if cnc._setup.motion_mode != motion {
		if cnc._setup.block1.x_flag == OFF {
			return inc.NCE_X_VALUE_UNSPECIFIED_IN_YZ_PLANE_CANNED_CYCLE
		}

	}
	cnc._setup.block1.x_number =
		inc.If(cnc._setup.block1.x_flag == ON, cnc._setup.block1.x_number, cnc._setup.cycle.cc).(float64)
	old_cc = cnc._setup.current.X

	if cnc._setup.distance_mode == inc.MODE_ABSOLUTE {
		aa_increment = 0.0
		bb_increment = 0.0
		r = cnc._setup.block1.r_number
		cc = cnc._setup.block1.x_number
		aa = inc.If(cnc._setup.block1.y_flag == ON, cnc._setup.block1.y_number, cnc._setup.current.Y).(float64)
		bb = inc.If(cnc._setup.block1.z_flag == ON, cnc._setup.block1.z_number, cnc._setup.current.Z).(float64)
	} else if cnc._setup.distance_mode == inc.MODE_INCREMENTAL {
		aa_increment = cnc._setup.block1.y_number
		bb_increment = cnc._setup.block1.z_number
		r = (cnc._setup.block1.r_number + old_cc)
		cc = (r + cnc._setup.block1.x_number) /* [NCMS, page 98] */
		aa = cnc._setup.current.Y
		bb = cnc._setup.current.Z
	} else {
		return inc.NCE_BUG_DISTANCE_MODE_NOT_G90_OR_G91

	}
	if r < cc {
		return inc.NCE_R_LESS_THAN_X_IN_CYCLE_IN_YZ_PLANE
	}
	if old_cc < r {
		cnc.canon.STRAIGHT_TRAVERSE(r, cnc._setup.current.Y, cnc._setup.current.Z,
			cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)
		old_cc = r
	}
	clear_cc = inc.If(cnc._setup.retract_mode == inc.R_PLANE, r, old_cc).(float64)

	save_mode = cnc.canon.GET_EXTERNAL_MOTION_CONTROL_MODE()
	if save_mode != CANON_EXACT_PATH {
		cnc.canon.SET_MOTION_CONTROL_MODE(CANON_EXACT_PATH)
	}

	switch motion {
	case inc.G_81:
		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g81(CANON_PLANE_YZ, aa, bb, clear_cc, cc)
			old_cc = clear_cc
		}

		break
	case inc.G_82:
		if (cnc._setup.motion_mode != inc.G_82) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G82
		}

	case inc.G_87:
		if cnc._setup.motion_mode != inc.G_87 {
			if cnc._setup.block1.i_flag == OFF {
				return inc.NCE_I_WORD_MISSING_WITH_G87
			}
			if cnc._setup.block1.j_flag == OFF {
				return inc.NCE_J_WORD_MISSING_WITH_G87
			}
			if cnc._setup.block1.k_flag == OFF {
				return inc.NCE_K_WORD_MISSING_WITH_G87
			}
		}
		i = inc.If(cnc._setup.block1.i_flag == ON, cnc._setup.block1.i_number, cnc._setup.cycle.i).(float64)
		j = inc.If(cnc._setup.block1.j_flag == ON, cnc._setup.block1.j_number, cnc._setup.cycle.j).(float64)
		k = inc.If(cnc._setup.block1.k_flag == ON, cnc._setup.block1.k_number, cnc._setup.cycle.k).(float64)
		cnc._setup.cycle.i = i
		cnc._setup.cycle.j = j
		cnc._setup.cycle.k = k
		if cnc._setup.distance_mode == inc.MODE_INCREMENTAL {
			i = (cc + i) /* i always absolute in function call below */
		}

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g87(CANON_PLANE_YZ, aa, (aa + j), bb, (bb + k), r, clear_cc, i, cc, cnc._setup.spindle_turning)
			old_cc = clear_cc
		}

		break
	case inc.G_88:
		if (cnc._setup.motion_mode != inc.G_88) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G88
		}
		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g88(CANON_PLANE_YZ, aa, bb, cc, cnc._setup.block1.p_number, cnc._setup.spindle_turning)
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	case inc.G_89:
		if (cnc._setup.motion_mode != inc.G_89) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G89
		}

		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			cnc.convert_cycle_g89(CANON_PLANE_YZ, aa, bb, clear_cc, cc, cnc._setup.block1.p_number)
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	default:
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	cnc._setup.current.Y = aa /* CYCLE_MACRO updates aa and bb */
	cnc._setup.current.Z = bb
	cnc._setup.current.X = clear_cc
	cnc._setup.cycle.cc = cnc._setup.block1.x_number

	if save_mode != CANON_EXACT_PATH {
		cnc.canon.SET_MOTION_CONTROL_MODE(save_mode)
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cycle_zx

   Returned Value: int
   If any of the specific functions called returns an error code,
   this returns that code.
   If any of the following errors occur, this returns the ERROR code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The y-value is not given the first time this code is called after
   some other motion mode has been in effect:
   NCE_Y_VALUE_UNSPECIFIED_IN_XZ_PLANE_CANNED_CYCLE
   2. The r clearance plane is below the bottom_y:
   NCE_R_LESS_THAN_Y_IN_CYCLE_IN_XZ_PLANE
   3. the distance mode is neither absolute or incremental:
   NCE_BUG_DISTANCE_MODE_NOT_G90_OR_G91
   4. G82, G86, G88, or G89 is called when it is not already in effect,
   and no p number is in the block:
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G82
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G86
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G88
   NCE_DWELL_TIME_P_WORD_MISSING_WITH_G89
   5. G83 is called when it is not already in effect,
   and no q number is in the block: NCE_Q_WORD_MISSING_WITH_G83
   6. G87 is called when it is not already in effect,
   and any of the i number, j number, or k number is missing:
   NCE_I_WORD_MISSING_WITH_G87
   NCE_J_WORD_MISSING_WITH_G87
   NCE_K_WORD_MISSING_WITH_G87
   7. the G code is not between G_81 and G_89.
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED

   Side effects:
   A number of moves are made to execute a canned cycle.

   Called by: convert_cycle

   See the documentation of convert_cycle_xy. This function is entirely
   similar. In this function aa represents z, bb represents x, and cc
   represents y.

   The CYCLE_MACRO is defined just before the convert_cycle_xy function.

   Tool length offsets work only when the tool axis is parallel to the
   Z-axis, so if this function is used, tool length offsets should be
   turned off, and the NC code written to take tool length into account.

   It is a little distracting that this function uses zx in some places
   and xz in others; uniform use of zx would be nice, since that is the
   order for a right-handed coordinate system. Also with that usage,
   permutation of the symbols x, y, and z would allow for automatically
   converting the convert_cycle_xy function (or convert_cycle_yz) into
   the convert_cycle_xz function. However, the canonical interface uses
   CANON_PLANE_XZ.

*/

func (cnc *rs274ngc_t) convert_cycle_zx( /* ARGUMENTS                                 */
	motion inc.GCodes) inc.STATUS { /* a g-code between G_81 and G_89, a canned cycle */

	//        static char name[] = "convert_cycle_zx";
	var (
		aa,
		aa_increment,
		bb,
		bb_increment,
		cc,
		clear_cc,
		i,
		j,
		k,
		old_cc float64
		plane CANON_PLANE
		r     float64
		//repeat    int
		save_mode CANON_MOTION_MODE
	)

	//
	plane = CANON_PLANE_XZ
	if cnc._setup.motion_mode != motion {
		if cnc._setup.block1.y_flag == OFF {
			return inc.NCE_Y_VALUE_UNSPECIFIED_IN_XZ_PLANE_CANNED_CYCLE
		}

	}
	cnc._setup.block1.y_number =
		inc.If(cnc._setup.block1.y_flag == ON, cnc._setup.block1.y_number, cnc._setup.cycle.cc).(float64)
	old_cc = cnc._setup.current.Y

	if cnc._setup.distance_mode == inc.MODE_ABSOLUTE {
		aa_increment = 0.0
		bb_increment = 0.0
		r = cnc._setup.block1.r_number
		cc = cnc._setup.block1.y_number
		aa = inc.If(cnc._setup.block1.z_flag == ON, cnc._setup.block1.z_number, cnc._setup.current.Z).(float64)
		bb = inc.If(cnc._setup.block1.x_flag == ON, cnc._setup.block1.x_number, cnc._setup.current.X).(float64)
	} else if cnc._setup.distance_mode == inc.MODE_INCREMENTAL {
		aa_increment = cnc._setup.block1.z_number
		bb_increment = cnc._setup.block1.x_number
		r = (cnc._setup.block1.r_number + old_cc)
		cc = (r + cnc._setup.block1.y_number) /* [NCMS, page 98] */
		aa = cnc._setup.current.Z
		bb = cnc._setup.current.X
	} else {
		return inc.NCE_BUG_DISTANCE_MODE_NOT_G90_OR_G91

	}
	if r < cc {
		return inc.NCE_R_LESS_THAN_Y_IN_CYCLE_IN_XZ_PLANE
	}

	if old_cc < r {
		cnc.canon.STRAIGHT_TRAVERSE(cnc._setup.current.X, r, cnc._setup.current.Z,
			cnc._setup.current.A, cnc._setup.current.B, cnc._setup.current.C)
		old_cc = r
	}
	clear_cc = inc.If(cnc._setup.retract_mode == inc.R_PLANE, r, old_cc).(float64)

	save_mode = cnc.canon.GET_EXTERNAL_MOTION_CONTROL_MODE()
	if save_mode != CANON_EXACT_PATH {
		cnc.canon.SET_MOTION_CONTROL_MODE(CANON_EXACT_PATH)

	}

	switch motion {
	case inc.G_81:
		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g81(CANON_PLANE_XZ, aa, bb, clear_cc, cc))
			old_cc = clear_cc
		}

		break
	case inc.G_82:
		if (cnc._setup.motion_mode != inc.G_82) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G82
		}

		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g82(CANON_PLANE_XZ, aa, bb, clear_cc, cc, cnc._setup.block1.p_number))
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	case inc.G_83:
		if (cnc._setup.motion_mode != inc.G_83) && (cnc._setup.block1.q_number == -1.0) {
			return inc.NCE_Q_WORD_MISSING_WITH_G83
		}

		cnc._setup.block1.q_number =
			inc.If(cnc._setup.block1.q_number == -1.0, cnc._setup.cycle.q, cnc._setup.block1.q_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g83(CANON_PLANE_XZ, aa, bb, r, clear_cc, cc, cnc._setup.block1.q_number))
			old_cc = clear_cc
		}

		cnc._setup.cycle.q = cnc._setup.block1.q_number
		break
	case inc.G_84:
		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g84(CANON_PLANE_XZ, aa, bb, clear_cc, cc,
				cnc._setup.spindle_turning, cnc._setup.speed_feed_mode))
			old_cc = clear_cc
		}

		break
	case inc.G_85:
		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g85(CANON_PLANE_XZ, aa, bb, clear_cc, cc))
			old_cc = clear_cc
		}

		break
	case inc.G_86:
		if (cnc._setup.motion_mode != inc.G_86) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G86
		}

		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g86(CANON_PLANE_XZ, aa, bb, clear_cc, cc,
				cnc._setup.block1.p_number, cnc._setup.spindle_turning))
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	case inc.G_87:
		if cnc._setup.motion_mode != inc.G_87 {
			if cnc._setup.block1.i_flag == OFF {
				return inc.NCE_I_WORD_MISSING_WITH_G87
			}
			if cnc._setup.block1.j_flag == OFF {
				return inc.NCE_J_WORD_MISSING_WITH_G87
			}
			if cnc._setup.block1.k_flag == OFF {
				return inc.NCE_K_WORD_MISSING_WITH_G87
			}
		}
		i = inc.If(cnc._setup.block1.i_flag == ON, cnc._setup.block1.i_number, cnc._setup.cycle.i).(float64)
		j = inc.If(cnc._setup.block1.j_flag == ON, cnc._setup.block1.j_number, cnc._setup.cycle.j).(float64)
		k = inc.If(cnc._setup.block1.k_flag == ON, cnc._setup.block1.k_number, cnc._setup.cycle.k).(float64)
		cnc._setup.cycle.i = i
		cnc._setup.cycle.j = j
		cnc._setup.cycle.k = k
		if cnc._setup.distance_mode == inc.MODE_INCREMENTAL {
			j = (cc + j) /* j always absolute in function call below */
		}
		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g87(CANON_PLANE_XZ, aa, (aa + k), bb,
				(bb + i), r, clear_cc, j, cc, cnc._setup.spindle_turning))
			old_cc = clear_cc
		}

		break
	case inc.G_88:
		if (cnc._setup.motion_mode != inc.G_88) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G88
		}

		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g88(CANON_PLANE_XZ, aa, bb, cc,
				cnc._setup.block1.p_number, cnc._setup.spindle_turning))
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	case inc.G_89:
		if (cnc._setup.motion_mode != inc.G_89) && (cnc._setup.block1.p_number == -1.0) {
			return inc.NCE_DWELL_TIME_P_WORD_MISSING_WITH_G89
		}

		cnc._setup.block1.p_number =
			inc.If(cnc._setup.block1.p_number == -1.0, cnc._setup.cycle.p, cnc._setup.block1.p_number).(float64)

		for repeat := cnc._setup.block1.l_number; repeat > 0; repeat-- {
			aa = (aa + aa_increment)
			bb = (bb + bb_increment)
			cnc.cycle_traverse(plane, aa, bb, old_cc)
			if old_cc != r {
				cnc.cycle_traverse(plane, aa, bb, r)
			}
			//CHP(call);
			(cnc.convert_cycle_g89(CANON_PLANE_XZ, aa, bb, clear_cc, cc,
				cnc._setup.block1.p_number))
			old_cc = clear_cc
		}

		cnc._setup.cycle.p = cnc._setup.block1.p_number
		break
	default:
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	cnc._setup.current.Z = aa /* CYCLE_MACRO updates aa and bb */
	cnc._setup.current.X = bb
	cnc._setup.current.Y = clear_cc
	cnc._setup.cycle.cc = cnc._setup.block1.y_number

	if save_mode != CANON_EXACT_PATH {
		cnc.canon.SET_MOTION_CONTROL_MODE(save_mode)
	}

	return inc.RS274NGC_OK
}

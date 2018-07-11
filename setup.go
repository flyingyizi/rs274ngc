package rs274ngc

import (
	"os"

	"github.com/flyingyizi/rs274ngc/inc"
)

type setup_i interface {
	Write_g_codes(block *Block_t) int
	Write_m_codes(block *Block_t) int
	Write_settings() int
}

var _ setup_i = &Setup_t{}

type Setup_t struct {
	axis_offset   inc.CANON_POSITION // g92offset
	current       inc.CANON_POSITION
	origin_offset inc.CANON_POSITION

	active_g_codes     [inc.RS274NGC_ACTIVE_G_CODES]inc.GCodes // array of active G codes
	active_m_codes     [inc.RS274NGC_ACTIVE_M_CODES]int        // array of active M codes
	active_settings    [inc.RS274NGC_ACTIVE_SETTINGS]float64   // array of feed, speed, etc.
	block1             Block_t                                 // parsed next block
	blocktext          string                                  // linetext downcased, white space gone
	control_mode       inc.CANON_MOTION_MODE                   // exact path or cutting mode
	current_slot       int                                     // carousel slot number of current tool
	cutter_comp_radius float64                                 // current cutter compensation radius
	cutter_comp_side   inc.CANON_SIDE                          // current cutter compensation side
	cycle              struct {
		cc float64 // cc-value (normal) for canned cycles
		i  float64 // i-value for canned cycles
		j  float64 // j-value for canned cycles
		k  float64 // k-value for canned cycles
		l  int     // l-value for canned cycles
		p  float64 // p-value (dwell) for canned cycles
		q  float64 // q-value for canned cycles
		r  float64 // r-value for canned cycles
	}
	distance_mode     inc.DISTANCE_MODE // absolute or incremental
	ijk_distance_mode inc.DISTANCE_MODE // absolute or incremental
	feed_mode         inc.FeedMode      // G_93 (inverse time) or G_94 units/min
	feed_override     ON_OFF            // whether feed override is enabled
	feed_rate         float64           // feed rate in current units/min

	filename     string   // name of currently open NC code file
	file_pointer *os.File // file pointer for open NC code file

	coolant struct {
		flood ON_OFF // whether flood coolant is on
		mist  ON_OFF // whether mist coolant is on
	}
	length_offset_index int             // for use with tool length offsets
	length_units        inc.CANON_UNITS // millimeters or inches
	line_length         uint            // length of line last read
	linetext            string          // text of most recent line read
	motion_mode         inc.GCodes      // active G-code for motion
	origin_index        int             // active origin (1=G54 to 9=G59.3)
	parameters          []float64       // system parameters
	percent_flag        ON_OFF          // ON means first line was percent sign

	plane              inc.CANON_PLANE  // active plane, XY-, YZ-, or XZ-plane
	probe_flag         ON_OFF           // flag indicating probing done
	program_x          float64          // program x, used when cutter comp on
	program_y          float64          // program y, used when cutter comp on
	retract_mode       inc.RETRACT_MODE // for cycles, old_z or r_plane
	selected_tool_slot int              // tool slot selected but not active
	sequence_number    int              // sequence number of line last read
	speed              float64          // current spindle speed in rpm
	spindle_mode       inc.SpindleMode
	speed_feed_mode    inc.CANON_SPEED_FEED_MODE                    // independent or synched
	speed_override     ON_OFF                                       // whether speed override is enabled
	spindle_turning    inc.CANON_DIRECTION                          // direction spindle is turning
	tool_length_offset float64                                      // current tool length offset
	tool_max           uint                                         // highest number tool slot in carousel
	tool_table         [inc.CANON_TOOL_MAX + 1]inc.CANON_TOOL_TABLE // index is slot number
	tool_table_index   int                                          // tool index used with cutter comp
	traverse_rate      float64                                      // rate for traverse motions

}

/* write_g_codes

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The active_g_codes in the settings are updated.

   Called by:
   rs274ngc_execute
   rs274ngc_init

   The block may be NULL.

   This writes active g_codes into the settings->active_g_codes array by
   examining the interpreter settings. The array of actives is composed
   of ints, so (to handle codes like 59.1) all g_codes are reported as
   ints ten times the actual value. For example, 59.1 is reported as 591.

   The correspondence between modal groups and array indexes is as follows
   (no apparent logic to it).

   The group 0 entry is taken from the block (if there is one), since its
   codes are not modal.

   group 0  - gez[2]  g4, g10, g28, g30, g53, g92 g92.1, g92.2, g92.3 - misc
   group 1  - gez[1]  g0, g1, g2, g3, g38.2, g80, g81, g82, g83, g84, g85,
   g86, g87, g88, g89 - motion
   group 2  - gez[3]  g17, g18, g19 - plane selection
   group 3  - gez[6]  g90, g91 - distance mode
   group 4  - no such group
   group 5  - gez[7]  g93, g94 - feed rate mode
   group 6  - gez[5]  g20, g21 - units
   group 7  - gez[4]  g40, g41, g42 - cutter radius compensation
   group 8  - gez[9]  g43, g49 - tool length offset
   group 9  - no such group
   group 10 - gez[10] g98, g99 - return mode in canned cycles
   group 11 - no such group
   group 12 - gez[8]  g54, g55, g56, g57, g58, g59, g59.1, g59.2, g59.3
   - coordinate system
   group 13 - gez[11] g61, g61.1, g64 - control mode

*/

func (settings *Setup_t) Write_g_codes(block *Block_t) int { /* pointer to a block of RS274/NGC instructions */

	gez := &settings.active_g_codes

	///todo gez[0] = settings.sequence_number
	gez[1] = settings.motion_mode
	if block == nil {
		gez[2] = -1
	} else {
		gez[2] = block.g_modes[GCodeMisc]
	}

	gez[3] =
		inc.If(settings.plane == inc.CANON_PLANE_XY, inc.G_17,
			inc.If(settings.plane == inc.CANON_PLANE_XZ, inc.G_18, inc.G_19).(inc.GCodes)).(inc.GCodes)
	gez[4] =
		inc.If(settings.cutter_comp_side == inc.CANON_SIDE_RIGHT, inc.G_42,
			inc.If(settings.cutter_comp_side == inc.CANON_SIDE_LEFT, inc.G_41, inc.G_40).(inc.GCodes)).(inc.GCodes)
	gez[5] =
		inc.If(settings.length_units == inc.CANON_UNITS_INCHES, inc.G_20, inc.G_21).(inc.GCodes)
	gez[6] =
		inc.If(settings.distance_mode == inc.MODE_ABSOLUTE, inc.G_90, inc.G_91).(inc.GCodes)
	gez[7] =
		inc.If(settings.feed_mode == inc.INVERSE_TIME, inc.G_93, inc.G_94).(inc.GCodes)
	gez[8] =
		inc.If(settings.origin_index < 7, (530 + (10 * settings.origin_index)),
			(584 + settings.origin_index)).(inc.GCodes)
	gez[9] =
		inc.If(settings.tool_length_offset == 0.0, inc.G_49, inc.G_43).(inc.GCodes)
	gez[10] =
		inc.If(settings.retract_mode == inc.OLD_Z, inc.G_98, inc.G_99).(inc.GCodes)
	gez[11] =
		inc.If(settings.control_mode == inc.CANON_CONTINUOUS, inc.G_64,
			inc.If(settings.control_mode == inc.CANON_EXACT_PATH, inc.G_61, inc.G_61_1).(inc.GCodes)).(inc.GCodes)

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* write_m_codes

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The settings->active_m_codes are updated.

   Called by:
   rs274ngc_execute
   rs274ngc_init

   This is testing only the feed override to see if overrides is on.
   Might add check of speed override.

*/
func (settings *Setup_t) Write_m_codes(block *Block_t) int { /* pointer to a block of RS274/NGC instructions */
	emz := settings.active_m_codes
	emz[0] = settings.sequence_number /* 0 seq number  */

	/* 1 stopping    */
	if block == nil {
		emz[1] = -1
	} else {
		emz[1] = block.m_modes[4]
	}

	emz[2] = inc.If(settings.spindle_turning == inc.CANON_STOPPED, 5,
		inc.If(settings.spindle_turning == inc.CANON_CLOCKWISE, 3, 4).(int)).(int) /* 2 spindle     */

	/* 3 tool change */
	if block == nil {
		emz[3] = -1
	} else {
		emz[3] = block.m_modes[6]
	}

	emz[4] = inc.If(settings.coolant.mist == ON, 7,
		inc.If(settings.coolant.flood == ON, -1, 9).(int)).(int) /* 4 mist        */
	emz[5] = inc.If(settings.coolant.flood == ON, 8, -1).(int)  /* 5 flood       */
	emz[6] = inc.If(settings.feed_override == ON, 48, 49).(int) /* 6 overrides   */

	return inc.RS274NGC_OK
}

/* write_settings

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The settings.active_settings array of doubles is updated with the
   sequence number, feed, and speed settings.

   Called by:
   rs274ngc_execute
   rs274ngc_init

*/
func (settings *Setup_t) Write_settings() int {
	//double * vals

	settings.active_settings[0] = float64(settings.sequence_number) /* 0 sequence number */
	settings.active_settings[1] = settings.feed_rate                /* 1 feed rate       */
	settings.active_settings[2] = settings.speed                    /* 2 spindle speed   */

	return inc.RS274NGC_OK
}

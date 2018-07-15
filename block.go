package rs274ngc

import (
	"math"
	"strconv"
	"strings"

	"github.com/flyingyizi/rs274ngc/inc"

	"github.com/flyingyizi/rs274ngc/ops"
)

/*
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
var _gees map[int]int = map[int]int{ /*key:code, value:group*/
	40: 0, 100: 0, 280: 0, 300: 0, 530: 0, 920: 0, 921: 0, 922: 0, 923: 0,
	0: 1, 10: 1, 20: 1, 30: 1, 382: 1, 800: 1, 810: 1, 820: 1, 830: 1, 840: 1, 850: 1,
	170: 2, 180: 2, 190: 2,
	900: 3, 910: 3,
	930: 5, 940: 5,
	200: 6, 210: 6,
	400: 7, 410: 7, 420: 7,
	430: 8, 490: 8,
	980: 10, 990: 10,
	540: 12, 550: 12, 560: 12, 570: 12, 580: 12, 590: 12, 591: 12, 592: 12, 593: 12,

	610: 13, 611: 13, 640: 13}

type GModalGroup int

const (
	GCodeMisc                     GModalGroup = 0
	GCodeMotion                               = 1
	GCodePlaneSelection                       = 2
	GCodeDistance                             = 3
	GCodeFeedRateMode                         = 5
	GCodeUnit                                 = 6
	GCodeCutterRadiusCompensation             = 7
	GCodeToolLengthOffset                     = 8
	GCodeReturnModeInCannedCycle              = 10
	GCodeCoordinateSystem                     = 12
	GCodeControlMode                          = 13
	// num of gcode modal group
	GModalGroupLen = 14
	//MCodeStoping              = 4
	//MCodeToolChange           = 6
	//MCodeSpindleTurning       = 7
	//MCodeCoolant              = 8
	//MCodeSpeedOverrideSwitch = 9
)

/*

   Modal groups and modal group numbers for M codes are not described in
   [Fanuc]. We have used the groups from [NCMS] and added M60, as an
   extension of the language for pallet shuttle and stop. This version has
   no codes related to axis clamping.

   The groups are:
   group 4 = {m0,m1,m2,m30,m60} - stopping
   group 6 = {m6}               - tool change
   group 7 = {m3,m4,m5}         - spindle turning
   group 8 = {m7,m8,m9}         - coolant
   group 9 = {m48,m49}          - feed and speed override switch bypass

*/
var _ems map[int]int = map[int]int{ /*key:code, value:group*/
	0: 4, 1: 4, 2: 4, 30: 4, 60: 4,
	6: 6,
	3: 7, 4: 7, 5: 7,
	7: 8, 8: 8, 9: 8,
	48: 9, 49: 9}

type ON_OFF bool

// on-off switch settings
const (
	OFF ON_OFF = false
	ON  ON_OFF = true
)

type block_i interface {
	Init_block() int
	Enhance_block(settings *Setup_t) int
	check_g_codes(settings *Setup_t) int
	Check_items(settings *Setup_t) int
	check_m_codes() int
	check_other_codes() int

	Read_items(tool_max uint, line string, parameters []float64) inc.STATUS
}

var _ block_i = &Block_t{}

type Block_t struct {
	a_flag   ON_OFF
	a_number float64

	b_flag   ON_OFF
	b_number float64

	c_flag   ON_OFF
	c_number float64

	comment  string
	d_number int
	f_number float64
	// g_modes array in the block keeps track of which G modal groups are used on a line of code
	g_modes  [GModalGroupLen]inc.GCodes
	h_number int

	i_flag   ON_OFF
	i_number float64
	j_flag   ON_OFF
	j_number float64
	k_flag   ON_OFF
	k_number float64

	l_number     int
	line_number  int
	motion_to_be inc.GCodes
	m_count      int
	m_modes      [10]int
	p_number     float64
	q_number     float64
	r_flag       ON_OFF
	r_number     float64
	s_number     float64
	t_number     int
	x_flag       ON_OFF
	x_number     float64
	y_flag       ON_OFF
	y_number     float64
	z_flag       ON_OFF
	z_number     float64

	Parameter_occurrence int64       // parameter buffer index
	Parameter_numbers    [50]int     // parameter number buffer
	Parameter_values     [50]float64 // parameter value buffer

}

/****************************************************************************/

/* init_block

   Returned Value: int (RS274NGC_OK)

   Side effects:
   Values in the block are reset as described below.

   Called by: parse_line

   This system reuses the same block over and over, rather than building
   a new one for each line of NC code. The block is re-initialized before
   each new line of NC code is read.

   The block contains many slots for values which may or may not be present
   on a line of NC code. For some of these slots, there is a flag which
   is turned on (at the time time value of the slot is read) if the item
   is present.  For slots whose values are to be read which do not have a
   flag, there is always some excluded range of values. Setting the
   initial value of these slot to some number in the excluded range
   serves to show that a value for that slot has not been read.

   The rules for the indicators for slots whose values may be read are:
   1. If the value may be an arbitrary real number (which is always stored
   internally as a double), a flag is needed to indicate if a value has
   been read. All such flags are initialized to OFF.
   Note that the value itself is not initialized; there is no point in it.
   2. If the value must be a non-negative real number (which is always stored
   internally as a double), a value of -1.0 indicates the item is not present.
   3. If the value must be an unsigned integer (which is always stored
   internally as an int), a value of -1 indicates the item is not present.
   (RS274/NGC does not use any negative integers.)
   4. If the value is a character string (only the comment slot is one), the
   first character is set to 0 (NULL).

*/

func (block *Block_t) Init_block() int {
	//int n;

	block.a_flag = OFF /*AA*/
	block.b_flag = OFF /*BB*/
	block.c_flag = OFF /*CC*/
	block.comment = ""
	block.d_number = -1
	block.f_number = -1.0
	for n := 0; n < GModalGroupLen; n++ {
		block.g_modes[n] = -1
	}

	block.h_number = -1
	block.i_flag = OFF
	block.j_flag = OFF
	block.k_flag = OFF
	block.l_number = -1
	block.line_number = -1
	block.motion_to_be = -1
	block.m_count = 0
	for n := 0; n < 10; n++ {
		block.m_modes[n] = -1
	}
	block.p_number = -1.0
	block.q_number = -1.0
	block.r_flag = OFF
	block.s_number = -1.0
	block.t_number = -1
	block.x_flag = OFF
	block.y_flag = OFF
	block.z_flag = OFF

	return inc.RS274NGC_OK
}

/* enhance_block

   Returned Value:
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. A g80 is in the block, no modal group 0 code that uses axes
   is in the block, and one or more axis values is given:
   NCE_CANNOT_USE_AXIS_VALUES_WITH_G80
   2. A g92 is in the block and no axis value is given:
   NCE_ALL_AXES_MISSING_WITH_G92
   3. One g-code from group 1 and one from group 0, both of which can use
   axis values, are in the block:
   NCE_CANNOT_USE_TWO_G_CODES_THAT_BOTH_USE_AXIS_VALUES
   4. A g-code from group 1 which can use axis values is in the block,
   but no axis value is given: NCE_ALL_AXES_MISSING_WITH_MOTION_CODE
   5. Axis values are given, but there is neither a g-code in the block
   nor an active previously given modal g-code that uses axis values:
   NCE_CANNOT_USE_AXIS_VALUES_WITHOUT_A_G_CODE_THAT_USES_THEM

   Side effects:
   The value of motion_to_be in the block is set.

   Called by: parse_line

   If there is a g-code for motion in the block (in g_modes[1]),
   set motion_to_be to that. Otherwise, if there is an axis value in the
   block and no g-code to use it (any such would be from group 0 in
   g_modes[0]), set motion_to_be to be the last motion saved (in
   settings->motion mode).

   This also make the checks described above.

*/

func (block *Block_t) Enhance_block(settings *Setup_t) int { /* pointer to machine settings       */

	//name := "enhance_block"
	var (
		axis_flag             ON_OFF
		mode_zero_covets_axes bool
	)

	axis_flag = ((block.x_flag == ON) ||
		(block.y_flag == ON) ||
		(block.a_flag == ON) || /*AA*/
		(block.b_flag == ON) || /*BB*/
		(block.c_flag == ON) || /*CC*/
		(block.z_flag == ON))
	mode_zero_covets_axes = ((block.g_modes[GCodeMisc] == inc.G_10) ||
		(block.g_modes[GCodeMisc] == inc.G_28) ||
		(block.g_modes[GCodeMisc] == inc.G_30) ||
		(block.g_modes[GCodeMisc] == inc.G_92))

	if block.g_modes[GCodeMotion] != -1 {
		if block.g_modes[GCodeMotion] == inc.G_80 {
			if axis_flag == ON && (!mode_zero_covets_axes) {
				return inc.NCE_CANNOT_USE_AXIS_VALUES_WITH_G80
			}

			if (!axis_flag) && (block.g_modes[GCodeMisc] == inc.G_92) {
				return inc.NCE_ALL_AXES_MISSING_WITH_G92
			}
		} else {
			if mode_zero_covets_axes {
				return inc.NCE_CANNOT_USE_TWO_G_CODES_THAT_BOTH_USE_AXIS_VALUES
			}
			if !axis_flag {
				return inc.NCE_ALL_AXES_MISSING_WITH_MOTION_CODE
			}
		}
		block.motion_to_be = block.g_modes[GCodeMotion]
	} else if mode_zero_covets_axes { /* other 3 can get by without axes but not G92 */
		if (!axis_flag) && (block.g_modes[GCodeMisc] == inc.G_92) {
			return inc.NCE_ALL_AXES_MISSING_WITH_G92
		}

	} else if axis_flag {
		if (settings.motion_mode == -1) || (settings.motion_mode == inc.G_80) {
			return inc.NCE_CANNOT_USE_AXIS_VALUES_WITHOUT_A_G_CODE_THAT_USES_THEM
		}
		block.motion_to_be = settings.motion_mode
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* check_g_codes

   Returned Value: int
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. NCE_DWELL_TIME_MISSING_WITH_G4
   2. NCE_MUST_USE_G0_OR_G1_WITH_G53
   3. NCE_CANNOT_USE_G53_INCREMENTAL
   4. NCE_LINE_WITH_G10_DOES_NOT_HAVE_L2
   5. NCE_P_VALUE_NOT_AN_INTEGER_WITH_G10_L2
   6. NCE_P_VALUE_OUT_OF_RANGE_WITH_G10_L2
   7. NCE_BUG_BAD_G_CODE_MODAL_GROUP_0

   Side effects: none

   Called by: check_items

   This runs checks on g_codes from a block of RS274/NGC instructions.
   Currently, all checks are on g_codes in modal group 0.

   The read_g function checks for errors which would foul up the reading.
   The enhance_block function checks for logical errors in the use of
   axis values by g-codes in modal groups 0 and 1.
   This function checks for additional logical errors in g_codes.

   [Fanuc, page 45, note 4] says there is no maximum for how many g_codes
   may be put on the same line, [NCMS] says nothing one way or the other,
   so the test for that is not used.

   We are suspending any implicit motion g_code when a g_code from our
   group 0 is used.  The implicit motion g_code takes effect again
   automatically after the line on which the group 0 g_code occurs.  It
   is not clear what the intent of [Fanuc] is in this regard. The
   alternative is to require that any implicit motion be explicitly
   cancelled.

   Not all checks on g_codes are included here. Those checks that are
   sensitive to whether other g_codes on the same line have been executed
   yet are made by the functions called by convert_g.

   Our reference sources differ regarding what codes may be used for
   dwell time.  [Fanuc, page 58] says use "p" or "x". [NCMS, page 23] says
   use "p", "x", or "u". We are allowing "p" only, since it is consistent
   with both sources and "x" would be confusing. However, "p" is also used
   with G10, where it must be an integer, so reading "p" values is a bit
   more trouble than would be nice.

*/

func (block *Block_t) check_g_codes(settings *Setup_t) int { /* pointer to machine settings      */
	//static char name[] SET_TO "check_g_codes";
	var (
		mode0 inc.GCodes
		p_int int
	)

	mode0 = block.g_modes[0]

	if mode0 == -1 {

	} else if mode0 == inc.G_4 {
		if block.p_number == -1.0 {
			return inc.NCE_DWELL_TIME_MISSING_WITH_G4
		}
	} else if mode0 == inc.G_10 {
		p_int = (int)(block.p_number + 0.0001)
		if block.l_number != 2 {
			return inc.NCE_LINE_WITH_G10_DOES_NOT_HAVE_L2
		}
		if ((block.p_number + 0.0001) - (float64)(p_int)) > 0.0002 {
			return inc.NCE_P_VALUE_NOT_AN_INTEGER_WITH_G10_L2
		}

		if (p_int < 1) || (p_int > 9) {
			return inc.NCE_P_VALUE_OUT_OF_RANGE_WITH_G10_L2
		}
	} else if mode0 == inc.G_28 {

	} else if mode0 == inc.G_30 {

	} else if mode0 == inc.G_53 {
		if (block.motion_to_be != inc.G_0) && (block.motion_to_be != inc.G_1) {
			return inc.NCE_MUST_USE_G0_OR_G1_WITH_G53
		}
		if (block.g_modes[GCodeDistance] == inc.G_91) ||
			((block.g_modes[GCodeDistance] != inc.G_90) &&
				(settings.distance_mode == inc.MODE_INCREMENTAL)) {
			return inc.NCE_CANNOT_USE_G53_INCREMENTAL
		}
	} else if mode0 == inc.G_92 {

	} else if (mode0 == inc.G_92_1) || (mode0 == inc.G_92_2) || (mode0 == inc.G_92_3) {

	} else {
		return inc.NCE_BUG_BAD_G_CODE_MODAL_GROUP_0
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/
/* check_items

   Returned Value: int
   If any one of check_g_codes, check_m_codes, and check_other_codes
   returns an error code, this returns that code.
   Otherwise, it returns RS274NGC_OK.

   Side effects: none

   Called by: parse_line

   This runs checks on a block of RS274 code.

   The functions named read_XXXX check for errors which would foul up the
   reading. This function checks for additional logical errors.

   A block has an array of g_codes, which are initialized to -1
   (meaning no code). This calls check_g_codes to check the g_codes.

   A block has an array of m_codes, which are initialized to -1
   (meaning no code). This calls check_m_codes to check the m_codes.

   Items in the block which are not m or g codes are checked by
   check_other_codes.

*/
func (block *Block_t) Check_items(settings *Setup_t) int {
	//static char name[] SET_TO "check_items";

	block.check_g_codes(settings)
	block.check_m_codes()
	block.check_other_codes()
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* check_m_codes

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. There are too many m codes in the block: NCE_TOO_MANY_M_CODES_ON_LINE

   Side effects: none

   Called by: check_items

   This runs checks on m_codes from a block of RS274/NGC instructions.

   The read_m function checks for errors which would foul up the
   reading. This function checks for additional errors in m_codes.

*/
func (block *Block_t) check_m_codes() int {
	//static char name[] SET_TO "check_m_codes";

	// max number of m codes on one line
	const MAX_EMS = 4

	if block.m_count > MAX_EMS {
		return inc.NCE_TOO_MANY_M_CODES_ON_LINE
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* check_other_codes

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. An A-axis value is given with a canned cycle (g80 to g89):
   NCE_CANNOT_PUT_AN_A_IN_CANNED_CYCLE
   2. A B-axis value is given with a canned cycle (g80 to g89):
   NCE_CANNOT_PUT_A_B_IN_CANNED_CYCLE
   3. A C-axis value is given with a canned cycle (g80 to g89):
   NCE_CANNOT_PUT_A_C_IN_CANNED_CYCLE
   4. A d word is in a block with no cutter_radius_compensation_on command:
   NCE_D_WORD_WITH_NO_G41_OR_G42
   5. An h_number is in a block with no tool length offset setting:
   NCE_H_WORD_WITH_NO_G43
   6. An i_number is in a block with no G code that uses it:
   NCE_I_WORD_WITH_NO_G2_OR_G3_OR_G87_TO_USE_IT
   7. A j_number is in a block with no G code that uses it:
   NCE_J_WORD_WITH_NO_G2_OR_G3_OR_G87_TO_USE_IT
   8. A k_number is in a block with no G code that uses it:
   NCE_K_WORD_WITH_NO_G2_OR_G3_OR_G87_TO_USE_IT
   9. A l_number is in a block with no G code that uses it:
   NCE_L_WORD_WITH_NO_CANNED_CYCLE_OR_G10
   10. A p_number is in a block with no G code that uses it:
   NCE_P_WORD_WITH_NO_G4_G10_G82_G86_G88_G89
   11. A q_number is in a block with no G code that uses it:
   NCE_Q_WORD_WITH_NO_G83
   12. An r_number is in a block with no G code that uses it:
   NCE_R_WORD_WITH_NO_G_CODE_THAT_USES_IT

   Side effects: none

   Called by: check_items

   This runs checks on codes from a block of RS274/NGC code which are
   not m or g codes.

   The functions named read_XXXX check for errors which would foul up the
   reading. This function checks for additional logical errors in codes.

*/
func (block *Block_t) check_other_codes() int {
	//static char name[] SET_TO "check_other_codes";
	var (
		motion inc.GCodes
	)

	motion = block.motion_to_be

	if block.a_flag != OFF {
		if (block.g_modes[GCodeMotion] > inc.G_80) && (block.g_modes[GCodeMotion] < inc.G_90) {
			return inc.NCE_CANNOT_PUT_AN_A_IN_CANNED_CYCLE
		}
	}
	if block.b_flag != OFF {
		if (block.g_modes[GCodeMotion] > inc.G_80) && (block.g_modes[GCodeMotion] < inc.G_90) {
			return inc.NCE_CANNOT_PUT_A_B_IN_CANNED_CYCLE
		}
	}
	if block.c_flag != OFF {
		if (block.g_modes[GCodeMotion] > inc.G_80) && (block.g_modes[GCodeMotion] < inc.G_90) {
			return inc.NCE_CANNOT_PUT_A_C_IN_CANNED_CYCLE
		}
	}
	if block.d_number != -1 {
		if (block.g_modes[GCodeCutterRadiusCompensation] != inc.G_41) &&
			(block.g_modes[GCodeCutterRadiusCompensation] != inc.G_42) {
			return inc.NCE_D_WORD_WITH_NO_G41_OR_G42
		}

	}
	if block.h_number != -1 {
		if block.g_modes[GCodeToolLengthOffset] != inc.G_43 {
			return inc.NCE_H_WORD_WITH_NO_G43
		}
	}
	if block.i_flag == ON { /* could still be useless if yz_plane arc */

		if (motion != inc.G_2) && (motion != inc.G_3) && (motion != inc.G_87) {
			return inc.NCE_I_WORD_WITH_NO_G2_OR_G3_OR_G87_TO_USE_IT
		}

	}
	if block.j_flag == ON { /* could still be useless if xz_plane arc */

		if (motion != inc.G_2) && (motion != inc.G_3) && (motion != inc.G_87) {
			return inc.NCE_J_WORD_WITH_NO_G2_OR_G3_OR_G87_TO_USE_IT
		}

	}
	if block.k_flag == ON { /* could still be useless if xy_plane arc */

		if (motion != inc.G_2) && (motion != inc.G_3) && (motion != inc.G_87) {
			return inc.NCE_K_WORD_WITH_NO_G2_OR_G3_OR_G87_TO_USE_IT
		}

	}
	if block.l_number != -1 {
		if ((motion < inc.G_81) || (motion > inc.G_89)) &&
			(block.g_modes[GCodeMisc] != inc.G_10) {
			return inc.NCE_L_WORD_WITH_NO_CANNED_CYCLE_OR_G10
		}

	}
	if block.p_number != -1.0 {
		if (block.g_modes[GCodeMisc] != inc.G_10) &&
			(block.g_modes[GCodeMisc] != inc.G_4) &&
			(motion != inc.G_82) && (motion != inc.G_86) &&
			(motion != inc.G_88) && (motion != inc.G_89) {
			return inc.NCE_P_WORD_WITH_NO_G4_G10_G82_G86_G88_G89
		}
	}
	if block.q_number != -1.0 {
		if motion != inc.G_83 {
			return inc.NCE_Q_WORD_WITH_NO_G83
		}
	}
	if block.r_flag == ON {
		if ((motion != inc.G_2) && (motion != inc.G_3)) &&
			((motion < inc.G_81) || (motion > inc.G_89)) {
			return inc.NCE_R_WORD_WITH_NO_G_CODE_THAT_USES_IT
		}
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_items

   Returned Value: int
   If read_line_number or read_one_item returns an error code,
   this returns that code.
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   One line of RS274 code is read and data inserted into a block.
   The counter which is passed around among the readers is initialized.
   System parameters may be reset.

   Called by: parse_line

*/

func (block *Block_t) Read_items( /* ARGUMENTS                                      */
	tool_max uint,
	line string, /* string: line of RS274/NGC code being processed */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	length := len(line)
	counter := 0
	s := inc.RS274NGC_OK

	l := []byte(line)

	if l[counter] == '/' { /* skip the slash character if first */
		counter++
	}
	if l[counter] == 'n' {
		if s = block.read_line_number(l, &counter); s != inc.RS274NGC_OK {
			return s
		}
	}
	for counter < length {
		///////////////////////
		switch line[counter] {
		case '#':
			block.read_parameter_setting(l, &counter, parameters)
			break
		case '(':
			block.read_comment(l, &counter, parameters)
			break
		case 'a': //A A-axis of machine
			block.read_a(l, &counter, parameters)
			break
		case 'b':
			block.read_b(l, &counter, parameters)
			break
		case 'c':
			block.read_c(l, &counter, parameters)
			break
		case 'd':
			block.read_d(l, &counter, parameters)
			break
		case 'f':
			block.read_f(l, &counter, parameters)
			break
		case 'g':
			block.read_g(l, &counter, parameters)
			break
		case 'h':
			block.read_h(tool_max, l, &counter, parameters)
			break
		case 'i':
			block.read_i(l, &counter, parameters)
			break
		case 'j':
			block.read_j(l, &counter, parameters)
			break
		case 'k':
			block.read_k(l, &counter, parameters)
			break
		case 'l':
			block.read_l(l, &counter, parameters)
			break
		case 'm':
			block.read_m(l, &counter, parameters)
			break
		case 'p':
			block.read_p(l, &counter, parameters)
			break
		case 'q':
			block.read_q(l, &counter, parameters)
			break
		case 'r':
			block.read_r(l, &counter, parameters)
			break
		case 's':
			block.read_s(l, &counter, parameters)
			break
		case 't':
			block.read_t(l, &counter, parameters)
			break
		case 'x':
			block.read_x(l, &counter, parameters)
			break
		case 'y':
			block.read_y(l, &counter, parameters)
			break
		case 'z':
			block.read_z(l, &counter, parameters)
			break
		default:
			return inc.NCE_BAD_CHARACTER_USED
		}

		///////////////////////
	}

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_line_number

   Returned Value: int
   If read_integer_unsigned returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not n:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. The line number is too large (more than 99999):
   NCE_LINE_NUMBER_GREATER_THAN_99999

   Side effects:
   counter is reset to the character following the line number.
   A line number is inserted in the block.

   Called by: read_items

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'n', indicating a line number.
   The function reads characters which give the (integer) value of the
   line number.

   Note that extra initial zeros in a line number will not cause the
   line number to be too large.

*/

func (block *Block_t) read_line_number( /* ARGUMENTS                               */
	line []byte, /* string: line of RS274    code being processed  */
	counter *int) inc.STATUS { /* pointer to a counter for position on the line  */

	if line[*counter] != 'n' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if value, s := block.read_integer_unsigned(line, counter); s != inc.RS274NGC_OK {
		return s
	} else if value > 99999 {
		return inc.NCE_LINE_NUMBER_GREATER_THAN_99999
	} else {
		block.line_number = int(value)
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_integer_unsigned

   Returned Value: int
   If any of the following errors occur, this returns the error shown.
   Otherwise, RS274NGC_OK is returned.
   1. The first character is not a digit: NCE_BAD_FORMAT_UNSIGNED_INTEGER
   2. sscanf fails: NCE_SSCANF_FAILED

   Side effects:
   The number read from the line is put into what integer_ptr points at.

   Called by: read_line_number

   This reads an explicit unsigned (positive) integer from a string,
   starting from the position given by *counter. It expects to find one
   or more digits. Any character other than a digit terminates reading
   the integer. Note that if the first character is a sign (+ or -),
   an error will be reported (since a sign is not a digit).

*/

func (block *Block_t) read_integer_unsigned( /* ARGUMENTS                       */
	line []byte, /* string: line of RS274 code being processed    */
	counter *int /* pointer to a counter for position on the line */) (vaule uint64, s inc.STATUS) { /*  the value being read               */

	//static char name[] SET_TO "read_integer_unsigned";
	//int n;
	var (
		temp string
		err  error
	)

	for i, v := range line[*counter:] {
		if (v < 48 /*'0'*/) || (v > 57 /*'9'*/) {
			if i == 0 {
				s = inc.NCE_BAD_FORMAT_UNSIGNED_INTEGER
				return
			}
			temp = string(line[*counter : *counter+i])

			break
		}
	}

	if vaule, err = strconv.ParseUint(temp, 10, 64); err != nil {
		s = inc.NCE_SSCANF_FAILED
		return
	}
	*counter = *counter + len(temp)
	s = inc.RS274NGC_OK
	return
}

/****************************************************************************/

/* read_a

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not a:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. An a_coordinate has already been inserted in the block:
   NCE_MULTIPLE_A_WORDS_ON_ONE_LINE.
   3. A values are not allowed: NCE_CANNOT_USE_A_WORD.

   Side effects:
   counter is reset.
   The a_flag in the block is turned on.
   An a_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'a', indicating an a_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   The counter is then set to point to the character following.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

   If the AA compiler flag is defined, the a_flag in the block is turned
   on and the a_number in the block is set to the value read. If the
   AA flag is not defined, (i) if the AXIS_ERROR flag is defined, that means
   A values are not allowed, and an error value is returned, (ii) if the
   AXIS_ERROR flag is not defined, nothing is done.

*/

func (block *Block_t) read_a( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'a' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.a_flag != OFF {
		return inc.NCE_MULTIPLE_A_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)
	block.a_flag = ON
	block.a_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_b

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not b:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A b_coordinate has already been inserted in the block:
   NCE_MULTIPLE_B_WORDS_ON_ONE_LINE.
   3. B values are not allowed: NCE_CANNOT_USE_B_WORD

   Side effects:
   counter is reset.
   The b_flag in the block is turned on.
   A b_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'b', indicating a b_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   The counter is then set to point to the character following.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

   If the BB compiler flag is defined, the b_flag in the block is turned
   on and the b_number in the block is set to the value read. If the
   BB flag is not defined, (i) if the AXIS_ERROR flag is defined, that means
   B values are not allowed, and an error value is returned, (ii) if the
   AXIS_ERROR flag is not defined, nothing is done.

*/

func (block *Block_t) read_b( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'b' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)

	if !!block.b_flag {
		return inc.NCE_MULTIPLE_B_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)
	block.b_flag = ON
	block.b_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_c

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not c:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. An c_coordinate has already been inserted in the block:
   NCE_MULTIPLE_C_WORDS_ON_ONE_LINE
   3. C values are not allowed: NCE_CANNOT_USE_C_WORD

   Side effects:
   counter is reset.
   The c_flag in the block is turned on.
   A c_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'c', indicating an c_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   The counter is then set to point to the character following.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

   If the CC compiler flag is defined, the c_flag in the block is turned
   on and the c_number in the block is set to the value read. If the
   CC flag is not defined, (i) if the AXIS_ERROR flag is defined, that means
   C values are not allowed, and an error value is returned, (ii) if the
   AXIS_ERROR flag is not defined, nothing is done.

*/
func (block *Block_t) read_c( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'c' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if !!block.c_flag {
		return inc.NCE_MULTIPLE_C_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)
	block.c_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_comment

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not '(' ,
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED

   Side effects:
   The counter is reset to point to the character following the comment.
   The comment string, without parentheses, is copied into the comment
   area of the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character '(', indicating a comment is
   beginning. The function reads characters of the comment, up to and
   including the comment closer ')'.

   It is expected that the format of a comment will have been checked (by
   read_text or read_keyboard_line) and bad format comments will
   have prevented the system from getting this far, so that this function
   can assume a close parenthesis will be found when an open parenthesis
   has been found, and that comments are not nested.

   The "parameters" argument is not used in this function. That argument is
   present only so that this will have the same argument list as the other
   "read_XXX" functions called using a function pointer by read_one_item.

*/
func (block *Block_t) read_comment( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	n := *counter

	if line[n] != '(' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	n++

	for ; line[n] != ')'; n++ {
	}
	block.comment = string(line[*counter : n+1])
	*counter = n + 1

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_d

   Returned Value: int
   If read_integer_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not d:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A d_number has already been inserted in the block:
   NCE_MULTIPLE_D_WORDS_ON_ONE_LINE
   3. The d_number is negative: NCE_NEGATIVE_D_WORD_TOOL_RADIUS_INDEX_USED
   4. The d_number is more than _setup.tool_max: NCE_TOOL_RADIUS_INDEX_TOO_BIG

   Side effects:
   counter is reset to the character following the tool number.
   A d_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'd', indicating an index into a
   table of tool diameters.  The function reads characters which give the
   (integer) value of the index. The value may not be more than
   _setup.tool_max and may not be negative, but it may be zero. The range
   is checked here.

   read_integer_value allows a minus sign, so a check for a negative value
   is made here, and the parameters argument is also needed.

*/

func (block *Block_t) read_d( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value int

	if line[*counter] != 'd' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}

	*counter = (*counter + 1)
	if block.d_number > -1 {
		return inc.NCE_MULTIPLE_D_WORDS_ON_ONE_LINE
	}
	block.read_integer_value(line, counter, &value, parameters)
	if value < 0 {
		return inc.NCE_NEGATIVE_D_WORD_TOOL_RADIUS_INDEX_USED
	}
	block.d_number = int(value)

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_f

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not f:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. An f_number has already been inserted in the block:
   NCE_MULTIPLE_F_WORDS_ON_ONE_LINE
   3. The f_number is negative: NCE_NEGATIVE_F_WORD_USED

   Side effects:
   counter is reset to point to the first character following the f_number.
   The f_number is inserted in block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'f'. The function reads characters
   which tell how to set the f_number, up to the start of the next item
   or the end of the line. This information is inserted in the block.

   The value may be a real number or something that evaluates to a real
   number, so read_real_value is used to read it. Parameters may be
   involved, so the parameters argument is required. The value is always
   a feed rate.

*/
func (block *Block_t) read_f( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */
	var value float64

	if line[*counter] != 'f' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.f_number > -1.0 {
		return inc.NCE_MULTIPLE_F_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)
	if value < 0.0 {
		return inc.NCE_NEGATIVE_F_WORD_USED
	}
	block.f_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_g

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not g:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. The value is negative: NCE_NEGATIVE_G_CODE_USED
   3. The value differs from a number ending in an even tenth by more
   than 0.0001: NCE_G_CODE_OUT_OF_RANGE
   4. The value is greater than 99.9: NCE_G_CODE_OUT_OF_RANGE
   5. The value is not the number of a valid g code: NCE_UNKNOWN_G_CODE_USED
   6. Another g code from the same modal group has already been
   inserted in the block: NCE_TWO_G_CODES_USED_FROM_SAME_MODAL_GROUP

   Side effects:
   counter is reset to the character following the end of the g_code.
   A g code is inserted as the value of the appropriate mode in the
   g_modes array in the block.
   The g code counter in the block is increased by 1.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'g', indicating a g_code.  The
   function reads characters which tell how to set the g_code.

   The RS274/NGC manual [NCMS, page 51] allows g_codes to be represented
   by expressions and provide [NCMS, 71 - 73] that a g_code must evaluate
   to to a number of the form XX.X (59.1, for example). The manual does not
   say how close an expression must come to one of the allowed values for
   it to be legitimate. Is 59.099999 allowed to mean 59.1, for example?
   In the interpreter, we adopt the convention that the evaluated number
   for the g_code must be within 0.0001 of a value of the form XX.X

   To simplify the handling of g_codes, we convert them to integers by
   multiplying by 10 and rounding down or up if within 0.001 of an
   integer. Other functions that deal with g_codes handle them
   symbolically, however. The symbols are defined in rs274NGC.hh
   where G_1 is 10, G_83 is 830, etc.

   This allows any number of g_codes on one line, provided that no two
   are in the same modal group.

*/
func (block *Block_t) read_g( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value_read float64

	if line[*counter] != 'g' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	block.read_real_value(line, counter, &value_read, parameters)
	value_read = (10.0 * value_read)
	value := int(math.Floor(value_read))

	diff := value_read - float64(value)
	if diff > 0.999 {
		value = int(math.Ceil(value_read))
	} else if diff > 0.001 {
		return inc.NCE_G_CODE_OUT_OF_RANGE
	}

	if value > 999 {
		return inc.NCE_G_CODE_OUT_OF_RANGE
	} else if value < 0 {
		return inc.NCE_NEGATIVE_G_CODE_USED
	}

	mode, ok := _gees[value]
	if ok == false {
		return inc.NCE_UNKNOWN_G_CODE_USED
	}
	if block.g_modes[mode] != -1 {
		return inc.NCE_TWO_G_CODES_USED_FROM_SAME_MODAL_GROUP
	}

	block.g_modes[mode] = inc.GCodes(value)

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_h

   Returned Value: int
   If read_integer_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not h:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. An h_number has already been inserted in the block:
   NCE_MULTIPLE_H_WORDS_ON_ONE_LINE
   3. The value is negative: NCE_NEGATIVE_H_WORD_TOOL_LENGTH_OFFSET_INDEX_USED
   4. The value is greater than _setup.tool_max:
   NCE_TOOL_LENGTH_OFFSET_INDEX_TOO_BIG

   Side effects:
   counter is reset to the character following the h_number.
   An h_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'h', indicating a tool length
   offset index.  The function reads characters which give the (integer)
   value of the tool length offset index (not the actual distance of the
   offset).

*/

func (block *Block_t) read_h( /* ARGUMENTS                                      */
	tool_max uint,
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value int

	if line[*counter] != 'h' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.h_number > -1 {
		return inc.NCE_MULTIPLE_H_WORDS_ON_ONE_LINE
	}
	block.read_integer_value(line, counter, &value, parameters)
	if value < 0 {
		return inc.NCE_NEGATIVE_H_WORD_TOOL_LENGTH_OFFSET_INDEX_USED
	}
	h := int(value)

	if h > int(tool_max) {
		return inc.NCE_TOOL_LENGTH_OFFSET_INDEX_TOO_BIG
	}
	block.h_number = h
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_i

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not i:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. An i_coordinate has already been inserted in the block:
   NCE_MULTIPLE_I_WORDS_ON_ONE_LINE

   Side effects:
   counter is reset.
   The i_flag in the block is turned on.
   An i_coordinate setting is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'i', indicating a i_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   This information is inserted in the block. The counter is then set to
   point to the character following.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

*/

func (block *Block_t) read_i( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'i' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}

	*counter = (*counter + 1)
	if !!block.i_flag {
		return inc.NCE_MULTIPLE_I_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)
	block.i_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_real_value

   Returned Value: int
   If one of the following functions returns an error code,
   this returns that code.
   read_real_expression
   read_parameter
   read_unary
   read_real_number
   If no characters are found before the end of the line this
   returns NCE_NO_CHARACTERS_FOUND_IN_READING_REAL_VALUE.
   Otherwise, this returns RS274NGC_OK.

   Side effects:
   The value read from the line is put into what double_ptr points at.
   The counter is reset to point to the first character after the
   characters which make up the value.

   Called by:
   read_a
   read_b
   read_c
   read_f
   read_g
   read_i
   read_integer_value
   read_j
   read_k
   read_p
   read_parameter_setting
   read_q
   read_r
   read_real_expression
   read_s
   read_x
   read_y
   read_z

   This attempts to read a real value out of the line, starting at the
   index given by the counter. The value may be a number, a parameter
   value, a unary function, or an expression. It calls one of four
   other readers, depending upon the first character.

*/

func (block *Block_t) read_real_value( /* ARGUMENTS                               */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	double_ptr *float64, /* pointer to double to be read                   */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	c := line[*counter]
	if c == 0 {
		return inc.NCE_NO_CHARACTERS_FOUND_IN_READING_REAL_VALUE
	}
	if c == '[' {
		block.read_real_expression(line, counter, double_ptr, parameters)
	} else if c == '#' {
		block.read_parameter(line, counter, double_ptr, parameters)
	} else if (c >= 'a') && (c <= 'z') {
		block.read_unary(line, counter, double_ptr, parameters)
	} else {
		read_real_number(line, counter, double_ptr)
	}

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_unary

   Returned Value: int
   If any of the following functions returns an error code,
   this returns that code.
   execute_unary
   read_atan
   read_operation_unary
   read_real_expression
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. the name of the unary operation is not followed by a left bracket:
   NCE_LEFT_BRACKET_MISSING_AFTER_UNARY_OPERATION_NAME

   Side effects:
   The value read from the line is put into what double_ptr points at.
   The counter is reset to point to the first character after the
   characters which make up the value.

   Called by:  read_real_value

   This attempts to read the value of a unary operation out of the line,
   starting at the index given by the counter. The atan operation is
   handled specially because it is followed by two arguments.

*/
func (block *Block_t) read_unary( /* ARGUMENTS                               */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	double_ptr *float64, /* pointer to double to be read                   */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	//static char name[] SET_TO "read_unary";
	var operation ops.Operation

	block.read_operation_unary(line, counter, &operation)
	if line[*counter] != '[' {
		return inc.NCE_LEFT_BRACKET_MISSING_AFTER_UNARY_OPERATION_NAME
	}

	block.read_real_expression(line, counter, double_ptr, parameters)

	if operation == ops.ATAN {
		block.read_atan(line, counter, double_ptr, parameters)
	} else {
		ops.Execute_unary(double_ptr, operation)
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_parameter

   Returned Value: int
   If read_integer_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, this returns RS274NGC_OK.
   1. The first character read is not # :
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. The parameter number is out of bounds:
   NCE_PARAMETER_NUMBER_OUT_OF_RANGE

   Side effects:
   The value of the given parameter is put into what double_ptr points at.
   The counter is reset to point to the first character after the
   characters which make up the value.

   Called by:  read_real_value

   This attempts to read the value of a parameter out of the line,
   starting at the index given by the counter.

   According to the RS274/NGC manual [NCMS, p. 62], the characters following
   # may be any "parameter expression". Thus, the following are legal
   and mean the same thing (the value of the parameter whose number is
   stored in parameter 2):
   ##2
   #[#2]

   Parameter setting is done in parallel, not sequentially. For example
   if #1 is 5 before the line "#1=10 #2=#1" is read, then after the line
   is is executed, #1 is 10 and #2 is 5. If parameter setting were done
   sequentially, the value of #2 would be 10 after the line was executed.

*/

func (block *Block_t) read_parameter( /* ARGUMENTS                               */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	double_ptr *float64, /* pointer to double to be read                   */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	// static char name[] SET_TO "read_parameter";
	var (
		index int
	)
	// int status;

	if line[*counter] != '#' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	block.read_integer_value(line, counter, &index, parameters)

	if (index < 1) || (index >= inc.RS274NGC_MAX_PARAMETERS) {
		return inc.NCE_PARAMETER_NUMBER_OUT_OF_RANGE
	}

	*double_ptr = parameters[index]
	return inc.RS274NGC_OK
}

/****************************************************************************/

/*

   The following version is stack-based and fully general. It is the
   classical stack-based version with left-to-right evaluation of
   operations of the same precedence. Separate stacks are used for
   operations and values, and the stacks are made with arrays
   rather than lists, but those are implementation details. Pushing
   and popping are implemented by increasing or decreasing the
   stack index.

   Additional levels of precedence may be defined easily by changing the
   precedence function. The size of MAX_STACK should always be at least
   as large as the number of precedence levels used. We are currently
   using four precedence levels (for right-bracket, plus-like operations,
   times-like operations, and power).

*/

const MAX_STACK = 5

func (block *Block_t) read_real_expression( /* ARGUMENTS                               */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	value *float64, /* pointer to double to be computed               */
	parameters []float64) (s inc.STATUS) { /* array of system parameters                     */

	//static char name[] SET_TO "read_real_expression";

	var (
		operators [MAX_STACK]ops.Operation
		values    [MAX_STACK]float64
	)

	if line[*counter] != '[' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)

	if s = block.read_real_value(line, counter, &values[0], parameters); s != inc.RS274NGC_OK {
		return
	}
	if s = block.read_operation(line, counter, &operators[0]); s != inc.RS274NGC_OK {
		return
	}

	stack_index := 1
	for operators[0] != ops.RIGHT_BRACKET {
		block.read_real_value(line, counter, &values[stack_index], parameters)
		block.read_operation(line, counter, &operators[stack_index])
		if ops.Precedence(operators[stack_index]) >
			ops.Precedence(operators[stack_index-1]) {
			stack_index++
		} else { /* precedence of latest operator is <= previous precedence */
			for ops.Precedence(operators[stack_index]) <= ops.Precedence(operators[stack_index-1]) {
				ops.Execute_binary(&(values[stack_index-1]),
					operators[stack_index-1],
					&(values[stack_index]))
				operators[stack_index-1] = operators[stack_index]
				if (stack_index > 1) &&
					(ops.Precedence(operators[stack_index-1]) <=
						ops.Precedence(operators[stack_index-2])) {
					stack_index--
				} else {
					break
				}
			}
		}
	}

	*value = values[0]
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_real_number

Returned Value: int
If any of the following errors occur, this returns the error shown.
Otherwise, it returns RS274NGC_OK.
1. The first character is not "+", "-", "." or a digit:
NCE_BAD_NUMBER_FORMAT
2. No digits are found after the first character and before the
end of the line or the next character that cannot be part of a real:
NCE_NO_DIGITS_FOUND_WHERE_REAL_NUMBER_SHOULD_BE
3. sscanf fails: NCE_SSCANF_FAILED

Side effects:
The number read from the line is put into what double_ptr points at.
The counter is reset to point to the first character after the real.

Called by:  read_real_value

This attempts to read a number out of the line, starting at the index
given by the counter. It stops when the first character that cannot
be part of the number is found.

The first character may be a digit, "+", "-", or "."
Every following character must be a digit or "." up to anything
that is not a digit or "." (a second "." terminates reading).

This function is not called if the first character is NULL, so it is
not necessary to check that.

The temporary insertion of a NULL character on the line is to avoid
making a format string like "%3lf" which the LynxOS compiler cannot
handle.

*/

func read_real_number( /* ARGUMENTS                               */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	double_ptr *float64) (s inc.STATUS) { /* pointer to double to be read                   */

	var (
		i          = 0
		flag_point = OFF
		flag_digit = OFF
	)

	/* check first character */
	temp := string(line[*counter:])
	c := temp[0]
	if c == '+' || c == '-' || c == '.' {
		i++
	} else if !((47 < c) && (c < 58)) { // not digit
		return inc.NCE_BAD_NUMBER_FORMAT
	}

	/* check out rest of characters (must be digit or decimal point) */
	for ; i < len(temp); i++ {
		c := temp[i]
		if (47 < c) && (c < 58) {
			flag_digit = ON
		} else if c == '.' {
			if flag_point == OFF {
				flag_point = ON
			} else {
				s = inc.NCE_SSCANF_FAILED
				return
			}
		} else {
			break
		}
	}
	*counter = *counter + i
	temp = temp[:i]

	if flag_digit == OFF {
		return inc.NCE_NO_DIGITS_FOUND_WHERE_REAL_NUMBER_SHOULD_BE
	}
	//line[n] = 0 /* temporary string termination for sscanf */
	if v, e := strconv.ParseFloat(temp, 64); e != nil {
		return inc.NCE_SSCANF_FAILED
	} else {
		s = inc.RS274NGC_OK
		*double_ptr = v
	}

	return
}

/****************************************************************************/

/****************************************************************************/

/* read_operation_unary

   Returned Value: int
   If the operation is not a known unary operation, this returns one of
   the following error codes:
   NCE_UNKNOWN_WORD_STARTING_WITH_A
   NCE_UNKNOWN_WORD_STARTING_WITH_C
   NCE_UNKNOWN_WORD_STARTING_WITH_E
   NCE_UNKNOWN_WORD_STARTING_WITH_F
   NCE_UNKNOWN_WORD_STARTING_WITH_L
   NCE_UNKNOWN_WORD_STARTING_WITH_R
   NCE_UNKNOWN_WORD_STARTING_WITH_S
   NCE_UNKNOWN_WORD_STARTING_WITH_T
   NCE_UNKNOWN_WORD_WHERE_UNARY_OPERATION_COULD_BE
   Otherwise, this returns RS274NGC_OK.

   Side effects:
   An integer code for the name of the operation read from the
   line is put into what operation points at.
   The counter is reset to point to the first character after the
   characters which make up the operation name.

   Called by:
   read_unary

   This attempts to read the name of a unary operation out of the line,
   starting at the index given by the counter. Known operations are:
   abs, acos, asin, atan, cos, exp, fix, fup, ln, round, sin, sqrt, tan.

*/
func (block *Block_t) read_operation_unary( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274/NGC code being processed */
	counter *int, /* pointer to a counter for position on the line  */
	operation *ops.Operation) inc.STATUS { /* array of system parameters                     */

	//static char name[] SET_TO "read_operation_unary";
	c := line[*counter]
	*counter = (*counter + 1)
	switch c {
	case 'a':
		if (line[*counter] == 'b') && (line[(*counter)+1] == 's') {
			*operation = ops.ABS
			*counter = (*counter + 2)
		} else if strings.HasPrefix(string(line[*counter:]), "cos") {
			*operation = ops.ACOS
			*counter = (*counter + 3)
		} else if strings.HasPrefix(string(line[*counter:]), "sin") {
			*operation = ops.ASIN
			*counter = (*counter + 3)
		} else if strings.HasPrefix(string(line[*counter:]), "tan") {
			*operation = ops.ATAN
			*counter = (*counter + 3)
		} else {
			return inc.NCE_UNKNOWN_WORD_STARTING_WITH_A
		}
	case 'c':
		if (line[*counter] == 'o') && (line[(*counter)+1] == 's') {
			*operation = ops.COS
			*counter = (*counter + 2)
		} else {
			return inc.NCE_UNKNOWN_WORD_STARTING_WITH_C
		}
	case 'e':
		if (line[*counter] == 'x') && (line[(*counter)+1] == 'p') {
			*operation = ops.EXP
			*counter = (*counter + 2)
		} else {
			return inc.NCE_UNKNOWN_WORD_STARTING_WITH_E
		}
	case 'f':
		if (line[*counter] == 'i') && (line[(*counter)+1] == 'x') {
			*operation = ops.FIX
			*counter = (*counter + 2)
		} else if (line[*counter] == 'u') && (line[(*counter)+1] == 'p') {
			*operation = ops.FUP
			*counter = (*counter + 2)
		} else {
			return inc.NCE_UNKNOWN_WORD_STARTING_WITH_F
		}
	case 'l':
		if line[*counter] == 'n' {
			*operation = ops.LN
			*counter = (*counter + 1)
		} else {
			return inc.NCE_UNKNOWN_WORD_STARTING_WITH_L
		}
	case 'r':
		if strings.HasPrefix(string(line[*counter:]), "ound") {
			*operation = ops.ROUND
			*counter = (*counter + 4)
		} else {
			return inc.NCE_UNKNOWN_WORD_STARTING_WITH_R
		}
	case 's':
		if (line[*counter] == 'i') && (line[(*counter)+1] == 'n') {
			*operation = ops.SIN
			*counter = (*counter + 2)
		} else if strings.HasPrefix(string(line[*counter:]), "qrt") {
			*operation = ops.SQRT
			*counter = (*counter + 3)
		} else {
			return inc.NCE_UNKNOWN_WORD_STARTING_WITH_S
		}
	case 't':
		if (line[*counter] == 'a') && (line[(*counter)+1] == 'n') {
			*operation = ops.TAN
			*counter = (*counter + 2)
		} else {
			return inc.NCE_UNKNOWN_WORD_STARTING_WITH_T
		}
	default:
		return inc.NCE_UNKNOWN_WORD_WHERE_UNARY_OPERATION_COULD_BE
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_j

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not j:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A j_coordinate has already been inserted in the block.
   NCE_MULTIPLE_J_WORDS_ON_ONE_LINE

   Side effects:
   counter is reset.
   The j_flag in the block is turned on.
   A j_coordinate setting is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'j', indicating a j_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   This information is inserted in the block. The counter is then set to
   point to the character following.

   The value may be a real number or something that evaluates to a real
   number, so read_real_value is used to read it. Parameters may be
   involved.

*/
func (block *Block_t) read_j( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */
	var value float64
	if line[*counter] != 'j' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.j_flag != OFF {
		return inc.NCE_MULTIPLE_J_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)
	block.j_flag = ON
	block.j_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_k

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not k:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A k_coordinate has already been inserted in the block:
   NCE_MULTIPLE_K_WORDS_ON_ONE_LINE

   Side effects:
   counter is reset.
   The k_flag in the block is turned on.
   A k_coordinate setting is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'k', indicating a k_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   This information is inserted in the block. The counter is then set to
   point to the character following.

   The value may be a real number or something that evaluates to a real
   number, so read_real_value is used to read it. Parameters may be
   involved.

*/
func (block *Block_t) read_k( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64
	if line[*counter] != 'k' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.k_flag != OFF {
		return inc.NCE_MULTIPLE_K_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)
	block.k_flag = ON
	block.k_number = value

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_l

   Returned Value: int
   If read_integer_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not l:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. An l_number has already been inserted in the block:
   NCE_MULTIPLE_L_WORDS_ON_ONE_LINE
   3. the l_number is negative: NCE_NEGATIVE_L_WORD_USED

   Side effects:
   counter is reset to the character following the l number.
   An l code is inserted in the block as the value of l.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'l', indicating an L code.
   The function reads characters which give the (integer) value of the
   L code.

   L codes are used for:
   1. the number of times a canned cycle should be repeated.
   2. a key with G10.

*/
func (block *Block_t) read_l( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value int
	if line[*counter] != 'l' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.l_number > -1 {
		return inc.NCE_MULTIPLE_L_WORDS_ON_ONE_LINE
	}
	block.read_integer_value(line, counter, &value, parameters)
	if value < 0 {
		return inc.NCE_NEGATIVE_L_WORD_USED
	}
	block.l_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_m

   Returned Value:
   If read_integer_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not m:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. The value is negative: NCE_NEGATIVE_M_CODE_USED
   3. The value is greater than 99: NCE_M_CODE_GREATER_THAN_99
   4. The m code is not known to the system: NCE_UNKNOWN_M_CODE_USED
   5. Another m code in the same modal group has already been read:
   NCE_TWO_M_CODES_USED_FROM_SAME_MODAL_GROUP

   Side effects:
   counter is reset to the character following the m number.
   An m code is inserted as the value of the appropriate mode in the
   m_modes array in the block.
   The m code counter in the block is increased by 1.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'm', indicating an m code.
   The function reads characters which give the (integer) value of the
   m code.

   read_integer_value allows a minus sign, so a check for a negative value
   is needed here, and the parameters argument is also needed.

*/
func (block *Block_t) read_m( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value int
	if line[*counter] != 'm' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)

	block.read_integer_value(line, counter, &value, parameters)
	if value < 0 {
		return inc.NCE_NEGATIVE_M_CODE_USED
	} else if value > 99 {
		return inc.NCE_M_CODE_GREATER_THAN_99
	}

	mode, ok := _ems[value]
	if ok == false {
		return inc.NCE_UNKNOWN_M_CODE_USED
	}
	if block.m_modes[mode] != -1 {
		return inc.NCE_TWO_M_CODES_USED_FROM_SAME_MODAL_GROUP
	}
	block.m_modes[mode] = value
	block.m_count++

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_p

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not p:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A p value has already been inserted in the block:
   NCE_MULTIPLE_P_WORDS_ON_ONE_LINE
   3. The p value is negative: NCE_NEGATIVE_P_WORD_USED

   Side effects:
   counter is reset to point to the first character following the p value.
   The p value setting is inserted in block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'p', indicating a p value
   setting. The function reads characters which tell how to set the p
   value, up to the start of the next item or the end of the line. This
   information is inserted in the block.

   P codes are used for:
   1. Dwell time in canned cycles g82, G86, G88, G89 [NCMS pages 98 - 100].
   2. A key with G10 [NCMS, pages 9, 10].

*/
func (block *Block_t) read_p( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'p' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.s_number > -1.0 {
		return inc.NCE_MULTIPLE_P_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)

	if block.p_number < 0.0 {
		return inc.NCE_NEGATIVE_P_WORD_USED
	}
	block.s_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_q

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not q:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A q value has already been inserted in the block:
   NCE_MULTIPLE_Q_WORDS_ON_ONE_LINE
   3. The q value is negative or zero: NCE_NEGATIVE_OR_ZERO_Q_VALUE_USED

   Side effects:
   counter is reset to point to the first character following the q value.
   The q value setting is inserted in block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'q', indicating a q value
   setting. The function reads characters which tell how to set the q
   value, up to the start of the next item or the end of the line. This
   information is inserted in the block.

   Q is used only in the G87 canned cycle [NCMS, page 98], where it must
   be positive.

*/

func (block *Block_t) read_q( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'q' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.q_number > -1.0 {
		return inc.NCE_MULTIPLE_Q_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)

	if block.q_number <= 0.0 {
		return inc.NCE_NEGATIVE_OR_ZERO_Q_VALUE_USED
	}
	block.q_number = value

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_r

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not r:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. An r_number has already been inserted in the block:
   NCE_MULTIPLE_R_WORDS_ON_ONE_LINE

   Side effects:
   counter is reset.
   The r_flag in the block is turned on.
   The r_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'r'. The function reads characters
   which tell how to set the coordinate, up to the start of the next item
   or the end of the line. This information is inserted in the block. The
   counter is then set to point to the character following.

   An r number indicates the clearance plane in canned cycles.
   An r number may also be the radius of an arc.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

*/
func (block *Block_t) read_r( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'r' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if !!block.r_flag {
		return inc.NCE_MULTIPLE_R_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)

	block.r_flag = ON
	block.r_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_s

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not s:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A spindle speed has already been inserted in the block:
   NCE_MULTIPLE_S_WORDS_ON_ONE_LINE
   3. The spindle speed is negative: NCE_NEGATIVE_SPINDLE_SPEED_USED

   Side effects:
   counter is reset to the character following the spindle speed.
   A spindle speed setting is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 's', indicating a spindle speed
   setting. The function reads characters which tell how to set the spindle
   speed.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

*/
func (block *Block_t) read_s( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 's' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.s_number > -1.0 {
		return inc.NCE_MULTIPLE_S_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)

	if block.s_number < 0.0 {
		return inc.NCE_NEGATIVE_SPINDLE_SPEED_USED
	}
	block.s_number = value

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_t

   Returned Value: int
   If read_integer_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not t:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A t_number has already been inserted in the block:
   NCE_MULTIPLE_T_WORDS_ON_ONE_LINE
   3. The t_number is negative: NCE_NEGATIVE_TOOL_ID_USED

   Side effects:
   counter is reset to the character following the t_number.
   A t_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 't', indicating a tool.
   The function reads characters which give the (integer) value of the
   tool code.

   The value must be an integer or something that evaluates to a
   real number, so read_integer_value is used to read it. Parameters
   may be involved.

*/
func (block *Block_t) read_t( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value int

	if line[*counter] != 't' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if block.t_number > -1 {
		return inc.NCE_MULTIPLE_T_WORDS_ON_ONE_LINE
	}
	block.read_integer_value(line, counter, &value, parameters)

	if block.t_number < 0 {
		return inc.NCE_NEGATIVE_TOOL_ID_USED
	}
	block.t_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_x

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not x:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A x_coordinate has already been inserted in the block:
   NCE_MULTIPLE_X_WORDS_ON_ONE_LINE

   Side effects:
   counter is reset.
   The x_flag in the block is turned on.
   An x_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'x', indicating a x_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   This information is inserted in the block. The counter is then set to
   point to the character following.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

*/
func (block *Block_t) read_x( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'x' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if !!block.x_flag {
		return inc.NCE_MULTIPLE_X_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)

	block.x_flag = ON
	block.x_number = value

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_y

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not y:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A y_coordinate has already been inserted in the block:
   NCE_MULTIPLE_Y_WORDS_ON_ONE_LINE

   Side effects:
   counter is reset.
   The y_flag in the block is turned on.
   A y_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'y', indicating a y_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   This information is inserted in the block. The counter is then set to
   point to the character following.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

*/
func (block *Block_t) read_y( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'y' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if !!block.y_flag {
		return inc.NCE_MULTIPLE_Y_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)

	block.y_flag = ON
	block.y_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_z

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not z:
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. A z_coordinate has already been inserted in the block:
   NCE_MULTIPLE_Z_WORDS_ON_ONE_LINE

   Side effects:
   counter is reset.
   The z_flag in the block is turned on.
   A z_number is inserted in the block.

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character 'z', indicating a z_coordinate
   setting. The function reads characters which tell how to set the
   coordinate, up to the start of the next item or the end of the line.
   This information is inserted in the block. The counter is then set to
   point to the character following.

   The value may be a real number or something that evaluates to a
   real number, so read_real_value is used to read it. Parameters
   may be involved.

*/
func (block *Block_t) read_z( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var value float64

	if line[*counter] != 'z' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	if !!block.z_flag {
		return inc.NCE_MULTIPLE_Z_WORDS_ON_ONE_LINE
	}
	block.read_real_value(line, counter, &value, parameters)

	block.z_flag = ON
	block.z_number = value

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_integer_value

   Returned Value: int
   If read_real_value returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The returned value is not close to an integer:
   NCE_NON_INTEGER_VALUE_FOR_INTEGER

   Side effects:
   The number read from the line is put into what integer_ptr points at.

   Called by:
   read_d
   read_l
   read_h
   read_m
   read_parameter
   read_parameter_setting
   read_t

   This reads an integer (positive, negative or zero) from a string,
   starting from the position given by *counter. The value being
   read may be written with a decimal point or it may be an expression
   involving non-integers, as long as the result comes out within 0.0001
   of an integer.

   This proceeds by calling read_real_value and checking that it is
   close to an integer, then returning the integer it is close to.

*/
func (block *Block_t) read_integer_value( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	integer_ptr *int, /* pointer to the value being read                */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var (
		float_value float64
	)

	block.read_real_value(line, counter, &float_value, parameters)
	*integer_ptr = int(math.Floor(float_value))
	if (float_value - float64(*integer_ptr)) > 0.9999 {
		*integer_ptr = int(math.Ceil(float_value))
	} else if (float_value - float64(*integer_ptr)) > 0.0001 {
		return inc.NCE_NON_INTEGER_VALUE_FOR_INTEGER
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_parameter_setting

   Returned Value: int
   If read_real_value or read_integer_value returns an error code,
   this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character read is not # :
   NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
   2. The parameter index is out of range: PARAMETER_NUMBER_OUT_OF_RANGE
   3. An equal sign does not follow the parameter expression:
   NCE_EQUAL_SIGN_MISSING_IN_PARAMETER_SETTING

   Side effects:
   counter is reset to the character following the end of the parameter
   setting. The parameter whose index follows "#" is set to the
   real value following "=".

   Called by: read_one_item

   When this function is called, counter is pointing at an item on the
   line that starts with the character '#', indicating a parameter
   setting when found by read_one_item.  The function reads characters
   which tell how to set the parameter.

   Any number of parameters may be set on a line. If parameters set early
   on the line are used in expressions farther down the line, the
   parameters have their old values, not their new values. This is
   usually called setting parameters in parallel.

   Parameter setting is not clearly described in [NCMS, pp. 51 - 62]: it is
   not clear if more than one parameter setting per line is allowed (any
   number is OK in this implementation). The characters immediately following
   the "#" must constitute a "parameter expression", but it is not clear
   what that is. Here we allow any expression as long as it evaluates to
   an integer.

   Parameters are handled in the interpreter by having a parameter table
   and a parameter buffer as part of the machine settings. The parameter
   table is passed to the reading functions which need it. The parameter
   buffer is used directly by functions that need it. Reading functions
   may set parameter values in the parameter buffer. Reading functions
   may obtain parameter values; these come from parameter table.

   The parameter buffer has three parts: (i) a counter for how many
   parameters have been set while reading the current line (ii) an array
   of the indexes of parameters that have been set while reading the
   current line, and (iii) an array of the values for the parameters that
   have been set while reading the current line; the nth value
   corresponds to the nth index. Any given index will appear once in the
   index number array for each time the parameter with that index is set
   on a line. There is no point in setting the same parameter more than
   one on a line because only the last setting of that parameter will
   take effect.

   The syntax recognized by this this function is # followed by an
   integer expression (explicit integer or expression evaluating to an
   integer) followed by = followed by a real value (number or
   expression).

   Note that # also starts a bunch of characters which represent a parameter
   to be evaluated. That situation is handled by read_parameter.

*/
func (block *Block_t) read_parameter_setting( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	var (
		index int
		value float64
	)

	if line[*counter] != '#' {
		return inc.NCE_BUG_FUNCTION_SHOULD_NOT_HAVE_BEEN_CALLED
	}
	*counter = (*counter + 1)
	block.read_integer_value(line, counter, &index, parameters)
	if (index < 1) ||
		index >= inc.RS274NGC_MAX_PARAMETERS {
		return inc.NCE_PARAMETER_NUMBER_OUT_OF_RANGE
	}
	if line[*counter] != '=' {
		return inc.NCE_EQUAL_SIGN_MISSING_IN_PARAMETER_SETTING
	}
	*counter = (*counter + 1)
	block.read_real_value(line, counter, &value, parameters)
	block.Parameter_numbers[block.Parameter_occurrence] = index
	block.Parameter_values[block.Parameter_occurrence] = value
	block.Parameter_occurrence++

	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_operation

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The operation is unknown:
   NCE_UNKNOWN_OPERATION_NAME_STARTING_WITH_A
   NCE_UNKNOWN_OPERATION_NAME_STARTING_WITH_M
   NCE_UNKNOWN_OPERATION_NAME_STARTING_WITH_O
   NCE_UNKNOWN_OPERATION_NAME_STARTING_WITH_X
   NCE_UNKNOWN_OPERATION
   2. The line ends without closing the expression: NCE_UNCLOSED_EXPRESSION

   Side effects:
   An integer representing the operation is put into what operation points
   at.  The counter is reset to point to the first character after the
   operation.

   Called by: read_real_expression

   This expects to be reading a binary operation (+, -, /, *, **, and,
   mod, or, xor) or a right bracket (]). If one of these is found, the
   value of operation is set to the symbolic value for that operation.
   If not, an error is reported as described above.

*/

func (block *Block_t) read_operation( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	operation *ops.Operation) inc.STATUS { /* pointer to operation to be read                */

	c := line[*counter]
	*counter = (*counter + 1)
	switch c {
	case '+':
		*operation = ops.PLUS
	case '-':
		*operation = ops.MINUS
	case '/':
		*operation = ops.DIVIDED_BY
	case '*':
		if line[*counter] == '*' {
			*operation = ops.POWER
			*counter = (*counter + 1)
		} else {
			*operation = ops.TIMES
		}
	case ']':
		*operation = ops.RIGHT_BRACKET
	case 'a':
		if (line[*counter] == 'n') && (line[(*counter)+1] == 'd') {
			*operation = ops.AND2
			*counter = (*counter + 2)
		} else {
			return inc.NCE_UNKNOWN_OPERATION_NAME_STARTING_WITH_A
		}
	case 'm':
		if (line[*counter] == 'o') && (line[(*counter)+1] == 'd') {
			*operation = ops.MODULO
			*counter = (*counter + 2)
		} else {
			return inc.NCE_UNKNOWN_OPERATION_NAME_STARTING_WITH_M
		}
	case 'o':
		if line[*counter] == 'r' {
			*operation = ops.NON_EXCLUSIVE_OR
			*counter = (*counter + 1)
		} else {
			return inc.NCE_UNKNOWN_OPERATION_NAME_STARTING_WITH_O
		}
	case 'x':
		if (line[*counter] == 'o') && (line[(*counter)+1] == 'r') {
			*operation = ops.EXCLUSIVE_OR
			*counter = (*counter + 2)
		} else {
			return inc.NCE_UNKNOWN_OPERATION_NAME_STARTING_WITH_X
		}
	case 0:
		return inc.NCE_UNCLOSED_EXPRESSION
	default:
		return inc.NCE_UNKNOWN_OPERATION
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* read_atan

   Returned Value: int
   If read_real_expression returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The first character to read is not a slash:
   NCE_SLASH_MISSING_AFTER_FIRST_ATAN_ARGUMENT
   2. The second character to read is not a left bracket:
   NCE_LEFT_BRACKET_MISSING_AFTER_SLASH_WITH_ATAN

   Side effects:
   The computed value is put into what double_ptr points at.
   The counter is reset to point to the first character after the
   characters which make up the value.

   Called by:
   read_unary

   When this function is called, the characters "atan" and the first
   argument have already been read, and the value of the first argument
   is stored in double_ptr.  This function attempts to read a slash and
   the second argument to the atan function, starting at the index given
   by the counter and then to compute the value of the atan operation
   applied to the two arguments.  The computed value is inserted into
   what double_ptr points at.

   The computed value is in the range from -180 degrees to +180 degrees.
   The range is not specified in the RS274/NGC manual [NCMS, page 51],
   although using degrees (not radians) is specified.

*/
func (block *Block_t) read_atan( /* ARGUMENTS                                      */
	line []byte, /* string: line of RS274 code being processed     */
	counter *int, /* pointer to a counter for position on the line  */
	double_ptr *float64, /* pointer to double to be read                   */
	parameters []float64) inc.STATUS { /* array of system parameters                     */

	//static char name[] SET_TO "read_atan";
	var argument2 float64

	if line[*counter] != '/' {
		return inc.NCE_SLASH_MISSING_AFTER_FIRST_ATAN_ARGUMENT
	}
	*counter = (*counter + 1)
	if line[*counter] != '[' {
		return inc.NCE_LEFT_BRACKET_MISSING_AFTER_SLASH_WITH_ATAN
	}

	block.read_real_expression(line, counter, &argument2, parameters)
	/* value in radians */
	*double_ptr = math.Atan2(*double_ptr, argument2)
	/* convert to degrees */
	*double_ptr = ((*double_ptr * 180.0) / inc.PI)
	return inc.RS274NGC_OK
}

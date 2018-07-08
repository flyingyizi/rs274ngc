package rs274ngc

import "github.com/flyingyizi/rs274ngc/inc"

/****************************************************************************/

/* convert_stop

   Returned Value: int
   When an m2 or m30 (program_end) is encountered, this returns RS274NGC_EXIT.
   If the code is not m0, m1, m2, m30, or m60, this returns
   NCE_BUG_CODE_NOT_M0_M1_M2_M30_M60
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   An m0, m1, m2, m30, or m60 in the block is executed.

   For m0, m1, and m60, this makes a function call to the PROGRAM_STOP
   canonical machining function (which stops program execution).
   In addition, m60 calls PALLET_SHUTTLE.

   For m2 and m30, this resets the machine and then calls PROGRAM_END.
   In addition, m30 calls PALLET_SHUTTLE.

   Called by: execute_block.

   This handles stopping or ending the program (m0, m1, m2, m30, m60)

   [NCMS] specifies how the following modes should be reset at m2 or
   m30. The descriptions are not collected in one place, so this list
   may be incomplete.

   G52 offsetting coordinate zero points [NCMS, page 10]
   G92 coordinate offset using tool position [NCMS, page 10]

   The following should have reset values, but no description of reset
   behavior could be found in [NCMS].
   G17, G18, G19 selected plane [NCMS, pages 14, 20]
   G90, G91 distance mode [NCMS, page 15]
   G93, G94 feed mode [NCMS, pages 35 - 37]
   M48, M49 overrides enabled, disabled [NCMS, pages 37 - 38]
   M3, M4, M5 spindle turning [NCMS, page 7]

   The following should be set to some value at machine start-up but
   not automatically reset by any of the stopping codes.
   1. G20, G21 length units [NCMS, page 15]. This is up to the installer.
   2. motion_control_mode. This is set in rs274ngc_init but not reset here.
   Might add it here.

   The following resets have been added by calling the appropriate
   canonical machining command and/or by resetting interpreter
   settings. They occur on M2 or M30.

   1. Axis offsets are set to zero (like g92.2) and      - SET_ORIGIN_OFFSETS
   origin offsets are set to the default (like G54)
   2. Selected plane is set to CANON_PLANE_XY (like G17) - SELECT_PLANE
   3. Distance mode is set to MODE_ABSOLUTE (like G90)   - no canonical call
   4. Feed mode is set to UNITS_PER_MINUTE (like G94)    - no canonical call
   5. Feed and speed overrides are set to ON (like M48)  - ENABLE_FEED_OVERRIDE
   - ENABLE_SPEED_OVERRIDE
   6. Cutter compensation is turned off (like G40)       - no canonical call
   7. The spindle is stopped (like M5)                   - STOP_SPINDLE_TURNING
   8. The motion mode is set to G_1 (like G1)            - no canonical call
   9. Coolant is turned off (like M9)                    - FLOOD_OFF & MIST_OFF

*/

func (cnc *rs274ngc_t) convert_stop() inc.STATUS {

	//static char name[] = "convert_stop";
	//        int index;
	//        char * line;
	//        int length;

	if cnc._setup.block1.m_modes[4] == 0 {
		cnc.canon.PROGRAM_STOP()
	} else if cnc._setup.block1.m_modes[4] == 60 {
		cnc.canon.PALLET_SHUTTLE()
		cnc.canon.PROGRAM_STOP()
	} else if cnc._setup.block1.m_modes[4] == 1 {
		cnc.canon.OPTIONAL_PROGRAM_STOP()
	} else if (cnc._setup.block1.m_modes[4] == 2) || (cnc._setup.block1.m_modes[4] == 30) {
		/* reset stuff here */
		/*1*/
		cnc._setup.current.X = cnc._setup.current.X + cnc._setup.origin_offset.X + cnc._setup.axis_offset.X
		cnc._setup.current.Y = cnc._setup.current.Y +
			cnc._setup.origin_offset.Y + cnc._setup.axis_offset.Y
		cnc._setup.current.Z = cnc._setup.current.Z +
			cnc._setup.origin_offset.Z + cnc._setup.axis_offset.Z

		cnc._setup.current.A = cnc._setup.current.A +
			cnc._setup.origin_offset.A + cnc._setup.axis_offset.A

		cnc._setup.current.B = cnc._setup.current.B +
			cnc._setup.origin_offset.B + cnc._setup.axis_offset.B

		cnc._setup.current.C = cnc._setup.current.C +
			cnc._setup.origin_offset.C + cnc._setup.axis_offset.C

		cnc._setup.origin_index = 1
		cnc._setup.parameters[5220] = 1.0
		cnc._setup.origin_offset.X = cnc._setup.parameters[5221]
		cnc._setup.origin_offset.Y = cnc._setup.parameters[5222]
		cnc._setup.origin_offset.Z = cnc._setup.parameters[5223]

		cnc._setup.origin_offset.A = cnc._setup.parameters[5224]

		cnc._setup.origin_offset.B = cnc._setup.parameters[5225]

		cnc._setup.origin_offset.C = cnc._setup.parameters[5226]

		cnc._setup.axis_offset.X = 0
		cnc._setup.axis_offset.X = 0
		cnc._setup.axis_offset.X = 0

		cnc._setup.axis_offset.A = 0 /*AA*/

		cnc._setup.axis_offset.B = 0 /*BB*/

		cnc._setup.axis_offset.C = 0 /*CC*/

		cnc._setup.current.X = cnc._setup.current.X -
			cnc._setup.origin_offset.X
		cnc._setup.current.Y = cnc._setup.current.Y -
			cnc._setup.origin_offset.Y
		cnc._setup.current.Z = cnc._setup.current.Z -
			cnc._setup.origin_offset.Z
		cnc._setup.current.A = cnc._setup.current.A -
			cnc._setup.origin_offset.A /*AA*/
		cnc._setup.current.B = cnc._setup.current.B -
			cnc._setup.origin_offset.B /*BB*/
		cnc._setup.current.C = cnc._setup.current.C -
			cnc._setup.origin_offset.C /*CC*/

		cnc.canon.SET_ORIGIN_OFFSETS(cnc._setup.origin_offset.X,
			cnc._setup.origin_offset.Y,
			cnc._setup.origin_offset.Z,
			cnc._setup.origin_offset.A,
			cnc._setup.origin_offset.B,
			cnc._setup.origin_offset.C)

		/*2*/
		if cnc._setup.plane != CANON_PLANE_XY {
			cnc.canon.SELECT_PLANE(CANON_PLANE_XY)
			cnc._setup.plane = CANON_PLANE_XY
		}

		/*3*/
		cnc._setup.distance_mode = inc.MODE_ABSOLUTE

		/*4*/
		cnc._setup.feed_mode = inc.UNITS_PER_MINUTE

		/*5*/
		if cnc._setup.feed_override != ON {
			cnc.canon.ENABLE_FEED_OVERRIDE()
			cnc._setup.feed_override = ON
		}
		if cnc._setup.speed_override != ON {
			cnc.canon.ENABLE_SPEED_OVERRIDE()
			cnc._setup.speed_override = ON
		}

		/*6*/
		cnc._setup.cutter_comp_side = inc.CANON_SIDE_OFF
		cnc._setup.program_x = inc.UNKNOWN

		/*7*/
		cnc.canon.STOP_SPINDLE_TURNING()
		cnc._setup.spindle_turning = CANON_STOPPED

		/*8*/
		cnc._setup.motion_mode = inc.G_1

		/*9*/
		if cnc._setup.coolant.mist == ON {
			cnc.canon.MIST_OFF()
			cnc._setup.coolant.mist = OFF
		}
		if cnc._setup.coolant.flood == ON {
			cnc.canon.FLOOD_OFF()
			cnc._setup.coolant.flood = OFF
		}

		if cnc._setup.block1.m_modes[4] == 30 {
			cnc.canon.PALLET_SHUTTLE()

		}
		cnc.canon.PROGRAM_END()
		return inc.RS274NGC_EXIT
	} else {
		return inc.NCE_BUG_CODE_NOT_M0_M1_M2_M30_M60
	}

	return inc.RS274NGC_OK
}

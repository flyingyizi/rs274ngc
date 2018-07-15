package rs274ngc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/flyingyizi/rs274ngc/inc"
)

const (
	RS274NGC_PARAMETER_FILE_NAME_DEFAULT = "rs274ngc.var"
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

/* rs274ngc.cc

   This rs274ngc.cc file contains the source code for (1) the kernel of
   several rs274ngc interpreters and (2) two of the four sets of interface
   functions declared in canon.hh:
   1. interface functions to call to tell the interpreter what to do.
   These all return a status value.
   2. interface functions to call to get information from the interpreter.

   Kernel functions call each other. A few kernel functions are called by
   interface functions.

   Interface function names all begin with "rs274ngc_".

   Error handling is by returning a status value of either a non-error
   code (RS274NGC_OK, RS274NGC_EXIT, etc.) or some specific error code
   from each function where there is a possibility of error.  If an error
   occurs, processing is always stopped, and control is passed back up
   through the function call hierarchy to an interface function; the
   error code is also passed back up. The stack of functions called is
   also recorded. The external program calling an interface function may
   then handle the error further as it sees fit.

   Since returned values are usually used as just described to handle the
   possibility of errors, an alternative method of passing calculated
   values is required. In general, if function A needs a value for
   variable V calculated by function B, this is handled by passing a
   pointer to V from A to B, and B calculates and sets V.

   There are a lot of functions named read_XXXX. All such functions read
   characters from a string using a counter. They all reset the counter
   to point at the character in the string following the last one used by
   the function. The counter is passed around from function to function
   by using pointers to it. The first character read by each of these
   functions is expected to be a member of some set of characters (often
   a specific character), and each function checks the first character.

   This version of the interpreter not saving input lines. A list of all
   lines will be needed in future versions to implement loops, and
   probably for other purposes.

   This version does not use any additional memory as it runs. No
   memory is allocated by the source code.

   This version does not suppress superfluous commands, such as a command
   to start the spindle when the spindle is already turning, or a command
   to turn on flood coolant, when flood coolant is already on.  When the
   interpreter is being used for direct control of the machining center,
   suppressing superfluous commands might confuse the user and could be
   dangerous, but when it is used to translate from one file to another,
   suppression can produce more concise output. Future versions might
   include an option for suppressing superfluous commands.

*/

/****************************************************************************/

type Rs274ngc_i interface {
	SetCanon(c inc.Canon_i)
	// open NC-program file
	Open(filename string) inc.STATUS
	// read the command
	Read(command []byte) inc.STATUS
	Close() inc.STATUS
	// execute a line of NC code
	Execute() int
	// stop running
	exit()
	// get ready to run
	Init() inc.STATUS
	// load a tool table
	load_tool_table() inc.STATUS
	// reset yourself
	reset()
	// restore interpreter variables from a file
	restore_parameters(filename string) inc.STATUS
	//TODO
	////default_parameters()
	// save interpreter variables to file
	save_parameters(filename string, parameters []float64) inc.STATUS
	// synchronize your internal model with the external world
	synch() inc.STATUS

	// copy active G codes into array [0]..[11]
	active_g_codes(codes []inc.GCodes)
	// copy active M codes into array [0]..[6]
	active_m_codes(codes []int)
	// copy active F, S settings into array [0]..[2]
	active_settings(settings []float64)

	// return the length of the most recently read line
	Line_length() uint
	// copy the text of the most recently read line into the line_text array,
	// but stop at max_size if the text is longer
	LineText() string
}
type Rs274ngc_t = rs274ngc_t

var _ Rs274ngc_i = &rs274ngc_t{}

type rs274ngc_t struct {
	_setup Setup_t

	canon inc.Canon_i
}

func (cnc *rs274ngc_t) SetCanon(c inc.Canon_i) {
	cnc.canon = c
}

/***********************************************************************/

/* rs274ngc_open

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise it returns RS274NGC_OK.
   1. A file is already open: NCE_A_FILE_IS_ALREADY_OPEN
   2. The name of the file is too long: NCE_FILE_NAME_TOO_LONG
   3. The file cannot be opened: NCE_UNABLE_TO_OPEN_FILE

   Side Effects: See below

   Called By: external programs

   The file is opened for reading and _setup.file_pointer is set.
   The file name is copied into _setup.filename.
   The _setup.sequence_number, is set to zero.
   rs274ngc_reset() is called, changing several more _setup attributes.

   The manual [NCMS, page 3] discusses the use of the "%" character at the
   beginning of a "tape". It is not clear whether it is intended that
   every NC-code file should begin with that character.

   In the following, "uses percents" means the first non-blank line
   of the file must consist of nothing but the percent sign, with optional
   leading and trailing white space, and there must be a second line
   of the same sort later on in the file. If a file uses percents,
   execution stops at the second percent line. Any lines after the
   second percent line are ignored.

   In this interpreter (recalling that M2 and M30 always ends execution):
   1. If execution of a file is ended by M2 or M30 (not necessarily on
   the last line of the file), then it is optional that the file
   uses percents.
   2. If execution of a file is not ended by M2 or M30, then it is
   required that the file uses percents.

   If the file being opened uses percents, this function turns on the
   _setup.percent flag, reads any initial blank lines, and reads the
   first line with the "%". If not, after reading enough to determine
   that, this function puts the file pointer back at the beginning of the
   file.

*/
func (cnc *rs274ngc_t) Open( /* ARGUMENTS                                     */
	filename string) inc.STATUS { /* string: the name of the input NC-program file */

	//static char name[] = "rs274ngc_open";

	if cnc._setup.file_pointer.IsInited() == true {
		return inc.NCE_A_FILE_IS_ALREADY_OPEN
	}

	if s := cnc._setup.file_pointer.Init(filename); s != inc.RS274NGC_OK {
		return s
	}

	cnc._setup.percent_flag = OFF
	for { /* skip blank lines */
		if l, _ := cnc._setup.file_pointer.R.ReadBytes('\n'); len(l) == 0 {
			return inc.NCE_FILE_ENDED_WITH_NO_PERCENT_SIGN
		} else if l = bytes.TrimSpace(l); len(l) != 0 {
			if l[0] == '%' {
				cnc._setup.percent_flag = ON
			}
			break
		}
	}
	if cnc._setup.percent_flag != ON {
		cnc._setup.file_pointer.Reset()
	}
	cnc._setup.filename = filename
	cnc._setup.sequence_number = 0

	cnc.reset()
	return inc.RS274NGC_OK
}

/***********************************************************************/

/* rs274ngc_close

   Returned Value: int (RS274NGC_OK)

   Side Effects:
   The NC-code file is closed if open.
   The _setup world model is reset.

   Called By: external programs

*/

func (cnc *rs274ngc_t) Close() inc.STATUS {
	cnc._setup.file_pointer.Close()
	cnc.reset()

	return inc.RS274NGC_OK
}

/***********************************************************************/

/* rs274ngc_read

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, this returns:
   a. RS274NGC_ENDFILE if the only non-white character on the line is %,
   b. RS274NGC_EXECUTE_FINISH if the first character of the
   close_and_downcased line is a slash, and
   c. RS274NGC_OK otherwise.
   1. The command and_setup.file_pointer are both NULL: NCE_FILE_NOT_OPEN
   2. The probe_flag is ON but the HME command queue is not empty:
   NCE_QUEUE_IS_NOT_EMPTY_AFTER_PROBING
   3. If read_text (which gets a line of NC code from file) or parse_line
   (which parses the line) returns an error code, this returns that code.

   Side Effects:
   _setup.sequence_number is incremented.
   The _setup.block1 is filled with data.

   Called By: external programs

   This reads a line of NC-code from the command string or, (if the
   command string is NULL) from the currently open file. The
   _setup.line_length will be set by read_text. This will be zero if the
   line is blank or consists of nothing but a slash. If the length is not
   zero, this parses the line into the _setup.block1.

*/

func (cnc *rs274ngc_t) Read(command []byte) inc.STATUS { /* a string to read */
	//static char name[] = "rs274ngc_read";

	//var status inc.STATUS

	if cnc._setup.probe_flag == ON {
		if 0 == cnc.canon.GET_EXTERNAL_QUEUE_EMPTY() {
			return inc.NCE_QUEUE_IS_NOT_EMPTY_AFTER_PROBING
		}
		cnc.set_probe_data()
		cnc._setup.probe_flag = OFF
	}
	if command == nil && false == cnc._setup.file_pointer.IsInited() {
		return inc.NCE_FILE_NOT_OPEN
	}
	var read_status inc.STATUS
	cnc._setup.linetext, cnc._setup.blocktext, cnc._setup.line_length, read_status =
		cnc.read_text(command)
	if read_status == inc.RS274NGC_EXECUTE_FINISH || read_status == inc.RS274NGC_OK {
		if cnc._setup.line_length != 0 {
			cnc.parse_line( /*cnc._setup.blocktext*/ )
		}

	} else if read_status == inc.RS274NGC_ENDFILE {

	} else {
		return (read_status)
	}
	return read_status
}

/***********************************************************************/

/* rs274ngc_execute

   Returned Value: int)
   If execute_block returns RS274NGC_EXIT, this returns that.
   If execute_block returns RS274NGC_EXECUTE_FINISH, this returns that.
   If execute_block returns an error code, this returns that code.
   Otherwise, this returns RS274NGC_OK.

   Side Effects:
   Calls to canonical machining commands are made.
   The interpreter variables are changed.
   At the end of the program, the file is closed.
   If using a file, the active G codes and M codes are updated.

   Called By: external programs

   This executes a previously parsed block.

*/

func (cnc *rs274ngc_t) Execute() inc.STATUS { /* NO ARGUMENTS */

	var status inc.STATUS

	if cnc._setup.line_length != 0 { /* line not blank */
		for n := int64(0); n < cnc._setup.block1.Parameter_occurrence; n++ {
			// copy parameter settings from parameter buffer into parameter table
			cnc._setup.parameters[cnc._setup.block1.Parameter_numbers[n]] = cnc._setup.block1.Parameter_values[n]
		}

		status = cnc.execute_block(&(cnc._setup.block1), &cnc._setup)
		cnc._setup.Write_g_codes(&(cnc._setup.block1))
		cnc._setup.Write_m_codes(&(cnc._setup.block1))
		cnc._setup.Write_settings()
		if (status != inc.RS274NGC_OK) &&
			(status != inc.RS274NGC_EXECUTE_FINISH) &&
			(status != inc.RS274NGC_EXIT) {
			return (status)
		}
	} else { /* blank line is OK */
		status = inc.RS274NGC_OK
	}
	return status
}

/****************************************************************************/

/* execute_block

   Returned Value: int
   If convert_stop returns RS274NGC_EXIT, this returns RS274NGC_EXIT.
   If any of the following functions is called and returns an error code,
   this returns that code.
   convert_comment
   convert_feed_mode
   convert_feed_rate
   convert_g
   convert_m
   convert_speed
   convert_stop
   convert_tool_select
   Otherwise, if the probe_flag in the settings is ON, this returns
   RS274NGC_EXECUTE_FINISH.
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   One block of RS274/NGC instructions is executed.

   Called by:
   rs274ngc_execute

   This converts a block to zero to many actions. The order of execution
   of items in a block is critical to safe and effective machine operation,
   but is not specified clearly in the RS274/NGC documentation.

   Actions are executed in the following order:
   1. any comment.
   2. a feed mode setting (g93, g94)
   3. a feed rate (f) setting if in units_per_minute feed mode.
   4. a spindle speed (s) setting.
   5. a tool selection (t).
   6. "m" commands as described in convert_m (includes tool change).
   7. any g_codes (except g93, g94) as described in convert_g.
   8. stopping commands (m0, m1, m2, m30, or m60).

   In inverse time feed mode, the explicit and implicit g code executions
   include feed rate setting with g1, g2, and g3. Also in inverse time
   feed mode, attempting a canned cycle cycle (g81 to g89) or setting a
   feed rate with g0 is illegal and will be detected and result in an
   error message.

*/

func (cnc *rs274ngc_t) execute_block( /* ARGUMENTS                                    */
	block *Block_t, /* pointer to a block of RS274/NGC instructions */
	settings *Setup_t) inc.STATUS { /* pointer to machine settings                  */

	//static char name[] = "execute_block";
	var status inc.STATUS

	if 0 == len(cnc._setup.block1.comment) {
		cnc.convert_comment(cnc._setup.block1.comment)
	}
	if cnc._setup.block1.g_modes[5] != -1 {
		cnc.convert_feed_mode(cnc._setup.block1.g_modes[5])
	}
	if cnc._setup.block1.f_number > -1.0 {
		/* handle elsewhere */
		if cnc._setup.feed_mode == inc.INVERSE_TIME {

		} else {
			cnc.convert_feed_rate()
		}
	}
	if cnc._setup.block1.s_number > -1.0 {
		cnc.convert_speed()
	}
	if cnc._setup.block1.t_number != -1 {
		cnc.convert_tool_select()
	}
	cnc.convert_m()
	cnc.convert_g()

	if cnc._setup.block1.m_modes[4] != -1 { /* converts m0, m1, m2, m30, or m60 */
		status = cnc.convert_stop()
		if status == inc.RS274NGC_EXIT {
			return inc.RS274NGC_EXIT

		} else if status != inc.RS274NGC_OK {
			return status
		}
	}
	return inc.If(cnc._setup.probe_flag == ON, inc.RS274NGC_EXECUTE_FINISH, inc.RS274NGC_OK).(inc.STATUS)

}

/****************************************************************************/

/* set_probe_data

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The current position is set.
   System parameters for probe position are set.

   Called by:  rs274ngc_read

*/

func (cnc *rs274ngc_t) set_probe_data() inc.STATUS { /* pointer to machine settings */
	//static char name[] = "set_probe_data";
	cnc._setup.current.X = cnc.canon.GET_EXTERNAL_POSITION_X()
	cnc._setup.current.Y = cnc.canon.GET_EXTERNAL_POSITION_Y()
	cnc._setup.current.Z = cnc.canon.GET_EXTERNAL_POSITION_Z()
	cnc._setup.current.X = cnc.canon.GET_EXTERNAL_POSITION_X()

	cnc._setup.current.A = cnc.canon.GET_EXTERNAL_POSITION_A()
	cnc._setup.current.B = cnc.canon.GET_EXTERNAL_POSITION_B()
	cnc._setup.current.C = cnc.canon.GET_EXTERNAL_POSITION_C()

	cnc._setup.parameters[5061] = cnc.canon.GET_EXTERNAL_PROBE_POSITION_X()
	cnc._setup.parameters[5062] = cnc.canon.GET_EXTERNAL_PROBE_POSITION_Y()
	cnc._setup.parameters[5063] = cnc.canon.GET_EXTERNAL_PROBE_POSITION_Z()

	cnc._setup.parameters[5064] = cnc.canon.GET_EXTERNAL_PROBE_POSITION_A()
	cnc._setup.parameters[5065] = cnc.canon.GET_EXTERNAL_PROBE_POSITION_B()
	cnc._setup.parameters[5066] = cnc.canon.GET_EXTERNAL_PROBE_POSITION_C()

	cnc._setup.parameters[5067] = cnc.canon.GET_EXTERNAL_PROBE_VALUE()
	return inc.RS274NGC_OK

}

/****************************************************************************/

/* read_text

   Returned Value: int
   If close_and_downcase returns an error code, this returns that code.
   If any of the following errors occur, this returns the error code shown.
   Otherwise, this returns:
   a. RS274NGC_ENDFILE if the percent_flag is ON and the only
   non-white character on the line is %,
   b. RS274NGC_EXECUTE_FINISH if the first character of the
   close_and_downcased line is a slash, and
   c. RS274NGC_OK otherwise.
   1. The end of the file is found and the percent_flag is ON:
   NCE_FILE_ENDED_WITH_NO_PERCENT_SIGN
   2. The end of the file is found and the percent_flag is OFF:
   NCE_FILE_ENDED_WITH_NO_PERCENT_SIGN_OR_program_end
   3. The command argument is not null and is too long or the command
   argument is null and the line read from the file is too long:
   NCE_COMMAND_TOO_LONG

   Side effects: See below

   Called by:  rs274ngc_read

   This reads a line of RS274 code from a command string or a file into
   the line array. If the command string is not null, the file is ignored.

   If the end of file is reached, an error is returned as described
   above. The end of the file should not be reached because (a) if the
   file started with a percent line, it must end with a percent line, and
   no more reading of the file should occur after that, and (b) if the
   file did not start with a percent line, it must have a program ending
   command (M2 or M30) in it, and no more reading of the file should
   occur after that.

   All blank space at the end of a line read from a file is removed and
   replaced here with NULL characters.

   This then calls close_and_downcase to downcase and remove tabs and
   spaces from everything on the line that is not part of a comment. Any
   comment is left as is.

   The length is set to zero if any of the following occur:
   1. The line now starts with a slash, but the second character is NULL.
   2. The first character is NULL.
   Otherwise, length is set to the length of the line.

   An input line is blank if the first character is NULL or it consists
   entirely of tabs and spaces and, possibly, a newline before the first
   NULL.

   Block delete is discussed in [NCMS, page 3] but the discussion makes
   no sense. Block delete is handled by having this function return
   RS274NGC_EXECUTE_FINISH if the first character of the
   close_and_downcased line is a slash. When the caller sees this,
   the caller is expected not to call rs274ngc_execute if the switch
   is on, but rather call rs274ngc_read again to overwrite and ignore
   what is read here.

   The value of the length argument is set to the number of characters on
   the reduced line.

*/

// inport *os.File,  a file pointer for an input file, or null
//out put
//raw_line []byte,  array to write raw input line into
// line []byte array for input line to be processed in
// length to be set
func (cnc *rs274ngc_t) read_text( /* ARGUMENTS                                   */
	command []byte /* a string which has input text */) (raw_line string, line string, length uint, s inc.STATUS) {

	var (
		err error
	)

	if command == nil {
		if raw_line, err = cnc._setup.file_pointer.R.ReadString('\n'); err != nil {
			if cnc._setup.percent_flag == ON {
				s = inc.NCE_FILE_ENDED_WITH_NO_PERCENT_SIGN
				return
			} else {
				s = inc.NCE_FILE_ENDED_WITH_NO_PERCENT_SIGN_OR_PROGRAM_END
				return
			}
		}

		line = strings.TrimSpace(raw_line)

		if line, s = close_and_downcase(line); s != inc.RS274NGC_OK {
			return
		}
		length = uint(len(line))
		if line[0] == '%' && cnc._setup.percent_flag == ON {
			s = inc.RS274NGC_ENDFILE
			return
		}
	} else {
		//if len(command) >= inc.RS274NGC_TEXT_SIZE {
		//	return inc.NCE_COMMAND_TOO_LONG
		//}
		raw_line = string(command)
		line = string(command)
		if line, s = close_and_downcase(line); s != inc.RS274NGC_OK {
			return
		}
		length = uint(len(line))
	}

	cnc._setup.sequence_number++
	cnc._setup.block1.Parameter_occurrence = 0 /* initialize parameter buffer */

	// an optional block delete character, which is a slash “/” .
	executeFinish := (line[0] == '/')
	if 0 == len(line) || (executeFinish && 1 == len(line)) {
		length = 0
	} else {
		length = uint(len(line))
	}
	s = inc.If(executeFinish, inc.RS274NGC_EXECUTE_FINISH, inc.RS274NGC_OK).(inc.STATUS)

	return //raw_line, line, length, inc.If(executeFinish, inc.RS274NGC_EXECUTE_FINISH, inc.RS274NGC_OK).(inc.STATUS)

}

/****************************************************************************/

/* parse_line

   Returned Value: int
   If any of the following functions returns an error code,
   this returns that code.
   init_block
   read_items
   enhance_block
   check_items
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   One RS274 line is read into a block and the block is checked for
   errors. System parameters may be reset.

   Called by:  rs274ngc_read

*/

// to fill block
func (cnc *rs274ngc_t) parse_line() inc.STATUS {

	cnc._setup.block1.Init_block()

	cnc._setup.block1.Read_items(cnc._setup.tool_max, cnc._setup.blocktext, cnc._setup.parameters)
	cnc._setup.block1.Enhance_block(&cnc._setup)
	cnc._setup.block1.Check_items(&cnc._setup)
	return inc.RS274NGC_OK
}

/***********************************************************************/

/* rs274ngc_active_g_codes

   Returned Value: none

   Side Effects: copies active G codes into the codes array

   Called By: external programs

   See documentation of write_g_codes.

*/

func (cnc *rs274ngc_t) active_g_codes( /* ARGUMENTS                   */
	codes []inc.GCodes) { /* array of codes to copy into */

	for n := 0; n < inc.RS274NGC_ACTIVE_G_CODES; n++ {
		codes[n] = cnc._setup.active_g_codes[n]
	}
}

/***********************************************************************/

/* rs274ngc_active_m_codes

   Returned Value: none

   Side Effects: copies active M codes into the codes array

   Called By: external programs

   See documentation of write_m_codes.

*/

func (cnc *rs274ngc_t) active_m_codes( /* ARGUMENTS                   */
	codes []int) { /* array of codes to copy into */

	for n := 0; n < inc.RS274NGC_ACTIVE_M_CODES; n++ {
		codes[n] = cnc._setup.active_m_codes[n]
	}
}

/***********************************************************************/

/* rs274ngc_active_settings

   Returned Value: none

   Side Effects: copies active F, S settings into array

   Called By: external programs

   See documentation of write_settings.

*/

func (cnc *rs274ngc_t) active_settings( /* ARGUMENTS                      */
	settings []float64) { /* array of settings to copy into */
	for n := 0; n < inc.RS274NGC_ACTIVE_SETTINGS; n++ {
		settings[n] = cnc._setup.active_settings[n]
	}
}

/***********************************************************************/

/* rs274_ngc_init

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, this returns RS274NGC_OK.
   1. rs274ngc_restore_parameters returns an error code.
   2. Parameter 5220, the work coordinate system index, is not in the range
   1 to 9: NCE_COORDINATE_SYSTEM_INDEX_PARAMETER_5220_OUT_OF_RANGE

   Side Effects:
   Many values in the _setup structure are reset.
   A USE_LENGTH_UNITS canonical command call is made.
   A SET_FEED_REFERENCE canonical command call is made.
   A SET_ORIGIN_OFFSETS canonical command call is made.
   An INIT_CANON call is made.

   Called By: external programs

   Currently we are running only in CANON_XYZ feed_reference mode.  There
   is no command regarding feed_reference in the rs274 language (we
   should try to get one added). The initialization routine, therefore,
   always calls SET_FEED_REFERENCE(CANON_XYZ).

*/

func (cnc *rs274ngc_t) Init() inc.STATUS { /* NO ARGUMENTS */

	//int k;                                    // starting index in parameters of origin offsets
	//int status;
	//char filename[RS274NGC_TEXT_SIZE];
	var (
	//name string = "rs274ngc_init"
	//pars     *float64
	) // short name for _setup.parameters
	cnc._setup.parameters = make([]float64, inc.RS274NGC_MAX_PARAMETERS)

	cnc.canon.INIT_CANON()
	cnc._setup.length_units = cnc.canon.GET_EXTERNAL_LENGTH_UNIT_TYPE()
	cnc.canon.USE_LENGTH_UNITS(cnc._setup.length_units)
	filename := cnc.canon.GET_EXTERNAL_PARAMETER_FILE_NAME()
	if len(filename) != 0 {
		cnc.restore_parameters(string(filename[:]))
	} else {
		cnc.restore_parameters(RS274NGC_PARAMETER_FILE_NAME_DEFAULT)
	}

	pars := (cnc._setup.parameters)
	cnc._setup.origin_index = (int)(pars[5220] + 0.0001)
	if (cnc._setup.origin_index < 1) || (cnc._setup.origin_index > 9) {
		return inc.NCE_COORDINATE_SYSTEM_INDEX_PARAMETER_5220_OUT_OF_RANGE
	}

	k := (5200 + (cnc._setup.origin_index * 20))
	cnc.canon.SET_ORIGIN_OFFSETS(
		(pars[k+1] + pars[5211]), /*x*/
		(pars[k+2] + pars[5212]), /*y*/
		(pars[k+3] + pars[5213]), /*z*/
		(pars[k+4] + pars[5214]), //AA
		(pars[k+5] + pars[5215]), //BB
		(pars[k+6] + pars[5216])) //CC

	cnc.canon.SET_FEED_REFERENCE(inc.CANON_XYZ)

	cnc._setup.axis_offset.A = pars[5214] /*AA*/

	//_setup.Aa_current set in rs274ngc_synch

	/*AA*/
	cnc._setup.origin_offset.A = pars[k+4]

	//_setup.active_g_codes initialized below
	//_setup.active_m_codes initialized below
	//_setup.active_settings initialized below
	cnc._setup.axis_offset.X = pars[5211]
	cnc._setup.axis_offset.Y = pars[5212]
	cnc._setup.axis_offset.Z = pars[5213]

	cnc._setup.axis_offset.B = pars[5215] /*BB*/

	//_setup.Bb_current set in rs274ngc_synch

	/*BB*/
	cnc._setup.origin_offset.B = pars[k+5]

	//_setup.block1 does not need initialization
	cnc._setup.blocktext = ""

	cnc._setup.axis_offset.C = pars[5216] /*CC*/

	//_setup.Cc_current set in rs274ngc_synch

	/*CC*/
	cnc._setup.origin_offset.C = pars[k+6]

	//_setup.current_slot set in rs274ngc_synch
	//_setup.current.X set in rs274ngc_synch
	//_setup.current.Y set in rs274ngc_synch
	//_setup.current.Z set in rs274ngc_synch
	cnc._setup.cutter_comp_side = inc.CANON_SIDE_OFF
	//_setup.cycle values do not need initialization
	cnc._setup.distance_mode = inc.MODE_ABSOLUTE
	cnc._setup.feed_mode = inc.UNITS_PER_MINUTE
	cnc._setup.feed_override = ON
	//_setup.feed_rate set in rs274ngc_synch
	cnc._setup.filename = ""
	//cnc._setup.file_pointer = nil
	//_setup.flood set in rs274ngc_synch
	cnc._setup.length_offset_index = 1
	//_setup.length_units set in rs274ngc_synch
	cnc._setup.line_length = 0
	cnc._setup.linetext = ""
	//_setup.mist set in rs274ngc_synch
	cnc._setup.motion_mode = inc.G_80
	//_setup.origin_index set above
	cnc._setup.origin_offset.X = pars[k+1]
	cnc._setup.origin_offset.Y = pars[k+2]
	cnc._setup.origin_offset.Z = pars[k+3]
	//_setup.parameters set above
	//_setup.parameter_occurrence does not need initialization
	//_setup.parameter_numbers does not need initialization
	//_setup.parameter_values does not need initialization
	//_setup.percent_flag does not need initialization
	//_setup.plane set in rs274ngc_synch
	cnc._setup.probe_flag = OFF
	cnc._setup.program_x = inc.UNKNOWN /* for cutter comp */
	cnc._setup.program_y = inc.UNKNOWN /* for cutter comp */
	//_setup.retract_mode does not need initialization
	//_setup.selected_tool_slot set in rs274ngc_synch
	cnc._setup.sequence_number = 0 /*DOES THIS NEED TO BE AT TOP? */
	//_setup.speed set in rs274ngc_synch
	cnc._setup.speed_feed_mode = inc.CANON_INDEPENDENT
	cnc._setup.speed_override = ON
	//_setup.spindle_turning set in rs274ngc_synch
	//_setup.stack does not need initialization
	//_setup.stack_index does not need initialization
	cnc._setup.tool_length_offset = 0.0
	//_setup.tool_max set in rs274ngc_synch
	//_setup.tool_table set in rs274ngc_synch
	cnc._setup.tool_table_index = 1
	//_setup.traverse_rate set in rs274ngc_synch

	cnc._setup.Write_g_codes(nil)
	cnc._setup.Write_m_codes(nil)
	cnc._setup.Write_settings()

	// Synch rest of settings to external world
	cnc.synch()

	return inc.RS274NGC_OK
}

/*

   This is an array of the index numbers of system parameters that must
   be included in a file used with the rs274ngc_restore_parameters
   function. The array is used by that function and by the
   rs274ngc_save_parameters function.

*/

var _required_parameters = [...]uint{
	5161 /*X*/, 5162 /*Y*/, 5163, /*Z*/ /* G28 home */
	5164 /*A*/, 5165 /*B*/, 5166, /*C*/

	5181, 5182, 5183, /* G30 home */
	5184, 5185, 5186,

	5211, 5212, 5213, /* G92 offsets */
	5214, 5215, 5216,

	5220, /* selected coordinate */

	5221, 5222, 5223, /* coordinate system 1 */
	5224, 5225, 5226,

	5241, 5242, 5243, /* coordinate system 2 */
	5244, 5245, 5246,

	5261, 5262, 5263, /* coordinate system 3 */
	5264, 5265, 5266,

	5281, 5282, 5283, /* coordinate system 4 */
	5284, 5285, 5286,

	5301, 5302, 5303, /* coordinate system 5 */
	5304, 5305, 5306,

	5321, 5322, 5323, /* coordinate system 6 */
	5324, 5325, 5326,

	5341, 5342, 5343, /* coordinate system 7 */
	5344, 5345, 5346,

	5361, 5362, 5363, /* coordinate system 8 */
	5364, 5365, 5366,

	5381, 5382, 5383, /* coordinate system 9 */
	5384, 5385, 5386,
	inc.RS274NGC_MAX_PARAMETERS}

/***********************************************************************/

/* rs274ngc_restore_parameters

   Returned Value:
   If any of the following errors occur, this returns the error code shown.
   Otherwise it returns RS274NGC_OK.
   1. The parameter file cannot be opened for reading: NCE_UNABLE_TO_OPEN_FILE
   2. A parameter index is out of range: NCE_PARAMETER_NUMBER_OUT_OF_RANGE
   3. A required parameter is missing from the file:
   NCE_REQUIRED_PARAMETER_MISSING
   4. The parameter file is not in increasing order:
   NCE_PARAMETER_FILE_OUT_OF_ORDER

   Side Effects: See below

   Called By:
   external programs
   rs274ngc_init

   This function restores the parameters from a file, modifying the
   parameters array. Usually parameters is _setup.parameters. The file
   contains lines of the form:

   <variable number> <value>

   e.g.,

   5161 10.456

   The variable numbers must be in increasing order, and certain
   parameters must be included, as given in the _required_parameters
   array. These are the axis offsets, the origin index (5220), and nine
   sets of origin offsets. Any parameter not given a value in the file
   has its value set to zero.

*/
func (cnc *rs274ngc_t) restore_parameters( /* ARGUMENTS                        */
	filename string) inc.STATUS { /* name of parameter file to read   */

	var (
		variable int
		//line     [256]byte
		k     uint
		value float64
	)

	// open original for reading
	infile, err := os.Open(filename)
	if err != nil {
		return inc.NCE_UNABLE_TO_OPEN_FILE
	}
	defer infile.Close()

	k = 0
	index := 0
	required := _required_parameters[index]
	index++

	r := bufio.NewReader(infile)
	for {
		//buf,err := r.ReadBytes('\n')
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		if n, _ := fmt.Sscanf(string(line[:]), "%d %f", &variable, &value); n == 2 {
			if (variable <= 0) || variable >= inc.RS274NGC_MAX_PARAMETERS {
				return inc.NCE_PARAMETER_NUMBER_OUT_OF_RANGE
			}

			for ; k < inc.RS274NGC_MAX_PARAMETERS; k++ {
				if int(k) > variable {
					return inc.NCE_PARAMETER_FILE_OUT_OF_ORDER
				} else if int(k) == variable {
					cnc._setup.parameters[k] = value
					if k == required {
						required = _required_parameters[index]
						index++
					}

					k++
					break
				} else { // if (k < variable)
					if k == required {
						return inc.NCE_REQUIRED_PARAMETER_MISSING
					} else {
						cnc._setup.parameters[k] = 0
					}

				}
			}
		}
	}

	if required != inc.RS274NGC_MAX_PARAMETERS {
		return inc.NCE_REQUIRED_PARAMETER_MISSING
	}
	for ; k < inc.RS274NGC_MAX_PARAMETERS; k++ {
		cnc._setup.parameters[k] = 0
	}

	return inc.RS274NGC_OK
}

/***********************************************************************/

/* rs274ngc_save_parameters

   Returned Value:
   If any of the following errors occur, this returns the error code shown.
   Otherwise it returns RS274NGC_OK.
   1. The existing file cannot be renamed:  NCE_CANNOT_CREATE_BACKUP_FILE
   2. The renamed file cannot be opened to read: NCE_CANNOT_OPEN_BACKUP_FILE
   3. The new file cannot be opened to write: NCE_CANNOT_OPEN_VARIABLE_FILE
   4. A parameter index is out of range: NCE_PARAMETER_NUMBER_OUT_OF_RANGE
   5. The renamed file is out of order: NCE_PARAMETER_FILE_OUT_OF_ORDER

   Side Effects: See below

   Called By:
   external programs
   rs274ngc_exit

   A file containing variable-value assignments is updated. The old
   version of the file is saved under a different name.  For each
   variable-value pair in the old file, a line is written in the new file
   giving the current value of the variable.  File lines have the form:

   <variable number> <value>

   e.g.,

   5161 10.456

   If a required parameter is missing from the input file, this does not
   complain, but does write it in the output file.

*/
func (cnc *rs274ngc_t) save_parameters( /* ARGUMENTS             */
	filename string, /* name of file to write */
	parameters []float64) inc.STATUS { /* parameters to save    */

	//    FILE * infile;
	//       FILE * outfile;
	//       char line[256];
	//       int variable;
	//       double value;
	//       unsigned int required;                             // number of next required parameter
	//       int index;                                // index into _required_parameters

	const (
		RS274NGC_PARAMETER_FILE_BACKUP_SUFFIX = ".bak"
	)
	// rename as .bak
	bak := filename + RS274NGC_PARAMETER_FILE_BACKUP_SUFFIX
	err := os.Rename(filename, bak)
	if err != nil {
		return inc.NCE_CANNOT_CREATE_BACKUP_FILE
	}

	// open backup for reading
	infile, err := os.Open(bak)
	if err != nil {
		return inc.NCE_CANNOT_OPEN_BACKUP_FILE
	}
	defer infile.Close()

	// open original for writing
	outfile, err := os.OpenFile(filename, os.O_RDWR, 0660)
	if err != nil {
		return inc.NCE_CANNOT_OPEN_VARIABLE_FILE
	}
	defer outfile.Close()

	var (
		variable int
		value    float64
		k        uint = 0
		index         = 0
		required      = _required_parameters[index]
	)
	index++

	r := bufio.NewReader(infile)
	for {
		//buf,err := r.ReadBytes('\n')
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		// try for a variable-value match
		if n, _ := fmt.Sscanf(string(line[:]), "%d %f", &variable, &value); n == 2 {
			if (variable <= 0) || variable >= inc.RS274NGC_MAX_PARAMETERS {
				return inc.NCE_PARAMETER_NUMBER_OUT_OF_RANGE
			}

			for ; k < inc.RS274NGC_MAX_PARAMETERS; k++ {
				if k > uint(variable) {
					return inc.NCE_PARAMETER_FILE_OUT_OF_ORDER
				} else if k == uint(variable) {
					line = fmt.Sprintf("%d\t%f\n", k, parameters[k])
					outfile.WriteString(line)
					//sprintf(line, "%d\t%f\n", k, parameters[k])
					//fputs(line, outfile)
					if k == required {
						required = _required_parameters[index]
						index++
					}

					k++
					break
				} else if k == required { // know (k < variable)
					line = fmt.Sprintf("%d\t%f\n", k, parameters[k])
					outfile.WriteString(line)
					//sprintf(line, "%d\t%f\n", k, parameters[k])
					//fputs(line, outfile)
					required = _required_parameters[index]
					index++
				}
			}
		}
	}

	for ; k < inc.RS274NGC_MAX_PARAMETERS; k++ {
		if k == required {
			line := fmt.Sprintf("%d\t%f\n", k, parameters[k])
			outfile.WriteString(line)
			//sprintf(line, "%d\t%f\n", k, parameters[k])
			//fputs(line, outfile)
			required = _required_parameters[index]
			index++
		}
	}
	return inc.RS274NGC_OK
}

/***********************************************************************/

/* rs274ngc_exit

   Returned Value: int (RS274NGC_OK)

   Side Effects: See below

   Called By: external programs

   The system parameters are saved to a file and some parts of the world
   model are reset. If GET_EXTERNAL_PARAMETER_FILE_NAME provides a
   non-empty file name, that name is used for the file that is
   written. Otherwise, the default parameter file name is used.

*/

func (cnc *rs274ngc_t) exit() { /* NO ARGUMENTS */

	file_name := cnc.canon.GET_EXTERNAL_PARAMETER_FILE_NAME()
	if len(file_name) == 0 {
		cnc.save_parameters(RS274NGC_PARAMETER_FILE_NAME_DEFAULT, cnc._setup.parameters)
	} else {
		cnc.save_parameters(string(file_name[:]), cnc._setup.parameters)
	}

	cnc.reset()
}

/***********************************************************************/

/* rs274ngc_line_length

   Returned Value: the length of the most recently read line

   Side Effects: none

   Called By: external programs

*/

func (cnc *rs274ngc_t) Line_length() uint {
	return cnc._setup.line_length
}

/***********************************************************************/

/* rs274ngc_line_text

Returned Value: none

Side Effects: See below

Called By: external programs

This copies at most (max_size - 1) non-null characters of the most
recently read line into the line_text string and puts a NULL after the
last non-null character.

*/

func (cnc *rs274ngc_t) LineText() string {
	return cnc._setup.linetext
}

/***********************************************************************/

/* rs274ngc_load_tool_table

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, this returns RS274NGC_OK.
   1. _setup.tool_max is larger than CANON_TOOL_MAX: NCE_TOOL_MAX_TOO_LARGE

   Side Effects:
   _setup.tool_table[] is modified.

   Called By:
   rs274ngc_synch
   external programs

   This function calls the canonical interface function GET_EXTERNAL_TOOL_TABLE
   to load the whole tool table into the _setup.

   The CANON_TOOL_MAX is an upper limit for this software. The
   _setup.tool_max is intended to be set for a particular machine.

*/

func (cnc *rs274ngc_t) load_tool_table() inc.STATUS { /* NO ARGUMENTS */

	var n int

	if cnc._setup.tool_max > inc.CANON_TOOL_MAX {
		return inc.NCE_TOOL_MAX_TOO_LARGE
	}
	for n = 0; uint(n) <= cnc._setup.tool_max; n++ {
		cnc._setup.tool_table[n] = cnc.canon.GET_EXTERNAL_TOOL_TABLE(n)
	}

	for ; n <= inc.CANON_TOOL_MAX; n++ {
		cnc._setup.tool_table[n] = inc.CANON_TOOL_TABLE{}

	}

	return inc.RS274NGC_OK
}

/***********************************************************************/

/* rs274ngc_reset

   Returned Value: int (RS274NGC_OK)

   Side Effects: See below

   Called By:
   external programs
   rs274ngc_close
   rs274ngc_exit
   rs274ngc_open

   This function resets the parts of the _setup model having to do with
   reading and interpreting one line. It does not affect the parts of the
   model dealing with a file being open; rs274ngc_open and rs274ngc_close
   do that.

   There is a hierarchy of resetting the interpreter. Each of the following
   calls does everything the ones above it do.

   rs274ngc_reset()
   rs274ngc_close()
   rs274ngc_init()

   In addition, rs274ngc_synch and rs274ngc_restore_parameters (both of
   which are called by rs274ngc_init) change the model.

*/

func (cnc *rs274ngc_t) reset() {
	cnc._setup.linetext = ""
	cnc._setup.blocktext = ""
	cnc._setup.line_length = 0
}

/****************************************************************************/

/* close_and_downcase

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. A left parenthesis is found inside a comment:
   NCE_NESTED_COMMENT_FOUND
   2. The line ends before an open comment is closed:
   NCE_UNCLOSED_COMMENT_FOUND
   3. A newline character is found that is not followed by null:
   NCE_NULL_MISSING_AFTER_NEWLINE

   Side effects: See below

   Called by:  read_text

   To simplify handling upper case letters, spaces, and tabs, this
   function removes spaces and and tabs and downcases everything on a
   line which is not part of a comment.

   Comments are left unchanged in place. Comments are anything
   enclosed in parentheses. Nested comments, indicated by a left
   parenthesis inside a comment, are illegal.

   The line must have a null character at the end when it comes in.
   The line may have one newline character just before the end. If
   there is a newline, it will be removed.

   Although this software system detects and rejects all illegal characters
   and illegal syntax, this particular function does not detect problems
   with anything but comments.

   We are treating RS274 code here as case-insensitive and spaces and
   tabs as if they have no meaning. [RS274D, page 6] says spaces and tabs
   are to be ignored by control.

   The KT and NGC manuals say nothing about case or spaces and tabs.

*/

func close_and_downcase( /* ARGUMENTS                   */
	line string) (string, inc.STATUS) { /* string: one line of NC code */

	var (
		comment = false
		t       string
	)

	//匹配一个或多个空白符的正则表达式
	reg := regexp.MustCompile("\\s+")
	str := reg.ReplaceAllString(line, "")

	str = strings.ToLower(str)

	for _, item := range str {
		if comment {
			t = t + string(item)
			if item == ')' {
				comment = false
			} else if item == '(' {
				return line, inc.NCE_NESTED_COMMENT_FOUND
			}
		} else if item == '(' { /* comment is starting */
			comment = true
			t = t + string(item)
		} else {
			t = t + string(item) /* copy anything else */
		}
	}

	if comment {
		return t, inc.NCE_UNCLOSED_COMMENT_FOUND
	}

	return t, inc.RS274NGC_OK
}

/***********************************************************************/

/* rs274ngc_synch

   Returned Value: int (RS274NGC_OK)

   Side Effects:
   sets the value of many attribute of _setup by calling various
   GET_EXTERNAL_xxx functions.

   Called By:
   rs274ngc_init
   external programs

   This function gets the _setup world model in synch with the rest of
   the controller.

*/

func (cnc *rs274ngc_t) synch() inc.STATUS { /* NO ARGUMENTS */

	cnc._setup.control_mode = cnc.canon.GET_EXTERNAL_MOTION_CONTROL_MODE()

	cnc._setup.current.A = cnc.canon.GET_EXTERNAL_POSITION_A()
	cnc._setup.current.B = cnc.canon.GET_EXTERNAL_POSITION_B()
	cnc._setup.current.C = cnc.canon.GET_EXTERNAL_POSITION_C()

	cnc._setup.current_slot = cnc.canon.GET_EXTERNAL_TOOL_SLOT()
	cnc._setup.current.X = cnc.canon.GET_EXTERNAL_POSITION_X()
	cnc._setup.current.Y = cnc.canon.GET_EXTERNAL_POSITION_Y()
	cnc._setup.current.Z = cnc.canon.GET_EXTERNAL_POSITION_Z()
	cnc._setup.feed_rate = cnc.canon.GET_EXTERNAL_FEED_RATE()
	cnc._setup.coolant.flood = inc.If(cnc.canon.GET_EXTERNAL_FLOOD() != 0, ON, OFF).(ON_OFF)
	cnc._setup.length_units = cnc.canon.GET_EXTERNAL_LENGTH_UNIT_TYPE()
	cnc._setup.coolant.mist = inc.If(cnc.canon.GET_EXTERNAL_MIST() != 0, ON, OFF).(ON_OFF)
	cnc._setup.plane = cnc.canon.GET_EXTERNAL_PLANE()
	cnc._setup.selected_tool_slot = cnc.canon.GET_EXTERNAL_TOOL_SLOT()
	cnc._setup.speed = cnc.canon.GET_EXTERNAL_SPEED()
	cnc._setup.spindle_turning = cnc.canon.GET_EXTERNAL_SPINDLE()
	cnc._setup.tool_max = uint(cnc.canon.GET_EXTERNAL_TOOL_MAX())
	cnc._setup.traverse_rate = cnc.canon.GET_EXTERNAL_TRAVERSE_RATE()

	cnc.load_tool_table() /*  must set  _setup.tool_max first */

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_comment

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The message function is called if the string starts with "MSG,".
   Otherwise, the comment function is called.

   Called by: execute_block

   To be a message, the first four characters of the comment after the
   opening left parenthesis must be "MSG,", ignoring the case of the
   letters and allowing spaces or tabs anywhere before the comma (to make
   the treatment of case and white space consistent with how it is
   handled elsewhere).

   Messages are not provided for in [NCMS]. They are implemented here as a
   subtype of comment. This is an extension to the rs274NGC language.

*/
func (cnc *rs274ngc_t) convert_comment(comment string /* string with comment */) inc.STATUS {

	//int m;
	//	int item;
	str := []byte(strings.TrimSpace(comment))

	str = bytes.TrimRight(str, "\x00")
	str = bytes.TrimSpace(str)
	if len(str) == 0 {
		return inc.RS274NGC_OK
	}
	item := str[0]

	if (item != 'M') && (item != 'm') {
		cnc.canon.COMMENT(comment)
		return inc.RS274NGC_OK
	} else {
		str = str[1:]
	}

	str = bytes.TrimSpace(str)
	item = str[0]
	if (item != 'S') && (item != 's') {
		cnc.canon.COMMENT(comment)
		return inc.RS274NGC_OK
	} else {
		str = str[1:]
	}
	str = bytes.TrimSpace(str)
	item = str[0]
	if (item != 'G') && (item != 'g') {
		cnc.canon.COMMENT(comment)
		return inc.RS274NGC_OK
	} else {
		str = str[1:]
	}
	str = bytes.TrimSpace(str)
	item = str[0]
	if item != ',' {
		cnc.canon.COMMENT(comment)
		return inc.RS274NGC_OK
	} else {
		cnc.canon.MESSAGE(str[1:])
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_feed_mode

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1.  g_code isn't G_93 or G_94: NCE_BUG_CODE_NOT_G93_OR_G94

   Side effects:
   The interpreter switches the machine settings to indicate the current
   feed mode (UNITS_PER_MINUTE or INVERSE_TIME).

   The canonical machine to which commands are being sent does not have
   a feed mode, so no command setting the distance mode is generated in
   this function. A comment function call is made (conditionally)
   explaining the change in mode, however.

   Called by: execute_block.

*/

func (cnc *rs274ngc_t) convert_feed_mode( /* ARGUMENTS                                 */
	g_code inc.GCodes) inc.STATUS { /* pointer to machine settings                  */

	//static char name[] = "convert_feed_mode";
	if g_code == inc.G_93 {

		cnc.canon.COMMENT(("interpreter: feed mode set to inverse time"))

		cnc._setup.feed_mode = inc.INVERSE_TIME
	} else if g_code == inc.G_94 {

		cnc.canon.COMMENT(("interpreter: feed mode set to units per minute"))

		cnc._setup.feed_mode = inc.UNITS_PER_MINUTE
	} else {
		return inc.NCE_BUG_CODE_NOT_G93_OR_G94
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_feed_rate

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The machine feed_rate is set to the value of f_number in the
   block by function call.
   The machine model feed_rate is set to that value.

   Called by: execute_block

   This is called only if the feed mode is UNITS_PER_MINUTE.

*/

func (cnc *rs274ngc_t) convert_feed_rate() inc.STATUS { /* pointer to machine settings              */

	cnc.canon.SET_FEED_RATE(cnc._setup.block1.f_number)
	cnc._setup.feed_rate = cnc._setup.block1.f_number
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_speed

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The machine spindle speed is set to the value of s_number in the
   block by a call to SET_SPINDLE_SPEED.
   The machine model for spindle speed is set to that value.

   Called by: execute_block.

*/

func (cnc *rs274ngc_t) convert_speed() inc.STATUS { /* pointer to machine settings              */

	cnc.canon.SET_SPINDLE_SPEED(cnc._setup.block1.s_number)
	cnc._setup.speed = cnc._setup.block1.s_number
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_m

   Returned Value: int
   If convert_tool_change returns an error code, this returns that code.
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   m_codes in the block are executed. For each m_code
   this consists of making a function call(s) to a canonical machining
   function(s) and setting the machine model.

   Called by: execute_block.

   This handles four separate types of activity in order:
   1. changing the tool (m6) - which also retracts and stops the spindle.
   2. Turning the spindle on or off (m3, m4, and m5)
   3. Turning coolant on and off (m7, m8, and m9)
   4. turning a-axis clamping on and off (m26, m27) - commented out.
   5. enabling or disabling feed and speed overrides (m49, m49).
   Within each group, only the first code encountered will be executed.

   This does nothing with m0, m1, m2, m30, or m60 (which are handled in
   convert_stop).

*/
func (cnc *rs274ngc_t) convert_m() inc.STATUS {
	//	static char name[] = "convert_m";

	if cnc._setup.block1.m_modes[6] != -1 {
		if e := cnc.convert_tool_change(); e != inc.RS274NGC_OK {
			return e
		}
	}

	if cnc._setup.block1.m_modes[7] == 3 {
		cnc.canon.START_SPINDLE_CLOCKWISE()
		cnc._setup.spindle_turning = inc.CANON_CLOCKWISE
	} else if cnc._setup.block1.m_modes[7] == 4 {
		cnc.canon.START_SPINDLE_COUNTERCLOCKWISE()
		cnc._setup.spindle_turning = inc.CANON_COUNTERCLOCKWISE
	} else if cnc._setup.block1.m_modes[7] == 5 {
		cnc.canon.STOP_SPINDLE_TURNING()
		cnc._setup.spindle_turning = inc.CANON_STOPPED
	}

	if cnc._setup.block1.m_modes[8] == 7 {
		cnc.canon.MIST_ON()
		cnc._setup.coolant.mist = ON
	} else if cnc._setup.block1.m_modes[8] == 8 {
		cnc.canon.FLOOD_ON()
		cnc._setup.coolant.flood = ON
	} else if cnc._setup.block1.m_modes[8] == 9 {
		cnc.canon.MIST_OFF()
		cnc._setup.coolant.mist = OFF
		cnc.canon.FLOOD_OFF()
		cnc._setup.coolant.flood = OFF
	}

	/* No axis clamps in this version
	     if (cnc._setup.block1.m_modes[2] == 26)
	       {
	   #ifdef DEBUG_EMC
	   COMMENT("interpreter: automatic A-axis clamping turned on");
	   #endif
	   cnc._setup.a_axis_clamping = ON;
	   }
	   else if (cnc._setup.block1.m_modes[2] == 27)
	   {
	   #ifdef DEBUG_EMC
	   COMMENT("interpreter: automatic A-axis clamping turned off");
	   #endif
	   cnc._setup.a_axis_clamping = OFF;
	   }
	*/

	if cnc._setup.block1.m_modes[9] == 48 {
		cnc.canon.ENABLE_FEED_OVERRIDE()
		cnc.canon.ENABLE_SPEED_OVERRIDE()
		cnc._setup.feed_override = ON
		cnc._setup.speed_override = ON
	} else if cnc._setup.block1.m_modes[9] == 49 {
		cnc.canon.DISABLE_FEED_OVERRIDE()
		cnc.canon.DISABLE_SPEED_OVERRIDE()
		cnc._setup.feed_override = OFF
		cnc._setup.speed_override = OFF
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_tool_change

   Returned Value: int (RS274NGC_OK)

   Side effects:
   This makes function calls to canonical machining functions, and sets
   the machine model as described below.

   Called by: convert_m

   This function carries out an m6 command, which changes the tool in the
   spindle. The only function call this makes is to the CHANGE_TOOL
   function. The semantics of this function call is that when it is
   completely carried out, the tool that was selected is in the spindle,
   the tool that was in the spindle (if any) is returned to its changer
   slot, the spindle will be stopped (but the spindle speed setting will
   not have changed) and the x, y, z, a, b, and c positions will be the same
   as they were before (although they may have moved around during the
   change).

   It would be nice to add more flexibility to this function by allowing
   more changes to occur (position changes, for example) as a result of
   the tool change. There are at least two ways of doing this:

   1. Require that certain machine settings always have a given fixed
   value after a tool change (which may be different from what the value
   was before the change), and record the fixed values somewhere (in the
   world model that is read at initialization, perhaps) so that this
   function can retrieve them and reset any settings that have changed.
   Fixed values could even be hard coded in this function.

   2. Allow the executor of the CHANGE_TOOL function to change the state
   of the world however it pleases, and have the interpreter read the
   executor's world model after the CHANGE_TOOL function is carried out.
   Implementing this would require a change in other parts of the EMC
   system, since calls to the interpreter would then have to be
   interleaved with execution of the function calls output by the
   interpreter.

   There may be other commands in the block that includes the tool change.
   They will be executed in the order described in execute_block.

   This implements the "Next tool in T word" approach to tool selection.
   The tool is selected when the T word is read (and the carousel may
   move at that time) but is changed when M6 is read.

   Note that if a different tool is put into the spindle, the current_z
   location setting may be incorrect for a time. It is assumed the
   program will contain an appropriate USE_TOOL_LENGTH_OFFSET command
   near the CHANGE_TOOL command, so that the incorrect setting is only
   temporary.

   In [NCMS, page 73, 74] there are three other legal approaches in addition
   to this one.

*/

func (cnc *rs274ngc_t) convert_tool_change() inc.STATUS {
	//static char name[] SET_TO "convert_tool_change";

	cnc.canon.CHANGE_TOOL(cnc._setup.selected_tool_slot)
	cnc._setup.current_slot = cnc._setup.selected_tool_slot
	cnc._setup.spindle_turning = inc.CANON_STOPPED

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_tool_select

   Returned Value: int
   If the tool slot given in the block is larger than allowed,
   this returns NCE_SELECTED_TOOL_SLOT_NUMBER_TOO_LARGE.
   Otherwise, it returns RS274NGC_OK.

   Side effects: See below

   Called by: execute_block

   A select tool command is given, which causes the changer chain to move
   so that the slot with the t_number given in the block is next to the
   tool changer, ready for a tool change.  The
   cnc._setup.selected_tool_slot is set to the given slot.

   An alternative in this function is to select by tool id. This was used
   in the K&T and VGER interpreters. It is easy to code.

   A check that the t_number is not negative has already been made in read_t.
   A zero t_number is allowed and means no tool should be selected.

*/
func (cnc *rs274ngc_t) convert_tool_select() inc.STATUS {

	//static char name[] = "convert_tool_select";

	if cnc._setup.block1.t_number > int(cnc._setup.tool_max) {
		return inc.NCE_SELECTED_TOOL_SLOT_NUMBER_TOO_LARGE
	}

	cnc.canon.SELECT_TOOL(cnc._setup.block1.t_number)
	cnc._setup.selected_tool_slot = cnc._setup.block1.t_number
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_gconvert_g

   Returned Value: int
   If one of the following functions is called and returns an error code,
   this returns that code.
   convert_control_mode
   convert_coordinate_system
   convert_cutter_compensation
   convert_distance_mode
   convert_dwell
   convert_length_units
   convert_modal_0
   convert_motion
   convert_retract_mode
   convert_set_plane
   convert_tool_length_offset
   Otherwise, it returns RS274NGC_OK.

   Side effects:
   Any g_codes in the block (excluding g93 and 94) and any implicit
   motion g_code are executed.

   Called by: execute_block.

   This takes a pointer to a block of RS274/NGC instructions (already
   read in) and creates the appropriate output commands corresponding to
   any "g" codes in the block.

   Codes g93 and g94, which set the feed mode, are executed earlier by
   execute_block before reading the feed rate.

   G codes are are executed in the following order.
   1.  mode 0, G4 only - dwell. Left here from earlier versions.
   2.  mode 2, one of (G17, G18, G19) - plane selection.
   3.  mode 6, one of (G20, G21) - length units.
   4.  mode 7, one of (G40, G41, G42) - cutter radius compensation.
   5.  mode 8, one of (G43, G49) - tool length offset
   6.  mode 12, one of (G54, G55, G56, G57, G58, G59, G59.1, G59.2, G59.3)
   - coordinate system selection.
   7.  mode 13, one of (G61, G61.1, G64) - control mode
   8.  mode 3, one of (G90, G91) - distance mode.
   9.  mode 10, one of (G98, G99) - retract mode.
   10. mode 0, one of (G10, G28, G30, G92, G92.1, G92.2, G92.3) -
   setting coordinate system locations, return to reference point 1,
   return to reference point 2, setting or cancelling axis offsets.
   11. mode 1, one of (G0, G1, G2, G3, G38.2, G80, G81 to G89) - motion or cancel.
   G53 from mode 0 is also handled here, if present.

   Some mode 0 and most mode 1 G codes must be executed after the length units
   are set, since they use coordinate values. Mode 1 codes also must wait
   until most of the other modes are set.

*/
func (cnc *rs274ngc_t) convert_g() inc.STATUS {

	//static char name[] = "convert_g";
	//int status;

	if cnc._setup.block1.g_modes[0] == inc.G_4 {
		cnc.convert_dwell(cnc._setup.block1.p_number)
	}
	if cnc._setup.block1.g_modes[2] != -1 {
		cnc.convert_set_plane(cnc._setup.block1.g_modes[2])
	}
	if cnc._setup.block1.g_modes[6] != -1 {
		cnc.convert_length_units(cnc._setup.block1.g_modes[6])
	}
	if cnc._setup.block1.g_modes[7] != -1 {
		cnc.convert_cutter_compensation(cnc._setup.block1.g_modes[7])
	}
	if cnc._setup.block1.g_modes[8] != -1 {
		cnc.convert_tool_length_offset(cnc._setup.block1.g_modes[8])
	}
	if cnc._setup.block1.g_modes[12] != -1 {
		cnc.convert_coordinate_system(cnc._setup.block1.g_modes[12])
	}
	if cnc._setup.block1.g_modes[13] != -1 {
		cnc.convert_control_mode(cnc._setup.block1.g_modes[13])
	}
	if cnc._setup.block1.g_modes[3] != -1 {
		cnc.convert_distance_mode(cnc._setup.block1.g_modes[3])
	}
	if cnc._setup.block1.g_modes[10] != -1 {
		cnc.convert_retract_mode(cnc._setup.block1.g_modes[10])
	}
	if cnc._setup.block1.g_modes[0] != -1 {
		cnc.convert_modal_0(cnc._setup.block1.g_modes[0])
	}
	if cnc._setup.block1.motion_to_be != -1 {
		cnc.convert_motion(cnc._setup.block1.motion_to_be)
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_dwell

   Returned Value: int (RS274NGC_OK)

   Side effects:
   A dwell command is executed.

   Called by: convert_g.

*/
func (cnc *rs274ngc_t) convert_dwell(time float64) inc.STATUS { /* time in seconds to dwell  */

	cnc.canon.DWELL(time)
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_set_plane

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. G_18 or G_19 is called when cutter radius compensation is on:
   NCE_CANNOT_USE_XZ_PLANE_WITH_CUTTER_RADIUS_COMP
   NCE_CANNOT_USE_YZ_PLANE_WITH_CUTTER_RADIUS_COMP
   2. The g_code is not G_17, G_18, or G_19:
   NCE_BUG_CODE_NOT_G17_G18_OR_G19

   Side effects:
   A canonical command setting the current plane is executed.

   Called by: convert_g.

*/

func (cnc *rs274ngc_t) convert_set_plane( /* ARGUMENTS                    */
	g_code inc.GCodes) inc.STATUS { /* pointer to machine settings  */

	//static char name[] = "convert_set_plane";
	if g_code == inc.G_17 {
		cnc.canon.SELECT_PLANE(inc.CANON_PLANE_XY)
		cnc._setup.plane = inc.CANON_PLANE_XY
	} else if g_code == inc.G_18 {
		if cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF {
			return inc.NCE_CANNOT_USE_XZ_PLANE_WITH_CUTTER_RADIUS_COMP
		}
		cnc.canon.SELECT_PLANE(inc.CANON_PLANE_XZ)
		cnc._setup.plane = inc.CANON_PLANE_XZ
	} else if g_code == inc.G_19 {
		if cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF {
			return inc.NCE_CANNOT_USE_YZ_PLANE_WITH_CUTTER_RADIUS_COMP
		}

		cnc.canon.SELECT_PLANE(inc.CANON_PLANE_YZ)
		cnc._setup.plane = inc.CANON_PLANE_YZ
	} else {
		return inc.NCE_BUG_CODE_NOT_G17_G18_OR_G19
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_length_units

   Returned Value: int
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. The g_code argument isnt G_20 or G_21:
   NCE_BUG_CODE_NOT_G20_OR_G21
   2. Cutter radius compensation is on:
   NCE_CANNOT_CHANGE_UNITS_WITH_CUTTER_RADIUS_COMP

   Side effects:
   A command setting the length units is executed. The machine
   settings are reset regarding length units and current position.

   Called by: convert_g.

   We are not changing tool length offset values or tool diameter values.
   Those values must be given in the table in the correct units. Thus it
   will generally not be feasible to switch units in the middle of a
   program.

   We are not changing the parameters that represent the positions
   of the nine work coordinate systems.

   We are also not changing feed rate values when length units are
   changed, so the actual behavior may change.

   Several other distance items in the settings (such as the various
   parameters for cycles) are also not reset.

   We are changing origin offset and axis offset values, which are
   critical. If this were not done, when length units are set and the new
   length units are not the same as the default length units
   (millimeters), and any XYZ origin or axis offset is not zero, then any
   subsequent change in XYZ origin or axis offset values will be
   incorrect.  Also, g53 (motion in absolute coordinates) will not work
   correctly.

*/
func (cnc *rs274ngc_t) convert_length_units( /* ARGUMENTS                    */
	g_code inc.GCodes) inc.STATUS { /* g_code being executed (must be G_20 or G_21) */

	//static char name[] = "convert_length_units";
	if cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF {
		return inc.NCE_CANNOT_CHANGE_UNITS_WITH_CUTTER_RADIUS_COMP
	}

	if g_code == inc.G_20 {
		cnc.canon.USE_LENGTH_UNITS(inc.CANON_UNITS_INCHES)
		if cnc._setup.length_units != inc.CANON_UNITS_INCHES {
			cnc._setup.length_units = inc.CANON_UNITS_INCHES
			cnc._setup.current.X = (cnc._setup.current.X * inc.INCH_PER_MM)
			cnc._setup.current.Y = (cnc._setup.current.Y * inc.INCH_PER_MM)
			cnc._setup.current.Z = (cnc._setup.current.Z * inc.INCH_PER_MM)
			cnc._setup.axis_offset.X =
				(cnc._setup.axis_offset.X * inc.INCH_PER_MM)
			cnc._setup.axis_offset.Y =
				(cnc._setup.axis_offset.Y * inc.INCH_PER_MM)
			cnc._setup.axis_offset.Z =
				(cnc._setup.axis_offset.Z * inc.INCH_PER_MM)
			cnc._setup.origin_offset.X =
				(cnc._setup.origin_offset.X * inc.INCH_PER_MM)
			cnc._setup.origin_offset.Y =
				(cnc._setup.origin_offset.Y * inc.INCH_PER_MM)
			cnc._setup.origin_offset.Z =
				(cnc._setup.origin_offset.Z * inc.INCH_PER_MM)
		}
	} else if g_code == inc.G_21 {
		cnc.canon.USE_LENGTH_UNITS(inc.CANON_UNITS_MM)
		if cnc._setup.length_units != inc.CANON_UNITS_MM {
			cnc._setup.length_units = inc.CANON_UNITS_MM
			cnc._setup.current.X = (cnc._setup.current.X * inc.MM_PER_INCH)
			cnc._setup.current.Y = (cnc._setup.current.Y * inc.MM_PER_INCH)
			cnc._setup.current.Z = (cnc._setup.current.Z * inc.MM_PER_INCH)
			cnc._setup.axis_offset.X =
				(cnc._setup.axis_offset.X * inc.MM_PER_INCH)
			cnc._setup.axis_offset.Y =
				(cnc._setup.axis_offset.Y * inc.MM_PER_INCH)
			cnc._setup.axis_offset.Z =
				(cnc._setup.axis_offset.Z * inc.MM_PER_INCH)
			cnc._setup.origin_offset.X =
				(cnc._setup.origin_offset.X * inc.MM_PER_INCH)
			cnc._setup.origin_offset.Y =
				(cnc._setup.origin_offset.Y * inc.MM_PER_INCH)
			cnc._setup.origin_offset.Z =
				(cnc._setup.origin_offset.Z * inc.MM_PER_INCH)
		}
	} else {
		return inc.NCE_BUG_CODE_NOT_G20_OR_G21
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cutter_compensation

   Returned Value: int
   If convert_cutter_compensation_on or convert_cutter_compensation_off
   is called and returns an error code, this returns that code.
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. g_code is not G_40, G_41, or G_42:
   NCE_BUG_CODE_NOT_G40_G41_OR_G42

   Side effects:
   The value of cutter_comp_side in the machine model mode is
   set to RIGHT, LEFT, or OFF. The currently active tool table index in
   the machine model (which is the index of the slot whose diameter
   value is used in cutter radius compensation) is updated.

   Since cutter radius compensation is performed in the interpreter, no
   call is made to any canonical function regarding cutter radius compensation.

   Called by: convert_g

*/
func (cnc *rs274ngc_t) convert_cutter_compensation( /* ARGUMENTS                    */
	g_code inc.GCodes) inc.STATUS { /* must be G_40, G_41, or G_42              */

	//static char name[] = "convert_cutter_compensation";

	if g_code == inc.G_40 {
		cnc.convert_cutter_compensation_off()
	} else if g_code == inc.G_41 {
		cnc.convert_cutter_compensation_on(inc.CANON_SIDE_LEFT)
	} else if g_code == inc.G_42 {
		cnc.convert_cutter_compensation_on(inc.CANON_SIDE_RIGHT)
	} else {
		return inc.NCE_BUG_CODE_NOT_G40_G41_OR_G42
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cutter_compensation_off

   Returned Value: int (RS274NGC_OK)

   Side effects:
   A comment is made that cutter radius compensation is turned off.
   The machine model of the cutter radius compensation mode is set to OFF.
   The value of program_x in the machine model is set to UNKNOWN.
   This serves as a flag when cutter radius compensation is
   turned on again.

   Called by: convert_cutter_compensation

*/
func (cnc *rs274ngc_t) convert_cutter_compensation_off() inc.STATUS {

	cnc.canon.COMMENT(("interpreter: cutter radius compensation off"))
	cnc._setup.cutter_comp_side = inc.CANON_SIDE_OFF
	cnc._setup.program_x = inc.UNKNOWN
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_cutter_compensation_on

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The selected plane is not the XY plane:
   NCE_CANNOT_TURN_CUTTER_RADIUS_COMP_ON_OUT_OF_XY_PLANE
   2. Cutter radius compensation is already on:
   NCE_CANNOT_TURN_CUTTER_RADIUS_COMP_ON_WHEN_ON

   Side effects:
   A COMMENT function call is made (conditionally) saying that the
   interpreter is switching mode so that cutter radius compensation is on.
   The value of cutter_comp_radius in the machine model mode is
   set to the absolute value of the radius given in the tool table.
   The value of cutter_comp_side in the machine model mode is
   set to RIGHT or LEFT. The currently active tool table index in
   the machine model is updated.

   Called by: convert_cutter_compensation

   check_other_codes checks that a d word occurs only in a block with g41
   or g42.

   Cutter radius compensation is carried out in the interpreter, so no
   call is made to a canonical function (although there is a canonical
   function, START_CUTTER_RADIUS_COMPENSATION, that could be called if
   the primitive level could execute it).

   This version uses a D word if there is one in the block, but it does
   not require a D word, since the sample programs which the interpreter
   is supposed to handle do not have them.  Logically, the D word is
   optional, since the D word is always (except in cases we have never
   heard of) the slot number of the tool in the spindle. Not requiring a
   D word is contrary to [Fanuc, page 116] and [NCMS, page 79], however.
   Both manuals require the use of the D-word with G41 and G42.

   This version handles a negative offset radius, which may be
   encountered if the programmed tool path is a center line path for
   cutting a profile and the path was constructed using a nominal tool
   diameter. Then the value in the tool table for the diameter is set to
   be the difference between the actual diameter and the nominal
   diameter. If the actual diameter is less than the nominal, the value
   in the table is negative. The method of handling a negative radius is
   to switch the side of the offset and use a positive radius. This
   requires that the profile use arcs (not straight lines) to go around
   convex corners.

*/

func (cnc *rs274ngc_t) convert_cutter_compensation_on( /* ARGUMENTS               */
	side inc.CANON_SIDE) inc.STATUS { /* side of path cutter is on (LEFT or RIGHT) */

	//static char name[] = "convert_cutter_compensation_on";

	if cnc._setup.plane != inc.CANON_PLANE_XY {
		return inc.NCE_CANNOT_TURN_CUTTER_RADIUS_COMP_ON_OUT_OF_XY_PLANE
	}

	if cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF {
		return inc.NCE_CANNOT_TURN_CUTTER_RADIUS_COMP_ON_WHEN_ON
	}

	index := inc.If(cnc._setup.block1.d_number != -1, cnc._setup.block1.d_number, cnc._setup.current_slot).(int)
	radius := ((cnc._setup.tool_table[index].Diameter) / 2.0)

	if radius < 0.0 { /* switch side & make radius positive if radius negative */
		radius = -radius
		if side == inc.CANON_SIDE_RIGHT {
			side = inc.CANON_SIDE_LEFT
		} else {
			side = inc.CANON_SIDE_RIGHT
		}
	}

	if side == inc.CANON_SIDE_RIGHT {
		cnc.canon.COMMENT(("interpreter: cutter radius compensation on right"))

	} else {
		cnc.canon.COMMENT(("interpreter: cutter radius compensation on left"))

	}

	cnc._setup.cutter_comp_radius = radius
	cnc._setup.tool_table_index = index
	cnc._setup.cutter_comp_side = side
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_tool_length_offset

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The block has no offset index (h number): NCE_OFFSET_INDEX_MISSING
   2. The g_code argument is not G_43 or G_49:
   NCE_BUG_CODE_NOT_G43_OR_G49

   Side effects:
   A USE_TOOL_LENGTH_OFFSET function call is made. Current_z,
   tool_length_offset, and length_offset_index are reset.

   Called by: convert_g

   This is called to execute g43 or g49.

   The g49 RS274/NGC command translates into a USE_TOOL_LENGTH_OFFSET(0.0)
   function call.

   The g43 RS274/NGC command translates into a USE_TOOL_LENGTH_OFFSET(length)
   function call, where length is the value of the entry in the tool length
   offset table whose index is the H number in the block.

   The H number in the block (if present) was checked for being a non-negative
   integer when it was read, so that check does not need to be repeated.

*/

func (cnc *rs274ngc_t) convert_tool_length_offset( /* ARGUMENTS                    */
	g_code inc.GCodes) inc.STATUS { /* g_code being executed (must be G_43 or G_49) */

	//static char name[] = "convert_tool_length_offset";
	var offset float64

	if g_code == inc.G_49 {
		cnc.canon.USE_TOOL_LENGTH_OFFSET(0.0)
		cnc._setup.current.Z = (cnc._setup.current.Z +
			cnc._setup.tool_length_offset)
		cnc._setup.tool_length_offset = 0.0
		cnc._setup.length_offset_index = 0
	} else if g_code == inc.G_43 {
		index := cnc._setup.block1.h_number
		if index == -1 {
			return inc.NCE_OFFSET_INDEX_MISSING
		}
		offset = cnc._setup.tool_table[index].Length
		cnc.canon.USE_TOOL_LENGTH_OFFSET(offset)
		cnc._setup.current.Z =
			(cnc._setup.current.Z + cnc._setup.tool_length_offset - offset)
		cnc._setup.tool_length_offset = offset
		cnc._setup.length_offset_index = index
	} else {
		return inc.NCE_BUG_CODE_NOT_G43_OR_G49
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_coordinate_system

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The value of the g_code argument is not 540, 550, 560, 570, 580, 590
   591, 592, or 593:
   NCE_BUG_CODE_NOT_IN_RANGE_G54_TO_G593

   Side effects:
   If the coordinate system selected by the g_code is not already in
   use, the canonical program coordinate system axis offset values are
   reset and the coordinate values of the current point are reset.

   Called by: convert_g.

   COORDINATE SYSTEMS (involves g10, g53, g54 - g59.3, g92)

   The canonical machining functions view of coordinate systems is:
   1. There are two coordinate systems: absolute and program.
   2. All coordinate values are given in terms of the program coordinate system.
   3. The offsets of the program coordinate system may be reset.

   The RS274/NGC view of coordinate systems, as given in section 3.2
   of [NCMS] is:
   1. there are ten coordinate systems: absolute and 9 program. The
   program coordinate systems are numbered 1 to 9.
   2. you can switch among the 9 but not to the absolute one. G54
   selects coordinate system 1, G55 selects 2, and so on through
   G56, G57, G58, G59, G59.1, G59.2, and G59.3.
   3. you can set the offsets of the 9 program coordinate systems
   using G10 L2 Pn (n is the number of the coordinate system) with
   values for the axes in terms of the absolute coordinate system.
   4. the first one of the 9 program coordinate systems is the default.
   5. data for coordinate systems is stored in parameters [NCMS, pages 59 - 60].
   6. g53 means to interpret coordinate values in terms of the absolute
   coordinate system for the one block in which g53 appears.
   7. You can offset the current coordinate system using g92. This offset
   will then apply to all nine program coordinate systems.

   The approach used in the interpreter mates the canonical and NGC views
   of coordinate systems as follows:

   During initialization, data from the parameters for the first NGC
   coordinate system is used in a SET_ORIGIN_OFFSETS function call and
   origin_index in the machine model is set to 1.

   If a g_code in the range g54 - g59.3 is encountered in an NC program,
   the data from the appropriate NGC coordinate system is copied into the
   origin offsets used by the interpreter, a SET_ORIGIN_OFFSETS function
   call is made, and the current position is reset.

   If a g10 is encountered, the convert_setup function is called to reset
   the offsets of the program coordinate system indicated by the P number
   given in the same block.

   If a g53 is encountered, the axis values given in that block are used
   to calculate what the coordinates are of that point in the current
   coordinate system, and a STRAIGHT_TRAVERSE or STRAIGHT_FEED function
   call to that point using the calculated values is made. No offset
   values are changed.

   If a g92 is encountered, that is handled by the convert_axis_offsets
   function. A g92 results in an axis offset for each axis being calculated
   and stored in the machine model. The axis offsets are applied to all
   nine coordinate systems. Axis offsets are initialized to zero.

*/

func (cnc *rs274ngc_t) convert_coordinate_system( /* ARGUMENTS                    */
	g_code inc.GCodes) inc.STATUS { /* g_code called (must be one listed above)      */

	//static char name[] = "convert_coordinate_system";
	var origin int

	parameters := cnc._setup.parameters
	switch g_code {
	case 540:
		origin = 1
	case 550:
		origin = 2
	case 560:
		origin = 3
	case 570:
		origin = 4
	case 580:
		origin = 5
	case 590:
		origin = 6
	case 591:
		origin = 7
	case 592:
		origin = 8
	case 593:
		origin = 9
	default:
		return inc.NCE_BUG_CODE_NOT_IN_RANGE_G54_TO_G593
	}

	if origin == cnc._setup.origin_index { /* already using this origin */
		cnc.canon.COMMENT(("interpreter: continuing to use same coordinate system"))
		return inc.RS274NGC_OK
	}

	cnc._setup.origin_index = origin
	parameters[5220] = float64(origin)

	/* axis offsets could be included in the two set of calculations for
	   current.X, current.Y, etc., but do not need to be because the results
	   would be the same. They would be added in then subtracted out. */
	cnc._setup.current.X =
		(cnc._setup.current.X + cnc._setup.origin_offset.X)
	cnc._setup.current.Y =
		(cnc._setup.current.Y + cnc._setup.origin_offset.Y)
	cnc._setup.current.Z =
		(cnc._setup.current.Z + cnc._setup.origin_offset.Z)
	cnc._setup.current.A = /*AA*/
		(cnc._setup.current.A + cnc._setup.origin_offset.A)
	cnc._setup.current.B = /*BB*/
		(cnc._setup.current.B + cnc._setup.origin_offset.B)
	cnc._setup.current.C = /*CC*/
		(cnc._setup.current.C + cnc._setup.origin_offset.C)

	x := parameters[5201+(origin*20)]
	y := parameters[5202+(origin*20)]
	z := parameters[5203+(origin*20)]
	a := parameters[5204+(origin*20)] /*AA*/
	b := parameters[5205+(origin*20)] /*BB*/
	c := parameters[5206+(origin*20)] /*CC*/

	cnc._setup.origin_offset.X = x
	cnc._setup.origin_offset.Y = y
	cnc._setup.origin_offset.Z = z
	cnc._setup.origin_offset.A = a /*AA*/
	cnc._setup.origin_offset.B = b /*BB*/
	cnc._setup.origin_offset.C = c /*CC*/

	cnc._setup.current.X = (cnc._setup.current.X - x)
	cnc._setup.current.Y = (cnc._setup.current.Y - y)
	cnc._setup.current.Z = (cnc._setup.current.Z - z)
	cnc._setup.current.A = (cnc._setup.current.A - a)
	cnc._setup.current.B = (cnc._setup.current.B - b)
	cnc._setup.current.C = (cnc._setup.current.C - c)

	cnc.canon.SET_ORIGIN_OFFSETS(x+cnc._setup.axis_offset.X,
		y+cnc._setup.axis_offset.Y,
		z+cnc._setup.axis_offset.Z,
		a+cnc._setup.axis_offset.A,
		b+cnc._setup.axis_offset.B,
		c+cnc._setup.axis_offset.C)
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_control_mode

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. g_code isn't G_61, G_61_1 or G_64: NCE_BUG_CODE_NOT_G61_G61_1_OR_G64

   Side effects: See below

   Called by: convert_g.

   The interpreter switches the machine settings to indicate the
   control mode (CANON_EXACT_STOP, CANON_EXACT_PATH or CANON_CONTINUOUS).

   A call is made to SET_MOTION_CONTROL_MODE(CANON_XXX), where CANON_XXX is
   CANON_EXACT_PATH if g_code is G_61, CANON_EXACT_STOP if g_code is G_61_1,
   and CANON_CONTINUOUS if g_code is G_64.

   Setting the control mode to CANON_EXACT_STOP on G_61 would correspond
   more closely to the meaning of G_61 as given in [NCMS, page 40], but
   CANON_EXACT_PATH has the advantage that the tool does not stop if it
   does not have to, and no evident disadvantage compared to
   CANON_EXACT_STOP, so it is being used for G_61. G_61_1 is not defined
   in [NCMS], so it is available and is used here for setting the control
   mode to CANON_EXACT_STOP.

   It is OK to call SET_MOTION_CONTROL_MODE(CANON_XXX) when CANON_XXX is
   already in force.

*/
func (cnc *rs274ngc_t) convert_control_mode( /* ARGUMENTS                    */
	g_code inc.GCodes) inc.STATUS { /* g_code being executed (G_61, G61_1, || G_64) */

	//static char name[] = "convert_control_mode";
	if g_code == inc.G_61 {
		cnc.canon.SET_MOTION_CONTROL_MODE(inc.CANON_EXACT_PATH)
		cnc._setup.control_mode = inc.CANON_EXACT_PATH
	} else if g_code == inc.G_61_1 {
		cnc.canon.SET_MOTION_CONTROL_MODE(inc.CANON_EXACT_STOP)
		cnc._setup.control_mode = inc.CANON_EXACT_STOP
	} else if g_code == inc.G_64 {
		cnc.canon.SET_MOTION_CONTROL_MODE(inc.CANON_CONTINUOUS)
		cnc._setup.control_mode = inc.CANON_CONTINUOUS
	} else {
		return inc.NCE_BUG_CODE_NOT_G61_G61_1_OR_G64
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_distance_mode

   Returned Value: int
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. g_code isn't G_90 or G_91: NCE_BUG_CODE_NOT_G90_OR_G91

   Side effects:
   The interpreter switches the machine settings to indicate the current
   distance mode (absolute or incremental).

   The canonical machine to which commands are being sent does not have
   an incremental mode, so no command setting the distance mode is
   generated in this function. A comment function call explaining the
   change of mode is made (conditionally), however, if there is a change.

   Called by: convert_g.

*/

func (cnc *rs274ngc_t) convert_distance_mode( /* ARGUMENTS                    */
	g_code inc.GCodes) inc.STATUS { /* g_code being executed (must be G_90 or G_91) */

	//static char name[] = "convert_distance_mode";
	if g_code == inc.G_90 {
		if cnc._setup.distance_mode != inc.MODE_ABSOLUTE {
			cnc.canon.COMMENT(("interpreter: distance mode changed to absolute"))
			cnc._setup.distance_mode = inc.MODE_ABSOLUTE
		}
	} else if g_code == inc.G_91 {
		if cnc._setup.distance_mode != inc.MODE_INCREMENTAL {
			cnc.canon.COMMENT(("interpreter: distance mode changed to incremental"))
			cnc._setup.distance_mode = inc.MODE_INCREMENTAL
		}
	} else {
		return inc.NCE_BUG_CODE_NOT_G90_OR_G91
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_retract_mode

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. g_code isn't G_98 or G_99: NCE_BUG_CODE_NOT_G98_OR_G99

   Side effects:
   The interpreter switches the machine settings to indicate the current
   retract mode for canned cycles (OLD_Z or R_PLANE).

   Called by: convert_g.

   The canonical machine to which commands are being sent does not have a
   retract mode, so no command setting the retract mode is generated in
   this function.

*/
func (cnc *rs274ngc_t) convert_retract_mode( /* ARGUMENTS                    */
	g_code inc.GCodes) inc.STATUS { /* g_code being executed (must be G_98 or G_99) */

	//static char name[] = "convert_retract_mode";
	if g_code == inc.G_98 {
		cnc.canon.COMMENT(("interpreter: retract mode set to old_z"))
		cnc._setup.retract_mode = inc.OLD_Z
	} else if g_code == inc.G_99 {
		cnc.canon.COMMENT(("interpreter: retract mode set to r_plane"))
		cnc._setup.retract_mode = inc.R_PLANE
	} else {
		return inc.NCE_BUG_CODE_NOT_G98_OR_G99
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_modal_0

   Returned Value: int
   If one of the following functions is called and returns an error code,
   this returns that code.
   convert_axis_offsets
   convert_home
   convert_setup
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. code is not G_4, G_10, G_28, G_30, G_53, G92, G_92_1, G_92_2, or G_92_3:
   NCE_BUG_CODE_NOT_G4_G10_G28_G30_G53_OR_G92_SERIES

   Side effects: See below

   Called by: convert_g

   If the g_code is g10, g28, g30, g92, g92.1, g92.2, or g92.3 (all are in
   modal group 0), it is executed. The other two in modal group 0 (G4 and
   G53) are executed elsewhere.

*/

func (cnc *rs274ngc_t) convert_modal_0( /* ARGUMENTS                                    */
	code inc.GCodes) inc.STATUS { /* G code, must be from group 0                 */

	//static char name[] = "convert_modal_0";

	if code == inc.G_10 {
		cnc.convert_setup()
	} else if (code == inc.G_28) || (code == inc.G_30) {
		cnc.convert_home(code)
	} else if (code == inc.G_92) || (code == inc.G_92_1) ||
		(code == inc.G_92_2) || (code == inc.G_92_3) {
		cnc.convert_axis_offsets(code)
	} else if (code == inc.G_4) || (code == inc.G_53) { /* handled elsewhere */
	} else {
		return inc.NCE_BUG_CODE_NOT_G4_G10_G28_G30_G53_OR_G92_SERIES
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_motion

   Returned Value: int
   If one of the following functions is called and returns an error code,
   this returns that code.
   convert_arc
   convert_cycle
   convert_probe
   convert_straight
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. The motion code is not 0,1,2,3,38.2,80,81,82,83,84,85,86,87, 88, or 89:
   NCE_BUG_UNKNOWN_MOTION_CODE

   Side effects:
   A g_code from the group causing motion (mode 1) is executed.

   Called by: convert_g.

*/

func (cnc *rs274ngc_t) convert_motion( /* ARGUMENTS                                 */
	motion inc.GCodes) inc.STATUS { /* g_code for a line, arc, canned cycle      */

	//static char name[] = "convert_motion";

	if (motion == inc.G_0) || (motion == inc.G_1) {
		cnc.convert_straight(motion)
	} else if (motion == inc.G_3) || (motion == inc.G_2) {
		cnc.convert_arc(motion)
	} else if motion == inc.G_38_2 {
		cnc.convert_probe()
	} else if motion == inc.G_80 {
		cnc.canon.COMMENT(("interpreter: motion mode set to none"))
		cnc._setup.motion_mode = inc.G_80
	} else if (motion > inc.G_80) && (motion < inc.G_90) {
		cnc.convert_cycle(motion)
	} else {
		return inc.NCE_BUG_UNKNOWN_MOTION_CODE
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_setup

   Returned Value: int (RS274NGC_OK)

   Side effects:
   SET_PROGRAM_ORIGIN is called, and the coordinate
   values for the program origin are reset.
   If the program origin is currently in use, the values of the
   the coordinates of the current point are updated.

   Called by: convert_modal_0.

   This is called only if g10 is called. g10 L2 may be used to alter the
   location of coordinate systems as described in [NCMS, pages 9 - 10] and
   [Fanuc, page 65]. [Fanuc] has only six coordinate systems, while
   [NCMS] has nine (the first six of which are the same as the six [Fanuc]
   has). All nine are implemented here.

   Being in incremental distance mode has no effect on the action of G10
   in this implementation. The manual is not explicit about what is
   intended.

   See documentation of convert_coordinate_system for more information.

*/

func (cnc *rs274ngc_t) convert_setup() inc.STATUS {

	//static char name[] = "convert_setup";
	//double x;
	//double y;
	//double z;
	var x, y, z, a, b, c float64

	parameters := cnc._setup.parameters
	p_int := int(cnc._setup.block1.p_number + 0.0001)

	if cnc._setup.block1.x_flag == ON {
		x = cnc._setup.block1.x_number
		parameters[5201+(p_int*20)] = x
	} else {
		x = parameters[5201+(p_int*20)]
	}

	if cnc._setup.block1.y_flag == ON {
		y = cnc._setup.block1.y_number
		parameters[5202+(p_int*20)] = y
	} else {
		y = parameters[5202+(p_int*20)]

	}
	if cnc._setup.block1.z_flag == ON {
		z = cnc._setup.block1.z_number
		parameters[5203+(p_int*20)] = z
	} else {
		z = parameters[5203+(p_int*20)]
	}

	if cnc._setup.block1.a_flag == ON {
		a = cnc._setup.block1.a_number
		parameters[5204+(p_int*20)] = a
	} else {
		a = parameters[5204+(p_int*20)]
	}

	if cnc._setup.block1.b_flag == ON {
		b = cnc._setup.block1.b_number
		parameters[5205+(p_int*20)] = b
	} else {
		b = parameters[5205+(p_int*20)]

	}

	if cnc._setup.block1.c_flag == ON {
		c = cnc._setup.block1.c_number
		parameters[5206+(p_int*20)] = c
	} else {
		c = parameters[5206+(p_int*20)]

	}

	/* axis offsets could be included in the two sets of calculations for
	   current.X, current.Y, etc., but do not need to be because the results
	   would be the same. They would be added in then subtracted out. */
	if p_int == cnc._setup.origin_index { /* system is currently used */
		cnc._setup.current.X =
			(cnc._setup.current.X + cnc._setup.origin_offset.X)
		cnc._setup.current.Y =
			(cnc._setup.current.Y + cnc._setup.origin_offset.Y)
		cnc._setup.current.Z =
			(cnc._setup.current.Z + cnc._setup.origin_offset.Z)
		cnc._setup.current.A = /*AA*/
			(cnc._setup.current.A + cnc._setup.origin_offset.A)
		cnc._setup.current.B = /*BB*/
			(cnc._setup.current.B + cnc._setup.origin_offset.B)
		cnc._setup.current.C = /*CC*/
			(cnc._setup.current.C + cnc._setup.origin_offset.C)

		cnc._setup.origin_offset.X = x
		cnc._setup.origin_offset.Y = y
		cnc._setup.origin_offset.Z = z
		cnc._setup.origin_offset.A = a /*AA*/
		cnc._setup.origin_offset.B = b /*BB*/
		cnc._setup.origin_offset.C = c /*CC*/

		cnc._setup.current.X = (cnc._setup.current.X - x)
		cnc._setup.current.Y = (cnc._setup.current.Y - y)
		cnc._setup.current.Z = (cnc._setup.current.Z - z)
		cnc._setup.current.A = (cnc._setup.current.A - a)
		cnc._setup.current.B = (cnc._setup.current.B - b)
		cnc._setup.current.C = (cnc._setup.current.C - c)

		cnc.canon.SET_ORIGIN_OFFSETS(x+cnc._setup.axis_offset.X,
			y+cnc._setup.axis_offset.Y,
			z+cnc._setup.axis_offset.Z,
			a+cnc._setup.axis_offset.A,
			b+cnc._setup.axis_offset.B,
			c+cnc._setup.axis_offset.C)
	} else {
		cnc.canon.COMMENT(("interpreter: setting coordinate system origin"))

	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_home

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. cutter radius compensation is on:
   NCE_CANNOT_USE_G28_OR_G30_WITH_CUTTER_RADIUS_COMP
   2. The code is not G28 or G30: NCE_BUG_CODE_NOT_G28_OR_G30

   Side effects:
   This executes a straight traverse to the programmed point, using
   the current coordinate system, tool length offset, and motion mode
   to interpret the coordinate values. Then it executes a straight
   traverse to the location of reference point 1 (if G28) or reference
   point 2 (if G30). It also updates the setting of the position of the
   tool point to the end point of the move.

   Called by: convert_modal_0.

   During the motion from the intermediate point to the home point, this
   function currently makes the A and C axes turn counterclockwise if a
   turn is needed.  This is not necessarily the most efficient way to do
   it. A check might be made of which direction to turn to have the least
   turn to get to the reference position, and the axis would turn that
   way.

*/

func (cnc *rs274ngc_t) convert_home(move inc.GCodes) inc.STATUS { /* G code, must be G_28 or G_30             */

	//static char name[] = "convert_home";
	var (
		end_x, end_y, end_z                               float64
		AA_end, BB_end, CC_end, AA_end2, BB_end2, CC_end2 float64
	)

	parameters := cnc._setup.parameters
	cnc.find_ends(&end_x, &end_y, &end_z, &AA_end, &BB_end, &CC_end)
	if cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF {
		return inc.NCE_CANNOT_USE_G28_OR_G30_WITH_CUTTER_RADIUS_COMP
	}

	cnc.canon.STRAIGHT_TRAVERSE(end_x, end_y, end_z, AA_end, BB_end, CC_end)
	if move == inc.G_28 {
		cnc.find_relative(parameters[5161], parameters[5162], parameters[5163],
			parameters[5164], /*AA*/
			parameters[5165], /*BB*/
			parameters[5166], /*CC*/
			&end_x, &end_y, &end_z, &AA_end2, &BB_end2, &CC_end2)
	} else if move == inc.G_30 {
		cnc.find_relative(parameters[5181], parameters[5182], parameters[5183],
			parameters[5184], /*AA*/
			parameters[5185], /*BB*/
			parameters[5186], /*CC*/
			&end_x, &end_y, &end_z, &AA_end2, &BB_end2, &CC_end2)
	} else {
		return inc.NCE_BUG_CODE_NOT_G28_OR_G30
	}

	cnc.canon.STRAIGHT_TRAVERSE(end_x, end_y, end_z, AA_end, BB_end, CC_end)
	cnc._setup.current.X = end_x
	cnc._setup.current.Y = end_y
	cnc._setup.current.Z = end_z

	cnc._setup.current.A = AA_end2 /*AA*/
	cnc._setup.current.B = BB_end2 /*BB*/
	cnc._setup.current.C = CC_end2 /*CC*/

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_axis_offsets

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. The function is called when cutter radius compensation is on:
   NCE_CANNOT_CHANGE_AXIS_OFFSETS_WITH_CUTTER_RADIUS_COMP
   2. The g_code argument is not G_92, G_92_1, G_92_2, or G_92_3
   NCE_BUG_CODE_NOT_IN_G92_SERIES

   Side effects:
   SET_PROGRAM_ORIGIN is called, and the coordinate
   values for the axis offsets are reset. The coordinates of the
   current point are reset. Parameters may be set.

   Called by: convert_modal_0.

   The action of G92 is described in [NCMS, pages 10 - 11] and {Fanuc,
   pages 61 - 63]. [NCMS] is ambiguous about the intent, but [Fanuc]
   is clear. When G92 is executed, an offset of the origin is calculated
   so that the coordinates of the current point with respect to the moved
   origin are as specified on the line containing the G92. If an axis
   is not mentioned on the line, the coordinates of the current point
   are not changed. The execution of G92 results in an axis offset being
   calculated and saved for each of the six axes, and the axis offsets
   are always used when motion is specified with respect to absolute
   distance mode using any of the nine coordinate systems (those designated
   by G54 - G59.3). Thus all nine coordinate systems are affected by G92.

   Being in incremental distance mode has no effect on the action of G92
   in this implementation. [NCMS] is not explicit about this, but it is
   implicit in the second sentence of [Fanuc, page 61].

   The offset is the amount the origin must be moved so that the
   coordinate of the controlled point has the specified value. For
   example, if the current point is at X=4 in the currently specified
   coordinate system and the current X-axis offset is zero, then "G92 x7"
   causes the X-axis offset to be reset to -3.

   Since a non-zero offset may be already be in effect when the G92 is
   called, that must be taken into account.

   In addition to causing the axis offset values in the _setup model to be
   set, G92 sets parameters 5211 to 5216 to the x,y,z,a,b,c axis offsets.

   The action of G92.2 is described in [NCMS, page 12]. There is no
   equivalent command in [Fanuc]. G92.2 resets axis offsets to zero.
   G92.1, also included in [NCMS, page 12] (but the usage here differs
   slightly from the spec), is like G92.2, except that it also causes
   the axis offset parameters to be set to zero, whereas G92.2 does not
   zero out the parameters.

   G92.3 is not in [NCMS]. It sets the axis offset values to the values
   given in the parameters.

*/
func (cnc *rs274ngc_t) convert_axis_offsets( /* ARGUMENTS                               */
	g_code inc.GCodes) inc.STATUS { /* g_code being executed (must be in G_92 series) */

	//static char name[] = "convert_axis_offsets";
	//double * pars;                            /* short name for cnc._setup.parameters            */

	if cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF { /* not "== ON" */
		return inc.NCE_CANNOT_CHANGE_AXIS_OFFSETS_WITH_CUTTER_RADIUS_COMP
	}
	pars := cnc._setup.parameters
	if g_code == inc.G_92 {
		if cnc._setup.block1.x_flag == ON {
			cnc._setup.axis_offset.X =
				(cnc._setup.current.X + cnc._setup.axis_offset.X - cnc._setup.block1.x_number)
			cnc._setup.current.X = cnc._setup.block1.x_number
		}

		if cnc._setup.block1.y_flag == ON {
			cnc._setup.axis_offset.Y =
				(cnc._setup.current.Y + cnc._setup.axis_offset.Y - cnc._setup.block1.y_number)
			cnc._setup.current.Y = cnc._setup.block1.y_number
		}

		if cnc._setup.block1.z_flag == ON {
			cnc._setup.axis_offset.Z =
				(cnc._setup.current.Z + cnc._setup.axis_offset.Z - cnc._setup.block1.z_number)
			cnc._setup.current.Z = cnc._setup.block1.z_number
		}

		if cnc._setup.block1.a_flag == ON { /*AA*/
			cnc._setup.axis_offset.A = (cnc._setup.current.A +
				cnc._setup.axis_offset.A - cnc._setup.block1.a_number)
			cnc._setup.current.A = cnc._setup.block1.a_number
		}

		if cnc._setup.block1.b_flag == ON { /*BB*/
			cnc._setup.axis_offset.B = (cnc._setup.current.B +
				cnc._setup.axis_offset.B - cnc._setup.block1.b_number)
			cnc._setup.current.B = cnc._setup.block1.b_number
		}

		if cnc._setup.block1.c_flag == ON { /*CC*/
			cnc._setup.axis_offset.C = (cnc._setup.current.C +
				cnc._setup.axis_offset.C - cnc._setup.block1.c_number)
			cnc._setup.current.C = cnc._setup.block1.c_number
		}

		cnc.canon.SET_ORIGIN_OFFSETS(cnc._setup.origin_offset.X+cnc._setup.axis_offset.X,
			cnc._setup.origin_offset.Y+cnc._setup.axis_offset.Y,
			cnc._setup.origin_offset.Z+cnc._setup.axis_offset.Z,
			(cnc._setup.origin_offset.A + cnc._setup.axis_offset.A),
			(cnc._setup.origin_offset.B + cnc._setup.axis_offset.B),
			(cnc._setup.origin_offset.C + cnc._setup.axis_offset.C))
		pars[5211] = cnc._setup.axis_offset.X
		pars[5212] = cnc._setup.axis_offset.Y
		pars[5213] = cnc._setup.axis_offset.Z

		pars[5214] = cnc._setup.axis_offset.A

		pars[5215] = cnc._setup.axis_offset.B

		pars[5216] = cnc._setup.axis_offset.C

	} else if (g_code == inc.G_92_1) || (g_code == inc.G_92_2) {
		cnc._setup.current.X =
			cnc._setup.current.X + cnc._setup.axis_offset.X
		cnc._setup.current.Y =
			cnc._setup.current.Y + cnc._setup.axis_offset.Y
		cnc._setup.current.Z =
			cnc._setup.current.Z + cnc._setup.axis_offset.Z

		cnc._setup.current.A = /*AA*/

			(cnc._setup.current.A + cnc._setup.axis_offset.A)

		cnc._setup.current.B = /*BB*/

			(cnc._setup.current.B + cnc._setup.axis_offset.B)

		cnc._setup.current.C = /*CC*/

			(cnc._setup.current.C + cnc._setup.axis_offset.C)

		cnc.canon.SET_ORIGIN_OFFSETS(cnc._setup.origin_offset.X,
			cnc._setup.origin_offset.Y,
			cnc._setup.origin_offset.Z,
			cnc._setup.origin_offset.A,
			cnc._setup.origin_offset.B,
			cnc._setup.origin_offset.C)
		cnc._setup.axis_offset.X = 0.0
		cnc._setup.axis_offset.Y = 0.0
		cnc._setup.axis_offset.Z = 0.0

		cnc._setup.axis_offset.A = 0.0 /*AA*/

		cnc._setup.axis_offset.B = 0.0 /*BB*/

		cnc._setup.axis_offset.C = 0.0 /*CC*/

		if g_code == inc.G_92_1 {
			pars[5211] = 0.0
			pars[5212] = 0.0
			pars[5213] = 0.0

			pars[5214] = 0.0 /*AA*/

			pars[5215] = 0.0 /*BB*/

			pars[5216] = 0.0 /*CC*/

		}
	} else if g_code == inc.G_92_3 {
		cnc._setup.current.X =
			cnc._setup.current.X + cnc._setup.axis_offset.X - pars[5211]
		cnc._setup.current.Y =
			cnc._setup.current.Y + cnc._setup.axis_offset.Y - pars[5212]
		cnc._setup.current.Z =
			cnc._setup.current.Z + cnc._setup.axis_offset.Z - pars[5213]

		cnc._setup.current.A = /*AA*/
			cnc._setup.current.A + cnc._setup.axis_offset.A - pars[5214]

		cnc._setup.current.B = /*BB*/
			cnc._setup.current.B + cnc._setup.axis_offset.B - pars[5215]

		cnc._setup.current.C = /*CC*/
			cnc._setup.current.C + cnc._setup.axis_offset.C - pars[5216]

		cnc._setup.axis_offset.X = pars[5211]
		cnc._setup.axis_offset.Y = pars[5212]
		cnc._setup.axis_offset.Z = pars[5213]

		cnc._setup.axis_offset.A = pars[5214]

		cnc._setup.axis_offset.B = pars[5215]

		cnc._setup.axis_offset.C = pars[5216]

		cnc.canon.SET_ORIGIN_OFFSETS(cnc._setup.origin_offset.X+cnc._setup.axis_offset.X,
			cnc._setup.origin_offset.Y+cnc._setup.axis_offset.Y,
			cnc._setup.origin_offset.Z+cnc._setup.axis_offset.Z,
			(cnc._setup.origin_offset.A + cnc._setup.axis_offset.A),
			(cnc._setup.origin_offset.B + cnc._setup.axis_offset.B),
			(cnc._setup.origin_offset.C + cnc._setup.axis_offset.C))
	} else {
		return inc.NCE_BUG_CODE_NOT_IN_G92_SERIES
	}

	return inc.RS274NGC_OK
}

/****************************************************************************/

/* convert_probe

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. No value is given in the block for any of X, Y, or Z:
   NCE_X_Y_AND_Z_WORDS_ALL_MISSING_WITH_G38_2
   2. feed mode is inverse time: NCE_CANNOT_PROBE_IN_INVERSE_TIME_FEED_MODE
   3. cutter radius comp is on: NCE_CANNOT_PROBE_WITH_CUTTER_RADIUS_COMP_ON
   4. Feed rate is zero: NCE_CANNOT_PROBE_WITH_ZERO_FEED_RATE
   5. Rotary axis motion is programmed:
   NCE_CANNOT_MOVE_ROTARY_AXES_DURING_PROBING
   6. The starting point for the probe move is within 0.01 inch or 0.254
   millimeters of the point to be probed:
   NCE_START_POINT_TOO_CLOSE_TO_PROBE_POINT

   Side effects:
   This executes a straight_probe command.
   The probe_flag in the settings is set to ON.
   The motion mode in the settings is set to G_38_2.

   Called by: convert_motion.

   The approach to operating in incremental distance mode (g91) is to
   put the the absolute position values into the block before using the
   block to generate a move.

   After probing is performed, the location of the probe cannot be
   predicted. This differs from every other command, all of which have
   predictable results. The next call to the interpreter (with either
   rs274ngc_read or rs274ngc_execute) will result in updating the
   current position by calls to get_external_position_x, etc.

*/

func (cnc *rs274ngc_t) convert_probe() inc.STATUS {

	//static char name[] = "convert_probe";
	var (
		end_x, end_y, end_z, AA_end, BB_end, CC_end float64
	)

	if ((cnc._setup.block1.x_flag == OFF) && (cnc._setup.block1.y_flag == OFF)) &&
		(cnc._setup.block1.z_flag == OFF) {
		return inc.NCE_X_Y_AND_Z_WORDS_ALL_MISSING_WITH_G38_2
	}
	if cnc._setup.feed_mode == inc.INVERSE_TIME {
		return inc.NCE_CANNOT_PROBE_IN_INVERSE_TIME_FEED_MODE
	}

	if cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF {
		return inc.NCE_CANNOT_PROBE_WITH_CUTTER_RADIUS_COMP_ON
	}
	if cnc._setup.feed_rate == 0.0 {
		return inc.NCE_CANNOT_PROBE_WITH_ZERO_FEED_RATE
	}

	cnc.find_ends(&end_x, &end_y,
		&end_z, &AA_end, &BB_end, &CC_end)
	if (AA_end != cnc._setup.current.A) /*AA*/ || (BB_end != cnc._setup.current.B) /*BB*/ || (CC_end != cnc._setup.current.C) /*CC*/ {
		return inc.NCE_CANNOT_MOVE_ROTARY_AXES_DURING_PROBING
	}

	distance := math.Sqrt(math.Pow((cnc._setup.current.X-end_x), 2) +
		math.Pow((cnc._setup.current.Y-end_y), 2) +
		math.Pow((cnc._setup.current.Z-end_z), 2))

	if distance < inc.If((cnc._setup.length_units == inc.CANON_UNITS_MM), 0.254, 0.01).(float64) {
		return inc.NCE_START_POINT_TOO_CLOSE_TO_PROBE_POINT
	}

	cnc.canon.TURN_PROBE_ON()
	cnc.canon.STRAIGHT_PROBE(end_x, end_y, end_z, AA_end, BB_end, CC_end)
	cnc.canon.TURN_PROBE_OFF()
	cnc._setup.motion_mode = inc.G_38_2
	cnc._setup.probe_flag = ON
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* find_ends

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The values of px, py, pz, aa_p, bb_p, and cc_p are set

   Called by:
   convert_arc
   convert_home
   convert_probe
   convert_straight

   This finds the coordinates of a point, "end", in the currently
   active coordinate system, and sets the values of the pointers to the
   coordinates (which are the arguments to the function).

   In all cases, if no value for the coodinate is given in the block, the
   current value for the coordinate is used. When cutter radius
   compensation is on, this function is called before compensation
   calculations are performed, so the current value of the programmed
   point is used, not the current value of the actual current_point.

   There are three cases for when the coordinate is included in the block:

   1. G_53 is active. This means to interpret the coordinates as machine
   coordinates. That is accomplished by adding the two offsets to the
   coordinate given in the block.

   2. Absolute coordinate mode is in effect. The coordinate in the block
   is used.

   3. Incremental coordinate mode is in effect. The coordinate in the
   block plus either (i) the programmed current position - when cutter
   radius compensation is in progress, or (2) the actual current position.

*/

func (cnc *rs274ngc_t) find_ends( /* ARGUMENTS                                    */
	px, /* pointer to end_x                             */
	py, /* pointer to end_y                             */
	pz, /* pointer to end_z                             */
	AA_p, /* pointer to end_a                       */ /*AA*/
	BB_p, /* pointer to end_b                       */ /*BB*/
	CC_p *float64) inc.STATUS { /* pointer to end_c                       */ /*CC*/

	mode := cnc._setup.distance_mode
	middle := (cnc._setup.program_x != inc.UNKNOWN)
	comp := (cnc._setup.cutter_comp_side != inc.CANON_SIDE_OFF)

	if cnc._setup.block1.g_modes[0] == inc.G_53 { /* distance mode is absolute in this case */
		cnc.canon.COMMENT(("interpreter: offsets temporarily suspended"))

		*px = inc.If(cnc._setup.block1.x_flag == ON, (cnc._setup.block1.x_number -
			(cnc._setup.origin_offset.X + cnc._setup.axis_offset.X)), cnc._setup.current.X).(float64)

		*py = inc.If(cnc._setup.block1.y_flag == ON, (cnc._setup.block1.y_number -
			(cnc._setup.origin_offset.Y + cnc._setup.axis_offset.Y)), cnc._setup.current.Y).(float64)
		*pz = inc.If(cnc._setup.block1.z_flag == ON, (cnc._setup.block1.z_number -
			(cnc._setup.tool_length_offset + cnc._setup.origin_offset.Z + cnc._setup.axis_offset.Z)), cnc._setup.current.Z).(float64)

		*AA_p = inc.If(cnc._setup.block1.a_flag == ON, (cnc._setup.block1.a_number -

			(cnc._setup.origin_offset.A + cnc._setup.axis_offset.A)), cnc._setup.current.A).(float64)

		*BB_p = inc.If(cnc._setup.block1.b_flag == ON, (cnc._setup.block1.b_number -

			(cnc._setup.origin_offset.B + cnc._setup.axis_offset.B)), cnc._setup.current.B).(float64)

		*CC_p = inc.If(cnc._setup.block1.c_flag == ON, (cnc._setup.block1.c_number -
			(cnc._setup.tool_length_offset + cnc._setup.origin_offset.C + cnc._setup.axis_offset.C)), cnc._setup.current.C).(float64)

	} else if mode == inc.MODE_ABSOLUTE {
		*px = inc.If(cnc._setup.block1.x_flag == ON, cnc._setup.block1.x_number,
			inc.If(comp && middle, cnc._setup.program_x, cnc._setup.current.X).(float64)).(float64)

		*py = inc.If(cnc._setup.block1.y_flag == ON, cnc._setup.block1.y_number,
			inc.If(comp && middle, cnc._setup.program_y, cnc._setup.current.Y).(float64)).(float64)

		*pz = inc.If(cnc._setup.block1.z_flag == ON, cnc._setup.block1.z_number, cnc._setup.current.Z).(float64)

		*AA_p = inc.If(cnc._setup.block1.a_flag == ON, cnc._setup.block1.a_number, cnc._setup.current.A).(float64) /*AA*/

		*BB_p = inc.If(cnc._setup.block1.b_flag == ON, cnc._setup.block1.b_number, cnc._setup.current.B).(float64) /*BB*/

		*CC_p = inc.If(cnc._setup.block1.c_flag == ON, cnc._setup.block1.c_number, cnc._setup.current.C).(float64) /*CC*/

	} else { /* mode is MODE_INCREMENTAL */

		*px = inc.If(cnc._setup.block1.x_flag == ON,
			inc.If(comp && middle, (cnc._setup.block1.x_number+cnc._setup.program_x), (cnc._setup.block1.x_number+cnc._setup.current.X)).(float64),
			inc.If((comp && middle), cnc._setup.program_x, cnc._setup.current.X).(float64)).(float64)

		*py = inc.If(cnc._setup.block1.y_flag == ON,
			inc.If(comp && middle, (cnc._setup.block1.y_number+cnc._setup.program_y), (cnc._setup.block1.y_number+cnc._setup.current.Y)).(float64),
			inc.If((comp && middle), cnc._setup.program_y, cnc._setup.current.Y).(float64)).(float64)

		*pz = inc.If(cnc._setup.block1.z_flag == ON,
			(cnc._setup.current.Z + cnc._setup.block1.z_number), cnc._setup.current.Z).(float64)
		*AA_p = inc.If(cnc._setup.block1.a_flag == ON, /*AA*/
			(cnc._setup.current.A + cnc._setup.block1.a_number), cnc._setup.current.A).(float64)
		*BB_p = inc.If(cnc._setup.block1.b_flag == ON, /*BB*/
			(cnc._setup.current.B + cnc._setup.block1.b_number), cnc._setup.current.B).(float64)
		*CC_p = inc.If(cnc._setup.block1.c_flag == ON, /*CC*/
			(cnc._setup.current.C + cnc._setup.block1.c_number), cnc._setup.current.C).(float64)
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* find_relative

   Returned Value: int (RS274NGC_OK)

   Side effects:
   The values of x2, y2, z2, aa_2, bb_2, and cc_2 are set.
   (NOTE: aa_2 etc. are written with lower case letters in this
   documentation because upper case would confuse the pre-preprocessor.)

   Called by:
   convert_home

   This finds the coordinates in the current system, under the current
   tool length offset, of a point (x1, y1, z1, aa_1, bb_1, cc_1) whose absolute
   coordinates are known.

   Don't confuse this with the inverse operation.

*/

func (cnc *rs274ngc_t) find_relative( /* ARGUMENTS                   */
	x1, /* absolute x position         */
	y1, /* absolute y position         */
	z1, /* absolute z position         */
	AA_1, /* absolute a position         */ /*AA*/
	BB_1, /* absolute b position         */ /*BB*/
	CC_1 float64, /* absolute c position         */ /*CC*/
	x2, /* pointer to relative x       */
	y2, /* pointer to relative y       */
	z2, /* pointer to relative z       */
	AA_2, /* pointer to relative a       */ /*AA*/
	BB_2, /* pointer to relative b       */ /*BB*/
	CC_2 *float64) inc.STATUS { /* pointer to relative c       */ /*CC*/

	*x2 = (x1 - (cnc._setup.origin_offset.X + cnc._setup.axis_offset.X))
	*y2 = (y1 - (cnc._setup.origin_offset.Y + cnc._setup.axis_offset.Y))
	*z2 = (z1 - (cnc._setup.tool_length_offset +
		cnc._setup.origin_offset.Z + cnc._setup.axis_offset.Z))

	/*AA*/
	*AA_2 = (AA_1 - (cnc._setup.origin_offset.A +
		cnc._setup.axis_offset.A)) /*AA*/

	/*BB*/
	*BB_2 = (BB_1 - (cnc._setup.origin_offset.B +
		cnc._setup.axis_offset.B)) /*BB*/

	/*CC*/
	*CC_2 = (CC_1 - (cnc._setup.origin_offset.C +

		cnc._setup.axis_offset.C)) /*CC*/

	return inc.RS274NGC_OK
}

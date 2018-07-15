package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"

	. "github.com/flyingyizi/rs274ngc"
	"github.com/flyingyizi/rs274ngc/example/canon"
	"github.com/flyingyizi/rs274ngc/inc"
)

/************************************************************************/

/* main

   The executable exits with either 0 (under all conditions not listed
   below) or 1 (under the following conditions):
   1. A fatal error occurs while interpreting from a file.
   2. Read_tool_file fails.
   3. An error occurs in rs274ngc_init.

   ***********************************************************************

   Here are three ways in which the rs274abc executable may be called.
   Any other sort of call to the executable will cause an error message
   to be printed and the interpreter will not run. Other executables
   may be called similarly.

   1. If the rs274abc stand-alone executable is called with no arguments,
   input is taken from the keyboard, and an error in the input does not
   cause the rs274abc executable to exit.

   EXAMPLE:

   1A. To interpret from the keyboard, enter:

   rs274abc


   ***********************************************************************

   2. If the executable is called with one argument, the argument is
   taken to be the name of an NC file and the file is interpreted as
   described in the documentation of interpret_from_file.

   EXAMPLES:

   2A. To interpret the file "cds.abc" and read the results on the
   screen, enter:

   rs274abc cds.abc

   2B. To interpret the file "cds.abc" and print the results in the file
   "cds.prim", enter:

   rs274abc cds.abc > cds.prim

   ***********************************************************************

   Whichever way the executable is called, this gives the user several
   choices before interpretation starts

   1 = start interpreting
   2 = choose parameter file
   3 = read tool file ...
   4 = turn block delete switch ON
   5 = adjust error handling...

   Interpretation starts when option 1 is chosen. Until that happens, the
   user is repeatedly given the five choices listed above.  Item 4
   toggles between "turn block delete switch ON" and "turn block delete
   switch OFF".  See documentation of adjust_error_handling regarding
   what option 5 does.

   User instructions are printed to stderr (with fprintf) so that output
   can be redirected to a file. When output is redirected and user
   instructions are printed to stdout (with printf), the instructions get
   redirected and the user does not see them.

*/

func main() {

	flag.Parse()

	if len(os.Args) > 3 {
		fmt.Printf("Usage %s  \n", os.Args[0])
		fmt.Printf("   or %s <input file>\n", os.Args[0])
		fmt.Printf("   or %s <input file> <output file>  \n", os.Args[0])
		return
	}

	var (
		block_delete ON_OFF = OFF
		print_stack  ON_OFF = OFF
		choice       int
		tool_flag    = 0
		do_next      = 2 /* 2=stop */
		default_name = "rs274ngc.var"
		err          error
		//_outfile     = os.Stdout /* may be reset below */
		status int
	)

	_parameter_file_name := default_name
	for {
		fmt.Println("enter a number:")
		fmt.Println("1 = start interpreting")
		fmt.Println("2 = choose parameter file ...")
		fmt.Println("3 = read tool file ...")
		fmt.Printf("4 = turn block delete switch %s \n",
			inc.If((block_delete == OFF), "ON", "OFF").(string))

		fmt.Println("5 = adjust error handling...")
		fmt.Println("enter choice => ")

		if _, err := fmt.Scanf("%d", &choice); err != nil {
			continue
		}

		if choice == 1 {
			break
		} else if choice == 2 {
			_, err := os.Stat(_parameter_file_name)
			if os.IsNotExist(err) {
				return
			}

		} else if choice == 3 {
			if cnc.read_tool_file("") != 0 {
				return
			}
			tool_flag = 1
		} else if choice == 4 {
			block_delete = inc.If((block_delete == OFF), ON, OFF).(ON_OFF)
		} else if choice == 5 {
			//todo adjust_error_handling(argc, &print_stack, &do_next);
		}

	}

	fmt.Println("executing")
	if tool_flag == 0 {
		if s := cnc.read_tool_file("rs274ngc.tool_default"); s != 0 {
			return
		}
	}

	if len(os.Args) == 3 {
		/*_outfile*/ _, err = os.OpenFile(os.Args[2], os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("could not open output file %s\n", os.Args[2])
			return
		}
	}

	cnc.SetCanon(&canon.Canon_t{})
	if status := cnc.Init(); status != inc.RS274NGC_OK {
		return
	}

	if len(os.Args) == 1 {
		status = cnc.interpret_from_keyboard(block_delete, print_stack)
	} else { /* if (argc == 2 or argc == 3) */
		status = cnc.Open(os.Args[1])
		if status != inc.RS274NGC_OK { /* do not need to close since not open */
			cnc.report_error(status, print_stack)
			return
		}
		status = cnc.interpret_from_file(do_next, block_delete, print_stack)
		//todo rs274ngc_file_name(buffer, 5)  /* called to exercise the function */
		//todo rs274ngc_file_name(buffer, 79) /* called to exercise the function */
		cnc.Close()

	}
	//	rs274ngc_line_length()         /* called to exercise the function */
	//	rs274ngc_sequence_number()     /* called to exercise the function */
	//	rs274ngc_active_g_codes(gees)  /* called to exercise the function */
	//	rs274ngc_active_m_codes(ems)   /* called to exercise the function */
	//	rs274ngc_active_settings(sets) /* called to exercise the function */
	//	rs274ngc_exit()                /* saves parameters */
	//	exit(status)
}

var (
	_tool_max = 68 /*Not static. Driver reads  */

)

type CNC struct {
	Rs274ngc_t
	_tools [inc.CANON_TOOL_MAX]inc.CANON_TOOL_TABLE /*Not static. Driver writes */
}

var cnc = CNC{}

/*********************************************************************/
func (c *CNC) init() {
	c.SetCanon(&canon.Canon_t{})
}

func (c *CNC) open(filename string) inc.STATUS {
	if len(filename) == 0 {
		return c.Open("cds-sample.ngc")
	} else {
		return c.Open(filename)
	}
}

/* interpret_from_file

   Returned Value: int (0 or 1)
   If any of the following errors occur, this returns 1.
   Otherwise, it returns 0.
   1. rs274ngc_read returns something other than RS274NGC_OK or
   RS274NGC_EXECUTE_FINISH, no_stop is off, and the user elects
   not to continue.
   2. rs274ngc_execute returns something other than RS274NGC_OK,
   EXIT, or RS274NGC_EXECUTE_FINISH, no_stop is off, and the user
   elects not to continue.

   Side Effects:
   An open NC-program file is interpreted.

   Called By:
   main

   This emulates the way the EMC system uses the interpreter.

   If the do_next argument is 1, this goes into MDI mode if an error is
   found. In that mode, the user may (1) enter code or (2) enter "quit" to
   get out of MDI. Once out of MDI, this asks the user whether to continue
   interpreting the file.

   If the do_next argument is 0, an error does not stop interpretation.

   If the do_next argument is 2, an error stops interpretation.

*/

func (c *CNC) interpret_from_file( /* ARGUMENTS                  */
	do_next int, /* what to do if error        */
	block_delete, /* switch which is ON or OFF  */
	print_stack ON_OFF) inc.STATUS { /* option which is ON or OFF  */

	var (
		status inc.STATUS
	)
	//char line[RS274NGC_TEXT_SIZE];
	stdin := bufio.NewReader(os.Stdin)

	for {
		status = c.Read(nil)
		if (status == inc.RS274NGC_EXECUTE_FINISH) && (block_delete == ON) {
			continue
		} else if status == inc.RS274NGC_ENDFILE {
			break
		}

		if (status != inc.RS274NGC_OK) && // should not be EXIT
			(status != inc.RS274NGC_EXECUTE_FINISH) {
			c.report_error(status, print_stack)
			if (status == inc.NCE_FILE_ENDED_WITH_NO_PERCENT_SIGN) ||
				(do_next == 2) { /* 2 means stop */
				status = 1
				break
			} else if do_next == 1 { /* 1 means MDI */
				fmt.Fprintf(os.Stderr, "starting MDI\n")
				c.interpret_from_keyboard(block_delete, print_stack)
				fmt.Fprintf(os.Stderr, "continue program? y/n =>")

				line, _, _ := stdin.ReadLine()
				if line[0] != 'y' {
					status = 1
					break
				} else {
					continue
				}
			} else { /* if do_next == 0 -- 0 means continue */
				continue
			}
		}
		status = c.Execute()
		if (status != inc.RS274NGC_OK) &&
			(status != inc.RS274NGC_EXIT) &&
			(status != inc.RS274NGC_EXECUTE_FINISH) {
			c.report_error(status, print_stack)
			status = 1
			if do_next == 1 { /* 1 means MDI */
				fmt.Fprintf(os.Stderr, "starting MDI\n")
				c.interpret_from_keyboard(block_delete, print_stack)
				fmt.Fprintf(os.Stderr, "continue program? y/n =>")
				line, _, _ := stdin.ReadLine()
				if line[0] != 'y' {
					break
				}
			} else if do_next == 2 { /* 2 means stop */
				break
			}
		} else if status == inc.RS274NGC_EXIT {
			break
		}
	}
	return inc.If(status == 1, 1, 0).(int)
}

/************************************************************************/

/* read_tool_file

   Returned Value: int
   If any of the following errors occur, this returns 1.
   Otherwise, it returns 0.
   1. The file named by the user cannot be opened.
   2. No blank line is found.
   3. A line of data cannot be read.
   4. A tool slot number is less than 1 or >= _tool_max

   Side Effects:
   Values in the tool table of the machine setup are changed,
   as specified in the file.

   Called By: main

   Tool File Format
   -----------------
   Everything above the first blank line is read and ignored, so any sort
   of header material may be used.

   Everything after the first blank line should be data. Each line of
   data should have four or more items separated by white space. The four
   required items are slot, tool id, tool length offset, and tool diameter.
   Other items might be the holder id and tool description, but these are
   optional and will not be read. Here is a sample line:

   20  1419  4.299  1.0   1 inch carbide end mill

   The tool_table is indexed by slot number.

*/

func (c *CNC) read_tool_file(file_name string) int { /* name of tool file */

	var (
		tool_file_port *os.File
	)
	//		char buffer[1000];
	//		int slot;
	//		int tool_id;
	//		double offset;
	//		double diameter;

	var (
		//length int
		err error
	)

	if len(file_name) == 0 { /* ask for name if given name is empty string */
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("name of tool file => ")
		data, _, _ := reader.ReadLine()

		if tool_file_port, err = os.Open(string(data)); err != nil {
			fmt.Println("Cannot open ", data)
			return -1
		}
	} else {
		if tool_file_port, err = os.Open(file_name); err != nil {
			fmt.Println("Cannot open ", file_name)
			return -1
		}

	}

	for slot := 0; slot <= _tool_max; slot++ /* initialize */ {

		//todo _tools[slot].id = -1
		c._tools[slot].Length = 0
		c._tools[slot].Diameter = 0
	}

	reader := bufio.NewReader(tool_file_port)
	for {
		if l, _, e := reader.ReadLine(); e != nil {
			fmt.Println("Bad tool file format")
		} else if len(l) == 0 {
			break
		}
	}

	l := make([]byte, 1024)
	var (
		slot, tool_id    int
		offset, diameter float64
	)
	for {
		if l, _, err = reader.ReadLine(); err != nil {
			break
		}
		if n, _ := fmt.Sscanf(string(l), "%d %d %f %f", &slot, &tool_id, &offset, &diameter); n < 4 {
			fmt.Printf("Bad input line \"%s\" in tool file", string(l))
			return -1
		}

		if (slot < 0) || (slot > _tool_max) { /* zero and max both OK */
			fmt.Printf("Out of range tool slot number %d\n", slot)
			return -1
		}
		//todo _tools[slot].id = tool_id
		c._tools[slot].Length = offset
		c._tools[slot].Diameter = diameter

	}
	return 0
}

/***********************************************************************/

/* interpret_from_keyboard

   Returned Value: int (0)

   Side effects:
   Lines of NC code entered by the user are interpreted.

   Called by:
   interpret_from_file
   main

   This prompts the user to enter a line of rs274 code. When the user
   hits <enter> at the end of the line, the line is executed.
   Then the user is prompted to enter another line.

   Any canonical commands resulting from executing the line are printed
   on the monitor (stdout).  If there is an error in reading or executing
   the line, an error message is printed on the monitor (stderr).

   To exit, the user must enter "quit" (followed by a carriage return).

*/
func (c *CNC) interpret_from_keyboard( /* ARGUMENTS                 */
	block_delete ON_OFF, /* switch which is ON or OFF */
	print_stack ON_OFF) int { /* option which is ON or OFF */

	//char line[RS274NGC_TEXT_SIZE];
	//int status;

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("READ => ")
		line, _, _ := reader.ReadLine()

		if bytes.ContainsAny(line, "quit") {
			return 0
		}

		status := c.Read(line)
		if (status == inc.RS274NGC_EXECUTE_FINISH) && (block_delete == ON) {

		} else if status == inc.RS274NGC_ENDFILE {

		} else if (status != inc.RS274NGC_EXECUTE_FINISH) &&
			(status != inc.RS274NGC_OK) {
			c.report_error(status, print_stack)
		} else {
			status = c.Execute()
			if (status == inc.RS274NGC_EXIT) ||
				(status == inc.RS274NGC_EXECUTE_FINISH) {

			} else if status != inc.RS274NGC_OK {
				c.report_error(status, print_stack)

			}
		}
	}
}

func (c *CNC) report_error( /* ARGUMENTS                            */
	error_code int, /* the code number of the error message */
	print_stack ON_OFF) { /* print stack if ON, otherwise not     */

	//char buffer[RS274NGC_TEXT_SIZE];
	//int k;

	buffer := inc.Rs274ngc_error_text(error_code) /* for coverage of code */
	if len(buffer) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", "Unknown error, bad error code")
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", buffer)
	}

	fmt.Fprintf(os.Stderr, "%s\n", c.LineText())

	//	if print_stack == rs274ngc.ON {
	//		for k := 0; ; k++ {
	//			rs274ngc_stack_name(k, buffer, RS274NGC_TEXT_SIZE)
	//			if (buffer[0] != 0); fprintf(stderr, "%s\n", buffer) {
	//
	//			} else {
	//				break
	//			}
	//		}
	//	}
}

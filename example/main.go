package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/flyingyizi/rs274ngc"
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

/*
命令行参数的格式可以是：
-flag xxx （使用空格，一个 - 符号）
--flag xxx （使用空格，两个 - 符号）
-flag=xxx （使用等号，一个 - 符号）
--flag=xxx （使用等号，两个 - 符号）
其中，布尔类型的参数防止解析时的二义性，应该使用等号的方式指定。
*/

func main() {
	//go run flag.go -id=2 -name="golang"
	//ok := flag.Bool("ok", false, "is ok")
	//id := flag.Int("id", 0, "id")
	//port := flag.String("port", ":8080", "http listen port")
	//var name string
	//flag.StringVar(&name, "name", "123", "name")
	//
	//flag.Parse()
	//
	//    // After parsing, the arguments after the flag are available as the slice flag.Args() or individually as flag.Arg(i). The arguments are indexed from 0 through flag.NArg()-1
	//    // Args returns the non-flag command-line arguments
	//    // NArg is the number of arguments remaining after flags have been processed
	//    fmt.Printf("args=%s, num=%d\n", flag.Args(), flag.NArg())
	//    for i := 0; i != flag.NArg(); i++ {
	//        fmt.Printf("arg[%d]=%s\n", i, flag.Arg(i))
	//    }

	//fmt.Println("ok:", *ok)
	//fmt.Println("id:", *id)
	//fmt.Println("port:", *port)
	//fmt.Println("name:", name)

	//	   int status;
	//	   int choice;
	//	   int do_next;                                  /* 0=continue, 1=mdi, 2=stop */
	//	   int block_delete;
	//	   char buffer[80];
	//	   int tool_flag;
	//	   int gees[RS274NGC_ACTIVE_G_CODES];
	//	   int ems[RS274NGC_ACTIVE_M_CODES];
	//	   double sets[RS274NGC_ACTIVE_SETTINGS];
	//	   char default_name[] = "rs274ngc.var";
	//	   int print_stack;

	flag.Parse()

	if len(os.Args) > 3 {
		fmt.Println("Usage \"%s\"\n", os.Args[0])
		fmt.Println("   or \"%s <input file>\"\n", os.Args[0])
		fmt.Println("   or \"%s <input file> <output file>\"\n", os.Args[0])
		return
	}

	var (
		choice int
		//do_next                      = 2 /* 2=stop */
		block_delete rs274ngc.ON_OFF = rs274ngc.OFF
		//print_stack  rs274ngc.ON_OFF = rs274ngc.OFF
		tool_flag    = 0
		default_name = "rs274ngc.var"
	)
	_parameter_file_name := default_name
	//todo _outfile = stdout;                       /* may be reset below */

	for {
		fmt.Println("enter a number:")
		fmt.Println("1 = start interpreting")
		fmt.Println("2 = choose parameter file ...")
		fmt.Println("3 = read tool file ...")
		fmt.Println("4 = turn block delete switch %s",
			inc.If((block_delete == rs274ngc.OFF), "ON", "OFF").(string))

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
			//TODO
			//if (read_tool_file("") != 0) {
			//   return
			//}
			//tool_flag = 1;
		} else if choice == 4 {
			block_delete = inc.If((block_delete == rs274ngc.OFF), rs274ngc.ON, rs274ngc.OFF).(rs274ngc.ON_OFF)
		} else if choice == 5 {
			//todo adjust_error_handling(argc, &print_stack, &do_next);
		}

	}

	fmt.Println("executing")
	if tool_flag == 0 {
		if s := read_tool_file("rs274ngc.tool_default"); s != 0 {
			return
		}
	}

	//TODO
	if len(os.Args) == 3 {

		_outfile = fopen(argv[2], "w")
		if _outfile == NULL {
			fmt.Println("could not open output file %s", argv[2])
			return
		}
	}

	//	status := rs274ngc.TEST.Init()
	//	if status != inc.RS274NGC_OK {
	//		return
	//	}
	//
	//	if argc == 1 {
	//		status = interpret_from_keyboard(block_delete, print_stack)
	//	} else { /* if (argc IS 2 or argc IS 3) */
	//		status = rs274ngc.TEST.Open(argv[1])
	//		if status != inc.RS274NGC_OK { /* do not need to close since not open */
	//			report_error(status, print_stack)
	//			exit(1)
	//		}
	//		status = interpret_from_file(do_next, block_delete, print_stack)
	//		rs274ngc_file_name(buffer, 5)  /* called to exercise the function */
	//		rs274ngc_file_name(buffer, 79) /* called to exercise the function */
	//		rs274ngc_close()
	//	}
	//	rs274ngc_line_length()         /* called to exercise the function */
	//	rs274ngc_sequence_number()     /* called to exercise the function */
	//	rs274ngc_active_g_codes(gees)  /* called to exercise the function */
	//	rs274ngc_active_m_codes(ems)   /* called to exercise the function */
	//	rs274ngc_active_settings(sets) /* called to exercise the function */
	//	rs274ngc_exit()                /* saves parameters */
	//	exit(status)
}

var (
	_tool_max = 68                                     /*Not static. Driver reads  */
	_tools    [inc.CANON_TOOL_MAX]inc.CANON_TOOL_TABLE /*Not static. Driver writes */

)

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

func read_tool_file(file_name string) int { /* name of tool file */

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
			fmt.Println("Cannot open %s", data)
			return -1
		}
	} else {
		if tool_file_port, err = os.Open(file_name); err != nil {
			fmt.Println("Cannot open %s", file_name)
			return -1
		}

	}

	for slot := 0; slot <= _tool_max; slot++ /* initialize */ {

		//todo _tools[slot].id = -1
		_tools[slot].Length = 0
		_tools[slot].Diameter = 0
	}

	reader := bufio.NewReader(tool_file_port)
	for {
		if l, isprefix, _ := reader.ReadLine(); isprefix == true {

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
		if n, _ := fmt.Sscanf(string(l), "%d %d %lf %lf", &slot, &tool_id, &offset, &diameter); n < 4 {
			fmt.Println("Bad input line \"%s\" in tool file", l)
			return -1
		}

		if (slot < 0) || (slot > _tool_max) { /* zero and max both OK */
			fmt.Println("Out of range tool slot number %d\n", slot)
			return -1
		}
		//todo _tools[slot].id = tool_id
		_tools[slot].Length = offset
		_tools[slot].Diameter = diameter

	}
	return 0
}

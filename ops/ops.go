package ops

import (
	"math"

	"github.com/flyingyizi/rs274ngc/inc"
)

// These are not enums because the "&" operator is used in
// reading the operation names and is illegal with an enum
type Operation int

const (
	_ Operation = iota
	// unary operations
	ABS   //  1
	ACOS  //  2
	ASIN  //  3
	ATAN  //  4
	COS   // 5
	EXP   // 6
	FIX   // 7
	FUP   // 8
	LN    //
	ROUND //  10
	SIN   //  11
	SQRT  // 12
	TAN   // 13

	// binary operations
	DIVIDED_BY
	MODULO
	POWER
	TIMES
	AND2
	EXCLUSIVE_OR
	MINUS
	NON_EXCLUSIVE_OR
	PLUS
	RIGHT_BRACKET
)

//void execute_binary(double* left, BinaryOperation operation, double* right);
//void execute_unary(double* double_ptr, UnaryOperation operation);
//
//int precedence(BinaryOperation an_operator);

/****************************************************************************/

/* execute binary

   Returned value: int
   If execute_binary1 or execute_binary2 returns an error code, this
   returns that code.
   Otherwise, it returns RS274NGC_OK.

   Side effects: The value of left is set to the result of applying
   the operation to left and right.

   Called by: read_real_expression

   This just calls either execute_binary1 or execute_binary2.

*/

func Execute_binary(
	left *float64,
	operation Operation,
	right *float64) inc.STATUS {
	//static char name[] = "execute_binary";
	//int status;

	if operation < AND2 {
		(execute_binary1(left, operation, right))
	} else {
		(execute_binary2(left, operation, right))
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* execute_binary1

   Returned Value: int
   If any of the following errors occur, this returns the error shown.
   Otherwise, it returns RS274NGC_OK.
   1. operation is unknown: NCE_BUG_UNKNOWN_OPERATION
   2. An attempt is made to divide by zero: NCE_ATTEMPT_TO_DIVIDE_BY_ZERO
   3. An attempt is made to raise a negative number to a non-integer power:
   NCE_ATTEMPT_TO_RAISE_NEGATIVE_TO_NON_INTEGER_POWER

   Side effects:
   The result from performing the operation is put into what left points at.

   Called by: read_real_expression.

   This executes the operations: DIVIDED_BY, MODULO, POWER, TIMES.

*/

func execute_binary1( /* ARGUMENTS                       */
	left *float64, /* pointer to the left operand     */
	operation Operation, /* integer code for the operation  */
	right *float64) inc.STATUS { /* pointer to the right operand    */
	//        static char name[] = "execute_binary1";
	switch operation {
	case DIVIDED_BY:
		if *right == 0.0 {
			return inc.NCE_ATTEMPT_TO_DIVIDE_BY_ZERO
		}

		*left = (*left / *right)
		break
	case MODULO: /* always calculates a positive answer */
		*left = math.Mod(*left, *right)
		if *left < 0.0 {
			*left = (*left + math.Abs(*right))
		}
		break
	case POWER:
		if (*left < 0.0) && (math.Floor(*right) != *right) {
			return inc.NCE_ATTEMPT_TO_RAISE_NEGATIVE_TO_NON_INTEGER_POWER
		}

		*left = math.Pow(*left, *right)
		break
	case TIMES:
		*left = (*left * *right)
		break
	default:
		return (inc.NCE_BUG_UNKNOWN_OPERATION)
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* execute_binary2

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. operation is unknown: NCE_BUG_UNKNOWN_OPERATION

   Side effects:
   The result from performing the operation is put into what left points at.

   Called by: read_real_expression.

   This executes the operations: AND2, EXCLUSIVE_OR, MINUS,
   NON_EXCLUSIVE_OR, PLUS. The RS274/NGC manual [NCMS] does not say what
   the calculated value of the three logical operations should be. This
   function calculates either 1.0 (meaning true) or 0.0 (meaning false).
   Any non-zero input value is taken as meaning true, and only 0.0 means
   false.

*/
func execute_binary2( /* ARGUMENTS                       */
	left *float64, /* pointer to the left operand     */
	operation Operation, /* integer code for the operation  */
	right *float64) inc.STATUS { /* pointer to the right operand    */
	//static char name[] = "execute_binary2";
	switch operation {
	case AND2:
		*left = inc.If((*left == 0.0) || (*right == 0.0), 0.0, 1.0).(float64)
		break
	case EXCLUSIVE_OR:
		*left = inc.If(((*left == 0.0) && (*right != 0.0)) ||
			((*left != 0.0) && (*right == 0.0)), 1.0, 0.0).(float64)
		break
	case MINUS:
		*left = (*left - *right)
		break
	case NON_EXCLUSIVE_OR:
		*left = inc.If((*left != 0.0) || (*right != 0.0), 1.0, 0.0).(float64)
		break
	case PLUS:
		*left = (*left + *right)
		break
	default:
		return (inc.NCE_BUG_UNKNOWN_OPERATION)
	}
	return inc.RS274NGC_OK
}

/* execute_unary

   Returned Value: int
   If any of the following errors occur, this returns the error code shown.
   Otherwise, it returns RS274NGC_OK.
   1. the operation is unknown: NCE_BUG_UNKNOWN_OPERATION
   2. the argument to acos is not between minus and plus one:
   NCE_ARGUMENT_TO_ACOS_OUT_RANGE
   3. the argument to asin is not between minus and plus one:
   NCE_ARGUMENT_TO_ASIN_OUT_RANGE
   4. the argument to the natural logarithm is not positive:
   NCE_ZERO_OR_NEGATIVE_ARGUMENT_TO_LN
   5. the argument to square root is negative:
   NCE_NEGATIVE_ARGUMENT_TO_SQRT

   Side effects:
   The result from performing the operation on the value in double_ptr
   is put into what double_ptr points at.

   Called by: read_unary.

   This executes the operations: ABS, ACOS, ASIN, COS, EXP, FIX, FUP, LN
   ROUND, SIN, SQRT, TAN

   All angle measures in the input or output are in degrees.

*/

func Execute_unary( /* ARGUMENTS                       */
	double_ptr *float64, /* pointer to the operand          */
	operation Operation) inc.STATUS { /* integer code for the operation  */

	switch operation {
	case ABS:
		if *double_ptr < 0.0 {
			*double_ptr = (-1.0 * *double_ptr)
		}
		break
	case ACOS:
		if (*double_ptr < -1.0) || (*double_ptr > 1.0) {
			return inc.NCE_ARGUMENT_TO_ACOS_OUT_OF_RANGE
		}
		*double_ptr = math.Acos(*double_ptr)
		*double_ptr = ((*double_ptr * 180.0) / inc.PI)
		break
	case ASIN:
		if (*double_ptr < -1.0) || (*double_ptr > 1.0) {
			return inc.NCE_ARGUMENT_TO_ASIN_OUT_OF_RANGE
		}

		*double_ptr = math.Asin(*double_ptr)
		*double_ptr = ((*double_ptr * 180.0) / inc.PI)
		break
	case COS:
		*double_ptr = math.Cos((*double_ptr * inc.PI) / 180.0)
		break
	case EXP:
		*double_ptr = math.Exp(*double_ptr)
		break
	case FIX:
		*double_ptr = math.Floor(*double_ptr)
		break
	case FUP:
		*double_ptr = math.Ceil(*double_ptr)
		break
	case LN:
		if *double_ptr <= 0.0 {
			return inc.NCE_ZERO_OR_NEGATIVE_ARGUMENT_TO_LN
		}
		*double_ptr = math.Log(*double_ptr)
		break
	case ROUND:
		*double_ptr = (float64)((int)(*double_ptr + (inc.If(*double_ptr < 0.0, -0.5, 0.5).(float64))))
		break
	case SIN:
		*double_ptr = math.Sin((*double_ptr * inc.PI) / 180.0)
		break
	case SQRT:
		if *double_ptr < 0.0 {
			return inc.NCE_NEGATIVE_ARGUMENT_TO_SQRT
		}
		*double_ptr = math.Sqrt(*double_ptr)
		break
	case TAN:
		*double_ptr = math.Tan((*double_ptr * inc.PI) / 180.0)
		break
	default:
		return (inc.NCE_BUG_UNKNOWN_OPERATION)
	}
	return inc.RS274NGC_OK
}

/****************************************************************************/

/* precedence

   Returned Value: int
   This returns an integer representing the precedence level of an_operator

   Side Effects: None

   Called by: read_real_expression

   To add additional levels of operator precedence, edit this function.

*/

func Precedence(an_operator Operation) uint {
	if an_operator == RIGHT_BRACKET {
		return 1
	} else if an_operator == POWER {
		return 4
	} else if an_operator >= AND2 {
		return 2
	} else {
		return 3
	}
}

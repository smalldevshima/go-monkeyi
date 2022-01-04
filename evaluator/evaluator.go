package evaluator

import (
	"fmt"

	"github.com/smalldevshima/go-monkey/ast"
	"github.com/smalldevshima/go-monkey/object"
)

// Constants / Variables

// Error format strings
const (
	ERR_PREFIX_UNKNOWN     ErrorFormat = "unknown operator: %s%s"
	ERR_INFIX_UNKNOWN      ErrorFormat = "unknown operator: %s %s %s"
	ERR_INFIX_MISMATCH     ErrorFormat = "type mismatch: %s %s %s"
	ERR_IDENTIFIER_UNKNOWN ErrorFormat = "unknown identifier: %s"
)

var (
	NULL = &object.Null{}

	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}

	// FALSY_VALUES is a list of all object values considered falsy in Monkey
	FALSY_VALUES = []object.Object{NULL, FALSE}
)

// Functions

// isTruthy defines which values are truthy in the Monkey language
func isTruthy(obj object.Object) bool {
	for _, falsyVal := range FALSY_VALUES {
		if falsyVal == obj {
			return false
		}
	}

	return true
}

func newError(format ErrorFormat, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(string(format), a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.O_ERROR
	}
	return false
}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// * Statements:
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)

	// * Literal expressions:
	case *ast.BooleanLiteral:
		return nativeBooleanToObject(node.Value)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}

	// * Operator expressions:
	case *ast.PrefixExpression:
		operand := Eval(node.Right, env)
		if isError(operand) {
			return operand
		}
		return evalPrefixExpression(node.Operator, operand)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	// * Control flow expressions:
	case *ast.IfExpression:
		return evalIfExpression(node, env)

	// * Identifiers, function calls:
	case *ast.Identifier:
		return evalIdentifier(node, env)
	}

	return nil
}

func nativeBooleanToObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalProgram(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range statements {
		result = Eval(stmt, env)

		// * return early, if result is an object.ReturnValue or an object.Error
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range statements {
		result = Eval(stmt, env)

		if result != nil {
			// * return early, if result type is object.O_RETURN_VALUE or object.O_ERRIR
			if result.Type() == object.O_RETURN_VALUE || isError(result) {
				return result
			}
		}
	}

	return result
}

func evalPrefixExpression(operator string, operand object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(operand)
	case "-":
		return evalDashOperatorExpression(operand)
	}
	return newError(ERR_PREFIX_UNKNOWN, operator, operand.Type())
}

// evalBangOperatorExpression returns the opposite object of the isTruthy(operand) result
func evalBangOperatorExpression(operand object.Object) object.Object {
	if isTruthy(operand) {
		return FALSE
	}
	return TRUE
}

func evalDashOperatorExpression(operand object.Object) object.Object {
	if operand.Type() != object.O_INTEGER {
		return newError(ERR_PREFIX_UNKNOWN, "-", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	// * need to switch on both the type of left and right
	case left.Type() != right.Type():
		return newError(ERR_INFIX_MISMATCH, left.Type(), operator, right.Type())
	case left.Type() == object.O_INTEGER && right.Type() == object.O_INTEGER:
		return evalIntegerInfixExpression(operator, left, right)

	// * special cases for infix operators '==' and '!='
	// * directly compare pointers, since booleans and null use global objects
	// * all other types are filtered out by preceding cases
	case operator == "==":
		return nativeBooleanToObject(left == right)
	case operator == "!=":
		return nativeBooleanToObject(left != right)
	}

	return newError(ERR_INFIX_UNKNOWN, left.Type(), operator, right.Type())
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftInt := left.(*object.Integer).Value
	rightInt := right.(*object.Integer).Value
	var newInt int64
	switch operator {
	case "+":
		newInt = leftInt + rightInt
	case "-":
		newInt = leftInt - rightInt
	case "*":
		newInt = leftInt * rightInt
	case "/":
		newInt = leftInt / rightInt
	case "==":
		return nativeBooleanToObject(leftInt == rightInt)
	case "!=":
		return nativeBooleanToObject(leftInt != rightInt)
	case "<":
		return nativeBooleanToObject(leftInt < rightInt)
	case ">":
		return nativeBooleanToObject(leftInt > rightInt)
	default:
		return newError(ERR_INFIX_UNKNOWN, left.Type(), operator, right.Type())
	}

	return &object.Integer{Value: newInt}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Then, env)
	} else if ie.Otherwise != nil {
		return Eval(ie.Otherwise, env)
	}

	return NULL
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError(ERR_IDENTIFIER_UNKNOWN, node.Value)
	}

	return val
}

/// Types

type ErrorFormat string

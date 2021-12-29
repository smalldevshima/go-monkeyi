package parser

import (
	"fmt"
	"testing"

	"github.com/smalldevshima/go-monkey/ast"
	"github.com/smalldevshima/go-monkey/lexer"
	"github.com/smalldevshima/go-monkey/token"
)

func TestLetStatements(t *testing.T) {
	input := `
		let x = 5;
		let y = 10;
		let foobar = 838383;
		let foo = bar;
		`

	p := New(lexer.New(input))

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.Statements) != 4 {
		t.Fatalf("program.Statements does not contain 4 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"x", 5},
		{"y", 10},
		{"foobar", 838383},
		{"foo", "bar"},
	}

	for index, test := range tests {
		stmt := program.Statements[index]
		ok := t.Run("statement"+fmt.Sprint(index+1), func(tt *testing.T) {
			checkLetStatement(tt, stmt, test.expectedIdentifier, test.expectedValue)
		})
		if !ok {
			t.Fail()
		}
	}
}

func TestReturnStatements(t *testing.T) {
	input := `
		return 5;
		return 10;
		return add(5, 10);
		return x + z;
		`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 4 {
		t.Fatalf("program.Statements does not contain 4 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("smt not *ast.ReturnStatement. got=%T", stmt)
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral not 'return', got=%q", returnStmt.TokenLiteral())
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := `foobar;`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.Identifier. got=%T", stmt.Expression)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value is not 'foobar'. got=%q", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral is not 'foobar'. got=%q", ident.TokenLiteral())
	}
}

func TestLiteralExpression(t *testing.T) {
	literalTests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"5;", 5},
		{"true;", true},
		{"false;", false},
	}

	for _, test := range literalTests {
		t.Run("literal/"+fmt.Sprint(test.expectedValue), func(tt *testing.T) {
			l := lexer.New(test.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(tt, p)

			if len(program.Statements) != 1 {
				tt.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
			}
			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				tt.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
			}
			checkLiteralExpression(tt, stmt.Expression, test.expectedValue)
		})
	}
}

func TestPrefixExpression(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"!true", "!", true},
		{"!false", "!", false},
	}

	for _, test := range prefixTests {
		t.Run("prefix"+test.operator, func(tt *testing.T) {
			l := lexer.New(test.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(tt, p)

			if len(program.Statements) != 1 {
				tt.Fatalf("program.Statements does not contain 1 statement. got=%d: %s", len(program.Statements), program.Statements)
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				tt.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
			}

			exp, ok := stmt.Expression.(*ast.PrefixExpression)
			if !ok {
				t.Fatalf("stmt.Expression is not *ast.PrefixExpression. got=%T", stmt.Expression)
			}
			if exp.Operator != test.operator {
				t.Fatalf("exp.Operator is not %q. got=%q", test.operator, exp.Operator)
			}
			checkLiteralExpression(tt, exp.Right, test.value)
		})
	}
}

func TestInfixExpression(t *testing.T) {
	infixTests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"1 + 2", 1, "+", 2},
		{"3 - 4", 3, "-", 4},
		{"5 * 6", 5, "*", 6},
		{"7 / 8", 7, "/", 8},
		{"9 > 10", 9, ">", 10},
		{"11 < 12", 11, "<", 12},
		{"13 == 14", 13, "==", 14},
		{"15 != 16", 15, "!=", 16},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, test := range infixTests {
		t.Run("infix"+test.operator, func(tt *testing.T) {
			l := lexer.New(test.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(tt, p)
			if len(program.Statements) != 1 {
				tt.Fatalf("program.Statements does not contain 1 statement. got=%d: %s", len(program.Statements), program.Statements)
			}

			stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				tt.Fatalf("program.Statements[0] is not *ast.ExpressionStatement. got=%T", program.Statements[0])
			}
			checkInfixExpression(tt, stmt.Expression, test.left, test.operator, test.right)
		})
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b);",
		},
		{
			"!-a",
			"(!(-a));",
		},
		{
			"a + b + c",
			"((a + b) + c);",
		},
		{
			"a + b - c",
			"((a + b) - c);",
		},
		{
			"a * b * c",
			"((a * b) * c);",
		},
		{
			"a * b / c",
			"((a * b) / c);",
		},
		{
			"a + b / c",
			"(a + (b / c));",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f);",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4);((-5) * 5);",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4));",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4));",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)));",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4);",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2);",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5));",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5));",
		},
		{
			"!(true == true)",
			"(!(true == true));",
		},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		actual := program.String()
		if actual != tt.expected {
			t.Errorf("\nexpected=\n\t%q\ngot=\n\t%q", tt.expected, actual)
		}
	}
}

func TestFunctionCallExpression(t *testing.T) {
	input := `add(5, 10, 20)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not have 1 statement. got=%d: %s", len(program.Statements), program.Statements)
	}
	// todo
}

/// helpers

func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors:", len(errors))
	for i, msg := range errors {
		if i >= 10 {
			t.Errorf("omitting more errors ...")
			break
		}
		t.Errorf("%3d: %s", i+1, msg)
	}
	t.FailNow()
}

func checkLetStatement(t *testing.T, s ast.Statement, name string, value interface{}) {
	t.Helper()
	if s.TokenLiteral() != "let" {
		t.Fatalf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Fatalf("s not *ast.LetStatement. got=%T", s)
	}
	checkIdentifier(t, letStmt.Name, name)
	checkLiteralExpression(t, letStmt.Value, value)
}

func checkIntegerLiteral(t *testing.T, exp ast.Expression, value int64) {
	t.Helper()
	intLit, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp is not *ast.IntegerLiteral. got=%T", exp)
		return
	}
	if intLit.Token.Type != token.INTEGER {
		t.Errorf("intLit.Token.Type is not %q. got=%q", token.INTEGER, intLit.Token.Type)
	}
	if intLit.Value != value {
		t.Errorf("intLit.Value is not %d. got=%d", value, intLit.Value)
	}
}

func checkBooleanLiteral(t *testing.T, exp ast.Expression, value bool) {
	t.Helper()
	boolLit, ok := exp.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("exp is not *ast.BooleanLiteral. got=%T", exp)
		return
	}
	if boolLit.Token.Type != token.TRUE && boolLit.Token.Type != token.FALSE {
		t.Errorf("boolLit.Token.Type is neither %q nor %q. got=%q", token.TRUE, token.FALSE, boolLit.Token.Type)
	}
	if boolLit.Value != value {
		t.Errorf("boolLit.Value is not %v. got=%v", value, boolLit.Value)
	}
}

func checkIdentifier(t *testing.T, exp ast.Expression, value string) {
	t.Helper()
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
	}
}

func checkLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) {
	t.Helper()
	switch v := expected.(type) {
	case int:
		checkIntegerLiteral(t, exp, int64(v))
	case int64:
		checkIntegerLiteral(t, exp, v)
	case string:
		checkIdentifier(t, exp, v)
	case bool:
		checkBooleanLiteral(t, exp, v)
	default:
		t.Errorf("type of expected not handled. got=%T", expected)
	}
}

func checkInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) {
	t.Helper()
	infixExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not *ast.InfixExpression. got=%T(%s)", exp, exp)
		return
	}
	checkLiteralExpression(t, infixExp.Left, left)
	if infixExp.Operator != operator {
		t.Errorf("infixExp.Operator is not %q. got=%q", operator, infixExp.Operator)
	}
	checkLiteralExpression(t, infixExp.Right, right)
}

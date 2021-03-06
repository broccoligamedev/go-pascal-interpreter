package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// GRAMMAR

/*
EXPR		TERM ((ADD | SUB) TERM)+
TERM		FACTOR ((MUL | DIV) FACTOR)+
FACTOR		INT | (LPAREN EXPR RPAREN)
*/

// note(ryan): There's no good way to implement tagged unions in Go so we're left with
// either fat structs or dummy interfaces or empty interfaces. I chose fat structs because I think it's a bit
// cleaner.

type TokenType int

const (
	INTEGER TokenType = iota
	PLUS
	MINUS
	MULTIPLY
	DIVIDE
	EOF
	LPAREN
	RPAREN
)

type Token struct {
	tokenType TokenType
	value     int
}

type ASTNodeType int

const (
	BIN_OP ASTNodeType = iota
	NUM
)

type ASTNode struct {
	nodeType ASTNodeType
	left     *ASTNode
	right    *ASTNode
	token    *Token
}

// note(ryan): this is used for debugging purposes because sometimes we want
// print out the actual token and not just the enum value. see: populateTokenMap()
var tokenMap map[TokenType]string

var text string
var pos int
var currentToken *Token
var currentCharacter rune
var eof = false

func populateTokenMap() {
	tokenMap = make(map[TokenType]string)
	tokenMap[INTEGER] = "INT"
	tokenMap[PLUS] = "+"
	tokenMap[MINUS] = "-"
	tokenMap[MULTIPLY] = "*"
	tokenMap[DIVIDE] = "/"
	tokenMap[EOF] = "EOF"
	tokenMap[LPAREN] = "("
	tokenMap[RPAREN] = ")"
}

func (t *Token) String() string {
	return strconv.Itoa(t.value)
}

func advance() {
	pos++
	if pos > len(text)-1 {
		eof = true
	} else {
		currentCharacter = rune(text[pos])
	}
}

func skipWhitespace() {
	for !eof && unicode.IsSpace(currentCharacter) {
		advance()
	}
}

func expr() (*ASTNode, error) {
	var err error
	node, err := term()
	if err != nil {
		return nil, err
	}
	for currentToken.tokenType == PLUS ||
		currentToken.tokenType == MINUS {
		token := currentToken
		switch token.tokenType {
		case PLUS:
			err = eat(PLUS)
			if err != nil {
				return nil, err
			}
		case MINUS:
			err = eat(MINUS)
			if err != nil {
				return nil, err
			}
		}
		right, err := term()
		if err != nil {
			return nil, err
		}
		node = &ASTNode{
			nodeType: BIN_OP,
			left:     node,
			token:    token,
			right:    right,
		}
	}
	return node, nil
}

func term() (*ASTNode, error) {
	var err error
	node, err := factor()
	if err != nil {
		return nil, err
	}
	for currentToken.tokenType == DIVIDE ||
		currentToken.tokenType == MULTIPLY {
		token := currentToken
		switch token.tokenType {
		case DIVIDE:
			err = eat(DIVIDE)
			if err != nil {
				return nil, err
			}
		case MULTIPLY:
			err = eat(MULTIPLY)
			if err != nil {
				return nil, err
			}
		}
		right, err := factor()
		if err != nil {
			return nil, err
		}
		node = &ASTNode{
			left:     node,
			right:    right,
			nodeType: BIN_OP,
			token:    token,
		}
	}
	return node, nil
}

func factor() (*ASTNode, error) {
	var err error
	// note(ryan): if we find an LPAREN here then we eat it and recursively call expr.
	// this lets us handle nested expressions with parenthesis. see the grammar for more
	// info.
	token := currentToken
	if token.tokenType == INTEGER {
		eat(INTEGER)
		return &ASTNode{
			nodeType: NUM,
			token:    token,
		}, nil
	} else if token.tokenType == LPAREN {
		err = eat(LPAREN)
		if err != nil {
			return nil, err
		}
		node, err := expr()
		err = eat(RPAREN)
		if err != nil {
			return nil, err
		}
		return node, nil
	}
	return nil, errors.New("bad token type in factor")
}

func integer() (int, error) {
	digits := []rune{}
	for !eof && unicode.IsDigit(currentCharacter) {
		digits = append(digits, currentCharacter)
		advance()
	}
	//fmt.Println("currentCharacter", strconv.QuoteRune(currentCharacter))
	if !eof {
		if currentCharacter != '+' &&
			currentCharacter != '-' &&
			currentCharacter != '/' &&
			currentCharacter != '*' &&
			currentCharacter != ')' &&
			!unicode.IsSpace(currentCharacter) {
			return 0, errors.New("bad syntax")
		}
	}
	value, err := strconv.Atoi(string(digits))
	if err != nil {
		return 0, err
	}
	return value, nil
}

func visit(node *ASTNode, visitor func(*ASTNode, interface{}, interface{}) (interface{}, error)) (interface{}, error) {
	// todo(ryan): we might want to encapsulate the visitor function in some kind of visitor struct
	// if we ever want to do something between visits, this could be relevant...
	var err error
	var leftVal interface{}
	var rightVal interface{}
	if node.left != nil {
		leftVal, err = visit(node.left, visitor)
		if err != nil {
			return 0, err
		}
	}
	if node.right != nil {
		rightVal, err = visit(node.right, visitor)
		if err != nil {
			return 0, err
		}
	}
	result, err := visitor(node, leftVal, rightVal)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func evalVisitor(node *ASTNode, leftVal interface{}, rightVal interface{}) (interface{}, error) {
	token := node.token
	var result int
	switch node.nodeType {
	case BIN_OP:
		switch token.tokenType {
		case PLUS:
			result = leftVal.(int) + rightVal.(int)
		case MINUS:
			result = leftVal.(int) - rightVal.(int)
		case MULTIPLY:
			result = leftVal.(int) * rightVal.(int)
		case DIVIDE:
			if rightVal.(int) == 0 {
				return nil, errors.New("can't divide by zero")
			}
			result = leftVal.(int) / rightVal.(int)
		}
	case NUM:
		return token.value, nil
	}
	return result, nil
}

func lispVisitor(node *ASTNode, leftVal interface{}, rightVal interface{}) (interface{}, error) {
	token := node.token
	var result string
	switch node.nodeType {
	case BIN_OP:
		result = "(" + tokenMap[token.tokenType] + " " + leftVal.(string) + " " + rightVal.(string) + ")"
	case NUM:
		result = strconv.Itoa(token.value)
	}
	return result, nil
}

func polishVisistor(node *ASTNode, leftVal interface{}, rightVal interface{}) (interface{}, error) {
	token := node.token
	var result string
	switch node.nodeType {
	case BIN_OP:
		result = leftVal.(string) + " " + rightVal.(string) + " " + tokenMap[token.tokenType]
	case NUM:
		result = strconv.Itoa(token.value)
	}
	return result, nil
}

func eat(tokenType TokenType) error {
	var err error
	//fmt.Println("eating " + tokenMap[tokenType])
	if currentToken.tokenType == tokenType {
		currentToken, err = getNextToken()
		if err != nil {
			return err
		}
	} else {
		return errors.New(
			"wrong token type. expected " +
				tokenMap[tokenType] +
				" but got " +
				tokenMap[currentToken.tokenType])
	}
	return nil
}

func getNextToken() (*Token, error) {
	var err error
	for !eof {
		if unicode.IsSpace(currentCharacter) {
			skipWhitespace()
			continue
		}
		if unicode.IsDigit(currentCharacter) {
			var value int
			value, err = integer()
			if err != nil {
				return nil, err
			}
			return &Token{
				tokenType: INTEGER,
				value:     value,
			}, nil
		}
		switch currentCharacter {
		case '+':
			advance()
			return &Token{
				tokenType: PLUS,
			}, nil
		case '-':
			advance()
			return &Token{
				tokenType: MINUS,
			}, nil
		case '*':
			advance()
			return &Token{
				tokenType: MULTIPLY,
			}, nil
		case '/':
			advance()
			return &Token{
				tokenType: DIVIDE,
			}, nil
		case '(':
			advance()
			return &Token{
				tokenType: LPAREN,
			}, nil
		case ')':
			advance()
			return &Token{
				tokenType: RPAREN,
			}, nil
		}
		return nil, errors.New("invalid token: " + string(currentCharacter))
	}
	return &Token{
		tokenType: EOF,
	}, nil
}

func main() {
	var err error
	reader := bufio.NewReader(os.Stdin)
	populateTokenMap()
	for {
		fmt.Print("calc> ")
		text, err = reader.ReadString('\n')
		text = strings.TrimSpace(text)
		eof = false
		pos = 0
		if len(text) == 0 {
			continue
		}
		currentCharacter = rune(text[0])
		currentToken, err = getNextToken()
		// note(ryan): the first token could be bad
		if err != nil {
			fmt.Println("error: " + err.Error())
			continue
		}
		// note(ryan): otherwise we continue as normal
		AST, err := expr()
		if err != nil {
			fmt.Println("error: " + err.Error())
			continue
		}
		err = eat(EOF)
		if err != nil {
			fmt.Println("error: " + err.Error())
			continue
		}
		result, err := visit(AST, evalVisitor)
		if err != nil {
			fmt.Println("error: " + err.Error())
			continue
		}
		fmt.Println(result)
	}
}

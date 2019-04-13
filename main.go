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
EXPR		TERM [[ADD | SUB] TERM]+
TERM		FACTOR [[MUL | DIV] FACTOR]+
FACTOR		INT | (EXPR)
*/

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

var text string
var pos int
var currentToken *Token
var currentCharacter rune
var eof = false

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

func expr() (int, error) {
	// todo(ryan): proper error handling
	result := term()
	for currentToken.tokenType == PLUS ||
		currentToken.tokenType == MINUS {
		token := currentToken
		switch token.tokenType {
		case PLUS:
			eat(PLUS)
			result += term()
		case MINUS:
			eat(MINUS)
			result -= term()
		}
	}
	return result, nil
}

func term() int {
	result := factor()
	for currentToken.tokenType == DIVIDE ||
		currentToken.tokenType == MULTIPLY {
		token := currentToken
		switch token.tokenType {
		case DIVIDE:
			eat(DIVIDE)
			result /= factor()
		case MULTIPLY:
			eat(MULTIPLY)
			result *= factor()
		}
	}
	return result
}

func factor() int {
	result := 0
	if currentToken.tokenType == LPAREN {
		eat(LPAREN)
		result, _ = expr()
		eat(RPAREN)
	} else {
		result = currentToken.value
		eat(INTEGER)
	}
	return result
}

func integer() int {
	digits := []rune{}
	for !eof && unicode.IsDigit(currentCharacter) {
		digits = append(digits, currentCharacter)
		advance()
	}
	value, err := strconv.Atoi(string(digits))
	if err != nil {
		panic(err)
	}
	return value
}

func eat(tokenType TokenType) {
	//fmt.Println("eating " + strconv.Itoa(int(tokenType)))
	if currentToken.tokenType == tokenType {
		currentToken = getNextToken()
	} else {
		fmt.Println(
			"expected " +
				strconv.Itoa(int(tokenType)) +
				" but got " +
				strconv.Itoa(int(currentToken.tokenType)))
		fmt.Println("pos", pos)
		fmt.Println("char", string(currentCharacter))
		panic(errors.New("wrong token type"))
	}
}

func getNextToken() *Token {
	for !eof {
		if unicode.IsSpace(currentCharacter) {
			skipWhitespace()
			continue
		}
		if unicode.IsDigit(currentCharacter) {
			return &Token{
				tokenType: INTEGER,
				value:     integer(),
			}
		}
		switch currentCharacter {
		case '+':
			advance()
			return &Token{
				tokenType: PLUS,
			}
		case '-':
			advance()
			return &Token{
				tokenType: MINUS,
			}
		case '*':
			advance()
			return &Token{
				tokenType: MULTIPLY,
			}
		case '/':
			advance()
			return &Token{
				tokenType: DIVIDE,
			}
		case '(':
			advance()
			return &Token{
				tokenType: LPAREN,
			}
		case ')':
			advance()
			return &Token{
				tokenType: RPAREN,
			}
		}
		panic(errors.New("invalid token: " + string(currentCharacter)))
	}
	return &Token{
		tokenType: EOF,
	}
}

func main() {
	var err error
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("calc> ")
		text, err = reader.ReadString('\n')
		text = strings.TrimSpace(text)
		eof = false
		pos = 0
		currentCharacter = rune(text[0])
		currentToken = getNextToken()
		if err != nil {
			panic(err)
		}
		if len(text) == 0 {
			continue
		}
		result, err := expr()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(result)
	}
}

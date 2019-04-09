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

type TokenType int

const (
	INTEGER TokenType = iota
	PLUS
	MINUS
	MULTIPLY
	DIVIDE
	EOF
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

func term() int {
	token := currentToken
	eat(INTEGER)
	return token.value
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
		}
		panic(errors.New("invalid token: " + string(currentCharacter)))
	}
	return &Token{
		tokenType: EOF,
	}
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

func expr() (int, error) {
	// todo(ryan): proper error handling
	currentToken = getNextToken()
	//fmt.Println("current token", currentToken.tokenType)
	result := term()
	//fmt.Println("current token", currentToken.tokenType)
	for currentToken.tokenType == PLUS ||
		currentToken.tokenType == MINUS ||
		currentToken.tokenType == MULTIPLY ||
		currentToken.tokenType == DIVIDE {
		token := currentToken
		switch token.tokenType {
		case PLUS:
			eat(PLUS)
			result += term()
		case MINUS:
			eat(MINUS)
			result -= term()
		case MULTIPLY:
			eat(MULTIPLY)
			result *= term()
		case DIVIDE:
			eat(DIVIDE)
			result /= term()
		default:
			panic("bad operator")
		}
		//fmt.Println("token :", currentToken.tokenType)
	}
	return result, nil
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

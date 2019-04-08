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
		if currentCharacter == '+' {
			advance()
			return &Token{
				tokenType: PLUS,
			}
		}
		if currentCharacter == '-' {
			advance()
			return &Token{
				tokenType: MINUS,
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

func expr() int {
	currentToken = getNextToken()
	left := currentToken
	eat(INTEGER)
	op := currentToken.tokenType
	if op == PLUS {
		eat(PLUS)
	} else if op == MINUS {
		eat(MINUS)
	}
	right := currentToken
	eat(INTEGER)
	result := 0
	if op == PLUS {
		result = left.value + right.value
	} else if op == MINUS {
		result = left.value - right.value
	}
	return result
}

func main() {
	var err error
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
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
		fmt.Println(expr())
	}
}

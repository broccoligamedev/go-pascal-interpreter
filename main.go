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
	EOF
	MINUS
)

type Token struct {
	tokenType TokenType
	value     int
}

var text string
var pos int
var currentToken *Token

func (t *Token) String() string {
	return strconv.Itoa(t.value)
}

func getNextToken() *Token {
	if pos > len(text)-1 {
		return &Token{
			tokenType: EOF,
			value:     0,
		}
	}
	currentChar := rune(text[pos])
	for unicode.IsSpace(currentChar) {
		pos++
		if pos > len(text)-1 {
			return &Token{
				tokenType: EOF,
				value:     0,
			}
		}
		currentChar = rune(text[pos])
	}
	//fmt.Println("current character", string(currentChar))
	// todo(ryan): is this the best way to do this?
	if unicode.IsDigit(currentChar) {
		// note(ryan): token is an integer so we need to collect all the digits
		digits := []rune{}
		for {
			digits = append(digits, currentChar)
			pos++
			if pos > len(text)-1 {
				break
			}
			currentChar = rune(text[pos])
			if !unicode.IsDigit(currentChar) {
				break
			}
		}
		value, err := strconv.Atoi(string(digits))
		if err != nil {
			panic(err)
		}
		return &Token{
			tokenType: INTEGER,
			value:     value,
		}
	}
	if currentChar == '+' {
		pos++
		return &Token{
			tokenType: PLUS,
		}
	}
	if currentChar == '-' {
		pos++
		return &Token{
			tokenType: MINUS,
		}
	}
	panic(errors.New("invalid token"))
}

func eat(tokenType TokenType) {
	fmt.Println("eating " + strconv.Itoa(int(tokenType)))
	if currentToken.tokenType == tokenType {
		currentToken = getNextToken()
	} else {
		fmt.Println(
			"expected " +
				strconv.Itoa(int(tokenType)) +
				" but got " +
				strconv.Itoa(int(currentToken.tokenType)))
		panic(errors.New("wrong token type"))
	}
}

func expr() int {
	currentToken = getNextToken()
	left := currentToken
	eat(INTEGER)
	eat(PLUS)
	right := currentToken
	eat(INTEGER)
	result := left.value + right.value
	return result
}

func main() {
	var err error
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, err = reader.ReadString('\n')
		text = strings.TrimSpace(text)
		pos = 0
		if err != nil {
			panic(err)
		}
		if len(text) == 0 {
			continue
		}
		fmt.Println(expr())
	}
}

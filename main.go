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

func expr() (int, error) {
	var err error
	result, err := term()
	if err != nil {
		return 0, err
	}
	for currentToken.tokenType == PLUS ||
		currentToken.tokenType == MINUS {
		token := currentToken
		switch token.tokenType {
		case PLUS:
			err = eat(PLUS)
			if err != nil {
				return 0, err
			}
			value, err := term()
			if err != nil {
				return 0, err
			}
			result += value
		case MINUS:
			err = eat(MINUS)
			if err != nil {
				return 0, err
			}
			value, err := term()
			if err != nil {
				return 0, err
			}
			result -= value
		}
	}
	return result, nil
}

func term() (int, error) {
	var err error
	result, err := factor()
	if err != nil {
		return 0, err
	}
	for currentToken.tokenType == DIVIDE ||
		currentToken.tokenType == MULTIPLY {
		token := currentToken
		switch token.tokenType {
		case DIVIDE:
			err = eat(DIVIDE)
			if err != nil {
				return 0, err
			}
			value, err := factor()
			if err != nil {
				return 0, err
			}
			result /= value
		case MULTIPLY:
			err = eat(MULTIPLY)
			if err != nil {
				return 0, err
			}
			value, err := factor()
			if err != nil {
				return 0, err
			}
			result *= value
		}
	}
	return result, nil
}

func factor() (int, error) {
	var err error
	result := 0
	// note(ryan): if we find an LPAREN here then we eat it and recursively call expr.
	// this lets us handle nested expressions with parenthesis. see the grammar for more
	// info.
	if currentToken.tokenType == LPAREN {
		err = eat(LPAREN)
		if err != nil {
			return 0, err
		}
		result, err = expr()
		if err != nil {
			return 0, err
		}
		err = eat(RPAREN)
		if err != nil {
			return 0, err
		}
	} else {
		result = currentToken.value
		err = eat(INTEGER)
		if err != nil {
			return 0, err
		}
	}
	return result, nil
}

func integer() (int, error) {
	digits := []rune{}
	for !eof && unicode.IsDigit(currentCharacter) {
		digits = append(digits, currentCharacter)
		advance()
	}
	value, err := strconv.Atoi(string(digits))
	if err != nil {
		return 0, err
	}
	return value, nil
}

func eat(tokenType TokenType) error {
	var err error
	//fmt.Println("eating " + strconv.Itoa(int(tokenType)))
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
		result, err := expr()
		if err != nil {
			fmt.Println("error: " + err.Error())
			continue
		}
		err = eat(EOF)
		if err != nil {
			fmt.Println("error: " + err.Error())
			continue
		}
		fmt.Println(result)
	}
}

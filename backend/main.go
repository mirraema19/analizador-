package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// TokenType define el tipo de un token (PR, ID, etc.)
type TokenType string

const (
	PR         TokenType = "PR"         // Palabra Reservada
	ID         TokenType = "ID"         // Identificador
	NUMERO     TokenType = "Numeros"    // Número
	SIMBOLO    TokenType = "Simbolos"   // Símbolo
	CADENA     TokenType = "Cadenas"    // Cadena de texto
	ERROR      TokenType = "Error"      // Token no reconocido
	COMENTARIO TokenType = "Comentario" // Comentario
)

// Token representa una unidad léxica
type Token struct {
	Type   TokenType `json:"type"`
	Lexeme string    `json:"lexeme"`
	Line   int       `json:"line"`
}

// AnalysisResult es la estructura que se enviará al frontend
type AnalysisResult struct {
	Tokens         []Token           `json:"tokens"`
	Counts         map[TokenType]int `json:"counts"`
	LexicalErrors  []string          `json:"lexicalErrors"`
	SyntaxErrors   []string          `json:"syntaxErrors"`
	SemanticErrors []string          `json:"semanticErrors"`
}

// Palabras reservadas de Java que consideraremos
var keywords = map[string]bool{
	"public": true, "class": true, "static": true, "void": true, "main": true,
	"String": true, "int": true, "if": true, "else": true, "true": true, "false": true, // <-- CORRECCIÓN AQUÍ
	"System": true, "out": true, "println": true, "equals": true,
}

// --- Analizador Léxico (Tokenizer) ---
func Lexer(code string) []Token {
	var tokens []Token
	lines := strings.Split(code, "\n")

	// Regex mejorada para capturar corchetes individuales y operadores dobles
	re := regexp.MustCompile(`(//.*)|(\"[^\"]*\")|([a-zA-Z_][a-zA-Z0-9_]*)|([0-9]+)|(\[|\]|;|\(|\)|\{|\}|=|<|>|\.|\!)|(\S+)`)

	for i, line := range lines {
		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			lexeme := ""
			for j := 1; j < len(match); j++ {
				if match[j] != "" {
					lexeme = match[j]
					break
				}
			}

			if lexeme == "" {
				continue
			}

			token := Token{Lexeme: lexeme, Line: i + 1}

			if strings.HasPrefix(lexeme, "//") {
				token.Type = COMENTARIO
			} else if strings.HasPrefix(lexeme, "\"") {
				token.Type = CADENA
			} else if _, isKeyword := keywords[lexeme]; isKeyword {
				token.Type = PR
			} else if ok, _ := regexp.MatchString(`^[0-9]+$`, lexeme); ok {
				token.Type = NUMERO
			} else if ok, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, lexeme); ok {
				token.Type = ID
			} else if ok, _ := regexp.MatchString(`^(\[|\]|;|\(|\)|\{|\}|=|<|>|\.|\!)$`, lexeme); ok {
				token.Type = SIMBOLO
			} else {
				token.Type = ERROR
			}
			tokens = append(tokens, token)
		}
	}
	return tokens
}

// --- Analizadores Sintáctico y Semántico (Simplificados) ---
func Analyze(tokens []Token) ([]string, []string) {
	var syntaxErrors []string
	var semanticErrors []string

	symbolTable := make(map[string]string)
	symbolTable["System"] = "Class"
	
	// Pre-declarar 'args' ya no es necesario con la lógica mejorada, pero no hace daño
	// symbolTable["args"] = "String[]" 

	parenBalance := 0
	braceBalance := 0
	for _, t := range tokens {
		if t.Lexeme == "(" { parenBalance++ }
		if t.Lexeme == ")" { parenBalance-- }
		if t.Lexeme == "{" { braceBalance++ }
		if t.Lexeme == "}" { braceBalance-- }
	}
	if parenBalance != 0 {
		syntaxErrors = append(syntaxErrors, "Error de sintaxis: Paréntesis desbalanceados.")
	}
	if braceBalance != 0 {
		syntaxErrors = append(syntaxErrors, "Error de sintaxis: Llaves desbalanceadas.")
	}
	
	// Verificar punto y coma (lógica simplificada para evitar falsos positivos)
	for i := 0; i < len(tokens)-1; i++ {
		// Regla: si un statement termina en un ID, número o ')' y el siguiente token no es ';', '{', '}', ')'
		// es un posible error.
		t := tokens[i]
		if t.Type == ID || t.Type == NUMERO || t.Lexeme == ")" {
			if i+1 < len(tokens) {
				next := tokens[i+1]
				// Si estamos en la misma línea y el siguiente no es un delimitador esperado
				if t.Line == next.Line && !strings.ContainsAny(next.Lexeme, ";){") {
					// Y tampoco es el inicio de una declaración como "int x"
					if !(next.Type == ID && (t.Lexeme == "int" || t.Lexeme == "String")) {
						// syntaxErrors = append(syntaxErrors, fmt.Sprintf("Error de sintaxis en línea %d: Posible falta de un ';' después de '%s'.", t.Line, t.Lexeme))
					}
				}
			}
		}
	}
	
	for i, token := range tokens {
		// Detectar declaración de variable (ej: "int edad" o "String[] args")
		if token.Lexeme == "int" || token.Lexeme == "String" {
			j := i + 1
			// Saltamos los corchetes si es un array (ej: String[] args)
			if j < len(tokens) && tokens[j].Lexeme == "[" {
				j++ // salta '['
				if j < len(tokens) && tokens[j].Lexeme == "]" {
					j++ // salta ']'
				}
			}

			if j < len(tokens) && tokens[j].Type == ID {
				varName := tokens[j].Lexeme
				if _, exists := symbolTable[varName]; exists {
					semanticErrors = append(semanticErrors, fmt.Sprintf("Error semántico en línea %d: La variable '%s' ya ha sido declarada.", tokens[j].Line, varName))
				} else {
					symbolTable[varName] = token.Lexeme
					fmt.Printf("Declarada variable '%s' de tipo '%s'\n", varName, token.Lexeme)
				}
			}
		}

		if token.Type == ID {
			// Ignorar si es un nombre de clase, miembro o parte de una declaración
			isDeclaration := i > 0 && (tokens[i-1].Lexeme == "int" || tokens[i-1].Lexeme == "String" || tokens[i-1].Lexeme == "class")
			// Corrección para `String[] args`: ignorar `args` si viene después de `[]`
			isPostArrayDeclaration := i > 2 && tokens[i-2].Lexeme == "String" && tokens[i-1].Lexeme == "]"
			isMemberAccess := i > 0 && tokens[i-1].Lexeme == "."
			
			if isDeclaration || isPostArrayDeclaration || isMemberAccess {
				continue
			}
			
			if _, exists := symbolTable[token.Lexeme]; !exists {
				if !keywords[token.Lexeme] {
					semanticErrors = append(semanticErrors, fmt.Sprintf("Error semántico en línea %d: La variable o identificador '%s' no ha sido declarado.", token.Line, token.Lexeme))
				}
			}
		}
	}

	return syntaxErrors, semanticErrors
}

// --- Servidor HTTP ---
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var requestBody struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens := Lexer(requestBody.Code)
	syntaxErrors, semanticErrors := Analyze(tokens)

	var lexicalErrors []string
	displayTokens := []Token{}
	counts := make(map[TokenType]int)

	for _, t := range tokens {
		if t.Type == ERROR {
			lexicalErrors = append(lexicalErrors, fmt.Sprintf("Error léxico en línea %d: Símbolo no reconocido '%s'.", t.Line, t.Lexeme))
		}
		if t.Type != COMENTARIO && t.Type != ERROR {
			counts[t.Type]++
			displayTokens = append(displayTokens, t)
		}
	}

	result := AnalysisResult{
		Tokens:         displayTokens,
		Counts:         counts,
		LexicalErrors:  lexicalErrors,
		SyntaxErrors:   syntaxErrors,
		SemanticErrors: semanticErrors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	fmt.Println("Servidor Go escuchando en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
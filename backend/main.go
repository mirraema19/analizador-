// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode"
)

// --- 1. Definición de Tipos y Estructuras (Sin cambios) ---
type TokenType int
const (TOKEN_UNKNOWN TokenType = iota; TOKEN_COMMAND; TOKEN_FLAG; TOKEN_PARAM)
func (t TokenType) String() string { return [...]string{"UNKNOWN", "COMMAND", "FLAG", "PARAM"}[t] }
type Token struct {Type  TokenType `json:"type"`; Value string `json:"value"`}
type AnalysisRequest struct {Command string `json:"command"`}
type AnalysisResult struct {Status string `json:"status"`; ErrorType string `json:"errorType,omitempty"`; Message string `json:"message"`; Tokens []Token `json:"tokens"`}


// --- 2. Analizador Léxico (Lexer) (Sin cambios) ---
func Lexer(input string) ([]Token, error) {
	var tokens []Token
	trimmedInput := strings.TrimSpace(input)
	if trimmedInput == "" {return nil, fmt.Errorf("El comando está vacío.")}
	parts := splitCommand(trimmedInput)
	if len(parts) == 0 || parts[0] != "git" {
		errorToken := Token{Type: TOKEN_UNKNOWN, Value: parts[0]}
		return []Token{errorToken}, fmt.Errorf("Error Léxico: Se esperaba el comando 'git', pero se encontró '%s'.", parts[0])
	}
	knownCommands := map[string]bool{"init": true, "config": true, "clone": true, "status": true, "add": true, "reset": true, "commit": true, "branch": true, "checkout": true, "merge": true, "push": true}
	for i, part := range parts {
		var tokenType TokenType
		if i == 0 {tokenType = TOKEN_COMMAND
		} else if i == 1 && knownCommands[part] {tokenType = TOKEN_COMMAND
		} else if strings.HasPrefix(part, "-") {tokenType = TOKEN_FLAG
		} else {tokenType = TOKEN_PARAM}
		tokens = append(tokens, Token{Type: tokenType, Value: part})
	}
	return tokens, nil
}
func splitCommand(input string) []string {
	var result []string; var current strings.Builder; inQuote := false
	for _, r := range input {
		if r == '"' { inQuote = !inQuote }
		if unicode.IsSpace(r) && !inQuote {if current.Len() > 0 { result = append(result, current.String()); current.Reset() }
		} else { current.WriteRune(r) }
	}
	if current.Len() > 0 { result = append(result, current.String()) }
	for i, part := range result {
		if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {result[i] = strings.Trim(part, "\"")}
	}
	return result
}


// --- 3. Analizador Sintáctico y Semántico (CON LAS NUEVAS REGLAS) ---
func ParseAndAnalyze(tokens []Token) AnalysisResult {
	if len(tokens) < 2 {return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Comando 'git' incompleto.", Tokens: tokens}}
	command := tokens[1].Value
	
	switch command {
	// ... (casos de init, clone, status se mantienen igual)
	case "init": if len(tokens) > 2 {return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: `git init` no admite parámetros adicionales.", Tokens: tokens}}; return AnalysisResult{Status: "Correcto", Message: "Comando `git init` válido.", Tokens: tokens}
	case "clone": if len(tokens) != 3 || tokens[2].Type != TOKEN_PARAM {return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: `git clone` requiere exactamente una URL de repositorio.", Tokens: tokens}}; url := tokens[2].Value; if !strings.HasPrefix(url, "http") && !strings.HasPrefix(url, "git@") {return AnalysisResult{Status: "Advertencia", ErrorType: "Semántico", Message: "Advertencia Semántica: El parámetro no parece una URL válida.", Tokens: tokens}}; return AnalysisResult{Status: "Correcto", Message: "Comando `git clone` válido.", Tokens: tokens}
	case "status": if len(tokens) > 2 {return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: `git status` no admite parámetros.", Tokens: tokens}}; return AnalysisResult{Status: "Correcto", Message: "Comando `git status` válido.", Tokens: tokens}

	case "config":
		if len(tokens) == 3 && tokens[2].Value == "--list" {return AnalysisResult{Status: "Correcto", Message: "Comando `git config --list` válido.", Tokens: tokens}}
		if len(tokens) == 4 && tokens[2].Value == "--global" && (tokens[3].Value == "user.name" || tokens[3].Value == "user.email") {
			return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Falta el valor para la configuración.", Tokens: tokens}
		}
		if len(tokens) == 5 && tokens[2].Value == "--global" && (tokens[3].Value == "user.name" || tokens[3].Value == "user.email") && tokens[4].Type == TOKEN_PARAM {
			if tokens[4].Value == "" {return AnalysisResult{Status: "Advertencia", ErrorType: "Semántico", Message: "Advertencia Semántica: El valor para la configuración no debe estar vacío.", Tokens: tokens}}
			return AnalysisResult{Status: "Correcto", Message: "Comando `git config` válido.", Tokens: tokens}
		}
		return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Uso incorrecto de `git config`.", Tokens: tokens}

	case "add":
		if len(tokens) < 3 {return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Falta el archivo a añadir.", Tokens: tokens}}
		if tokens[2].Type != TOKEN_PARAM {return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: `git add` requiere un parámetro (ej: '.', 'archivo.txt').", Tokens: tokens}}
		if tokens[2].Value == "." {return AnalysisResult{Status: "Advertencia", ErrorType: "Semántico", Message: "Advertencia Semántica: `git add .` puede incluir archivos no deseados. Se recomienda revisar con `git status` primero.", Tokens: tokens}}
		return AnalysisResult{Status: "Correcto", Message: "Comando `git add` válido.", Tokens: tokens}

	case "reset":
		if len(tokens) == 3 {
			if tokens[2].Type == TOKEN_PARAM {return AnalysisResult{Status: "Correcto", Message: "Comando `git reset <archivo>` válido para quitarlo del staging.", Tokens: tokens}}
			if tokens[2].Value == "--hard" {return AnalysisResult{Status: "Advertencia", ErrorType: "Semántico", Message: "Advertencia Semántica: `git reset --hard` borra los cambios locales sin recuperación. Es una operación muy destructiva.", Tokens: tokens}}
		}
		if len(tokens) == 4 && tokens[2].Value == "--soft" && tokens[3].Type == TOKEN_PARAM {return AnalysisResult{Status: "Correcto", Message: "Comando `git reset --soft` válido.", Tokens: tokens}}
		return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Uso no reconocido de `git reset`.", Tokens: tokens}

	case "commit":
		// El caso de `git commit "sin -m"` se detecta aquí
		if len(tokens) == 3 && tokens[2].Type == TOKEN_PARAM {
			return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Falta el flag -m.", Tokens: tokens}
		}
		if len(tokens) == 3 && tokens[2].Value == "--amend" {return AnalysisResult{Status: "Correcto", Message: "Comando `git commit --amend` válido.", Tokens: tokens}}
		if len(tokens) == 4 && (tokens[2].Value == "-m" || tokens[2].Value == "-am") && tokens[3].Type == TOKEN_PARAM {
			if tokens[3].Value == "" {return AnalysisResult{Status: "Advertencia", ErrorType: "Semántico", Message: "Advertencia Semántica: Realizar un commit sin un mensaje descriptivo es una mala práctica.", Tokens: tokens}}
			return AnalysisResult{Status: "Correcto", Message: "Comando `git commit` válido.", Tokens: tokens}
		}
		return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Uso incorrecto de `git commit`.", Tokens: tokens}
	
	case "push":
		// Error sintáctico: git push origin (falta la rama)
		if len(tokens) == 3 && tokens[2].Type == TOKEN_PARAM {
			return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Falta la rama a subir (ej: git push origin main).", Tokens: tokens}
		}
		// Formas sintácticamente válidas
		isSyntacticallyCorrect := (len(tokens) == 2) || (len(tokens) == 4 && tokens[2].Type == TOKEN_PARAM && tokens[3].Type == TOKEN_PARAM) || (len(tokens) == 3 && tokens[2].Type == TOKEN_FLAG) || (len(tokens) == 5 && tokens[3].Type == TOKEN_FLAG)
		if !isSyntacticallyCorrect {
			return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Estructura de `git push` no reconocida.", Tokens: tokens}
		}
		// Si es sintácticamente correcto, buscar advertencias semánticas
		for _, t := range tokens {
			if t.Value == "--force" {
				return AnalysisResult{Status: "Advertencia", ErrorType: "Semántico", Message: "Advertencia Semántica: Forzar el push puede sobrescribir cambios remotos. Úsalo con extrema precaución.", Tokens: tokens}
			}
		}
		return AnalysisResult{Status: "Correcto", Message: "Comando `git push` válido.", Tokens: tokens}

	case "branch":
		if len(tokens) == 2 {return AnalysisResult{Status: "Correcto", Message: "Comando `git branch` (listar) válido.", Tokens: tokens}}
		if len(tokens) == 3 && tokens[2].Type == TOKEN_PARAM {return AnalysisResult{Status: "Correcto", Message: "Comando `git branch <nombre-rama>` (crear) válido.", Tokens: tokens}}
		if len(tokens) == 4 && tokens[2].Value == "-d" && tokens[3].Type == TOKEN_PARAM {
			if tokens[3].Value == "main" || tokens[3].Value == "master" {
				return AnalysisResult{Status: "Advertencia", ErrorType: "Semántico", Message: "Advertencia Semántica: Borrar la rama principal puede afectar el proyecto.", Tokens: tokens}
			}
			return AnalysisResult{Status: "Correcto", Message: "Comando `git branch -d` (eliminar) válido.", Tokens: tokens}
		}
		return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Uso incorrecto de `git branch`.", Tokens: tokens}

	case "checkout":
		if len(tokens) == 2 {
			return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Falta la rama a la que se quiere cambiar.", Tokens: tokens}
		}
		if len(tokens) == 3 && tokens[2].Type == TOKEN_PARAM {return AnalysisResult{Status: "Correcto", Message: "Comando `git checkout <rama>` (cambiar) válido.", Tokens: tokens}}
		if len(tokens) == 4 && tokens[2].Value == "-b" && tokens[3].Type == TOKEN_PARAM {return AnalysisResult{Status: "Correcto", Message: "Comando `git checkout -b` (crear y cambiar) válido.", Tokens: tokens}}
		return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: Uso incorrecto de `git checkout`.", Tokens: tokens}

	case "merge":
		if len(tokens) != 3 || tokens[2].Type != TOKEN_PARAM {return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: "Error de Sintaxis: `git merge` requiere el nombre de una rama para unir.", Tokens: tokens}}
		return AnalysisResult{Status: "Correcto", Message: "Comando `git merge` válido.", Tokens: tokens}
		
	default:
		return AnalysisResult{Status: "Error", ErrorType: "Sintáctico", Message: fmt.Sprintf("Error de Sintaxis: Comando 'git %s' no es reconocido por este analizador.", command), Tokens: tokens}
	}
}


// --- 4. Servidor HTTP y Manejadores (Sin cambios) ---
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {http.Error(w, "Método no permitido", http.StatusMethodNotAllowed); return}
	var req AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {http.Error(w, "Cuerpo de la petición inválido", http.StatusBadRequest); return}
	tokens, err := Lexer(req.Command)
	if err != nil {
		result := AnalysisResult{Status: "Error", ErrorType: "Léxico", Message: err.Error(), Tokens: tokens}
		w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(result); return
	}
	result := ParseAndAnalyze(tokens)
	w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(result)
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*"); w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS"); w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {w.WriteHeader(http.StatusOK); return}; next.ServeHTTP(w, r)
	})
}
func main() {
	mux := http.NewServeMux(); finalHandler := http.HandlerFunc(analyzeHandler); mux.Handle("/analyze", corsMiddleware(finalHandler)); port := "8080"
	log.Printf("Servidor Go con reglas avanzadas escuchando en el puerto %s", port); log.Println("Endpoint disponible en: POST http://localhost:8080/analyze")
	if err := http.ListenAndServe(":"+port, mux); err != nil {log.Fatalf("No se pudo iniciar el servidor: %s\n", err)}
}
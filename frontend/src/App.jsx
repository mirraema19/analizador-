import { useState } from 'react';
import './style.css';

// Código inicial para probar el analizador (con errores de cada tipo)
const initialCode = `public class Main { // 'publi' seria error semantico, no se reconoce ID
    public static void main(String[] args) { // Sintaxis correcta
        String escuela = "upchiapas";
        int edad = 20;

        // Error Semántico: 'ed' no está declarado
        if (ed > 18) {
            System.out.println("Es mayor de edad.");
        }

        // Error Sintáctico: falta ')' en la condición del if
        if (escuela.equals("upchiapas") {
            System.out.println("Estudia en la UPChiapas.");
        }

        // Error Sintáctico: falta ';'
        int x = 5

        // Error Léxico: '@' no es un símbolo válido
        int y = 10 @ 20;
    }
}`;

// Código sin errores para verificar que no se muestren alertas
const correctCode = `public class Main {
    public static void main(String[] args) {
        String escuela = "upchiapas";
        int edad = 20;

        if (edad >= 18) {
            System.out.println("Es mayor de edad.");
        }

        if (escuela.equals("upchiapas")) {
            System.out.println("Estudia en la UPChiapas.");
        }

        int x = 5;
        int y = 10;
    }
}`;

function App() {
  const [code, setCode] = useState(correctCode); // Empezamos con el código correcto
  const [result, setResult] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleAnalyze = async () => {
    setIsLoading(true);
    setError('');
    setResult(null);

    try {
      const response = await fetch('http://localhost:8080/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`La respuesta del servidor no fue OK: ${response.status} ${errorText}`);
      }
      
      const data = await response.json();
      setResult(data);

    } catch (err) {
      setError('Error al conectar con el backend. Asegúrate de que está en ejecución en el puerto 8080.');
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };
  
  const renderTable = () => {
    if (!result || !result.tokens || result.tokens.length === 0) return null;

    const { tokens, counts } = result;
    // Las claves deben coincidir con los TokenType de Go
    const headers = ["Tokens", "PR", "ID", "Numeros", "Simbolos", "Cadenas"];
    const totalRow = headers.slice(1).map(headerKey => counts[headerKey] || 0);

    return (
      <div className="card">
        <h2>Analizador Léxico</h2>
        <div className="table-wrapper">
          <table>
            <thead>
              <tr>{headers.map(h => <th key={h}>{h}</th>)}</tr>
            </thead>
            <tbody>
              {tokens.map((token, index) => (
                <tr key={index}>
                  <td className="token-lexeme">{token.lexeme}</td>
                  {headers.slice(1).map(headerKey => (
                    <td key={`${index}-${headerKey}`}>{token.type === headerKey ? 'x' : ''}</td>
                  ))}
                </tr>
              ))}
              <tr className="total-row">
                <td><strong>Total</strong></td>
                {totalRow.map((count, index) => (
                  <td key={`total-${index}`}><strong>{count}</strong></td>
                ))}
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    );
  };

  const renderErrors = () => {
    if (!result) return null;
    
    // Ahora 'lexicalErrors' viene del backend
    const { lexicalErrors, syntaxErrors, semanticErrors } = result;
    const hasErrors = (lexicalErrors?.length > 0) ||
                      (syntaxErrors?.length > 0) ||
                      (semanticErrors?.length > 0);

    if (!hasErrors) {
      return (
        <div className="card success-card">
          <h2>Análisis Completado</h2>
          <p>✅ ¡Felicidades! No se encontraron errores léxicos, sintácticos o semánticos.</p>
        </div>
      );
    }

    return (
      <div className="card errors-card">
        <h2>Análisis de Errores</h2>
        {lexicalErrors?.length > 0 && (
          <div className="error-section">
            <h3>Errores Léxicos</h3>
            <ul className="error-list">{lexicalErrors.map((err, i) => <li key={`lex-${i}`}>{err}</li>)}</ul>
          </div>
        )}
        {syntaxErrors?.length > 0 && (
          <div className="error-section">
            <h3>Errores Sintácticos</h3>
            <ul className="error-list">{syntaxErrors.map((err, i) => <li key={`syn-${i}`}>{err}</li>)}</ul>
          </div>
        )}
        {semanticErrors?.length > 0 && (
          <div className="error-section">
            <h3>Errores Semánticos</h3>
            <ul className="error-list">{semanticErrors.map((err, i) => <li key={`sem-${i}`}>{err}</li>)}</ul>
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>Analizador (Go + React)</h1>
        <p>Un analizador léxico, sintáctico y semántico para un subconjunto de Java.</p>
      </header>
      <main className="app-main">
        <div className="editor-panel card">
          <h2>Código Fuente (Java)</h2>
          <div className="code-selectors">
            <button onClick={() => setCode(correctCode)}>Cargar Código Correcto</button>
            <button onClick={() => setCode(initialCode)}>Cargar Código con Errores</button>
          </div>
          <textarea
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder="Escribe tu código Java aquí..."
            spellCheck="false"
          />
          <button onClick={handleAnalyze} disabled={isLoading}>
            {isLoading ? 'Analizando...' : 'Analizar Código'}
          </button>
          {error && <p className="api-error">{error}</p>}
        </div>
        <div className="output-panels">
          {result && (
            <>
              <div className="results-panel">{renderTable()}</div>
              <div className="errors-panel">{renderErrors()}</div>
            </>
          )}
        </div>
      </main>
    </div>
  );
}

export default App;
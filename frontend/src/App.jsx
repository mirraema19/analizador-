// src/App.jsx
import React, { useState } from 'react';
import './style.css';

const API_URL = 'http://localhost:8080/analyze';

function App() {
  const [command, setCommand] = useState('git commit "falta el flag -m"');
  const [analysisResult, setAnalysisResult] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setIsLoading(true);
    setError(null);
    setAnalysisResult(null);

    try {
      const response = await fetch(API_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: command.trim() }),
      });

      if (!response.ok) {
        throw new Error(`Error del servidor: ${response.statusText}`);
      }

      const data = await response.json();
      setAnalysisResult(data);

    } catch (err) {
      console.error("Error al conectar con el backend:", err);
      setError('No se pudo conectar al servidor. Asegúrate de que el backend en Go esté funcionando.');
    } finally {
      setIsLoading(false);
    }
  };
  
  const getTokenTypeClass = (tokenType) => {
    switch (tokenType) {
      case 1: return 'token-command';
      case 2: return 'token-flag';
      case 3: return 'token-param';
      default: return 'token-unknown';
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'Correcto': return '✅';
      case 'Advertencia': return '⚠️';
      case 'Error': return '❌';
      default: return '';
    }
  };

  return (
    <div className="app-container">
      <header>
        <h1>Analizador de Comandos Git</h1>
        <p>Escribe un comando de Git para analizar su sintaxis y buenas prácticas.</p>
      </header>

      <main>
        <form onSubmit={handleSubmit}>
          <textarea
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            placeholder="Ej: git commit -m 'feat: nuevo componente'"
            rows="3"
            disabled={isLoading}
          />
          <button type="submit" disabled={isLoading}>
            {isLoading ? 'Analizando...' : 'Analizar Comando'}
          </button>
        </form>

        {error && <div className="network-error">{error}</div>}

        {analysisResult && (
          <div className={`result-box ${analysisResult.status.toLowerCase()}`}>
            <h3>
              {getStatusIcon(analysisResult.status)} {analysisResult.status}
              {/* AQUÍ ESTÁ EL CAMBIO: Muestra la etiqueta del tipo de error si existe */}
              {analysisResult.errorType && (
                <span className="error-type-badge">
                  {analysisResult.errorType}
                </span>
              )}
            </h3>
            <p className="result-message">{analysisResult.message}</p>
            
            {analysisResult.tokens && analysisResult.tokens.length > 0 && (
              <div className="token-view">
                <h4>Desglose del Comando (Análisis Léxico):</h4>
                <div className="tokenized-command">
                  {analysisResult.tokens.map((token, index) => (
                    <span key={index} className={`token ${getTokenTypeClass(token.type)}`}>
                      {token.value}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </main>

      <footer>
        <p>Proyecto con Backend en Go y Frontend en React.jsx</p>
      </footer>
    </div>
  );
}

export default App;
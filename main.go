package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var galaxy map[Symbol]Expr

func init() {
	program, err := parseProgram("./galaxy.txt")
	if err != nil {
		log.Fatalf("failed to parse program: %v", err)
	}
	galaxy = program
}

type InteractRequest struct {
	Expression string `json:"expression"`
}

type InteractResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

func interactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(InteractResponse{Error: "Method not allowed"})
		return
	}

	var req InteractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(InteractResponse{Error: "Invalid JSON"})
		return
	}

	// Parse and evaluate the expression
	if strings.TrimSpace(req.Expression) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(InteractResponse{Error: "Invalid expression"})
		return
	}

	tokens := strings.Split(strings.TrimSpace(req.Expression), " ")
	expr, remaining := parseExpr(tokens)
	if expr == nil || len(remaining) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(InteractResponse{Error: "Invalid expression"})
		return
	}

	// Evaluate the expression
	result := eval(expr, galaxy)

	// Try to convert to a value, handle panics for unsupported expressions
	var value interface{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				// If toValue panics, return the string representation
				value = printExpr(result)
			}
		}()
		value = toValue(result)
	}()

	json.NewEncoder(w).Encode(InteractResponse{Result: value})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Galaxy Interpreter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        textarea { width: 100%; height: 100px; margin: 10px 0; }
        button { padding: 10px 20px; font-size: 16px; }
        .result { margin-top: 20px; padding: 10px; background: #f0f0f0; border-radius: 5px; }
        .error { background: #ffe6e6; color: #cc0000; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Galaxy Interpreter</h1>
        <p>Enter a galaxy expression to evaluate:</p>
        <textarea id="expression" placeholder="Example: ap add 1 2"></textarea>
        <br>
        <button onclick="evaluateExpression()">Evaluate</button>
        <div id="result" class="result" style="display: none;"></div>
    </div>

    <script>
        async function evaluateExpression() {
            const expression = document.getElementById('expression').value.trim();
            const resultDiv = document.getElementById('result');
            
            if (!expression) {
                showResult('Please enter an expression', true);
                return;
            }

            try {
                const response = await fetch('/interact', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ expression: expression })
                });

                const data = await response.json();
                
                if (data.error) {
                    showResult('Error: ' + data.error, true);
                } else {
                    showResult('Result: ' + JSON.stringify(data.result, null, 2), false);
                }
            } catch (error) {
                showResult('Network error: ' + error.message, true);
            }
        }

        function showResult(message, isError) {
            const resultDiv = document.getElementById('result');
            resultDiv.textContent = message;
            resultDiv.className = 'result' + (isError ? ' error' : '');
            resultDiv.style.display = 'block';
        }

        // Allow Enter key to evaluate
        document.getElementById('expression').addEventListener('keypress', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                evaluateExpression();
            }
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/interact", interactHandler)

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

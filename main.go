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

type EvalRequest struct {
	Expression string `json:"expression"`
}

type EvalResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

type InteractRequest struct {
	State string `json:"state"`
	Point struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"point"`
}

type InteractResponse struct {
	NewState string        `json:"newstate"`
	Images   [][]PointPair `json:"images"`
	Error    string        `json:"error,omitempty"`
}

type PointPair struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func evalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(EvalResponse{Error: "Method not allowed"})
		return
	}

	var req EvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(EvalResponse{Error: "Invalid JSON"})
		return
	}

	// Parse and evaluate the expression
	if strings.TrimSpace(req.Expression) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(EvalResponse{Error: "Invalid expression"})
		return
	}

	tokens := strings.Split(strings.TrimSpace(req.Expression), " ")
	expr, remaining := parseExpr(tokens)
	if expr == nil || len(remaining) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(EvalResponse{Error: "Invalid expression"})
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

	json.NewEncoder(w).Encode(EvalResponse{Result: value})
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

	// Parse the state expression
	if strings.TrimSpace(req.State) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(InteractResponse{Error: "Invalid state"})
		return
	}

	stateTokens := strings.Split(strings.TrimSpace(req.State), " ")
	stateExpr, remaining := parseExpr(stateTokens)
	if stateExpr == nil || len(remaining) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(InteractResponse{Error: "Invalid state expression"})
		return
	}

	// Create point expression: ap ap cons x y
	pointExpr := &Ap{
		Left: &Ap{
			Left:  Symbol("cons"),
			Right: Number(req.Point.X),
		},
		Right: Number(req.Point.Y),
	}

	// Call interact function with state and point: ap ap interact state point
	// Check if interact function exists in galaxy
	if _, exists := galaxy[Symbol("interact")]; !exists {
		// If interact function doesn't exist, return a default response
		json.NewEncoder(w).Encode(InteractResponse{
			NewState: req.State,       // Return the same state
			Images:   [][]PointPair{}, // Empty images
		})
		return
	}

	interactExpr := &Ap{
		Left: &Ap{
			Left:  Symbol("interact"),
			Right: stateExpr,
		},
		Right: pointExpr,
	}

	// Evaluate the interaction
	result := eval(interactExpr, galaxy)

	// Try to convert the result to the expected format
	var newState string
	var images [][]PointPair
	var responseSent bool

	func() {
		defer func() {
			if r := recover(); r != nil {
				// If parsing fails, return error
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(InteractResponse{Error: "Failed to process interaction result"})
				responseSent = true
			}
		}()

		// Parse the result which should be a cons of (newState, images)
		resultValue := toValue(result)
		if resultSlice, ok := resultValue.([]interface{}); ok && len(resultSlice) == 2 {
			// First element is the new state
			if stateResult, ok := resultSlice[0].([]interface{}); ok {
				// Convert state back to expression string
				stateExprResult := valueToExpr(stateResult)
				newState = printExpr(stateExprResult)
			} else {
				newState = fmt.Sprintf("%v", resultSlice[0])
			}

			// Second element is the images
			if imagesResult, ok := resultSlice[1].([]interface{}); ok {
				images = parseImages(imagesResult)
			}
		} else {
			// If result doesn't match expected format, return string representation
			newState = printExpr(result)
			images = [][]PointPair{}
		}
	}()

	if !responseSent {
		json.NewEncoder(w).Encode(InteractResponse{
			NewState: newState,
			Images:   images,
		})
	}
}

// Helper function to convert a value back to an expression
func valueToExpr(value interface{}) Expr {
	switch v := value.(type) {
	case int64:
		return Number(v)
	case float64:
		return Number(int64(v))
	case string:
		return Symbol(v)
	case []interface{}:
		if len(v) == 2 {
			return &Ap{Left: valueToExpr(v[0]), Right: valueToExpr(v[1])}
		}
	}
	return Symbol("nil")
}

// Helper function to parse images from the result
func parseImages(imagesValue []interface{}) [][]PointPair {
	var images [][]PointPair

	for _, imageValue := range imagesValue {
		if imageSlice, ok := imageValue.([]interface{}); ok {
			var points []PointPair
			for _, pointValue := range imageSlice {
				if pointStruct, ok := pointValue.(struct{ Left, Right interface{} }); ok {
					if x, ok := pointStruct.Left.(int64); ok {
						if y, ok := pointStruct.Right.(int64); ok {
							points = append(points, PointPair{X: int(x), Y: int(y)})
						}
					}
				}
			}
			images = append(images, points)
		}
	}

	return images
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
                const response = await fetch('/eval', {
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
	http.HandleFunc("/eval", evalHandler)
	http.HandleFunc("/interact", interactHandler)

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

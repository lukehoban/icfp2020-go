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

	// Call galaxy function with state and point: ap ap galaxy state point
	// Check if galaxy function exists
	if _, exists := galaxy[Symbol("galaxy")]; !exists {
		// If galaxy function doesn't exist, return a default response
		json.NewEncoder(w).Encode(InteractResponse{
			NewState: req.State,       // Return the same state
			Images:   [][]PointPair{}, // Empty images
		})
		return
	}

	interactExpr := &Ap{
		Left: &Ap{
			Left:  Symbol("galaxy"),
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

		// Parse the result which should be a galaxy response: [flag, newState, images]
		resultValue := toValue(result)
		if resultSlice, ok := resultValue.([]interface{}); ok && len(resultSlice) == 3 {
			// First element is a flag (usually 0)
			// Second element is the new state
			if stateResult := resultSlice[1]; stateResult != nil {
				// Convert state back to expression string
				if stateSlice, ok := stateResult.([]interface{}); ok {
					stateExprResult := valueToExpr(stateSlice)
					newState = printExpr(stateExprResult)
				} else {
					// For simple values like numbers or symbols
					if stateExpr := valueToExpr(stateResult); stateExpr != nil {
						newState = printExpr(stateExpr)
					} else {
						newState = fmt.Sprintf("%v", stateResult)
					}
				}
			} else {
				newState = "nil"
			}

			// Third element is the images
			if imagesResult, ok := resultSlice[2].([]interface{}); ok {
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
		if len(v) == 0 {
			return Symbol("nil")
		} else if len(v) == 2 {
			return &Ap{Left: valueToExpr(v[0]), Right: valueToExpr(v[1])}
		} else {
			// Convert multi-element array to nested cons structure
			// [a, b, c, d] becomes cons(a, cons(b, cons(c, d)))
			result := valueToExpr(v[len(v)-1]) // Start with the last element
			for i := len(v) - 2; i >= 0; i-- {
				result = &Ap{
					Left: &Ap{
						Left:  Symbol("cons"),
						Right: valueToExpr(v[i]),
					},
					Right: result,
				}
			}
			return result
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
        .container { max-width: 1200px; margin: 0 auto; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        .section h2 { margin-top: 0; }
        textarea { width: 100%; height: 100px; margin: 10px 0; }
        input[type="number"] { width: 80px; margin: 5px; }
        button { padding: 10px 20px; font-size: 16px; margin: 5px; }
        .result { margin-top: 20px; padding: 10px; background: #f0f0f0; border-radius: 5px; }
        .error { background: #ffe6e6; color: #cc0000; }
        .canvas-container { margin: 20px 0; text-align: center; }
        canvas { border: 2px solid #333; background: white; }
        .controls { margin: 10px 0; }
        .state-display { background: #f8f8f8; padding: 10px; border-radius: 5px; font-family: monospace; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Galaxy Interpreter</h1>
        
        <div class="section">
            <h2>Expression Evaluator</h2>
            <p>Enter a galaxy expression to evaluate:</p>
            <textarea id="expression" placeholder="Example: ap ap add 1 2"></textarea>
            <br>
            <button onclick="evaluateExpression()">Evaluate</button>
            <div id="evalResult" class="result" style="display: none;"></div>
        </div>

        <div class="section">
            <h2>Galaxy Interaction</h2>
            <p>Interact with the galaxy by providing a state and clicking coordinates:</p>
            
            <div class="controls">
                <label>State: <textarea id="state" placeholder="nil">nil</textarea></label>
                <br>
                <label>X: <input type="number" id="pointX" value="0"></label>
                <label>Y: <input type="number" id="pointY" value="0"></label>
                <button onclick="interact()">Interact</button>
                <button onclick="resetState()">Reset State</button>
            </div>

            <div class="canvas-container">
                <canvas id="galaxyCanvas" width="400" height="400"></canvas>
                <p><em>Click on the canvas to interact with those coordinates</em></p>
            </div>

            <div>
                <h3>Current State:</h3>
                <div id="stateDisplay" class="state-display">nil</div>
            </div>
            
            <div id="interactResult" class="result" style="display: none;"></div>
        </div>
    </div>

    <script>
        const canvas = document.getElementById('galaxyCanvas');
        const ctx = canvas.getContext('2d');
        
        // Colors for different image layers
        const colors = [
            '#FF0000', '#00FF00', '#0000FF', '#FFFF00', '#FF00FF', '#00FFFF',
            '#800080', '#FFA500', '#008000', '#000080', '#800000', '#808000'
        ];

        let currentState = 'nil';
        let canvasScale = 1;
        let canvasOffsetX = 0;
        let canvasOffsetY = 0;

        // Expression evaluator (existing functionality)
        async function evaluateExpression() {
            const expression = document.getElementById('expression').value.trim();
            const resultDiv = document.getElementById('evalResult');
            
            if (!expression) {
                showEvalResult('Please enter an expression', true);
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
                    showEvalResult('Error: ' + data.error, true);
                } else {
                    showEvalResult('Result: ' + JSON.stringify(data.result, null, 2), false);
                }
            } catch (error) {
                showEvalResult('Network error: ' + error.message, true);
            }
        }

        function showEvalResult(message, isError) {
            const resultDiv = document.getElementById('evalResult');
            resultDiv.textContent = message;
            resultDiv.className = 'result' + (isError ? ' error' : '');
            resultDiv.style.display = 'block';
        }

        // Galaxy interaction functionality
        async function interact() {
            const state = document.getElementById('state').value.trim() || 'nil';
            const x = parseInt(document.getElementById('pointX').value) || 0;
            const y = parseInt(document.getElementById('pointY').value) || 0;
            
            await performInteraction(state, x, y);
        }

        async function performInteraction(state, x, y) {
            const resultDiv = document.getElementById('interactResult');
            
            try {
                const response = await fetch('/interact', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        state: state,
                        point: { x: x, y: y }
                    })
                });

                const data = await response.json();
                
                if (data.error) {
                    showInteractResult('Error: ' + data.error, true);
                } else {
                    currentState = data.newstate;
                    document.getElementById('state').value = currentState;
                    document.getElementById('stateDisplay').textContent = currentState;
                    
                    renderImages(data.images);
                    
                    showInteractResult('Interaction successful. Images: ' + data.images.length + ' layers', false);
                }
            } catch (error) {
                showInteractResult('Network error: ' + error.message, true);
            }
        }

        function showInteractResult(message, isError) {
            const resultDiv = document.getElementById('interactResult');
            resultDiv.textContent = message;
            resultDiv.className = 'result' + (isError ? ' error' : '');
            resultDiv.style.display = 'block';
        }

        function renderImages(images) {
            if (!images || images.length === 0) {
                clearCanvas();
                return;
            }

            // Calculate bounding box for all points
            let minX = Infinity, maxX = -Infinity;
            let minY = Infinity, maxY = -Infinity;

            for (const image of images) {
			    if (image == null) continue;
                for (const point of image) {
                    minX = Math.min(minX, point.x);
                    maxX = Math.max(maxX, point.x);
                    minY = Math.min(minY, point.y);
                    maxY = Math.max(maxY, point.y);
                }
            }

            // Add padding
            const padding = 20;
            const width = Math.max(maxX - minX + 2 * padding, 100);
            const height = Math.max(maxY - minY + 2 * padding, 100);

            // Calculate scale to fit canvas
            const maxCanvasSize = 400;
            canvasScale = Math.min(maxCanvasSize / width, maxCanvasSize / height);
            
            // Update canvas size
            canvas.width = Math.ceil(width * canvasScale);
            canvas.height = Math.ceil(height * canvasScale);
            
            // Calculate offset to center the content
            canvasOffsetX = (minX - padding) * canvasScale;
            canvasOffsetY = (minY - padding) * canvasScale;

            // Clear canvas
            clearCanvas();

            // Render each image layer with different colors
            images.forEach((image, layerIndex) => {
                if (!image) return; // Skip null/undefined images
                
                const color = colors[layerIndex % colors.length];
                ctx.fillStyle = color;
                ctx.strokeStyle = color;
                
                // Draw points
                image.forEach(point => {
                    const screenX = point.x * canvasScale - canvasOffsetX;
                    const screenY = point.y * canvasScale - canvasOffsetY;
                    
                    // Draw a small circle for each point
                    ctx.beginPath();
                    ctx.arc(screenX, screenY, 2, 0, 2 * Math.PI);
                    ctx.fill();
                });
            });
        }

        function clearCanvas() {
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            
            // Draw grid for reference
            ctx.strokeStyle = '#f0f0f0';
            ctx.lineWidth = 1;
            
            const gridSize = 20 * canvasScale;
            if (gridSize > 5) {
                for (let x = 0; x < canvas.width; x += gridSize) {
                    ctx.beginPath();
                    ctx.moveTo(x, 0);
                    ctx.lineTo(x, canvas.height);
                    ctx.stroke();
                }
                for (let y = 0; y < canvas.height; y += gridSize) {
                    ctx.beginPath();
                    ctx.moveTo(0, y);
                    ctx.lineTo(canvas.width, y);
                    ctx.stroke();
                }
            }
        }

        function resetState() {
            currentState = 'nil';
            document.getElementById('state').value = 'nil';
            document.getElementById('stateDisplay').textContent = 'nil';
            clearCanvas();
            document.getElementById('interactResult').style.display = 'none';
        }

        // Canvas click handler
        canvas.addEventListener('click', function(event) {
            const rect = canvas.getBoundingClientRect();
            const clickX = event.clientX - rect.left;
            const clickY = event.clientY - rect.top;
            
            // Convert screen coordinates to galaxy coordinates
            const galaxyX = Math.round((clickX + canvasOffsetX) / canvasScale);
            const galaxyY = Math.round((clickY + canvasOffsetY) / canvasScale);
            
            document.getElementById('pointX').value = galaxyX;
            document.getElementById('pointY').value = galaxyY;
            
            // Automatically perform interaction
            performInteraction(currentState, galaxyX, galaxyY);
        });

        // Allow Enter key to evaluate expressions
        document.getElementById('expression').addEventListener('keypress', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                evaluateExpression();
            }
        });

        // Initialize canvas
        clearCanvas();
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

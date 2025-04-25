package rest_easy

import (
	"context"
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Code       int    `json:"code,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

type FuncReq struct {
	Body    any
	Request *http.Request
	HasBody bool
}

type HandlerFunc[In any, Out any] func(ctx context.Context, in FuncReq) (Out, error)

// NewHandler creates a new http.Handler using the given `funcIn`.
func NewHandler[In FuncReq, Out any](f HandlerFunc[In, Out]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var funcIn FuncReq
		funcIn.Request = r

		// Handle Request Input
		// TODO Consider parsing headers, query params, etc and adding them to `FuncReq`
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			err := json.NewDecoder(r.Body).Decode(&funcIn.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		// Handle the response from the request func
		out, err := f(r.Context(), funcIn)
		if err != nil {

			sc := http.StatusInternalServerError

			// Check if the returned error is our custom ErrorResponse
			if e, ok := err.(*ErrorResponse); ok {

				w.WriteHeader(e.StatusCode)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(e)

				// Make sure to return to make sure the rest of this code
				// doesn't run and mess with the return.
				return
			}

			message := err.Error()
			w.WriteHeader(sc)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"error": message})

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if funcIn.HasBody {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(out)
		}

	})
}

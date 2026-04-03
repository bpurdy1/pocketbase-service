package webhooks

import (
	"fmt"
	"net/http"
)

func Handle(providerName string, handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provider, ok := Get(providerName)
		if !ok {
			http.Error(w, "unknown provider", http.StatusBadRequest)
			return
		}

		// 1. verify signature
		if err := provider.Verify(r); err != nil {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		// 2. parse event
		event, err := provider.Parse(r)
		if err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		// 3. handle event
		if err := handler(*event); err != nil {
			fmt.Println("webhook handler error:", err)
			http.Error(w, "handler error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

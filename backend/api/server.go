package api

import (
	"net/http"
	"fmt"
)

func StartServer(){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Noted V4 API is running 🚀")
	})

	fmt.Println("Server is running on port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
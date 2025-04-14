package handlers

/*func TestDBHandler(db *mongo.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.Client().Ping(ctx, nil); err != nil {
			http.Error(w, "DB down", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("MongoDB is live ✅"))
	}
}*/

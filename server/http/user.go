package handlers

import (
	"encoding/json"
	"net/http"
	"server/pkg"

	"go.mongodb.org/mongo-driver/mongo"
)

func Add_friend(client *mongo.Client, adder *pkg.UserManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		friend := r.URL.Query().Get("friend")
		if username == "" || friend == "" {
			http.Error(w, "username and friend are required", http.StatusBadRequest)
			return
		}
		clientObj := &pkg.Client{Username: username}
		adder.AddFriend(clientObj, friend)
		w.Write([]byte("Friend added successfully"))
	}
}

func RegisterServer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		pkg.RegisterUser(client, username)
		w.Write([]byte("User registered successfully"))
	}
}

func GetFriends(client *mongo.Client, userManager *pkg.UserManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &pkg.Client{Username: username}
		friends := userManager.GetFriends(clientObj)
		if friends == nil {
			http.Error(w, "Failed to get friends", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"friends":  friends,
		})
	}
}

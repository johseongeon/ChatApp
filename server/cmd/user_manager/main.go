package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/pkg"
)

var UserManagerInstance = &pkg.UserManager{}

var Collection = &pkg.MessageCollection{}

// curl http://localhost:8082/register?username=A
// curl "http://localhost:8082/addFriend?username=A&friend=B"
func main() {
	client, err := pkg.ConnectMongoDB()
	if err != nil {
		log.Fatal("MongoDB 연결 실패:", err)
	}

	userManager := &pkg.UserManager{Client: client}
	Collection.MessageCol = client.Database("ChatDB").Collection("users")

	http.HandleFunc("/register", pkg.RegisterServer(client))

	http.HandleFunc("/addFriend", func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		friend := r.URL.Query().Get("friend")
		if username == "" || friend == "" {
			http.Error(w, "username and friend are required", http.StatusBadRequest)
			return
		}
		clientObj := &pkg.Client{Username: username}
		userManager.AddFriend(clientObj, friend)
		w.Write([]byte("Friend added successfully"))
	})

	http.HandleFunc("/getFriends", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/getRooms", func(w http.ResponseWriter, r *http.Request) {
		pkg.EnableCORS(w)
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username is required", http.StatusBadRequest)
			return
		}
		clientObj := &pkg.Client{Username: username}
		rooms := userManager.GetChatRooms(clientObj)
		if rooms == nil {
			http.Error(w, "Failed to get rooms", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username": username,
			"rooms":    rooms,
		})
	})

	fmt.Println("Server started on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

package main

import (
	"fmt"
	"food-backend/internal/config"
	"food-backend/internal/models"
)

func main() {
	config.ConnectDatabase()
	var notifs []models.Notification
	config.DB.Find(&notifs)
	fmt.Printf("Banyak notif: %d\n", len(notifs))
	for _, n := range notifs {
		fmt.Printf("ID: %d, UserID: %d, Title: %s, Message: %s\n", n.ID, n.UserID, n.Title, n.Message)
	}

    var orders []models.Order
    config.DB.Order("id desc").Limit(2).Find(&orders)
    for _, o := range orders {
		var uid uint
		if o.UserID != nil { uid = *o.UserID }
		fmt.Printf("OrderID: %d, UserID: %d, GuestName: %s\n", o.ID, uid, o.GuestName)
	}
}

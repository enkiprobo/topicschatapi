package main

import (
	"log"
	"net/http"
	"os"
	"topicschatapi/database"
	"topicschatapi/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("no port error")
	}

	err := database.InitDB()
	defer database.MainDB.Close()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	handler.Hubnya = handler.NewHub()
	go handler.Hubnya.Run()
	http.HandleFunc("/topicslivechat", handler.LiveChatHandler)
	// kumpulan handler

	//========================================================================
	http.HandleFunc("/createuser", handler.CreateUser)
	// enki := database.User{
	// 	Username:     "enkiprobo",
	// 	Password:     "IMPACT",
	// 	ProfileImage: "kosong",
	// 	FullName:     "Enki Probo Sidhi",
	// 	BirthDate:    "12-12-2016",
	// }
	// err = database.CreateUser(enki)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	//========================================================================
	http.HandleFunc("/login", handler.Login)
	// user := database.GetUser("enkiprobo")
	// fmt.Println(user)
	//========================================================================
	http.HandleFunc("/creategroup", handler.CreateGroup)
	// err = database.InitTX()
	// if err != nil {
	// 	database.StopTX()
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// id, err := database.CreateGroup("SD", "kosong", "enkiprobo")
	// if err != nil {
	// 	database.RollBackTX()
	// 	database.StopTX()
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// err = database.InsertMember(id, "enkiprobo")
	// if err != nil {
	// 	database.RollBackTX()
	// 	database.StopTX()
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// err = database.CreateTopic("All", id)
	// if err != nil {
	// 	database.RollBackTX()
	// 	database.StopTX()
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// database.CommitTX()
	// database.StopTX()
	//========================================================================
	http.HandleFunc("/createchat", handler.CreateChat)
	// err = database.CreateChat("hai", 2, 1)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	//========================================================================
	http.HandleFunc("/createmute", handler.CreateMute)
	// err = database.CreateMute("enkiprobo", 1)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	//========================================================================
	http.HandleFunc("/pinchat", handler.PinChat)
	// err = database.PinChat(true, 1)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	//=======================================================================
	http.HandleFunc("/deletemute", handler.DeleteMute)
	// err = database.DeleteMute("enkiprobo", 1)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	//=======================================================================
	http.HandleFunc("/getmutelist", handler.GetMuteList)
	// arr, err := database.GetMuteList("enkiprobo")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// fmt.Println(arr)
	//=======================================================================
	http.HandleFunc("/getgrouptopic", handler.GetGroupTopic)
	// topics, err := database.GetGroupTopicQuery(3)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// fmt.Println(topics)
	//=======================================================================
	http.HandleFunc("/getgroupchat", handler.GetGroupChat)
	// groupschat, err := database.GetGroupChat(1)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// fmt.Println(groupschat)
	//=======================================================================
	http.HandleFunc("/getusergroup", handler.GetUserGroup)
	// usergroup, err := database.GetUserGroup("enkiprobo")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }
	// fmt.Println(usergroup)
	//=======================================================================
	http.HandleFunc("/insertmember", handler.InsertMember)
	http.HandleFunc("/createtopic", handler.CreateTopic)
	http.HandleFunc("/getchatgroupall", handler.GetChatGroupAll)

	log.Println(http.ListenAndServe(":"+port, nil))
}

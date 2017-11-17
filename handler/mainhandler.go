package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"topicschatapi/database"
)

type (
	UserResponse struct {
		User   database.User `json:"user"`
		Status string        `json:"status"`
	}
	UsersGroupResponse struct {
		UsersGroup []database.UsersGroup `json:"group_list"`
		Status     string                `json:"status"`
	}
	GroupsChatResponse struct {
		GroupsChat []database.GroupsChat `json:"chat_list"`
		Status     string                `json:"status"`
	}
	MuteListResponse struct {
		MuteList []int64 `json:"topic_id_list_mute"`
		Status   string  `json:"status"`
	}
	TopicResponse struct {
		Topics []database.Topic `json:"topic_list"`
		Status string           `json:"status"`
	}
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}
	user := database.User{
		Username:     r.FormValue("username"),
		Password:     r.FormValue("password"),
		FullName:     r.FormValue("fullname"),
		ProfileImage: r.FormValue("profileimage"),
		BirthDate:    r.FormValue("birthdate"),
	}

	checkUser := database.GetUser(user.Username)
	if checkUser.Username != "" {
		w.Write([]byte(`{"status": "username already exist"}`))
		return
	}

	err := database.CreateUser(user)
	if err != nil {
		log.Println("handle create user: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mapResponse := map[string]string{
		"status": "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	user := database.GetUser(username)
	if user.Username == "" {
		w.Write([]byte(`{"status": "user not exist"}`))
		return
	}
	if user.Password != password {
		w.Write([]byte(`{"status": "password not match"}`))
		return
	}

	mapResponse := UserResponse{
		User:   user,
		Status: "OK",
	}
	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}
	groupName := r.FormValue("groupname")
	groupImage := r.FormValue("groupimage")
	username := r.FormValue("username")

	err := database.InitTX()
	if err != nil {
		database.StopTX()
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := database.CreateGroup(groupName, groupImage, username)
	if err != nil {
		database.RollBackTX()
		database.StopTX()
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	idgm, err := database.InsertMember(id, username)
	if err != nil {
		database.RollBackTX()
		database.StopTX()
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = database.CreateTopic("All", id)
	if err != nil {
		database.RollBackTX()
		database.StopTX()
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	database.CommitTX()
	database.StopTX()

	// create response
	mapResponse := UsersGroupResponse{
		UsersGroup: []database.UsersGroup{
			{
				IDGroup:       id,
				GroupName:     groupName,
				GroupImage:    groupImage,
				IDGroupMember: idgm,
				CreatedBy:     username,
			},
		},
		Status: "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}
func CreateChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	idgm, _ := strconv.ParseInt(r.FormValue("idgm"), 10, 64)
	idTopic, _ := strconv.ParseInt(r.FormValue("idtopic"), 10, 64)

	idgcd, err := database.CreateChat(message, idgm, idTopic)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupChat, err := database.GetGroupChat(idTopic, idgcd, "")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create response
	mapResponse := WebsocketResponse{
		Category: "chat",
		Chat:     groupChat[0],
		Status:   "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Hubnya.broadcast <- response
	w.Write(response)
	return
}
func CreateMute(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	idTopic, _ := strconv.ParseInt(r.FormValue("idtopic"), 10, 64)

	err := database.CreateMute(username, idTopic)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapResponse := map[string]string{
		"status": "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}

func PinChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	pin, _ := strconv.ParseBool(r.FormValue("pin"))
	idgcd, _ := strconv.ParseInt(r.FormValue("idgcd"), 10, 64)

	err := database.PinChat(pin, idgcd)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapResponse := map[string]string{
		"status": "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}
func DeleteMute(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	idTopic, _ := strconv.ParseInt(r.FormValue("idtopic"), 10, 64)

	err := database.DeleteMute(username, idTopic)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapResponse := map[string]string{
		"status": "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}

func GetMuteList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")

	muteList, err := database.GetMuteList(username)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create response
	mapResponse := MuteListResponse{
		MuteList: muteList,
		Status:   "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}

func GetGroupTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	idGroup, _ := strconv.ParseInt(r.FormValue("idgroup"), 10, 64)

	groupTopics, err := database.GetGroupTopic(idGroup)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapResponse := TopicResponse{
		Topics: groupTopics,
		Status: "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}
func GetGroupChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	idTopic, _ := strconv.ParseInt(r.FormValue("idtopic"), 10, 64)

	groupChats, err := database.GetGroupChat(idTopic, -1, "")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create response
	mapResponse := GroupsChatResponse{
		GroupsChat: groupChats,
		Status:     "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}

func GetUserGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")

	userGroups, err := database.GetUserGroup(username)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create response
	mapResponse := UsersGroupResponse{
		UsersGroup: userGroups,
		Status:     "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}
func InsertMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	groupID, _ := strconv.ParseInt(r.FormValue("idgroup"), 10, 64)
	username := r.FormValue("username")

	idgm, err := database.InsertMember(groupID, username)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	group := database.GetGroupByIDGM(idgm)
	group.IDGroupMember = idgm

	// create response
	mapResponse := WebsocketResponse{
		Category: "group",
		Group:    group,
		Status:   "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Hubnya.broadcast <- response
	w.Write(response)
	return
}

func CreateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	topicname := r.FormValue("topicname")
	groupID, _ := strconv.ParseInt(r.FormValue("idgroup"), 10, 64)

	topic, err := database.CreateTopic(topicname, groupID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create response
	mapResponse := WebsocketResponse{
		Category: "chat",
		Topic:    topic,
		Status:   "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Hubnya.broadcast <- response
	w.Write(response)
	return
}

func GetChatGroupAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "not POST request", http.StatusBadRequest)
		return
	}

	idGroup, _ := strconv.ParseInt(r.FormValue("idgroup"), 10, 64)

	topicList, err := database.GetGroupTopic(idGroup)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	whereClause := " WHERE "
	for i, topic := range topicList {
		if i == 0 {
			whereClause += "id_topic = " + strconv.FormatInt(topic.IDTopic, 10)
		} else {
			whereClause += " OR id_topic = " + strconv.FormatInt(topic.IDTopic, 10)
		}
	}

	groupChats, err := database.GetGroupChat(-1, -1, whereClause)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create response
	mapResponse := GroupsChatResponse{
		GroupsChat: groupChats,
		Status:     "OK",
	}

	response, err := json.Marshal(mapResponse)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
	return
}

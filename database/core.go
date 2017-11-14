package database

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

const (
	createUserQuery string = `
		INSERT INTO 
			accounts(username, password, full_name, profile_image, birth_date)
		VALUES
			($1, $2, $3, $4, to_date( $5,'dd-mm-yyyy'))`
	getUserQuery string = `
		SELECT 
			username, password, full_name, profile_image, birth_date
		FROM
			accounts
		WHERE 
			username = $1`
	createGroupQuery string = `
		INSERT INTO 
			groups(group_name, group_image, created_by)
		VALUES
			($1, $2, $3)`
	insertMemberQuery string = `
		INSERT INTO
			group_members(id_group, username)
		VALUES
			($1, $2)`
	createTopicQuery string = `
		INSERT INTO
			topics(topic_name, id_group)
		VALUES
			($1, $2)`
	createChatQuery string = `
		INSERT INTO
			group_chat_details(chat_message, id_gm, id_topic)
		VALUES
			($1, $2, $3)`
	createMuteQuery string = `
		INSERT INTO
			mutes(username, id_topic)
		VALUES
			($1, $2)`
	pinChatQuery string = `
		UPDATE 
			group_chat_details
		SET
			pin = $1
		WHERE
			id_gcd = $2`
	deleteMuteQuery string = `
		DELETE FROM
			mutes
		WHERE
			username = $1
			AND
			id_topic = $2`
	getMuteListQuery string = `
		SELECT 
			id_topic
		FROM
			mutes
		WHERE
			username = $1`
	getGroupTopicQuery string = `
		SELECT 
			id_topic, topic_name, id_group
		FROM
			topics
		WHERE
			id_group = $1`
	getGroupChatQuery string = `
		SELECT
			id_gcd, chat_message, id_topic, pin, created_time, username, id_gm
		FROM
			group_chat_details
			JOIN group_members USING(id_gm)
		WHERE
			id_topic = $1
		ORDER BY 
			created_time 
			DESC`
	getGroupChatUsingIDCHATQuery string = `
			SELECT
				id_gcd, chat_message, id_topic, pin, created_time, username, id_gm
			FROM
				group_chat_details
				JOIN group_members USING(id_gm)
			WHERE
				id_gcd = $1`
	getUserGroupQuery string = ` 
		SELECT 
			id_group, group_name, group_image, created_by, id_gm
		FROM
			group_members
			JOIN 
				groups USING(id_group)
		WHERE
			username = $1`
)

type (
	User struct {
		Username     string `json:"username"`
		Password     string `json:"-"`
		FullName     string `json:"full_name"`
		ProfileImage string `json:"profile_image"`
		BirthDate    string `json:"birth_date"`
	}
	UsersGroup struct {
		IDGroup       int64  `json:"id_group"`
		GroupName     string `json:"group_name"`
		GroupImage    string `json:"group_image"`
		CreatedBy     string `json:"created_by"`
		IDGroupMember int64  `json:"id_gm"`
	}
	Topic struct {
		IDTopic   int64  `json:"id_topic"`
		TopicName string `json:"topic_name"`
		IDGroup   int64  `json:"id_group"`
	}
	GroupsChat struct {
		IDGroupsChat  int64  `json:"id_gcd"`
		ChatMessage   string `json:"chat_message"`
		IDTopic       int64  `json:"id_topic"`
		Pin           bool   `json:"pin"`
		CreatedTime   string `json:"created_time"`
		Username      string `json:"username"`
		IDGroupMember int64  `json:"id_gm"`
	}
)

var (
	MainDB *sql.DB
	MainTX *sql.Tx
)

func InitDB() error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	MainDB = db
	return nil
}
func InitTX() error {
	tx, err := MainDB.Begin()
	if err != nil {
		return err
	}
	MainTX = tx
	return nil
}
func StopTX() {
	MainTX = nil
}
func RollBackTX() {
	MainTX.Rollback()
}
func CommitTX() {
	MainTX.Commit()
}

func CreateUser(user User) error {
	_, err := MainDB.Exec(createUserQuery, user.Username, user.Password, user.FullName, user.ProfileImage, user.BirthDate)

	return err
}

func GetUser(username string) User {
	user := User{}

	row := MainDB.QueryRow(getUserQuery, username)

	row.Scan(&user.Username, &user.Password, &user.FullName, &user.ProfileImage, &user.BirthDate)

	return user
}

func CreateGroup(groupname, groupimage, username string) (int64, error) {
	lastGroupID := int64(-1)

	var err error
	returningID := " RETURNING id_group"
	if MainTX != nil {
		err = MainTX.QueryRow(createGroupQuery+returningID, groupname, groupimage, username).Scan(&lastGroupID)
	} else {
		err = MainDB.QueryRow(createGroupQuery+returningID, groupname, groupimage, username).Scan(&lastGroupID)
	}

	return lastGroupID, err
}
func InsertMember(groupID int64, username string) (int64, error) {
	lastMemberID := int64(-1)

	var err error
	returningID := " RETURNING id_gm"
	if MainTX != nil {
		err = MainTX.QueryRow(insertMemberQuery+returningID, groupID, username).Scan(&lastMemberID)
	} else {
		err = MainDB.QueryRow(insertMemberQuery+returningID, groupID, username).Scan(&lastMemberID)
	}

	return lastMemberID, err
}
func CreateTopic(topicsname string, groupID int64) error {
	var err error
	if MainTX != nil {
		_, err = MainTX.Exec(createTopicQuery, topicsname, groupID)
	} else {
		_, err = MainDB.Exec(createTopicQuery, topicsname, groupID)
	}
	return err
}

func CreateChat(message string, idgm, idtopic int64) (int64, error) {
	lastIDGroupsChat := int64(-1)

	returningID := " RETURNING id_gcd"
	err := MainDB.QueryRow(createChatQuery+returningID, message, idgm, idtopic).Scan(&lastIDGroupsChat)

	return lastIDGroupsChat, err
}

func CreateMute(username string, idtopic int64) error {
	_, err := MainDB.Exec(createMuteQuery, username, idtopic)
	return err
}

func PinChat(pin bool, idgcd int64) error {
	_, err := MainDB.Exec(pinChatQuery, pin, idgcd)
	return err
}

func DeleteMute(username string, idtopic int64) error {
	_, err := MainDB.Exec(deleteMuteQuery, username, idtopic)
	return err
}

func GetMuteList(username string) ([]int64, error) {
	var topicMuteList = []int64{}

	rows, err := MainDB.Query(getMuteListQuery, username)
	if err != nil {
		return topicMuteList, err
	}
	for rows.Next() {
		var topicid int64
		rows.Scan(&topicid)
		topicMuteList = append(topicMuteList, topicid)
	}
	return topicMuteList, nil
}
func GetGroupTopicQuery(idgroup int64) ([]Topic, error) {
	var topicList []Topic

	rows, err := MainDB.Query(getGroupTopicQuery, idgroup)
	if err != nil {
		return topicList, err
	}
	for rows.Next() {
		var topic Topic
		rows.Scan(&topic.IDTopic, &topic.TopicName, &topic.IDGroup)

		topicList = append(topicList, topic)
	}

	return topicList, nil
}

func GetGroupChat(idtopic, idgcd int64) ([]GroupsChat, error) {
	var groupsChatList []GroupsChat

	var rows *sql.Rows
	var err error
	if idgcd != -1 {
		rows, err = MainDB.Query(getGroupChatUsingIDCHATQuery, idgcd)
	} else {
		rows, err = MainDB.Query(getGroupChatQuery, idtopic)
	}

	if err != nil {
		return groupsChatList, err
	}
	for rows.Next() {
		var groupChat GroupsChat
		rows.Scan(&groupChat.IDGroupsChat, &groupChat.ChatMessage, &groupChat.IDTopic, &groupChat.Pin, &groupChat.CreatedTime, &groupChat.Username, &groupChat.IDGroupMember)

		groupsChatList = append(groupsChatList, groupChat)
	}
	return groupsChatList, nil
}

func GetUserGroup(username string) ([]UsersGroup, error) {
	var usersGroupList []UsersGroup

	rows, err := MainDB.Query(getUserGroupQuery, username)
	if err != nil {
		return usersGroupList, err
	}
	for rows.Next() {
		var usersGroup UsersGroup
		rows.Scan(&usersGroup.IDGroup, &usersGroup.GroupName, &usersGroup.GroupImage, &usersGroup.CreatedBy, &usersGroup.IDGroupMember)

		usersGroupList = append(usersGroupList, usersGroup)
	}
	return usersGroupList, nil
}
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

var groupID = -1001288115081

type reqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func HandlerSendMessage(res http.ResponseWriter, req *http.Request) {
	body := &reqBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		fmt.Println("could not decode request body", err)
		return
	}

	if err := sendMessage(body.ChatID, body.Text); err != nil {
		fmt.Println("error in sending reply:", err)
		return
	}
}

func HandlerUpdateMembers(res http.ResponseWriter, req *http.Request) {
	if err := updateMembers(); err != nil {
		fmt.Println("error in sending reply:", err)
		return
	}
}

func sendMessage(chatID int64, text string) error {
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   text,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	res, err := http.Post("https://api.telegram.org/bot1880447222:AAFJkmhX4V9u7mK5mUtTQoJTeQr4YjSTWPg/sendMessage", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status" + res.Status)
	}

	return nil
}

func updateMembers() error {

	res, err := http.Get("https://api.telegram.org/bot1880447222:AAFJkmhX4V9u7mK5mUtTQoJTeQr4YjSTWPg/getUpdates")
	if err != nil {
		return err
	}
	var data map[string][]map[string]map[string]map[string]int64
	if res.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal([]byte(bodyBytes), &data)
		listUser := getListUser()
		var left_chat_participant_id, new_chat_participant_id int64
		left_chat_participant_id = data["result"][len(data["result"])-1]["message"]["left_chat_participant"]["id"]
		new_chat_participant_id = data["result"][len(data["result"])-2]["message"]["new_chat_participant"]["id"]
		new_chat_participant_id = new_chat_participant_id + 2
		left_chat_participant_id += 1
		if new_chat_participant_id > 0 {
			index := find(listUser, new_chat_participant_id)
			if index == -1 {
				listUser = append(listUser, new_chat_participant_id)
			}
		}
		if left_chat_participant_id > 0 {
			index := find(listUser, left_chat_participant_id)
			if index != -1 {
				listUser = removeIndex(listUser, index)
			}
		}
		updateListUser(listUser)

	} else {
		return errors.New("unexpected status" + res.Status)
	}
	return nil
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/sendMessage", HandlerSendMessage)
	myRouter.HandleFunc("/updateMembers", HandlerUpdateMembers)
	log.Fatal(http.ListenAndServe(":3000", myRouter))
}

func removeIndex(arr []int64, index int) []int64 {
	return append(arr[:index], arr[index+1:]...)
}

func find(arr []int64, elementToFind int64) int {
	for i, n := range arr {
		if n == elementToFind {
			return i
		}
	}
	return -1
}

func getListUser() []int64 {
	csvFile, err := os.Open("list_users.csv")
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(csvFile)
	record, err := r.Read()
	fmt.Println(record)
	// for {
	// record, err := r.Read()
	// if err == io.EOF {
	// 	break
	// }
	new_record := []int64{}
	for i := range record {
		str_number := record[i]
		number, _ := strconv.ParseInt(str_number, 10, 64)
		new_record = append(new_record, number)
	}
	csvFile.Close()
	return new_record
}

func updateListUser(listUser []int64) {
	record_to_update := []string{}
	for i := range listUser {
		number := listUser[i]
		str_number := strconv.Itoa(int(number))
		record_to_update = append(record_to_update, str_number)
	}
	fmt.Println(record_to_update)
	csvFile, err := os.Create("list_users.csv")
	if err != nil {
		log.Fatal(err)
	}
	w := csv.NewWriter(csvFile)
	err = w.Write(record_to_update)
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()
	csvFile.Close()
}

// FInally, the main funtion starts our server on port 3000
func main() {
	handleRequests()
}

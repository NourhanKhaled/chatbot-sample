package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NourhanKhaled/chatbot-sample/tasklistAPI"
	"github.com/ramin0/chatbot"
)

// Autoload environment variables in .env
import _ "github.com/joho/godotenv/autoload"

func chatbotProcess(session chatbot.Session, message string) (string, error) {
	// Make sure a history key is defined in the session which points to a slice of strings
	_, historyFound := session["history"]
	if !historyFound {
		session["history"] = []string{}
	}
	fmt.Println("SESSION")
	fmt.Println(session)
	fmt.Printf("%#v\n", session)

	// Fetch the history from session and cast it to an array of strings
	history, _ := session["history"].([]string)

	// Add the message in the parsed body to the messages in the session
	history = append(history, message)

	s := strings.Split(message, " ")
	key := strings.ToLower(s[0])

	fmt.Println(message)

	if key == "token:" {
		accesstok := s[1]
		// refreshtok := s[2]
		// date := s[3]
		session["token"] = accesstok
		fmt.Println("TOKEN")
		fmt.Println(session)
		fmt.Printf("%#v\n", session)
		message, err := tasklistAPI.PostCode(accesstok)
		if err != nil {
			return "", err
		}
		return message, nil
	} else if key == "delete:" {
		taskNumber := s[1]
		fmt.Println("IN DELETE")
		fmt.Printf("%#v\n", session)
		message, err := tasklistAPI.DeleteTask(taskNumber, session["token"].(string))
		if err != nil {
			return "", err
		}
		return message, nil
	} else if key == "completed:" {
		taskNumber := s[1]
		message, err := tasklistAPI.TaskCompleted(taskNumber, session["token"].(string))
		if err != nil {
			return "", err
		}
		return message, nil
	} else if key == "view" && len(s) == 1 {
		message, err := tasklistAPI.GetTasks(session["token"].(string))
		if err != nil {
			return "", err
		}
		return message, nil
	} else if key == "create:" {
		title := ""
		notes := ""
		due := ""

		for i := 1; i < len(s); i++ {
			curr := strings.ToLower(s[i])

			if curr == "title:" {
				for j := i + 1; j < len(s); j++ {
					curr1 := s[j]
					temp := strings.ToLower(curr1)
					if temp == "notes:" || temp == "due:" {
						i = j - 1
						break
					}
					title += curr1 + " "
				}
			}

			if curr == "notes:" {
				for j := i + 1; j < len(s); j++ {
					curr1 := s[j]
					temp := strings.ToLower(curr1)
					if temp == "title:" || temp == "due:" {
						i = j - 1
						break
					}
					notes += curr1 + " "
				}
			}

			if curr == "due:" {
				for j := i + 1; j < len(s); j++ {
					curr1 := s[j]
					temp := strings.ToLower(curr1)
					if temp == "notes:" || temp == "title:" {
						i = j - 1
						break
					}
					due += curr1
				}
			}
		}

		message, err := tasklistAPI.CreateTask(title, notes, due, session["token"].(string))
		if err != nil {
			return "", err
		}
		return message, nil
	} else if key == "update:" {
		taskNumber := s[1]
		title := ""
		notes := ""
		due := ""

		for i := 2; i < len(s); i++ {
			curr := s[i]

			if curr == "title:" {
				for j := i + 1; j < len(s); j++ {
					curr1 := s[j]
					temp := strings.ToLower(curr1)
					if temp == "notes:" || temp == "due:" {
						i = j - 1
						break
					}
					title += curr1 + " "
				}
			}

			if curr == "notes:" {
				for j := i + 1; j < len(s); j++ {
					curr1 := s[j]
					temp := strings.ToLower(curr1)
					if temp == "title:" || temp == "due:" {
						i = j - 1
						break
					}
					notes += curr1 + " "
				}
			}

			if curr == "due:" {
				for j := i + 1; j < len(s); j++ {
					curr1 := s[j]
					temp := strings.ToLower(curr1)
					if temp == "notes:" || temp == "title:" {
						i = j - 1
						break
					}
					due += curr1
				}
			}
		}

		message, err := tasklistAPI.UpdateTask(taskNumber, title, notes, due, session["token"].(string))
		if err != nil {
			return "", err
		}
		return message, nil
	} else {
		message := "You have entered an invalid message."
		return message, nil
	}
}

func main() {
	// Uncomment the following lines to customize the chatbot
	chatbot.WelcomeMessage = tasklistAPI.SendAuthURL()
	chatbot.ProcessFunc(chatbotProcess)

	// Use the PORT environment variable
	port := os.Getenv("PORT")
	// Default to 3000 if no PORT environment variable was defined
	if port == "" {
		port = "3000"
	}

	// Start the server
	fmt.Printf("Listening on port %s...\n", port)
	log.Fatalln(chatbot.Engage(":" + port))
}

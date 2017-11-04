
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/ramin0/chatbot"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/NourhanKhaled/chatbot-sample/tasklistAPI"
	cors "github.com/heppu/simple-cors"
)

// Autoload environment variables in .env
import _ "github.com/joho/godotenv/autoload"

func chatbotProcess(session chatbot.Session, message string) (string, error) {
	if strings.EqualFold(message, "chatbot") {
		return "", fmt.Errorf("This can't be, I'm the one and only %s!", message)
	}

	var questionMarksCount int
	// Try fetching the count of question marks
	count, found := session["questionMarksCount"]
	// If a count is saved in the session
	if found {
		// Cast it into an int (since sessions values are generic)
		questionMarksCount = count.(int)
	} else {
		// Otherwise, initialize the count to 1
		questionMarksCount = 1
	}

	// Build the question marks string according to the question marks count
	var questionMarks string
	for i := 1; i <= questionMarksCount; i++ {
		questionMarks += "?"
	}

	// Save the updated question marks count to the session
	session["questionMarksCount"] = questionMarksCount + 1

	// Return the response with an extra question mark
	return fmt.Sprintf("Hello <b>%s</b>, my name is chatbot. What was yours again%s", message, questionMarks), nil
}



var (
	// WelcomeMessage A constant to hold the welcome message
	WelcomeMessage = "Welcome, what do you want to order?"

	// sessions = {
	//   "uuid1" = Session{...},
	//   ...
	// }
	sessions = map[string]Session{}

	processor = sampleProcessor
)

type (
	// Session Holds info about a session
	Session map[string]interface{}

	// JSON Holds a JSON object
	JSON map[string]interface{}

	// Processor Alias for Process func
	Processor func(session Session, message string) (string, error)
)

func sampleProcessor(session Session, message string) (string, error) {
	// Make sure a history key is defined in the session which points to a slice of strings
	_, historyFound := session["history"]
	if !historyFound {
		session["history"] = []string{}
	}

	// Fetch the history from session and cast it to an array of strings
	history, _ := session["history"].([]string)

	// Add the message in the parsed body to the messages in the session
	history = append(history, message)

	s := strings.Split(message, " ")
	key := s[0]

	if(key == "token:") {
		tok := s[1]
		session["token"] = tok
		message,err := tasklistAPI.PostCode(tok)
		if err!=nil {
			return "",err
		}
		return message,nil
	}
	if(key == "delete:") {
		taskNumber := s[1]
		message,err := tasklistAPI.DeleteTask(taskNumber)
		if err!=nil {
			return "",err
		}
		return message,nil
	}
	if(key == "completed:") {
		taskNumber := s[1]
		message,err := tasklistAPI.TaskCompleted(taskNumber)
		if err!=nil {
			return "",err
		}
		return message,nil
	}
	if(key == "view") {
		message,err := tasklistAPI.GetTasks()
		if err!=nil {
			return "",err
		}
		return message,nil
	}
	if(key == "create:") {
		title := ""
		notes := ""
		due := ""

		for i := 1; i < len(s); i++ {
			curr := s[i]

			if(curr == "title:") {
				for j := i+1; j < len(s); j++ {
					curr1 := s[j]
					if(curr1 == "notes:" || curr1 == "due:") {
						i = j-1
						break
					}
					title += curr1
				}
			}

			if(curr == "notes:") {
				for j := i+1; j < len(s); j++ {
					curr1 := s[j]
					if(curr1 == "title:" || curr1 == "due:") {
						i = j-1
						break
					}
					notes += curr1
				}
			}

			if(curr == "due:") {
				for j := i+1; j < len(s); j++ {
					curr1 := s[j]
					if(curr1 == "notes:" || curr1 == "title:") {
						i = j-1
						break
					}
					due += curr1
				}
			}
		}

		message,err := tasklistAPI.CreateTask(title, notes, due)
		if err!=nil {
			return "",err
		}
		return message,nil
	}

	if(key == "update:") {
		taskNumber := s[1]
		title := ""
		notes := ""
		due := ""

		for i := 2; i < len(s); i++ {
			curr := s[i]

			if(curr == "title:") {
				for j := i+1; j < len(s); j++ {
					curr1 := s[j]
					if(curr1 == "notes:" || curr1 == "due:") {
						i = j-1
						break
					}
					title += curr1+" "
				}
			}

			if(curr == "notes:") {
				for j := i+1; j < len(s); j++ {
					curr1 := s[j]
					if(curr1 == "title:" || curr1 == "due:") {
						i = j-1
						break
					}
					notes += curr1+" "
				}
			}

			if(curr == "due:") {
				for j := i+1; j < len(s); j++ {
					curr1 := s[j]
					if(curr1 == "notes:" || curr1 == "title:") {
						i = j-1
						break
					}
					due += curr1
				}
			}
		}

		message,err := tasklistAPI.UpdateTask(taskNumber,title, notes, due)
		if err!=nil {
			return "",err
		}
		return message,nil
	}

	// Form a sentence out of the history in the form Message 1, Message 2, and Message 3
	l := len(history)
	wordsForSentence := make([]string, l)
	copy(wordsForSentence, history)
	if l > 1 {
		wordsForSentence[l-1] = "and " + wordsForSentence[l-1]
	}
	sentence := strings.Join(wordsForSentence, ", ")

	// Save the updated history to the session
	session["history"] = history

	return fmt.Sprintf("So, you want %s! What else?", strings.ToLower(sentence)), nil
}

func chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	//Make sure a UUID exists in the Authorization header
	uuid := r.Header.Get("Authorization")
	if uuid == "" {
		http.Error(w, "Missing or empty Authorization header.", http.StatusUnauthorized)
		return
	}

	//Make sure a session exists for the extracted UUID
	session, sessionFound := sessions[uuid]
	if !sessionFound {
		http.Error(w, fmt.Sprintf("No session found for: %v.", uuid), http.StatusUnauthorized)
		return
	}

  data := JSON{}
  if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
    http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
    return
  }
  defer r.Body.Close()

  // Make sure a message key is defined in the body of the request
  _, messageFound := data["message"]
  if !messageFound {
    http.Error(w, "Missing message key in body.", http.StatusBadRequest)
    return
  }

	message, err := processor(session, data["message"].(string))
	if err != nil {
		http.Error(w, err.Error(), 422 /* http.StatusUnprocessableEntity */)
		return
	}
	writeJSON(w, JSON{
		"uuid": uuid,
		"message": message,
	})

}

// withLog Wraps HandlerFuncs to log requests to Stdout
func withLog(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := httptest.NewRecorder()
		fn(c, r)
		log.Printf("[%d] %-4s %s\n", c.Code, r.Method, r.URL.Path)

		for k, v := range c.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(c.Code)
		c.Body.WriteTo(w)
	}
}

// writeJSON Writes the JSON equivilant for data into ResponseWriter w
func writeJSON(w http.ResponseWriter, data JSON) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ProcessFunc Sets the processor of the chatbot
func ProcessFunc(p Processor) {
	processor = p
}

// handleWelcome Handles /welcome and responds with a welcome message and a generated UUID
func handleWelcome(w http.ResponseWriter, r *http.Request) {
	// Generate a UUID.
	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))
	message := tasklistAPI.SendAuthURL()
	// Create a session for this UUID
	sessions[uuid] = Session{}

	// Write a JSON containg the welcome message and the generated UUID

	writeJSON(w, JSON {
		"uuid":    uuid,
		"message": message,
	})
	// fmt.Fprintln(w, message)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	// Make sure only POST requests are handled
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
		return
	}

	//Make sure a UUID exists in the Authorization header
	uuid := r.Header.Get("Authorization")
	if uuid == "" {
		http.Error(w, "Missing or empty Authorization header.", http.StatusUnauthorized)
		return
	}

	//Make sure a session exists for the extracted UUID
	session, sessionFound := sessions[uuid]
	if !sessionFound {
		http.Error(w, fmt.Sprintf("No session found for: %v.", uuid), http.StatusUnauthorized)
		return
	}

	// Parse the JSON string in the body of the request
	data := JSON{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Make sure a message key is defined in the body of the request
	_, messageFound := data["message"]
	if !messageFound {
		http.Error(w, "Missing message key in body.", http.StatusBadRequest)
		return
	}

	// Process the received message
	message, err := processor(session, data["message"].(string))
	if err != nil {
		http.Error(w, err.Error(), 422 /* http.StatusUnprocessableEntity */)
		return
	}

	// Write a JSON containg the processed response
	writeJSON(w, JSON {
		"message": message,
	})
}

// handle Handles /
func handle(w http.ResponseWriter, r *http.Request) {
	body :=
		"<!DOCTYPE html><html><head><title>Chatbot</title></head><body><pre style=\"font-family: monospace;\">\n" +
			"Available Routes:\n\n" +
			"  GET  /welcome -> handleWelcome\n" +
			"  POST /chat    -> handleChat\n" +
			"  GET  /        -> handle        (current)\n" +
			"</pre></body></html>"
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintln(w, body)
}

// Engage Gives control to the chatbot
func Engage(addr string) error {
	// HandleFuncs
	mux := http.NewServeMux()
	mux.HandleFunc("/welcome", withLog(handleWelcome))
	mux.HandleFunc("/chat", withLog(chat))
	mux.HandleFunc("/", withLog(handle))



	// Start the server
	return http.ListenAndServe(addr, cors.CORS(mux))
}


func main() {
	// Uncomment the following lines to customize the chatbot
	// chatbot.WelcomeMessage = "What's your name?"
	// chatbot.ProcessFunc(chatbotProcess)

	// Use the PORT environment variable
	port := os.Getenv("PORT")
	// Default to 3000 if no PORT environment variable was defined
	if port == "" {
		port = "3000"
	}


	// Start the server
	fmt.Printf("Listening on port %s...\n", port)
	log.Fatalln(Engage(":" + port))
}

package tasklistAPI

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"
	//"strings"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

type (

	// JSON Holds a JSON object
	JSON map[string]interface{}
)

var client *http.Client

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func SendAuthURL() string {
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Println(authURL)
	return "Go to the following link in your browser then type the " +
		"authorization code in the form token: `write code here` </br>" + authURL

}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("tasks-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func CreateTask(title string, notes string, due string) (string, error) {
	fmt.Println("In create task")
	fmt.Println(title)
	fmt.Println(due)
	fmt.Println(notes)
	taskapi, err := tasks.New(client)
	if err != nil {
		log.Fatalf("Unable to create Tasks service: %v", err)
	}

	tasklistId, err := GetTaskList()
	if err != nil {
		log.Fatalf("Unable to get tasklist: %v", err)
	}

	// Make sure a message key is defined in the body of the request
	if len(title) == 0 {
		return "Missing title in body.", nil
	}

	if len(due) == 0 {
		return "Missing due date in body.", nil
	}

	date, err := time.Parse("2/1/2006", due)
	fmt.Println(date)
	if err != nil {
		return "wrong date format", nil
	}
	now := time.Now()
	if date.Before(now) && !(date.Day() == now.Day() && date.Month() == now.Month() && date.Year() == now.Year()) {
		return "Invalid date", nil
	}

	newformat := date.Format("2006-01-02T15:04:05Z")
	task, err := taskapi.Tasks.Insert(tasklistId, &tasks.Task{
		Title: title,
		Notes: notes,
		Due:   newformat,
	}).Do()

	if err != nil {
		return "Error in inserting the task", nil
	}
	// fmt.Printf("Got task, err: %#v, %v", task, err)

	comp := "No"
	if task.Completed != nil {
		comp = "Yes"
	}

	newformat1 := ""

	if len(task.Due) > 0 {
		date, _ := time.Parse("2006-01-02T15:04:05Z", task.Due)
		if err != nil {
			return "", err
		}
		newformat1 = date.Format("Mon 02/01/2006")
	}

	message := "Task inserted. \n"
	message += "Title: " + task.Title + "\n" +
		"Notes: " + task.Notes + "\n" +
		"Due: " + newformat1 + "\n" +
		"Completed: " + comp + "\n"

	return message, nil
}

func UpdateTask(taskNumber string, title string, notes string, due string) (string, error) {
	fmt.Println("In update task")
	fmt.Println(title)
	fmt.Println(due)
	fmt.Println(notes)
	taskapi, err := tasks.New(client)
	if err != nil {
		log.Fatalf("Unable to update Tasks service: %v", err)
	}

	tasklistId, err := GetTaskList()
	if err != nil {
		log.Fatalf("Unable to get tasklist: %v", err)
	}
	taskIndex, err := strconv.Atoi(taskNumber)

	if err != nil {
		return "Invalid index", nil
	}

	tasksarr, err := taskapi.Tasks.List(tasklistId).Do()

	if len(tasksarr.Items) < taskIndex {
		return "Invalid task number", nil
	}

	taskId := tasksarr.Items[taskIndex].Id

	updatedTitle := tasksarr.Items[taskIndex].Title
	updatedNotes := tasksarr.Items[taskIndex].Notes
	updatedDue := tasksarr.Items[taskIndex].Due

	if len(title) != 0 {
		updatedTitle = title
	}

	if len(due) != 0 {
		date, err := time.Parse("2/1/2006", due)
		fmt.Println(date)
		if err != nil {
			return "wrong date format", nil
		}
		now := time.Now()
		fmt.Println((date.Day() == now.Day() && date.Month() == now.Month() && date.Year() == now.Year()))
		if date.Before(now) && !(date.Day() == now.Day() && date.Month() == now.Month() && date.Year() == now.Year()) {
			return "", fmt.Errorf("Invalid date")
		}
		newformat := date.Format("2006-01-02T15:04:05Z")
		updatedDue = newformat
	}

	if len(notes) != 0 {
		updatedNotes = notes
	}

	task, err := taskapi.Tasks.Patch(tasklistId, taskId, &tasks.Task{
		Title: updatedTitle,
		Due:   updatedDue,
		Notes: updatedNotes,
	}).Do()

	fmt.Printf("Got task, err: %#v, %v", task, err)
	if err != nil {
		return "Error in updating task", nil
	}

	comp := "No"
	if task.Completed != nil {
		comp = "Yes"
	}

	newformat := ""

	if len(task.Due) > 0 {
		date, _ := time.Parse("2006-01-02T15:04:05Z", task.Due)
		if err != nil {
			return "", err
		}
		newformat = date.Format("Mon 02/01/2006")
	}

	message := "Task is updated \n"
	message += "Task Number: " + taskNumber + "\n" +
		"Title: " + task.Title + "\n" +
		"Notes: " + task.Notes + "\n" +
		"Due: " + newformat + "\n" +
		"Completed: " + comp + "\n"

	return message, nil

}

func DeleteTask(index string) (string, error) {
	//index := strings.TrimPrefix(r.URL.Path, "/delete/")
	taskIndex, err := strconv.Atoi(index)

	if err != nil {
		return "", fmt.Errorf("Invalid index")
	}

	srv, err := tasks.New(client)
	if err != nil {
		log.Fatalf("Unable to create Tasks service: %v", err)
	}

	tasklistId, err := GetTaskList()

	if err != nil {
		log.Fatalf("Unable to get tasklist: %v", err)
	}
	tasks, err := srv.Tasks.List(tasklistId).Do()

	if len(tasks.Items) < taskIndex || taskIndex < 0 {
		return "Invalid task number", nil
	}

	taskId := tasks.Items[taskIndex].Id
	err = srv.Tasks.Delete(tasklistId, taskId).Do()

	if err != nil {
		return "Unable to delete task", nil
	}
	return "Task is deleted", nil

}
func TaskCompleted(index string) (string, error) {
	//index := strings.TrimPrefix(r.URL.Path, "/delete/")
	taskIndex, err := strconv.Atoi(index)

	if err != nil {
		return "Invalid task number.", nil
	}

	srv, err := tasks.New(client)
	if err != nil {
		log.Fatalf("Unable to create Tasks service: %v", err)
	}

	tasklistId, err := GetTaskList()

	if err != nil {
		log.Fatalf("Unable to get tasklist: %v", err)
	}
	taskarr, err := srv.Tasks.List(tasklistId).Do()

	if len(taskarr.Items) < taskIndex || taskIndex < 0 {
		return "Invalid task number.", nil
	}

	taskId := taskarr.Items[taskIndex].Id
	task, err := srv.Tasks.Patch(tasklistId, taskId, &tasks.Task{
		Status: "completed",
	}).Do()
	fmt.Println(task)
	if err != nil {
		return "Error in updating task.", nil
	} else {
		comp := "No"
		if task.Completed != nil {
			comp = "Yes"
		}

		newformat := ""

		if len(task.Due) > 0 {
			date, _ := time.Parse("2006-01-02T15:04:05Z", task.Due)
			if err != nil {
				return "Oops! Something went wrong.", nil
			}
			newformat = date.Format("Mon 02/01/2006")
		}

		message := "Task is updated \n"
		message += "Task Number: " + index + "\n" +
			"Title: " + task.Title + "\n" +
			"Notes: " + task.Notes + "\n" +
			"Due: " + newformat + "\n" +
			"Completed: " + comp + "\n"

		return message, nil
	}

}

func PostCode(token string, refreshtoken string, date string) (string, error) {
	fmt.Println("in post code")
	ctx := context.Background()
	//
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	//
	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// tok, err := config.Exchange(oauth2.NoContext, token)
	// if err != nil {
	//   return "",err
	// }
	// fmt.Println("token")
	// yyyy-MM-dd'T'HH:mm:ss.SSSZ
	// date1, err := time.Parse(time.RFC3339Nano, date)

	// if err != nil {
	// 	fmt.Println(err)
	// 	return "", err
	// }

	tok := oauth2.Token{AccessToken: token}
	fmt.Printf("%#v\n", tok)
	client = config.Client(ctx, &tok)

	name, err := username()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	welcomeMessage := "Welcome " + name + " to Todo Brownies.\n To create a brownie click here http://allrecipes.com/recipe/10549/best-brownies/  \n But if you are super busy, super important and want to organize your tasks here you go."

	fmt.Println(welcomeMessage)
	return welcomeMessage, nil
}

// writeJSON Writes the JSON equivilant for data into ResponseWriter w
func writeJSON(w http.ResponseWriter, data JSON) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func GetTasks() (string, error) {
	tasklistId, err := GetTaskList()
	if err != nil {
		return "", err
	}

	srv, err := tasks.New(client)
	tasks, err := srv.Tasks.List(tasklistId).Do()
	if err != nil {
		return "", err
	}

	if len(tasks.Items) > 0 {
		//arr := make([]JSON,x)
		message := ""
		for c, i := range tasks.Items {
			t := strconv.Itoa(c)
			comp := "No"
			if i.Completed != nil {
				comp = "Yes"
			}

			newformat := ""

			if len(i.Due) > 0 {
				date, _ := time.Parse("2006-01-02T15:04:05Z", i.Due)
				if err != nil {
					return "", err
				}
				newformat = date.Format("Mon 02/01/2006")
			}

			message += "Task Number: " + t + "\n" +
				"Title: " + i.Title + "\n" +
				"Notes: " + i.Notes + "\n" +
				"Due: " + newformat + "\n" +
				"Completed: " + comp + "\n\n"
		}
		return message, nil
	} else {
		return "No tasks", nil
	}

}

func GetTaskList() (string, error) {
	srv, err := tasks.New(client)

	if err != nil {
		return "", fmt.Errorf("Unable to retrieve tasks Client %v!", err)
	}

	r, err := srv.Tasklists.List().MaxResults(1).Do()
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve task lists %v!", err)
	}

	fmt.Println("Task Lists:")
	if len(r.Items) > 0 {
		return r.Items[0].Id, nil
	} else {
		return "", fmt.Errorf("Task list not found")
	}
}

func username() (string, error) {
	srv, err := tasks.New(client)

	if err != nil {
		return "", fmt.Errorf("Unable to retrieve tasks Client %v!", err)
	}

	r, err := srv.Tasklists.List().MaxResults(1).Do()
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve task lists %v!", err)
	}

	fmt.Println("Task Lists:")
	if len(r.Items) > 0 {
		title := r.Items[0].Title
		title = title[:len(title)-7]
		return title, nil
	} else {
		return "", fmt.Errorf("Task list not found")
	}
}

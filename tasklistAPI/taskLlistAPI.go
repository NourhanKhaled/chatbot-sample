package tasklistAPI

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "os"
  "os/user"
  "path/filepath"

  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/tasks/v1"

)

type (

	// JSON Holds a JSON object
	JSON map[string]interface{}

)

var client *http.Client

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config, tok *oauth2.Token) *http.Client {

 // cacheFile, err := tokenCacheFile()
 //  if err != nil {
 //    log.Fatalf("Unable to get path to cached credential file. %v", err)
 //  }
 // tok, err := tokenFromFile(cacheFile)
 // if err != nil {
 //   getTokenFromWeb(config, w)
 //   saveToken(cacheFile, tok)
 // }
  fmt.Println("22111")
  return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func sendAuthURL(config *oauth2.Config, w http.ResponseWriter)  {
  fmt.Println("1111")

  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  fmt.Fprintln(w, "Go to the following link in your browser then type the "+
    "authorization code: " + authURL + " \n%v\n")
  return

  // var code string
  // if _, err := fmt.Scan(&code); err != nil {
  //   log.Fatalf("Unable to read authorization code %v", err)
  // }
  //
  // tok, err := config.Exchange(oauth2.NoContext, code)
  // if err != nil {
  //   log.Fatalf("Unable to retrieve token from web %v", err)
  // }
  // return tok
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

func CreateTask(w http.ResponseWriter, r *http.Request) {

  if r.Method == http.MethodPost {
    taskapi, err := tasks.New(client)
  	if err != nil {
  		log.Fatalf("Unable to create Tasks service: %v", err)
  	}

    tasklistId,err := GetTaskList()
    if err != nil {
  		log.Fatalf("Unable to get tasklist: %v", err)
  	}

    data := JSON{}
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
      http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
      return
    }
    defer r.Body.Close()

    // Make sure a message key is defined in the body of the request
    title, titleFound := data["title"]
    if !titleFound {
      http.Error(w, "Missing token key in body.", http.StatusBadRequest)
      return
    }

    due, dueFound := data["due"]
    if !dueFound {
      http.Error(w, "Missing token key in body.", http.StatusBadRequest)
      return
    }

    // date, err := time.Parse("06/1/2006 04:05", due.(string))
    // fmt.Println(date)
    // if err != nil {
    //   http.Error(w, "wrong date format", http.StatusBadRequest)
    //   return
    // }

  	task, err := taskapi.Tasks.Insert(tasklistId, &tasks.Task{
  		Title: title.(string),
  		Notes: data["notes"].(string),
  		Due:   due.(string),
  	}).Do()
  	fmt.Printf("Got task, err: %#v, %v", task, err)
  }else{
    http.Error(w, "Only POST requests are allowed.", http.StatusMethodNotAllowed)
  }
}

func PostCode(w http.ResponseWriter, r *http.Request) {
  ctx := context.Background()

  b, err := ioutil.ReadFile("client_secret.json")
  if err != nil {
    log.Fatalf("Unable to read client secret file: %v", err)
  }

  config, err := google.ConfigFromJSON(b, tasks.TasksScope)
  if err != nil {
    log.Fatalf("Unable to parse client secret file to config: %v", err)
  }

  cacheFile, err := tokenCacheFile()
   if err != nil {
     log.Fatalf("Unable to get path to cached credential file. %v", err)
   }

  if r.Method == http.MethodGet {
    tok, err := tokenFromFile(cacheFile)
    if err != nil {
      sendAuthURL(config, w)
    } else {
      client = getClient(ctx, config, tok)
    }
  }

  if r.Method == http.MethodPost {
    data := JSON{}
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
      http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
      return
    }
    defer r.Body.Close()

    // Make sure a message key is defined in the body of the request
    token, tokenFound := data["token"]
    if !tokenFound {
      http.Error(w, "Missing token key in body.", http.StatusBadRequest)
      return
    }

    tok, err := config.Exchange(oauth2.NoContext, token.(string))
    if err != nil {
      log.Fatalf("Unable to retrieve token from web %v", err)
    }

     client = getClient(ctx, config, tok)
     saveToken(cacheFile, tok)

    fmt.Println(tok)
    fmt.Println(client)
  }
}

// writeJSON Writes the JSON equivilant for data into ResponseWriter w
func writeJSON(w http.ResponseWriter, data JSON) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
  if r.Method == http.MethodGet {
    tasklistId, err := GetTaskList()
    fmt.Println(tasklistId)
    if(err != nil) {
      log.Fatalf("Unable to read task list: %v", err)
      return
    }

    srv, err := tasks.New(client)

    tasks, err := srv.Tasks.List(tasklistId).Do()
    if err != nil {
      log.Fatalf("Unable to retrieve tasks %v!", err)
    }
    fmt.Println("Type")
    // fmt.Println(reflect.TypeOf(tasks))
    // writeJSON(w, tasks)
    if len(tasks.Items) > 0 {
      x := len(tasks.Items)
      arr := make([]JSON,x)
      for c, i := range tasks.Items {
        arr[c] = JSON {
          "index": c,
          "title": i.Title,
          "updated": i.Updated,
          "notes": i.Notes,
          "status": i.Status,
          "due": i.Due,
          "completed": i.Completed,
      	}
      }
      writeJSON(w,JSON{
        "tasks": arr,
      })
    } else {
      fmt.Fprintln(w, "no tasks")
    }
  }else{
    http.Error(w, "Only GET requests are allowed.", http.StatusMethodNotAllowed)
  }
}

func GetTaskList() (string,error) {
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

func Main() {

  // ctx := context.Background()
  //
  // b, err := ioutil.ReadFile("client_secret.json")
  // if err != nil {
  //   log.Fatalf("Unable to read client secret file: %v", err)
  // }
  //
  // // If modifying these scopes, delete your previously saved credentials
  // // at ~/.credentials/tasks-go-quickstart.json
  // config, err := google.ConfigFromJSON(b, tasks.TasksScope)
  // if err != nil {
  //   log.Fatalf("Unable to parse client secret file to config: %v", err)
  // }
  //

  // // CreateTask(client)


}

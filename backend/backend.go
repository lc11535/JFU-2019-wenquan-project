package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const (
	serverAddr     = "127.0.0.1"
	port           = "80"
	adminPrivilege = 1
	userPrivilege  = 2
)

const (
	dbUser = "gay"
	dbPwd  = "12341234"
	dbPort = "8302"
	dbName = "project"
)

const (
	DB_CONNECTION_FAILED       = 600
	QUERY_COMPILICATION_FAILED = 601
	DUPLICATE_NAME             = 602
	INSERTION_ERROR            = 603
	SELECTION_ERROR            = 604
	USER_NOT_EXIST             = 605

	UNKNOWN_DB_ERROR = 700
)

const (
	JSON_HANDLE_ERROR = 800
	POTENTIAL_HACK    = 801
)

var (
	//register SESSION_KEY environment var before running code on server
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
)

type userInfo struct {
	Alias   string
	Pwd     string
	School  string
	Academy string
	Major   string
	Grade   string
	Auth    int
}

type wrappedInfo struct {
	Alias     string
	School    string
	Academy   string
	Major     string
	Grade     string
	Auth      int
	SessionID string
}

type postBack struct {
	Status int
	Data   interface{}
	Msg    string
}

type singleComment struct {
	Content     string
	CommentTime string
	Commentator string
}

type singleNotification struct {
	Title      string
	UploadTime string
	Content    string
}

type fileMetaInfo struct {
	Filename string
	School string
	Academy string
	Major string
	Course string
	Uploader string
	UploadTime string
	Pointer string
	Auth int
	Description string
	Size int64
	DownloadTime int
}

func createHandle(w http.ResponseWriter, r *http.Request) {

	//Assume it's a POST request
	curUser := userInfo{
		r.PostFormValue("username"),
		r.PostFormValue("password"),
		r.PostFormValue("school"),
		r.PostFormValue("academy"),
		r.PostFormValue("major"),
		r.PostFormValue("grade"),
		userPrivilege,
	}

	var backWrapper postBack

	if curUser.Alias == "" || curUser.Pwd == "" || curUser.School == "" || curUser.Academy == "" || curUser.Major == "" || curUser.Grade == "" {
		backWrapper = postBack{POTENTIAL_HACK, nil, "fucking bitch"}
	} else if err := insertUser(curUser); err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "duplicate") {
			//duplicate name
			backWrapper = postBack{DUPLICATE_NAME, nil, s}
		} else if strings.Contains(s, "insertion error") {
			//insertion failure
			backWrapper = postBack{INSERTION_ERROR, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		backWrapper = postBack{http.StatusOK, nil, "create user successful"}
	}

	data, err := json.Marshal(backWrapper)
	if err != nil {
		//unable to wrap as json, so write http.Error directly
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
	} else {
		w.Write(data)
	}
}

func loginHandle(w http.ResponseWriter, r *http.Request) {
	curUser := userInfo{
		r.PostFormValue("username"),
		r.PostFormValue("password"),
		"dummy",
		"dummy",
		"dummy",
		"dummy",
		userPrivilege,
	}
	pwd := curUser.Pwd

	storedPwd, err := retrievePwd(&curUser)

	var backWrapper postBack
	var data []byte

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "selection error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else if strings.Contains(s, "user not exist") {
			//user not exist
			backWrapper = postBack{USER_NOT_EXIST, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else if pwd != storedPwd {
		backWrapper = postBack{http.StatusForbidden, nil, "password authentication failed"}
	} else {
		//create session
		session, err := store.Get(r, "sess")
		session.ID = createPasswd()
		session.Options.MaxAge = 3600
		if err != nil {
			backWrapper = postBack{http.StatusInternalServerError, nil, err.Error()}
		} else {
			session.Values["validUser"] = true
			session.Save(r, w)

			//Send user info to the frontend
			wrapped := wrappedInfo{
				curUser.Alias,
				curUser.School,
				curUser.Academy,
				curUser.Major,
				curUser.Grade,
				curUser.Auth,
				session.ID,
			}
			backWrapper.Status = http.StatusOK
			backWrapper.Data = wrapped
			backWrapper.Msg = "login successful"

		}
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
		return
	}
	w.Write(data)
}

func schoolListHandle(w http.ResponseWriter, r *http.Request) {
	schoolList, err := retrieveSchoolList()

	var backWrapper postBack
	var data []byte

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "selection error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		//Send user info to the frontend
		backWrapper.Status = http.StatusOK
		backWrapper.Data = schoolList
		backWrapper.Msg = "get school list successful"
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
		return
	}
	w.Write(data)
}

func acaMajorHandle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	targetSchool := query["school"][0]
	res, err := retrieveAcademyAndMajor(targetSchool)

	var backWrapper postBack
	var data []byte

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "select error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		//Send user info to the frontend
		backWrapper.Status = http.StatusOK
		backWrapper.Data = res
		backWrapper.Msg = "get academy and major list successful"
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
		return
	}
	w.Write(data)
}

func fileListHandle(w http.ResponseWriter, r *http.Request) {
	//Retrieve [filename, school, academy, major, course, upload_date, auth]
	//auth == 0 represents public, auth == 1 represents school private
	//UploadDate is a string like "2019/5/1", "2019/12/11"
	fileList, err := retrieveFileList()

	var backWrapper postBack
	var data []byte

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "selection error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		//Send user info to the frontend
		backWrapper.Status = http.StatusOK
		backWrapper.Data = fileList
		backWrapper.Msg = "get file list successful"
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
		return
	}
	w.Write(data)
}

func commentHandle(w http.ResponseWriter, r *http.Request) {
	//front end give me a signle file name
	//get file id from table file_info
	//then select all comments from table comments
	//write all comments into writer
	//front-end give me "filename" via GET method
	//comments should be witnin 200 ASCI characters
	query := r.URL.Query()
	targetFile := query["filename"][0]
	res, err := retrieveComments(targetFile)

	var backWrapper postBack
	var data []byte

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "selection error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		//Send user info to the frontend
		backWrapper.Status = http.StatusOK
		backWrapper.Data = res
		backWrapper.Msg = "get comments successful"
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
		return
	}
	w.Write(data)
}

func notificationHandle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	targetSchool := query["school"][0]
	res, err := retrieveNotifications(targetSchool)

	var backWrapper postBack
	var data []byte

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "select error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		//Send user info to the frontend
		backWrapper.Status = http.StatusOK
		backWrapper.Data = res
		backWrapper.Msg = "get notifications successful"
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
		return
	}
	w.Write(data)
}

func uploadHandle(w http.ResponseWriter, r *http.Request) {
	//front-end give me some infos:
	//[file_name, academy, major, course, description, uploader, upload_date]
	//Alloc one path, and store the path string as "pointer"
	//open an iowriter in that path
	//write the content of file to that place
	//insert one record for this file
	//return some status
	//here, auth 0 means private, auth 1 means public
	fuck := r.PostFormValue("auth")
	tmpint, _ := strconv.Atoi(fuck)
	curFile := fileMetaInfo{
		r.PostFormValue("filename"),
		r.PostFormValue("school"),
		r.PostFormValue("academy"),
		r.PostFormValue("major"),
		r.PostFormValue("course"),
		r.PostFormValue("uploader"),
		r.PostFormValue("upload-time"),
		"dummy",
		tmpint,
		r.PostFormValue("description"),
		0,
		0,
	}
	fmt.Println(curFile.Filename)
	filename := curFile.School + "-" + curFile.Academy + "-" + curFile.Major + "-" + curFile.Filename
	curFile.Pointer = "files/" + filename

	file, _, err := r.FormFile("uploaded-file")

	var backWrapper postBack
	var data []byte

	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fw, err := os.Create(curFile.Pointer)
	if err != nil {
		fmt.Println("create file failed")
		fmt.Println(err)
		return
	}
	defer fw.Close()
	_, err = io.Copy(fw, file)
	if err != nil {
		fmt.Println("copy file failed")
		fmt.Println(err)
		return
	}
	info, err := os.Stat("files/" + filename)
	curFile.Size = info.Size()
	curFile.Filename = filename
	err = insertFileInfo(curFile)
	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "duplicate") {
			//duplicate name
			backWrapper = postBack{DUPLICATE_NAME, nil, s}
		} else if strings.Contains(s, "insertion error") {
			//insertion failure
			backWrapper = postBack{INSERTION_ERROR, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		backWrapper = postBack{http.StatusOK, curFile, "insert file info successful"}
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		//unable to wrap as json, so write http.Error directly
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
	} else {
		w.Write(data)
	}
}

func downloadHandle(w http.ResponseWriter, r *http.Request) {
	//front-end give me a file name via get method
	//use this name to select from database for pointer
	//use pointer to create a reader
	//write into ResponseWriter via reader
	//add one download time
	query := r.URL.Query()
	targetName := query["filename"][0]
	ptr, err := retrievePtr(targetName)

	var backWrapper postBack

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "select error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else if strings.Contains(s, "file not exists") {
			backWrapper = postBack{http.StatusNotFound, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		//Send user info to the frontend
		backWrapper.Status = http.StatusOK
		backWrapper.Msg = "get file successful"

		backWrapper.Data = []byte(targetName)
		addDownloadTime(targetName)
	}

	http.ServeFile(w, r, ptr)
}

func postCommentHandle(w http.ResponseWriter, r *http.Request) {
	targetName := r.PostFormValue("filename")
	id, err := retrieveID(targetName)

	var backWrapper postBack
	var data []byte

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "select error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else if strings.Contains(s, "file not exists") {
			backWrapper = postBack{http.StatusNotFound, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		var curComment singleComment
		curComment.Content = r.PostFormValue("content")
		curComment.Commentator = r.PostFormValue("poster")
		curComment.CommentTime = r.PostFormValue("post-time")
		err = insertComment(curComment, id)
		backWrapper.Data = curComment
		backWrapper.Status = http.StatusOK
		backWrapper.Msg = "insert comment successful"
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		//unable to wrap as json, so write http.Error directly
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
	} else {
		w.Write(data)
	}
}

func exactListHandle(w http.ResponseWriter, r *http.Request) {
	//Retrieve [filename, school, academy, major, course, upload_date, auth]
	//auth == 0 represents public, auth == 1 represents school private
	//UploadDate is a string like "2019/5/1", "2019/12/11"
	school := r.PostFormValue("school")
	academy := r.PostFormValue("academy")
	major := r.PostFormValue("major")

	fileList, err := retrieveExactFileList(school, academy, major)

	var backWrapper postBack
	var data []byte

	if err != nil {
		s := fmt.Sprint(err)
		if strings.Contains(s, "open database error") {
			//connection failed
			backWrapper = postBack{DB_CONNECTION_FAILED, nil, s}
		} else if strings.Contains(s, "compile") {
			//compile error
			backWrapper = postBack{QUERY_COMPILICATION_FAILED, nil, s}
		} else if strings.Contains(s, "selection error") {
			//insertion failure
			backWrapper = postBack{SELECTION_ERROR, nil, s}
		} else {
			//unknown error
			backWrapper = postBack{UNKNOWN_DB_ERROR, nil, s}
		}
	} else {
		//Send user info to the frontend
		backWrapper.Status = http.StatusOK
		backWrapper.Data = fileList
		backWrapper.Msg = "get exact file list successful"
	}

	data, err = json.Marshal(backWrapper)
	if err != nil {
		http.Error(w, err.Error(), JSON_HANDLE_ERROR)
		return
	}
	w.Write(data)
}

func insertComment(cm singleComment, id int) error {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	//failed connection
	if err := dbInstance.Ping(); err != nil {
		return errors.New("open database error")
	}

	//Attempt to insert
	insertStatement, err := dbInstance.Prepare("INSERT INTO comments VALUES(?, ?, ?, ?)")
	if err != nil {
		return errors.New("compile sql insert statement error")
	}

	_, err = insertStatement.Query(id, cm.CommentTime, cm.Commentator, cm.Content)
	if err != nil {
		return errors.New("insertion error")
	}

	return nil
}

func insertUser(cu userInfo) error {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	//failed connection
	if err := dbInstance.Ping(); err != nil {
		return errors.New("open database error")
	}

	//check duplicate
	selectStatement, err := dbInstance.Prepare("SELECT COUNT(*) FROM user_info WHERE alias LIKE '?'")
	if err != nil {
		return errors.New("compile sql select statement error")
	}

	//If the SELECT result is not empty, then this name is already in use
	rows, err := selectStatement.Query(cu.Alias)
	if rows != nil {
		return errors.New("duplicate alias is used")
	}

	//Attempt to insert
	insertStatement, err := dbInstance.Prepare("INSERT INTO user_info VALUES(?, ?, ?, ?, ?, ?, 2, 0)")
	if err != nil {
		return errors.New("compile sql insert statement error")
	}

	_, err = insertStatement.Query(cu.Alias, cu.Pwd, cu.School, cu.Academy, cu.Major, cu.Grade)
	if err != nil {
		return errors.New("insertion error")
	}

	return nil
}

func insertFileInfo(info fileMetaInfo) (e error) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	//failed connection
	if err := dbInstance.Ping(); err != nil {
		return errors.New("open database error")
	}

	//Attempt to insert
	insertStatement, err := dbInstance.Prepare("INSERT INTO file_store VALUES(?, ?, ?, ?, ?, ?, ?, ?, 0, ?, 0, ?, ?);")
	if err != nil {
		return errors.New("compile sql insert statement error")
	}

	_, err = insertStatement.Query(info.Filename, info.Pointer, info.Academy, info.Major, info.Course, info.Description, info.Auth, info.Uploader, info.UploadTime, info.Size, info.School)
	if err != nil {
		return errors.New("insertion error")
	}

	return nil
}

//side effect: fill other empty fields of curUser
func retrievePwd(curUser *userInfo) (s string, e error) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	statement, err := dbInstance.Prepare("SELECT * FROM user_info WHERE alias = ?")
	if err != nil {
		e = errors.New("compile sql statement error")
		return
	}

	rows, err := statement.Query(curUser.Alias)
	if err != nil {
		e = errors.New("select error")
		return
	}
	if rows == nil {
		e = errors.New("user not exist")
		return
	}

	for rows.Next() {
		var dummy1 int
		var dummy2 int
		err = rows.Scan(
			&curUser.Alias,
			&curUser.Pwd,
			&curUser.School,
			&curUser.Academy,
			&curUser.Major,
			&curUser.Grade,
			&dummy1, &dummy2,
		)
		if err != nil {
			e = err
			return
		}

		s = curUser.Pwd
	}

	return
}

func retrieveSchoolList() (s []string, e error) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	statement, err := dbInstance.Prepare("SELECT DISTINCT school FROM school_info")
	if err != nil {
		e = errors.New("compile sql statement error")
		return
	}

	rows, err := statement.Query()
	if err != nil {
		e = errors.New("select error")
		return
	}

	for rows.Next() {
		var curSchool string

		err = rows.Scan(&curSchool)
		if err != nil {
			e = err
			return
		}

		s = append(s, curSchool)
	}

	return
}

func retrieveFileList() (s []fileMetaInfo, e error) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	statement, err := dbInstance.Prepare("SELECT filename, school, file_store.academy, file_store.major, course, file_store.auth, upload_date, size, description, download_time FROM file_store;")
	if err != nil {
		e = errors.New("compile sql statement error")
		return
	}

	rows, err := statement.Query()
	if err != nil {
		e = errors.New("select error")
		return
	}

	for rows.Next() {
		var curFile fileMetaInfo

		err = rows.Scan(&curFile.Filename, &curFile.School, &curFile.Academy, &curFile.Major, &curFile.Course, &curFile.Auth, &curFile.UploadTime, &curFile.Size, &curFile.Description, &curFile.DownloadTime)
		if err != nil {
			e = err
			return
		}

		s = append(s, curFile)
	}

	return
}

func retrieveAcademyAndMajor(school string) (s map[string][]string, e error) {
	schoolList := make([]string, 0)

	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	academyStatement, err := dbInstance.Prepare("SELECT DISTINCT academy FROM school_info WHERE school = ?")
	if err != nil {
		e = errors.New("compile sql statement error")
		return
	}

	rows, err := academyStatement.Query(school)
	if err != nil {
		e = errors.New("select error")
		return
	}

	for rows.Next() {
		var curSchool string

		err = rows.Scan(&curSchool)
		if err != nil {
			e = err
			return
		}

		schoolList = append(schoolList, curSchool)
	}

	majorStatement, err := dbInstance.Prepare("SELECT DISTINCT major FROM school_info WHERE school = ? AND academy = ?;")

	if err != nil {
		e = errors.New("compile sql statement error")
		return
	}

	s = make(map[string][]string)

	for _, academy := range schoolList {
		rows, err := majorStatement.Query(school, academy)
		if err != nil {
			e = errors.New("select error")
			return
		}
		majorList := make([]string, 0)
		for rows.Next() {
			var curMajor string
			rows.Scan(&curMajor)
			if err != nil {
				e = err
				return
			}
			majorList = append(majorList, curMajor)
		}
		s[academy] = majorList
	}

	return
}

func retrieveComments(filename string) (s []singleComment, e error) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	statement, err := dbInstance.Prepare("SELECT content, comment_time, commentator FROM comments WHERE file_id IN(SELECT id FROM file_store WHERE filename = ?);")
	if err != nil {
		e = errors.New("compile sql statement error")
		return
	}

	rows, err := statement.Query(filename)
	if err != nil {
		e = errors.New("select error")
		return
	}

	for rows.Next() {
		var curComment singleComment

		err = rows.Scan(&curComment.Content, &curComment.CommentTime, &curComment.Commentator)
		if err != nil {
			e = err
			return
		}

		s = append(s, curComment)
	}

	return
}

func retrieveNotifications(school string) (s []singleNotification, e error) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	academyStatement, err := dbInstance.Prepare("SELECT title, upload_date, content FROM notifications WHERE school = ?;")
	if err != nil {
		e = errors.New("compile sql statement error")
		return
	}

	rows, err := academyStatement.Query(school)
	if err != nil {
		e = errors.New("select error")
		return
	}

	for rows.Next() {
		var curNoti singleNotification

		err = rows.Scan(&curNoti.Title, &curNoti.UploadTime, &curNoti.Content)
		if err != nil {
			e = err
			return
		}

		s = append(s, curNoti)
	}

	return
}

func retrievePtr(filename string) (s string, e error) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	statement, err := dbInstance.Prepare("SELECT pointer FROM file_store WHERE filename = ?")
	if err != nil {
		e = errors.New("compile sql statement error")
		fmt.Println(err)
		return
	}

	rows, err := statement.Query(filename)
	if err != nil {
		e = errors.New("select error")
		return
	}

	for rows.Next() {
		err = rows.Scan(&s)
		if err != nil {
			e = err
			return
		}
	}
	if(s == "") {
		e = errors.New("File not exists")
	}

	return
}

func retrieveID(filename string) (id int, e error) {
	id = -1
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	statement, err := dbInstance.Prepare("SELECT id FROM file_store WHERE filename = ?")
	if err != nil {
		e = errors.New("compile sql statement error")
		fmt.Println(err)
		return
	}

	rows, err := statement.Query(filename)
	if err != nil {
		e = errors.New("select error")
		return
	}

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			e = err
			return
		}
	}
	if(id == -1) {
		e = errors.New("File not exists")
	}

	return
}

func retrieveExactFileList(school string, academy string, major string) (s []fileMetaInfo, e error) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	if err := dbInstance.Ping(); err != nil {
		e = errors.New("open database error")
		return
	}

	if major == "全部专业" {
		statement, err := dbInstance.Prepare("SELECT filename, school, file_store.academy, file_store.major, course, file_store.auth, upload_date, size, description, download_time FROM file_store WHERE school = ? AND academy = ?;")
		if err != nil {
			fmt.Println(err)
			e = errors.New("compile sql statement error")
			return
		}

		rows, err := statement.Query(school, academy)
		if err != nil {
			e = errors.New("select error")
			return
		}

		for rows.Next() {
			var curFile fileMetaInfo

			err = rows.Scan(&curFile.Filename, &curFile.School, &curFile.Academy, &curFile.Major, &curFile.Course, &curFile.Auth, &curFile.UploadTime, &curFile.Size, &curFile.Description, &curFile.DownloadTime)
			if err != nil {
				e = err
				return
			}

			s = append(s, curFile)
		}

		return
	} else {
		statement, err := dbInstance.Prepare("SELECT filename, school, file_store.academy, file_store.major, course, file_store.auth, upload_date, size, description, download_time FROM file_store WHERE school = ? AND academy = ? AND major = ?;")
		if err != nil {
			fmt.Println(err)
			e = errors.New("compile sql statement error")
			return
		}

		rows, err := statement.Query(school, academy, major)
		if err != nil {
			e = errors.New("select error")
			return
		}

		for rows.Next() {
			var curFile fileMetaInfo

			err = rows.Scan(&curFile.Filename, &curFile.School, &curFile.Academy, &curFile.Major, &curFile.Course, &curFile.Auth, &curFile.UploadTime, &curFile.Size, &curFile.Description, &curFile.DownloadTime)
			if err != nil {
				e = err
				return
			}

			s = append(s, curFile)
		}

		return
	}


}

func addDownloadTime(filename string) {
	loc := strings.Join([]string{dbUser, ":", dbPwd, "@/", dbName}, "")
	dbInstance, _ := sql.Open("mysql", loc)
	defer dbInstance.Close()

	statement, _ := dbInstance.Prepare("UPDATE file_store SET download_time = download_time + 1 WHERE filename = ?;")

	_, _ = statement.Query(filename)
}

func createPasswd() string {
	t := time.Now()
	h := md5.New()
	io.WriteString(h, "ylink")
	io.WriteString(h, t.String())
	passwd := fmt.Sprintf("%x", h.Sum(nil))
	return passwd
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/api/create", createHandle)
	r.HandleFunc("/api/login", loginHandle)
	r.HandleFunc("/api/school-list", schoolListHandle)
	r.HandleFunc("/api/academy-major", acaMajorHandle)
	r.HandleFunc("/api/file-list", fileListHandle)
	r.HandleFunc("/api/comment-set", commentHandle)
	r.HandleFunc("/api/notification", notificationHandle)
	r.HandleFunc("/api/upload", uploadHandle)
	r.HandleFunc("/api/download", downloadHandle)
	r.HandleFunc("/api/post-comment", postCommentHandle)
	r.HandleFunc("/api/exact-file-list", exactListHandle)
	http.Handle("/api/", r)
	http.Handle("/", http.FileServer(http.Dir("")))
	err := http.ListenAndServe(":"+port, nil)
	fmt.Println(err)
}

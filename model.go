package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	validator "gopkg.in/go-playground/validator.v9"
)

type CreateRecordStr struct {
	Record byte `validate:"required,numeric"`
}

type MathRecord struct {
	Record    rune   `validate:"required,numeric"`
	Operation string `validate:"required,max=6"`
}

type GetRecords struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ForkID    string    `json:"fork_id"`
	Freeze    bool      `json:"freeze"`
	CreatedAt time.Time `json:"created_at"`
}

type GetRecordHistroy struct {
	ID        string    `json:"id"`
	RecordID  string    `json:"recoed_id"`
	Step      rune      `json:"step"`
	Operation string    `json:"operation"`
	OpValue   string    `json:"op_value"`
	Value     rune      `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type RollBack struct {
	Step rune `validate:"required,min=1"`
}

type Records []GetRecords
type RecordHistory []GetRecordHistroy

var value rune

func OpenDB() *sql.DB {
	db, _ := sql.Open("mysql", DB_USER+":"+DB_PASS+"@tcp("+DB_HOST+":"+DB_PORT+")/"+DB_NAME+"?parseTime=true")
	return db
}

func CreateRecord(w http.ResponseWriter, r *http.Request) {

	// Set request
	var record CreateRecordStr
	_ = json.NewDecoder(r.Body).Decode(&record)

	// Get User ID
	authorizationHeader := r.Header.Get("Authorization")
	userID := Auth(authorizationHeader, "id")

	// Create UUID
	UUID := uuid.NewV4()

	// Validate
	validate = validator.New()

	err := validate.Struct(record)

	if err != nil {

		// count err
		n := len(err.(validator.ValidationErrors))
		// create array
		validateArray := make([]string, n)
		// add all err to array
		for index, err := range err.(validator.ValidationErrors) {
			validateArray[index] = err.Field() + ":Field validation for " + err.Field() + " failed on the " + err.Tag() + " tag"
		}

		mapArr := map[string]interface{}{"validationError": validateArray}
		resError, _ := json.Marshal(mapArr)

		// Return err response
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resError)
		return
	}

	// Connect DB
	db := OpenDB()

	// Count record if less than 5 return true
	var count byte
	db.QueryRow("SELECT COUNT(id) FROM `records` WHERE user_id = ?", userID).Scan(&count)

	if count >= 5 {
		// Return success response
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 400,  message": "You cannot add more than 5 record"}`)
		w.Write(resErr)
		return
	}

	// Create record
	stmtCreateRecord, err := db.Prepare("INSERT INTO records(id, user_id, created_at) VALUES(?,?, NOW())")
	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500,  message": "` + err.Error() + `"}`)
		w.Write(resErr)
		return
	}
	stmtCreateRecord.Exec(UUID, userID)
	defer db.Close()

	// Add record to history
	stmtCreateRecordHis, err := db.Prepare("INSERT INTO records_history(record_id, step, operation, value, created_at) VALUES(?,?,?,?, NOW())")
	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500,  message": "` + err.Error() + `"}`)
		w.Write(resErr)
		return
	}

	stmtCreateRecordHis.Exec(UUID, 0, "create", record.Record)
	defer db.Close()

	// Return success response
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	mapArr := map[string]interface{}{"code": 200, "data": map[string]interface{}{"step": 0, "opeation": "create", "id": UUID}}
	resSucc, _ := json.Marshal(mapArr)
	w.Write(resSucc)
	return

}

func UpdateRecord(w http.ResponseWriter, r *http.Request) {

	// Set request
	var record MathRecord
	_ = json.NewDecoder(r.Body).Decode(&record)

	// Get Record ID
	vars := mux.Vars(r)
	recordID := vars["id"]

	// Get User ID
	authorizationHeader := r.Header.Get("Authorization")
	userID := Auth(authorizationHeader, "id")

	// Create UUID
	UUID := uuid.NewV4()

	// Validate
	validate = validator.New()

	err := validate.Struct(record)

	if err != nil {

		// count err
		n := len(err.(validator.ValidationErrors))
		// create array
		validateArray := make([]string, n)
		// add all err to array
		for index, err := range err.(validator.ValidationErrors) {
			validateArray[index] = err.Field() + ":Field validation for " + err.Field() + " failed on the " + err.Tag() + " tag"
		}

		mapArr := map[string]interface{}{"validationError": validateArray}
		resError, _ := json.Marshal(mapArr)

		// Return err response
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resError)
		return
	}

	// Connect DB
	db := OpenDB()

	// Check if record exist
	var recordIDExist string
	var recordFreeze bool
	err = db.QueryRow("SELECT id, freeze FROM `records` WHERE user_id = ? AND id = ?", userID, recordID).Scan(&recordIDExist, &recordFreeze)

	if err != nil {
		// Return success response
		w.WriteHeader(404)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 404,  message": "Record not found"}`)
		w.Write(resErr)
		return
	}

	// Check if record freeze

	if recordFreeze {
		// Return success response
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 422,  message": "Record is freeze"}`)
		w.Write(resErr)
		return
	}

	// Count Record History
	var recordStep, recordValue rune
	err = db.QueryRow("SELECT step, value FROM `records_history` WHERE record_id = ? ORDER BY step DESC LIMIT 1", recordID).Scan(&recordStep, &recordValue)

	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500, "message": ` + err.Error() + `}`)
		w.Write(resErr)
		return
	}

	// Check operation
	if record.Operation == "add" {
		value = recordValue + record.Record
	} else if record.Operation == "sub" {
		value = recordValue - record.Record

	} else if record.Operation == "div" {
		value = recordValue / record.Record

	} else if record.Operation == "multi" {
		value = recordValue * record.Record
	} else {
		// Return success response
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 422, "validationError": "wrong operation"}`)
		w.Write(resErr)
		return
	}

	// Add record to history
	stmtCreateRecordHis, err := db.Prepare("INSERT INTO records_history(record_id, step, operation,op_value, value, created_at) VALUES(?,?,?,?,?, NOW())")
	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500,  message": "` + err.Error() + `"}`)
		w.Write(resErr)
		return
	}

	stmtCreateRecordHis.Exec(recordIDExist, recordStep+1, record.Operation, record.Record, value)
	defer db.Close()

	// Return success response
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	mapArr := map[string]interface{}{"code": 200, "data": map[string]interface{}{"step": recordStep + 1, "opeation": record.Operation, "opValue": record.Record, "value": value, "id": UUID}}
	resSucc, _ := json.Marshal(mapArr)
	w.Write(resSucc)
	return

}

func FreezeRecored(w http.ResponseWriter, r *http.Request) {

	// Get Record ID
	vars := mux.Vars(r)
	recordID := vars["id"]

	// Get User ID
	authorizationHeader := r.Header.Get("Authorization")
	userID := Auth(authorizationHeader, "id")

	// Connect DB
	db := OpenDB()

	// Check if record exist
	var recordIDExist string
	var freezeRes bool
	err := db.QueryRow("SELECT id, freeze FROM `records` WHERE user_id = ? AND id = ?", userID, recordID).Scan(&recordIDExist, &freezeRes)

	if err != nil {
		// Return success response
		w.WriteHeader(404)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 404,  message": "Record not found"}`)
		w.Write(resErr)
		return
	}

	// Update
	stmtUpdate, err := db.Prepare("UPDATE `records` SET freeze = ? WHERE id = ? ")
	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500,  message": "` + err.Error() + `"}`)
		w.Write(resErr)
		return
	}

	if !freezeRes {
		stmtUpdate.Exec(1, recordIDExist)
		// Return success response
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		mapArr := map[string]interface{}{"code": 200, "data": map[string]interface{}{"freeze": true, "message": "successful freeze record"}}
		resSucc, _ := json.Marshal(mapArr)
		w.Write(resSucc)
		return

	} else {
		stmtUpdate.Exec(0, recordIDExist)
		// Return success response
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		mapArr := map[string]interface{}{"code": 200, "data": map[string]interface{}{"freeze": false, "message": "successful unfreeze record"}}
		resSucc, _ := json.Marshal(mapArr)
		w.Write(resSucc)
		return
	}

	defer db.Close()

}

func GetAllRecords(w http.ResponseWriter, r *http.Request) {

	// Get User ID
	authorizationHeader := r.Header.Get("Authorization")
	userID := Auth(authorizationHeader, "id")

	db := OpenDB()

	stmtGetRows, err := db.Query("SELECT id, user_id,fork_id,freeze,created_at FROM records WHERE user_id = ?", userID)
	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500, "message": ` + err.Error() + `}`)
		w.Write(resErr)
		return
	}
	resArray := Records{}

	for stmtGetRows.Next() {
		var record GetRecords
		var ID string
		var UserID string
		var ForkID sql.NullString
		var Freeze bool
		var CreatedAt time.Time

		err = stmtGetRows.Scan(&ID, &UserID, &ForkID, &Freeze, &CreatedAt)

		record.ID = ID
		record.UserID = UserID
		record.ForkID = ForkID.String
		record.Freeze = Freeze
		record.CreatedAt = CreatedAt
		resArray = append(resArray, record) // keep

	}
	mapArr := map[string]interface{}{"code": 200, "data": resArray}

	resSucc, _ := json.Marshal(mapArr)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resSucc)

}

func GetRecordHistory(w http.ResponseWriter, r *http.Request) {

	// Get Record ID
	vars := mux.Vars(r)
	recordID := vars["id"]

	// Get User ID
	authorizationHeader := r.Header.Get("Authorization")
	userID := Auth(authorizationHeader, "id")

	db := OpenDB()

	// Check if record exist
	var recordIDExist string
	err := db.QueryRow("SELECT id FROM `records` WHERE user_id = ? AND id = ?", userID, recordID).Scan(&recordIDExist)

	if err != nil {
		// Return success response
		w.WriteHeader(404)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 404,  message": "Record not found"}`)
		w.Write(resErr)
		return
	}

	stmtGetRows, err := db.Query("SELECT records_history.id, `record_id`, `step`, `operation`, `op_value`, `value`, records.created_at FROM `records` JOIN records_history ON records_history.record_id = records.id WHERE records.user_id = ? AND records.id = ?", userID, recordID)
	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500, "message": ` + err.Error() + `}`)
		w.Write(resErr)
		return
	}
	resArray := RecordHistory{}

	for stmtGetRows.Next() {
		var recordHis GetRecordHistroy
		var ID string
		var RecordID string
		var Step rune
		var Operation string
		var OpValue sql.NullString
		var Value rune
		var CreatedAt time.Time

		err = stmtGetRows.Scan(&ID, &RecordID, &Step, &Operation, &OpValue, &Value, &CreatedAt)

		recordHis.ID = ID
		recordHis.RecordID = RecordID
		recordHis.Step = Step
		recordHis.Operation = Operation
		recordHis.OpValue = OpValue.String
		recordHis.Value = Value
		recordHis.CreatedAt = CreatedAt
		resArray = append(resArray, recordHis) // keep

	}

	mapArr := map[string]interface{}{"code": 200, "data": resArray}

	resSucc, _ := json.Marshal(mapArr)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resSucc)

}

func DeleteRecord(w http.ResponseWriter, r *http.Request) {

	// Get Record ID
	vars := mux.Vars(r)
	recordID := vars["id"]

	// Get User ID
	authorizationHeader := r.Header.Get("Authorization")
	userID := Auth(authorizationHeader, "id")

	db := OpenDB()

	// Check if record exist
	var recordIDExist string
	err := db.QueryRow("SELECT id FROM `records` WHERE user_id = ? AND id = ?", userID, recordID).Scan(&recordIDExist)

	if err != nil {
		// Return success response
		w.WriteHeader(404)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 404,  message": "Record not found"}`)
		w.Write(resErr)
		return
	}

	//Delete
	stmtDelete, err := db.Prepare("DELETE FROM `records` WHERE user_id = ? AND id = ?")
	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500, "message": ` + err.Error() + `}`)
		w.Write(resErr)
		return
	}

	stmtDelete.Exec(userID, recordID)

	mapArr := map[string]interface{}{"code": 200, "message": "Successful delete record"}

	resSucc, _ := json.Marshal(mapArr)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resSucc)

}

func RollBackRecord(w http.ResponseWriter, r *http.Request) {

	// Set request
	var record RollBack
	_ = json.NewDecoder(r.Body).Decode(&record)

	// Get Record ID
	vars := mux.Vars(r)
	recordID := vars["id"]

	// Get User ID
	authorizationHeader := r.Header.Get("Authorization")
	userID := Auth(authorizationHeader, "id")

	// Validate
	validate = validator.New()

	err := validate.Struct(record)

	if err != nil {

		// count err
		n := len(err.(validator.ValidationErrors))
		// create array
		validateArray := make([]string, n)
		// add all err to array
		for index, err := range err.(validator.ValidationErrors) {
			validateArray[index] = err.Field() + ":Field validation for " + err.Field() + " failed on the " + err.Tag() + " tag"
		}

		mapArr := map[string]interface{}{"validationError": validateArray}
		resError, _ := json.Marshal(mapArr)

		// Return err response
		w.WriteHeader(422)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resError)
		return
	}

	db := OpenDB()

	// Check If Record Exist
	var recordIDExist string
	err = db.QueryRow("SELECT records.id FROM `records` JOIN records_history ON records_history.record_id = records.id WHERE records.user_id = ? AND records.id = ? AND records_history.step = ?", userID, recordID, record.Step).Scan(&recordIDExist)

	if err != nil {
		// Return success response
		w.WriteHeader(404)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 404,  message": "Record not found"}`)
		w.Write(resErr)
		return
	}

	//Delete
	stmtDelete, err := db.Prepare("DELETE FROM `records_history` WHERE records_history.record_id = ? AND records_history.step > ?")
	if err != nil {
		// Return success response
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		resErr := []byte(`{"code": 500, "message": ` + err.Error() + `}`)
		w.Write(resErr)
		return
	}

	stmtDelete.Exec(recordID, record.Step)

	mapArr := map[string]interface{}{"code": 200, "message": "Successful rollback"}

	resSucc, _ := json.Marshal(mapArr)

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resSucc)

}

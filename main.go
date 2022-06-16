package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrOperationFlgIsNotSpecified  = errors.New("-operation flag has to be specified")
	ErrFileNameFlgIsNotSpecified   = errors.New("-fileName flag has to be specified")
	ErrItemIsEmpty                 = errors.New("-item flag has to be specified")
	ErrItemShouldntbeused          = errors.New("-item shouldn't be used")
	ErrIdIsPresent                 = errors.New("-id is used but shouldn't be")
	ErrIdFlagHastoBeSpecify        = errors.New("-id flag has to be specified")
	ErrOperationFlagHasToBeSpecify = errors.New("-operation flag has to be specified")
)

type Arguments map[string]string

// stryuct of json record
type UserRecord struct {
	Id    string
	Email string
	Age   int
}

// func for parse console options
func parseArgs() Arguments {
	var operationOption *string
	var itemOption *string
	var idOption *string
	var fileNameOption *string

	//var args Arguments = make(map[string]string)

	operationOption = flag.String("operation", "default", "add item into users list")
	itemOption = flag.String("item", "", "completely or part of users list in json format")
	idOption = flag.String("id", "", "identify of user in list")
	fileNameOption = flag.String("fileName", "isnotspecified", "full file name json file for saving users list")
	flag.Parse()

	return map[string]string{"operation": *operationOption, "item": *itemOption, "id": *idOption, "fileName": *fileNameOption}
}

func Perform(args Arguments, writer io.Writer) error {
	// analyse of args
	var errPwd error
	var currentPwd string
	var filePointer *os.File
	var errFile error
	//var user UserRecord
	//var arrayOfUserList []UserRecord
	var errJson error
	var itemBytes []byte
	var fileByte []byte
	var fileErr error

	legalOperations := [5]string{"add", "list", "findById", "remove", "default"}

	// checking file anme flag is not specified
	if len(args["fileName"]) <= 0 {
		return ErrFileNameFlgIsNotSpecified
	}
	// checking operation flag
	if len(strings.TrimSpace(args["operation"])) <= 0 {
		return ErrOperationFlagHasToBeSpecify
	}
	if args["operation"] == "default" {
		return ErrOperationFlgIsNotSpecified
	} else {
		// cjecking operation flag value there is in list of allowed flags
		triggerThereIsIntoArray := false
		for _, legalOperationVal := range legalOperations {
			if args["operation"] == legalOperationVal {
				triggerThereIsIntoArray = true
			}
		}
		if !triggerThereIsIntoArray {
			return fmt.Errorf("Operation %v not allowed!", args["operation"])
		}
	}

	// take default path
	currentPwd, errPwd = os.Getwd()

	// Normalize of path to json file
	if args["fileName"] != "isnotspecified" {
		// checking that there is directory in full file name
		pathTofile, fileName := filepath.Split(args["fileName"])
		if len(pathTofile) <= 0 && len(fileName) > 0 {
			// checking currentPwd is correct
			if errPwd != nil {
				return fmt.Errorf("Current path is not get, error: %w\n", errPwd)
			}
			// there is not directory in full file name
			args["fileName"] = concat(currentPwd, concat("/", fileName))
		} else if len(pathTofile) > 0 && len(fileName) > 0 {
			// there is directory in full file name
			// checking there is the directory
			if _, err := os.Stat(pathTofile); os.IsNotExist(err) {
				// checking currentPwd is correct
				if errPwd != nil {
					return fmt.Errorf("Current path is not get, error: %w\n", errPwd)
				}
				// there is not file path
				args["fileName"] = concat(currentPwd, concat("/", fileName))
			}
		}
	} else {
		return ErrFileNameFlgIsNotSpecified
	}

	// it's not permitten options with operaton add
	if args["operation"] == "add" {
		user := []UserRecord{}
		usersFromFile := []UserRecord{}
		// checking id flag
		if len(args["id"]) > 0 {
			// id shouldn't be exist
			return ErrIdIsPresent
		}
		// checking item flag
		if args["item"] == "" {
			// item is empty
			return ErrItemIsEmpty
		}
		// Is there a json file
		if _, err := os.Stat(args["fileName"]); os.IsNotExist(err) {
			// there is not the json file
			// create and open file
			filePointer, errFile = os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0644)
			if errFile != nil {
				return fmt.Errorf("Error is happened when to trying create file %v, error: %w\n", args["fileName"], errFile)
			}
			// So file is new that's way we can write item to file
			// unmarshal string to json
			args["item"] = concat("[", concat(args["item"], "]"))
			//fmt.Println(args["item"])
			errJson = json.Unmarshal([]byte(args["item"]), &user)
			if errJson != nil {
				return fmt.Errorf("Ummarshal json from -item %v, finished with error: %w\n", args["item"], errJson)
			}
			// Marshal record
			itemBytes, errJson = json.Marshal(&user)
			if errJson != nil {
				return fmt.Errorf("Marshal json %v to string finished with error: %w\n", user, errJson)
			}
			// write to file
			if _, err := io.WriteString(filePointer, strings.ToLower(string(itemBytes))); err != nil {
				return fmt.Errorf("Write json %v to file %v finished with error: %w\n", string(itemBytes), args["fileName"], err)
			} else {
				return nil
			}
		} else {
			// there is the json file
			// read json file
			filePointer, errFile = os.OpenFile(args["fileName"], os.O_RDONLY, 0644)
			if errFile != nil {
				return fmt.Errorf("Error is happened when to trying create file %v, error: %w\n", args["fileName"], errFile)
			}
			fileByte, fileErr = ioutil.ReadAll(filePointer)
			if fileErr != nil {
				return fmt.Errorf("Error is happened when file %v was reading, error: %w", args["fileName"], fileErr)
			}
			// close json file
			filePointer.Close()
			// Unmarshal the json from -item
			args["item"] = concat("[", concat(args["item"], "]"))
			errJson = json.Unmarshal([]byte(args["item"]), &user)
			if errJson != nil {
				return fmt.Errorf("Ummarshal json from -item %v, finished with error: %w\n", args["item"], errJson)
			}
			// Unmarshal the json from the json file
			errJson = json.Unmarshal(fileByte, &usersFromFile)
			if errJson != nil {
				return fmt.Errorf("Ummarshal json from file %v, finished with error: %w\n", args["fileName"], errJson)
			}
			// looking for id into file
			for _, valUser := range user {
				for _, valUserFromFile := range usersFromFile {
					if valUser.Id == valUserFromFile.Id {
						writer.Write([]byte(concat("Item with id ", concat(valUserFromFile.Id, " already exists"))))
						return nil
						//return fmt.Errorf("Item with id %v already exists", valUserFromFile.Id)
					}
				}
				// append new record to array
				usersFromFile = append(usersFromFile, valUser)
			}
			// create and open file
			filePointer, errFile = os.OpenFile(args["fileName"], os.O_WRONLY, 0644)
			if errFile != nil {
				return fmt.Errorf("Error is happened when to trying create file %v, error: %w\n", args["fileName"], errFile)
			}
			// So file is new that's way we can write item to file
			// Marshal record
			itemBytes, errJson = json.Marshal(&usersFromFile)
			if errJson != nil {
				return fmt.Errorf("Marshal json %v to string finished with error: %w\n", user, errJson)
			}
			// write to file
			if _, err := io.WriteString(filePointer, strings.ToLower(string(itemBytes))); err != nil {
				return fmt.Errorf("Write json %v to file %v finished with error: %w\n", string(itemBytes), args["fileName"], err)
			} else {
				return nil
			}
		}
	}

	// it's not permitten options with operaton remove, findById
	if args["operation"] == "remove" || args["operation"] == "findById" {
		user := []UserRecord{}
		if args["item"] != "" {
			// there is item
			return ErrItemShouldntbeused
		}
		if len(args["id"]) <= 0 {
			return ErrIdFlagHastoBeSpecify
		}
		// checking file
		// Is there a json file
		if _, err := os.Stat(args["fileName"]); os.IsNotExist(err) {
			// There is not the json file
			return fmt.Errorf("-fileName json file %v is not there, please\nspecify the correct path,\nor make application with -operation add,\nerror: %w", args["fileName"], err)
		} else {
			// there is the json file
			if args["operation"] == "findById" {
				// open file for read only
				filePointer, errFile = os.OpenFile(args["fileName"], os.O_RDONLY, 0644)
				if errFile != nil {
					return fmt.Errorf("Error is happened when to trying create file %v, error: %w\n", args["fileName"], errFile)
				}
			} else {
				// open file for re-write
				filePointer, errFile = os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0644)
				if errFile != nil {
					return fmt.Errorf("Error is happened when to trying create file %v, error: %w\n", args["fileName"], errFile)
				}
			}
			// read form the json file
			fileByte, fileErr = ioutil.ReadAll(filePointer)
			if fileErr != nil {
				return fmt.Errorf("Error is happened when file %v was reading, error: %w", args["fileName"], fileErr)
			}
			// close the json file
			filePointer.Close()
			// Unmarshal the json from the json file
			//fmt.Println(string(fileByte))
			errJson = json.Unmarshal(fileByte, &user)
			if errJson != nil {
				return fmt.Errorf("Ummarshal json from file %v, finished with error: %w\n", args["fileName"], errJson)
			}
			// sort through users list
			arrayOfUserListNew := []UserRecord{}
			triggerThereIsNotId := false
			for _, valUser := range user {
				// if ids are equal
				//if arrayOfUserList[valUser].id == args["id"] {
				if valUser.Id == args["id"] {
					// id is found
					if args["operation"] == "findById" {
						arrayOfUserListNew = append(arrayOfUserListNew, valUser)
						// print user
						itemBytes, errJson = json.Marshal(&valUser)
						if errJson != nil {
							return fmt.Errorf("Marshal json %v to string finished with error: %w\n", valUser, errJson)
						}
						writer.Write([]byte(strings.ToLower(string(itemBytes))))
						//fmt.Printf("{\"id\":\"%v\",\"email\":\"%v\",\"age\":%v}\n", valUser.Id, valUser.Email, valUser.Age)
						return nil
					}
					triggerThereIsNotId = true
				} else {
					arrayOfUserListNew = append(arrayOfUserListNew, valUser)
				}
			}
			// findById didn't find
			if !triggerThereIsNotId {
				//return nil
				//return fmt.Errorf("Item with id %s not found", args["id"])
				if args["operation"] == "findById" {
					return nil
				}
				if args["operation"] == "remove" {
					writer.Write([]byte(concat("Item with id ", concat(args["id"], " not found"))))
					return nil
				}

			}
			// save result to file if operation remove
			if args["operation"] == "remove" {
				// write to file
				// Marshal record
				//fmt.Println("THERE IS")
				//fmt.Println(arrayOfUserListNew)
				itemBytes, errJson = json.Marshal(arrayOfUserListNew)
				if errJson != nil {
					return fmt.Errorf("Marshal json from file %v to string finished with error: %w\n", args["fileName"], errJson)
				}
				// open the json file
				filePointer, errFile = os.OpenFile(args["fileName"], os.O_TRUNC|os.O_WRONLY, 0644)
				// clear of file content
				filePointer.Seek(0, io.SeekStart)
				filePointer.Truncate(0)
				if errFile != nil {
					return fmt.Errorf("Error is happened when to trying create file %v, error: %w\n", args["fileName"], errFile)
				}
				if _, err := io.WriteString(filePointer, strings.ToLower(string(itemBytes))); err != nil {
					return fmt.Errorf("Write json %v to file %v finished with error: %w\n", string(itemBytes), args["fileName"], err)
				} else {
					return nil
				}
			}
		}
	}

	// it's not permitten options with operaton list
	if args["operation"] == "list" {
		//user := []UserRecord{}
		if args["item"] != "" {
			// there is item
			return ErrItemShouldntbeused
		}
		if len(args["id"]) > 0 {
			// id shouldn't be exist
			return ErrIdIsPresent
		}
		// Is there a json file
		if _, err := os.Stat(args["fileName"]); os.IsNotExist(err) {
			// There is not the json file
			return fmt.Errorf("-fileName json file %v is not there, please\nspecify the correct path,\nor make application with -operation add,\nerror: %w", args["fileName"], err)
		} else {
			// there is the json file
			filePointer, errFile = os.OpenFile(args["fileName"], os.O_RDONLY, 0644)
			if errFile != nil {
				return fmt.Errorf("Error is happened when to trying create file %v, error: %w\n", args["fileName"], errFile)
			}
			// read form the json file
			fileByte, fileErr = ioutil.ReadAll(filePointer)
			if fileErr != nil {
				return fmt.Errorf("Error is happened when file %v was reading, error: %w", args["fileName"], fileErr)
			}
			// Unmarshal the json from the json file
			//errJson = json.Unmarshal(fileByte, &user)
			//if errJson != nil {
			//	return fmt.Errorf("Ummarshal json from file %v, finished with error: %w\n", args["fileName"], errJson)
			//}
		}
		writer.Write(fileByte)
		//fmt.Println(user)
		return nil
	}
	return fmt.Errorf("Unknown error")
}

// concat function
func concat(src string, applystr string) string {
	// the best of method of concat
	var concatBytes []byte
	var applyBytes []byte
	var byteVal byte

	// the best of method of concat
	concatBytes = make([]byte, len(src))
	copy(concatBytes[0:], src)
	applyBytes = make([]byte, len(applystr))
	copy(applyBytes[0:], applystr)
	for _, byteVal = range []byte(applystr) {
		concatBytes = append(concatBytes, byteVal)
	}
	return string(concatBytes)
}

func main() {

	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

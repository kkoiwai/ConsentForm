package main

import (
	"errors"
	"fmt"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"regexp"
)


//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}

// Customer Reference data. Each CUSTID has 1 CustRef_Holder in Keyvalue, where many CustRefs are stored
type CustRef struct {
	EntityId  string `json:"entity_id"`
	CustomerRef string `json:"customer_ref"`
}
type CustRef_Holder struct {
	CustRefs 	[]CustRef `json:"custrefs"`
}

type Entity struct {
	EntityId  string `json:"entity_id"`
	EntityName string `json:"entity_name"`
	EntityPublicKey string `json:"entity_public_key"`
}
type Entity_Holder struct {
	Entities []Entity `json:"entities"`
}

type CustomerData struct {
	CustomerId  string `json:"customer_id"`
	SenderId string `json:"sender_id"`
	ReceiverId string `json:"receiver_id"`
	Content string `json:"content"`
}
type CustomerData_Holder struct {
	Entries []CustomerData `json:"entries"`
}

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	return nil, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	//caller, caller_affiliation, err := t.get_caller_data(stub)

	//if err != nil { return nil, errors.New("Error retrieving caller information")}


	if function == "register_customer" {

		if len(args) != 4 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		customer_id := args[0]
		receiver_id := args[1]
		sender_id := args[2]
		json_data := args[3]

		return t.register_customer(stub, customer_id, receiver_id, sender_id, json_data)

	} else if function == "delete_customer" {

		if len(args) != 2 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		customer_id := args[0]
		sender_id := args[1]

		return t.delete_customer(stub, customer_id, sender_id)

	} else if function == "register_customer_crossref" {

		if len(args) != 3 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		customer_id := args[0]
		entity_id := args[1]
		customer_ref := args[2]

		return t.register_customer_crossref(stub, customer_id,  entity_id , customer_ref )

	} else if function == "delete_customer_crossref" {

		if len(args) != 2 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		entity_id := args[0]
		customer_ref := args[1]

		return t.delete_customer_crossref(stub,  entity_id , customer_ref )

	} else if function == "register_entity" {

		if len(args) != 3 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		entity_id := args[0]
		entity_name := args[1]
		entity_public_key := args[2]

		return t.register_entity(stub,  entity_id , entity_name, entity_public_key )

	}else if function == "delete_entity" {

		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		entity_id := args[0]

		return t.delete_entity(stub,  entity_id )
	}

	return nil, errors.New("Function of that name doesn't exist.")
}
//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	if function == "get_customer" {

		if len(args) != 2 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		customer_id := args[0]
		receiver_id := args[1]

		return t.get_customer(stub, customer_id, receiver_id)

	} else if function == "get_all" {

		if len(args) != 0 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		return t.get_all(stub)

	} else if function == "get_customer_crossref"{
		if len(args) != 2 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}

		entity_id := args[0]
		customer_ref := args[1]

		return t.get_customer_crossref(stub, entity_id, customer_ref)

	} else if function == "get_all_entities" {

		if len(args) != 0 {
		fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}
		return t.get_all_entities(stub)

	}else if function == "get_customers_by_sender_id" {

		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}
		sender_id := args[0]
		return t.get_customers_by_sender_id(stub,sender_id)

	}else if function == "get_customers_by_receiver_id" {

		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments passed"); return nil, errors.New("QUERY: Incorrect number of arguments passed")
		}
		receiver_id := args[0]
		return t.get_customers_by_receiver_id(stub,receiver_id)

	}
	return nil, errors.New("QUERY: No such function.")

}

//=================================================================================================================================
//	 Register Function
//======================================================================================================

func (t *SimpleChaincode) register_customer(stub *shim.ChaincodeStub, customer_id string, receiver_id string, sender_id string, json_data string) ([]byte, error) {


	if(!valid_key(customer_id)||!valid_key(receiver_id)||!valid_key(sender_id)){
		return nil, errors.New("Invalid arguments")
	}

	var data_key, scr_key, src_key, rsc_key, csr_key, crs_key string;
	data_key, scr_key, src_key, rsc_key, csr_key, crs_key = create_keys(customer_id, receiver_id, sender_id)
	var err error
	// register the value to KVS
	fmt.Println("[DEBUG] PutState " + data_key + " , " + json_data)
	err = stub.PutState(data_key, []byte(json_data))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	// then, create index data
	err = stub.PutState(scr_key, []byte(data_key))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}
	err = stub.PutState(src_key, []byte(data_key))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}
	err = stub.PutState(rsc_key, []byte(data_key))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}
	err = stub.PutState(csr_key, []byte(data_key))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}
	err = stub.PutState(crs_key, []byte(data_key))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	return nil, nil

}

func (t *SimpleChaincode) delete_customer(stub *shim.ChaincodeStub, customer_id string, sender_id string) ([]byte, error) {


	if(!valid_key(customer_id)||!valid_key(sender_id)){
		return nil, errors.New("Invalid arguments")
	}

	keysIter, err := stub.RangeQueryState("SCR/" + sender_id + "/" + customer_id + "/", "SCR/" + sender_id + "/" + customer_id + "/" + "|")
	if err != nil {
		return nil, errors.New("Unable to start the iterator")
	}

	defer keysIter.Close()

	for keysIter.HasNext() {
		key, _, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		customer_id, receiver_id, sender_id := parse_key(key)
		data_key, scr_key, src_key, rsc_key, csr_key, crs_key := create_keys(customer_id, receiver_id, sender_id)

		// remove the value in KVS
		err = stub.DelState(data_key)
		if err != nil {
			return nil, errors.New("Unable to delete the state")
		}

		// then, remove index data
		err = stub.DelState(scr_key)
		if err != nil {
			return nil, errors.New("Unable to delete the state")
		}
		err = stub.DelState(src_key)
		if err != nil {
			return nil, errors.New("Unable to delete the state")
		}
		err = stub.DelState(rsc_key)
		if err != nil {
			return nil, errors.New("Unable to delete the state")
		}
		err = stub.DelState(csr_key)
		if err != nil {
			return nil, errors.New("Unable to delete the state")
		}
		err = stub.DelState(crs_key)
		if err != nil {
			return nil, errors.New("Unable to delete the state")
		}
	}
	return nil, nil

}

func (t *SimpleChaincode) register_customer_crossref(stub *shim.ChaincodeStub,customer_id string, entity_id string, customer_ref string) ([]byte, error) {

	if(!valid_key(customer_id)||!valid_key(entity_id)||!valid_key(customer_ref)){
		return nil, errors.New("Invalid arguments")
	}
	// check first to see if the crossref is already registered
	ckey:="CUSTREF/"+entity_id+"/"+customer_ref
	cval, err := stub.GetState(ckey)
	if err != nil { return nil, errors.New("Error in GetState: " + err.Error())	}
	if len(cval) > 0 { //found
		return nil, errors.New("Duplicate CustRef record")
	}

	var cust_refs CustRef_Holder

	key := "CUSTID/"+customer_id
	bytes, err := stub.GetState(key)
	if err != nil { return nil, errors.New("Error in GetState: " + err.Error())	}

	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &cust_refs)
		if err != nil {
			return nil, errors.New("Corrupt CustRef record: " + err.Error() + string(bytes))
		}
		//find duplicate
		for _, ref := range cust_refs.CustRefs {
			if (ref.CustomerRef == customer_ref && ref.EntityId == entity_id) {
				return nil, errors.New("Duplicate CustRef record")
			}
		}
	}

	cust_refs.CustRefs = append(cust_refs.CustRefs, CustRef{EntityId:entity_id, CustomerRef:customer_ref})

	bytes, err = json.Marshal(cust_refs)
	if err != nil { return nil, errors.New("Error creating CustRef record") }

	err = stub.PutState(key, bytes)
	if err != nil { return nil, errors.New("Unable to put the state") }

	// register ref key
	ref_key := "CUSTREF/"+entity_id+"/"+customer_ref

	err = stub.PutState(ref_key, []byte(key))
	if err != nil { return nil, errors.New("Unable to put the state") }
	return nil, nil

}

func (t *SimpleChaincode) delete_customer_crossref(stub *shim.ChaincodeStub, entity_id string, customer_ref string) ([]byte, error) {


	if(!valid_key(entity_id)||!valid_key(customer_ref)){
		return nil, errors.New("Invalid arguments")
	}

	ckey:="CUSTREF/"+entity_id+"/"+customer_ref
	datakeyAsbytes, err := stub.GetState(ckey)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + ckey + "\"}"
		return nil, errors.New(jsonResp)
	}

	key:=string(datakeyAsbytes)

	//key := "CUSTID/"+customer_id
	bytes, err := stub.GetState(key)
	if err != nil {
		return nil, errors.New("Corrupt CustRef record / customer record not found")
	}

	var cust_refs CustRef_Holder
	err = json.Unmarshal(bytes, &cust_refs)
	if err != nil {	return nil, errors.New("Corrupt CustRef record:" + string(bytes)) }

	//find entry
	for i := len(cust_refs.CustRefs) - 1; i >= 0; i-- {
		ref:=cust_refs.CustRefs[i]
		if (ref.CustomerRef == customer_ref && ref.EntityId == entity_id) {
			// found, take this element from cust_ref
			cust_refs.CustRefs = append(cust_refs.CustRefs[:i],cust_refs.CustRefs[i+1:]...)
			break
		}
	}

	if len(cust_refs.CustRefs) == 0 {
		err = stub.DelState(key)
		if err != nil { return nil, errors.New("Unable to delete the state") }
	} else {
		bytes, err = json.Marshal(cust_refs)
		if err != nil {
			return nil, errors.New("Error creating CustRef record")
		}

		err = stub.PutState(key, bytes)
		if err != nil {
			return nil, errors.New("Unable to put the state")
		}
	}
	// delete ref key
	ref_key := "CUSTREF/"+entity_id+"/"+customer_ref

	err = stub.DelState(ref_key)
	if err != nil { return nil, errors.New("Unable to delete the state") }
	return nil, nil

}

func (t *SimpleChaincode) register_entity(stub *shim.ChaincodeStub, entity_id string, entity_name string, entity_public_key string) ([]byte, error) {

	if(!valid_key(entity_id)||len(entity_public_key)==0){
		return nil, errors.New("Invalid arguments")
	}
	ekey:= "ENTID/"+entity_id


	// check if the record already exists.
	// If exists, further check if customer data that was sent to the entity exists.
	// You can"t delete or modify the public key if such customer data exists.
	eval, err := stub.GetState(ekey)
	if err != nil { return nil, errors.New("Error in GetState: " + err.Error())	}
	if len(eval) > 0 { //found
		var entity_existed Entity
		err = json.Unmarshal(eval, &entity_existed)
		if err != nil { return nil, errors.New("Corrupt Entity record: " + err.Error() + string(eval))}
		receiver_id := entity_existed.EntityId
		keysIter, err := stub.RangeQueryState("D/" + receiver_id + "/" , "D/" + receiver_id + "/" + "|")
		if err != nil { return nil, errors.New("Unable to start the iterator")}
		defer keysIter.Close()
		if keysIter.HasNext() && entity_existed.EntityPublicKey != entity_public_key{
			return nil, errors.New("You can't modify existing entity record if customer data exists")
		}
	}

	entity_data:= Entity{ EntityId:entity_id , EntityName:entity_name , EntityPublicKey:entity_public_key }

	bytes, err := json.Marshal(entity_data)
	if err != nil { return nil, errors.New("Error creating Entity record") }

	err = stub.PutState(ekey, []byte(bytes))
	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	return nil, nil

}

func (t *SimpleChaincode) delete_entity(stub *shim.ChaincodeStub, entity_id string) ([]byte, error) {

	if(!valid_key(entity_id)){
		return nil, errors.New("Invalid arguments")
	}
	ekey:= "ENTID/"+entity_id

	// check if the record already exists.
	// If exists, further check if customer data that was sent to the entity exists.
	// You can"t delete or modify the public key if such customer data exists.
	eval, err := stub.GetState(ekey)
	if err != nil { return nil, errors.New("Error in GetState: " + err.Error())	}
	if len(eval) > 0 { //found
		var entity_existed Entity
		err = json.Unmarshal(eval, &entity_existed)
		if err != nil { return nil, errors.New("Corrupt Entity record: " + err.Error() + string(eval))}
		receiver_id := entity_existed.EntityId
		keysIter, err := stub.RangeQueryState("D/" + receiver_id + "/" , "D/" + receiver_id + "/" + "|")
		if err != nil { return nil, errors.New("Unable to start the iterator")}
		defer keysIter.Close()
		if keysIter.HasNext() {
			return nil, errors.New("You can't delete existing entity record if customer data exists")
		}
	}

	err = stub.DelState(ekey)
	if err != nil {
		return nil, errors.New("Unable to delete the state")
	}

	return nil, nil

}

//=================================================================================================================================
//	 Query functions
//=================================================================================================================================

func (t *SimpleChaincode) get_customer(stub *shim.ChaincodeStub, customer_id string, receiver_id string) ([]byte, error) {

	var entries CustomerData_Holder
	var ent CustomerData

	keysIter, err := stub.RangeQueryState("D/" + receiver_id + "/" + customer_id + "/", "D/" + receiver_id + "/" + customer_id + "/" + "~")

	if err != nil {
		return nil, errors.New("Unable to start the iterator")
	}

	defer keysIter.Close()

	for keysIter.HasNext() {
		datakeyAsbytes, dataAsBytes, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		datakey:=string(datakeyAsbytes)

		customer_id , receiver_id, sender_id := parse_key(datakey)

		ent = CustomerData{ CustomerId:customer_id , ReceiverId:receiver_id, SenderId:sender_id, Content:string(dataAsBytes)}

		entries.Entries = append(entries.Entries,ent)
	}

	bytes, err := json.Marshal(entries)
	if err != nil {
		return nil, errors.New("Error creating CustomerData record")
	}
	return []byte(bytes), nil

}

func (t *SimpleChaincode) get_customers_by_sender_id(stub *shim.ChaincodeStub, sender_id string) ([]byte, error) {

	var entries CustomerData_Holder
	var ent CustomerData

	keysIter, err := stub.RangeQueryState("SCR/"+sender_id+"/", "SCR/"+sender_id+"/~")
	if err != nil {
		return nil, errors.New("Unable to start the iterator")
	}

	defer keysIter.Close()

	for keysIter.HasNext() {
		_, datakeyAsbytes, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		datakey:=string(datakeyAsbytes)
		valAsbytes, err := stub.GetState(datakey)
		if err != nil {
			return nil, errors.New("Error getting customer data of "+datakey)
		}
		customer_id , receiver_id, sender_id := parse_key(datakey)

		ent = CustomerData{ CustomerId:customer_id , ReceiverId:receiver_id, SenderId:sender_id, Content:string(valAsbytes)}

		entries.Entries = append(entries.Entries,ent)
	}

	bytes, err := json.Marshal(entries)
	if err != nil {
		return nil, errors.New("Error creating CustomerData record")
	}
	return []byte(bytes), nil
}

func (t *SimpleChaincode) get_customers_by_receiver_id(stub *shim.ChaincodeStub, receiver_id string) ([]byte, error) {

	var entries CustomerData_Holder
	var ent CustomerData

	keysIter, err := stub.RangeQueryState("D/"+receiver_id+"/", "D/"+receiver_id+"/~")
	if err != nil {
		return nil, errors.New("Unable to start the iterator")
	}

	defer keysIter.Close()

	for keysIter.HasNext() {
		datakeyAsbytes, dataAsBytes, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		datakey:=string(datakeyAsbytes)

		customer_id , receiver_id, sender_id := parse_key(datakey)
		if customer_id == ""||receiver_id==""|| sender_id=="" {
			return nil, fmt.Errorf("parse_key operation failed: %s %s %s %s",datakey, customer_id , receiver_id , sender_id)
		}

		ent = CustomerData{ CustomerId: customer_id , SenderId:sender_id, ReceiverId:receiver_id,  Content:string(dataAsBytes)}

		entries.Entries = append(entries.Entries,ent)
	}

	bytes, err := json.Marshal(entries)
	if err != nil {
		return nil, errors.New("Error creating CustomerData record")
	}
	return []byte(bytes), nil

}

func (t *SimpleChaincode) get_all(stub *shim.ChaincodeStub) ([]byte, error) {

	result := "["

	keysIter, err := stub.RangeQueryState("", "~")
	if err != nil {
		return nil, errors.New("Unable to start the iterator")
	}

	defer keysIter.Close()

	for keysIter.HasNext() {
		key, val, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		result += " [ " + key + " , " + string(val) + " ] ,"

	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result) - 1] + "]"
	}

	return []byte(result), nil
}

func (t *SimpleChaincode) get_customer_crossref(stub *shim.ChaincodeStub, entity_id string, customer_ref string) ([]byte, error) {

	var jsonResp = ""
	key:="CUSTREF/"+entity_id+"/"+customer_ref
	datakeyAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}
	datakey:=string(datakeyAsbytes)
	valAsbytes, err := stub.GetState(datakey)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + datakey + "\"}"
		return nil, errors.New(jsonResp)
	}

	return []byte(valAsbytes), nil
}

func (t *SimpleChaincode) get_all_entities(stub *shim.ChaincodeStub) ([]byte, error) {

	var entities Entity_Holder
	var ent Entity

	keysIter, err := stub.RangeQueryState("ENTID/", "ENTID/~")
	if err != nil {
		return nil, errors.New("Unable to start the iterator")
	}

	defer keysIter.Close()

	for keysIter.HasNext() {
		_, val, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		err = json.Unmarshal(val,&ent)
		if err != nil { return nil, errors.New("Error creating Entity record") }
		entities.Entities = append(entities.Entities,ent)
	}

	bytes, err := json.Marshal(entities)
	if err != nil {
		return nil, errors.New("Error creating Entities record")
	}
	return []byte(bytes), nil
}


//=================================================================================================================================
//	 Utility functions
//=================================================================================================================================
func create_keys(customer_id string, receiver_id string, sender_id string) (data_key, scr_key, src_key, rsc_key, csr_key, crs_key string) {
	data_key = "D/" + receiver_id + "/" + customer_id + "/" + sender_id
	scr_key = "SCR/" + sender_id + "/" + customer_id + "/" + receiver_id
	src_key = "SRC/" + sender_id + "/" + receiver_id + "/" + customer_id
	rsc_key = "RSC/" + receiver_id + "/" + sender_id + "/" + customer_id
	csr_key = "CSR/" + customer_id + "/" + sender_id + "/" + receiver_id
	crs_key = "CRS/" + customer_id + "/" + receiver_id + "/" + sender_id
	return
}

func get_key(key_type string, customer_id string, receiver_id string, sender_id string) (string) {

	switch key_type {
	case "data_key" :
		return "D/" + receiver_id + "/" + customer_id + "/" + sender_id
	case "scr_key" :
		return "SCR/" + sender_id + "/" + customer_id + "/" + receiver_id
	case "src_key" :
		return "SRC/" + sender_id + "/" + receiver_id + "/" + customer_id
	case "rsc_key" :
		return "RSC/" + receiver_id + "/" + sender_id + "/" + customer_id
	case "csr_key" :
		return "CSR/" + customer_id + "/" + sender_id + "/" + receiver_id
	case "crs_key" :
		return "CRS/" + customer_id + "/" + receiver_id + "/" + sender_id
	}
	return ""
}

func convert_key(key_type string, current_key string) (string) {

	customer_id, receiver_id, sender_id := parse_key(current_key)
	return get_key(key_type, customer_id, receiver_id, sender_id)

}

func parse_key(key string) (customer_id string, receiver_id string, sender_id string) {
	str := strings.Split("/", key)
	if len(str) != 4 {
		return "", "", ""
	}

	switch str[0] {
	case "D" :
		receiver_id = str[1]; customer_id = str[2]; sender_id = str[3]
	case "SCR" :
		sender_id = str[1]; customer_id = str[2]; receiver_id = str[3]
	case "SRC" :
		sender_id = str[1]; receiver_id = str[2]; customer_id = str[3]
	case "RSC" :
		receiver_id = str[1]; sender_id = str[2]; customer_id = str[3]
	case "CSR" :
		customer_id = str[1]; sender_id = str[2]; receiver_id = str[3]
	case "CRS" :
		customer_id = str[1]; receiver_id = str[2]; sender_id = str[3]
	}
	return customer_id , receiver_id , sender_id
}

func valid_key(key string)(bool){
	match, _ :=  regexp.MatchString("\\w", key)
	return match
}


//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================
func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Chaincode: %s", err)
	}
}


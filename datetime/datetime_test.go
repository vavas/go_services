package datetime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/telemetryapp/go_log/services/database"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"testing"
)

type testDateTimeJSON struct {
	Now DT `json:"now"`
}

func TestMarshalDateTime(t *testing.T) {
	now := time.Now()
	data := testDateTimeJSON{DT{now}}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	jsonString := string(jsonBytes)
	expected := fmt.Sprintf(`{"now":%.9f}`, float64(now.UnixNano())/float64(1e9))

	if jsonString != expected {
		t.Errorf("Unexpected value: %+v, expected %+v", jsonString, expected)
	}
}

func TestUnmarshalDateTimeUnix(t *testing.T) {
	now := time.Now()
	jsonString := fmt.Sprintf(`{"now":%d}`, now.Unix())

	result := testDateTimeJSON{}

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	if result.Now.Unix() != now.Unix() {
		t.Errorf("Unexpected value: %+v, expected %+v", result.Now, now)
	}
}

func TestUnmarshalDateTimeGoString(t *testing.T) {
	now := time.Now()
	jsonString := fmt.Sprintf(`{"now":"%s"}`, now.Local().String())

	result := testDateTimeJSON{}

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	if result.Now.Unix() != now.Unix() {
		t.Errorf("Unexpected value: %+v, expected %+v", result.Now, now)
	}
}

func TestUnmarshalDateTimeJSString(t *testing.T) {
	now := time.Now()
	jsonString := fmt.Sprintf(`{"now":"%s"}`, now.Format("Mon Jan 2 2006 15:04:05 GMT-0700 (MST)"))

	result := testDateTimeJSON{}

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	if result.Now.Unix() != now.Unix() {
		t.Errorf("Unexpected value: %+v, expected %+v", result.Now, now)
	}
}

func TestUnmarshalDateTimeISO(t *testing.T) {
	now := time.Now()
	jsonString := fmt.Sprintf(`{"now":"%s"}`, now.Format("2006-01-02T15:04:05Z07:00"))

	result := testDateTimeJSON{}

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	if result.Now.Unix() != now.Unix() {
		t.Errorf("Unexpected value: %+v, expected %+v", result.Now, now)
	}
}

type testDateTimePtrJSON struct {
	Now *DT `json:"now,omitempty"`
}

func TestMarshalDateTimePtr(t *testing.T) {
	now := time.Now()
	data := testDateTimePtrJSON{&DT{now}}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	jsonString := string(jsonBytes)
	expected := fmt.Sprintf(`{"now":%.9f}`, float64(now.UnixNano())/float64(1e9))

	if jsonString != expected {
		t.Errorf("Unexpected value: %+v, expected %+v", jsonString, expected)
	}
}

func TestMarshalDateTimePtrEmpty(t *testing.T) {
	data := testDateTimePtrJSON{}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	jsonString := string(jsonBytes)
	expected := "{}"

	if jsonString != expected {
		t.Errorf("Unexpected value: %+v, expected %+v", jsonString, expected)
	}
}

func TestMarshalDateTimePtrNil(t *testing.T) {
	data := testDateTimePtrJSON{nil}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	jsonString := string(jsonBytes)
	expected := "{}"

	if jsonString != expected {
		t.Errorf("Unexpected value: %+v, expected %+v", jsonString, expected)
	}
}

func TestUnmarshalDateTimePtrGoString(t *testing.T) {
	now := time.Now()
	jsonString := fmt.Sprintf(`{"now":"%s"}`, now.Local().String())

	result := testDateTimePtrJSON{}

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	if result.Now.Unix() != now.Unix() {
		t.Errorf("Unexpected value: %+v, expected %+v", result.Now, now)
	}
}

func TestUnmarshalDateTimePtrJSString(t *testing.T) {
	now := time.Now()
	jsonString := fmt.Sprintf(`{"now":"%s"}`, now.Format("Mon Jan 2 2006 15:04:05 GMT-0700 (MST)"))

	result := testDateTimePtrJSON{}

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	if result.Now.Unix() != now.Unix() {
		t.Errorf("Unexpected value: %+v, expected %+v", result.Now, now)
	}
}

func TestUnmarshalDateTimePtrISO(t *testing.T) {
	now := time.Now()
	jsonString := fmt.Sprintf(`{"now":"%s"}`, now.Format("2006-01-02T15:04:05Z07:00"))

	result := testDateTimePtrJSON{}

	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	if result.Now.Unix() != now.Unix() {
		t.Errorf("Unexpected value: %+v, expected %+v", result.Now, now)
	}
}

func TestBSON(t *testing.T) {
	database.TestConnect()
	dbc := database.DB()
	defer dbc.Client().Disconnect(nil)

	insertData := testDateTimeJSON{Now: DT{time.Now()}}

	collection := dbc.Collection("datetimetest")

	inserted, err := collection.InsertOne(context.Background(), insertData)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	insertedID, _ := inserted.InsertedID.(primitive.ObjectID)
	if len(insertedID.Hex()) != 24 {
		t.Errorf("Invalid id: %+v", insertedID.Hex())
	}

	primitiveData := primitive.M{}
	err = collection.FindOne(context.Background(), primitive.M{"_id": insertedID}).Decode(&primitiveData)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	// mongodb has millisecond precision

	primitiveNow, _ := primitiveData["now"].(primitive.DateTime)
	if primitiveNow.Time().UnixNano()/1000000 != insertData.Now.UnixNano()/1000000 {
		t.Errorf("Unexpected value: %+v, expected: %+v", primitiveNow, insertData.Now.Time)
	}

	updatedNow := time.Now()

	_, err = collection.UpdateOne(context.Background(), primitive.M{"_id": insertedID}, primitive.M{"$set": primitive.M{"now": updatedNow}})
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	structData := testDateTimeJSON{}
	err = collection.FindOne(context.Background(), primitive.M{"_id": insertedID}).Decode(&structData)
	if err != nil {
		t.Errorf("Unexpected error: %+v, expected nil", err)
	}

	if structData.Now.UnixNano()/1000000 != updatedNow.UnixNano()/1000000 {
		t.Errorf("Unexpected value: %+v, expected: %+v", structData.Now.Time, updatedNow.UnixNano()/1000000)
	}
}

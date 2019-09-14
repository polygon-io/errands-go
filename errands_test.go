package errands

import (
	"testing"
	"time"

	schemas "github.com/polygon-io/errands-server/schemas"
	log "github.com/sirupsen/logrus"
)

func TestGetErrands(t *testing.T) {
	api := New("http://localhost:5555")
	errands, err := api.GetErrands()
	if err != nil {
		t.Error(err)
	}
	log.Println("Got Errands:", errands.Results)
}

var createdId string

func TestCreateErrand(t *testing.T) {
	api := New("http://localhost:5555")
	errand := &schemas.Errand{}
	errand.Name = "Test Errand"
	errand.Type = "tester"
	// errand.Data = map[string]interface{}{
	// 	"testkey": "testvalue",
	// }
	errandRes, err := api.CreateErrand(errand)
	if err != nil {
		t.Error(err)
	}
	log.Println("Created Errand:", errandRes)
	createdId = errandRes.Results.ID
}

func TestProcessErrand(t *testing.T) {
	TestCreateErrand(t)
	api := New("http://localhost:5555")
	wait := make(chan int, 2)
	fn := func(errand *schemas.Errand) (map[string]interface{}, error) {
		log.Println("Processing:", errand.Name, " - ID: ", errand.ID)
		time.Sleep(5 * time.Second)
		wait <- 1
		return map[string]interface{}{
			"results": "OK",
		}, nil
	}
	processor, err := api.NewProcessor("tester", 1, fn)
	if err != nil {
		t.Error(err)
	}
	<-wait
	<-wait
	log.Println("Processed Errand..", processor)
}

func TestDeleteErrand(t *testing.T) {
	api := New("http://localhost:5555")
	errandRes, err := api.DeleteErrand(createdId)
	if err != nil {
		t.Error(err)
	}
	log.Println("Deleted Errand:", errandRes)
}

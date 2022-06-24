package main

import (
	pb "aagmon/kafkaClientP/proto"
	"context"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

func getMessagePayload(recordID string, vector []float32) ([]byte, error) {

	req := pb.PredictRequest{
		RecordID:       recordID,
		FeaturesVector: vector,
	}
	return proto.Marshal(&req)
}

func loadVectorsListFromCSV(filename string, embSize int) map[string][]float32 {

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot read vectors file")
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	vectorsMap := make(map[string][]float32)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error reading records from file %v", err)
		}

		floatVector := strArrayToFloat(record, embSize+1)
		vector_id := strconv.Itoa(int(floatVector[0]))
		vector_data := floatVector[1:]
		vectorsMap[vector_id] = vector_data
	}

	return vectorsMap
}

func strArrayToFloat(strAr []string, arrSize int) []float32 {
	size := len(strAr)
	if size != arrSize {
		log.Fatalf("recieved array of size %v instead of %v", size, arrSize)
	}
	floatArr := make([]float32, arrSize)
	for i, v := range strAr {
		floatVal, err := strconv.ParseFloat(v, 32)
		if err != nil {
			log.Fatal("cannot convert value to float 32")
		}
		floatArr[i] = float32(floatVal)
	}
	return floatArr
}

func main() {

	log.Println("hello world!")

	filename := "/Users/alonagmon/MyData/work/golang-projects/vectors_model/test_set.csv"
	dataset := loadVectorsListFromCSV(filename, 30)
	ctx := context.Background()
	numWorkers := 10
	var wg sync.WaitGroup
	workChan := make(chan string)

	client, err := kgo.NewClient(
		kgo.SeedBrokers([]string{"127.0.0.1:62036"}...),
	)
	if err != nil {
		log.Fatalf("error creating NewClient %v", err)
	}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for recordID := range workChan {
				payload, err := getMessagePayload(recordID, dataset[recordID])
				if err != nil {
					log.Fatalf("error marshalling %v", err)
				}
				record := &kgo.Record{
					Key:   []byte(recordID),
					Value: payload,
					Topic: "test-topic-1",
				}
				result := client.ProduceSync(ctx, record)
				err = result.FirstErr()
				if err != nil {
					log.Fatalf("ProduceSync failed: %v", err)
				}

			}

		}()

	}

	for k := range dataset {
		workChan <- k
	}
	close(workChan)
	wg.Wait()

}

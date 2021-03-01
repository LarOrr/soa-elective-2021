package main

import (
	"fmt"
	"reflect"
	"time"
)

//'words':"""Lorem ipsum dolor sit amet, consectetur adipiscingelit. Mauris adipiscing adipiscing placerat.Vestibulum augue augue,pellentesque quis sollicitudin id, adipiscing.""",
//'list':range(100),
//'dict':dict((str(i),'a') fori in iter(range(100))),
//'int':100,'
//float':100.123456

type KeyValue struct {
	Key   string
	Value string
}

type TestStruct struct {
	Words string // `json:"fieldA"`
	List  []int  // `json:"fieldB"`
	// Не работает с XML!!! вместо этого использую слайс из KeyValue
	//Dict map[string]int
	Dict  []KeyValue
	Int   int32
	Float float64
}

func initTestData() TestStruct {
	// List
	list := make([]int, 100)
	for i := 0; i < 100; i++ {
		list[i] = i
	}
	// Dict
	//dict := make(map[string]int)
	//for i := 0; i < 100; i++ {
	//	dict[fmt.Sprintf("key %v", i)] = i
	//}
	dict := make([]KeyValue, 100)
	for i := 0; i < 100; i++ {
		dict[i] = KeyValue{Key: fmt.Sprintf("key %v", i),
			Value: fmt.Sprintf("value %v", i)}
		//	dict[fmt.Sprintf("key %v", i)] = i
	}

	data := TestStruct{
		Words: `Lorem ipsum dolor sit amet, consectetur adipiscingelit. 
				Mauris adipiscing adipiscing placerat.
				Vestibulum augue augue,pellentesque quis sollicitudin id, adipiscing.`,
		List:  list,
		Dict:  dict,
		Int:   1234567890,
		Float: 123456.654321,
	}
	return data
}

func benchmarkSerializer(s Serializer) {
	loops := 1000
	data := initTestData()

	// Marshal
	start := time.Now()
	for i := 0; i < loops; i++ {
		_, err := s.Marshal(data)
		if err != nil {
			fmt.Printf("Error with marshaling %v", err)
			return
		}
	}
	elapsedMs := float64(time.Since(start).Microseconds()) / float64(loops)
	serResult, _ := s.Marshal(data)
	serialSize := len(serResult)
	fmt.Printf("Serialization time = %v ms | memory = %v bytes\n", elapsedMs, serialSize)

	// Unmarshal
	start = time.Now()
	for i := 0; i < loops; i++ {
		err := s.Unmarshal(serResult, &TestStruct{})
		if err != nil {
			fmt.Printf("Error with deserialization from bytes: %v", err)
			return
		}
	}
	// Отдельно от цикла чтобы не наполнять slice каждый раз
	unserData := TestStruct{}
	_ = s.Unmarshal(serResult, &unserData)
	elapsedMs = float64(time.Since(start).Microseconds()) / float64(loops)
	// Провекрка на правильность десериализации
	// Не проверяю "Apache Avro", т.к. там Dict сохраняется как map и нельзя восставновить порядок
	if s.Name() != "Apache Avro" && !reflect.DeepEqual(unserData, data) {
		panic("Unrealized data and initial data are not equal!")
	}
	fmt.Printf("Deserialization time = %v ms\n", elapsedMs)
}

func main() {
	serializers := []Serializer{
		GobSerializer{},
		JsonSerializer{},
		XmlSerializer{},
		ProtobufSerializer{},
		NewAvroSerializer(),
		YamlSerializer{},
		MsgPackSerializer{},
	}

	for _, ser := range serializers {
		fmt.Printf("\n======= %v =======\n", ser.Name())
		benchmarkSerializer(ser)
	}
}

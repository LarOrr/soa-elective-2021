package main

import (
	"bytes"
	"encoding/gob"
	"sort"

	// My module
	"ser_bench/model"
	// better to use this "https://github.com/mailru/easyjson"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"gopkg.in/yaml.v2" // https://github.com/go-yaml/yaml
	//Official version of protobuf here "github.com/golang/protobuf/proto"
	// Better (faster and easier to use) version "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/linkedin/goavro/v2"
)

type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	Name() string
}

// encoding/json
type JsonSerializer struct{}

func (s JsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s JsonSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (s JsonSerializer) Name() string {
	return "Json"
}

// encoding/xml
type XmlSerializer struct{}

func (s XmlSerializer) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (s XmlSerializer) Unmarshal(data []byte, v interface{}) error {

	return xml.Unmarshal(data, v)
}

func (s XmlSerializer) Name() string {
	return "XML"
}

// encoding/gob. Src: https://gist.github.com/mr4x/842d58079750c72d976bc62b7d4987b4

type GobSerializer struct{}

func (g GobSerializer) Marshal(v interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (g GobSerializer) Unmarshal(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	return gob.NewDecoder(b).Decode(v)
}

func (g GobSerializer) Name() string {
	return "Native Go Binary (Gob)"
}

// yaml

type YamlSerializer struct{}

func (s YamlSerializer) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (s YamlSerializer) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func (s YamlSerializer) Name() string {
	return "YAML"
}

// msgpack

type MsgPackSerializer struct{}

func (s MsgPackSerializer) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (s MsgPackSerializer) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

func (s MsgPackSerializer) Name() string {
	return "MessagePack"
}

// protobuf

type ProtobufSerializer struct{}

func (m ProtobufSerializer) Marshal(v interface{}) ([]byte, error) {
	protoMessage, err := toProto(v)
	if err != nil {
		fmt.Printf("Can't use this datatype for protobuf %v", err)
	}
	return proto.Marshal(protoMessage)
}

func (m ProtobufSerializer) Unmarshal(data []byte, v interface{}) error {
	protoMessage := &model.TestStruct{}
	err := proto.Unmarshal(data, protoMessage)
	if err != nil {
		return err
	}
	err = fromProto(protoMessage, v)
	// Возвращает потобаф версию - нехорошо, но не важно для теста
	return err
}

func (s ProtobufSerializer) Name() string {
	return "ProtoBuf"
}

// Костыльное решение - просто преобразую TestStruct в model.TestStruct (proto версия)
// Не уверен, ксть ли более красивый алтернативы
func toProto(v interface{}) (*model.TestStruct, error) {
	testStruct, ok := v.(TestStruct)
	if !ok {
		return nil, errors.New("Unsupported type for ProtoBuf")
	}
	protoList := make([]int32, len(testStruct.List))
	for i := 0; i < len(testStruct.List); i++ {
		protoList[i] = int32(testStruct.List[i])
	}

	protoDict := make([]*model.TestStruct_KeyValue, len(testStruct.Dict))
	for i := 0; i < len(testStruct.Dict); i++ {
		protoDict[i] = &model.TestStruct_KeyValue{
			Key:   testStruct.Dict[i].Key,
			Value: testStruct.Dict[i].Value,
		}
	}

	return &model.TestStruct{
		Words: testStruct.Words,
		List:  protoList,
		Dict:  protoDict,
		Int:   testStruct.Int,
		Float: testStruct.Float,
	}, nil
}

func fromProto(protoMessage *model.TestStruct, v interface{}) error {
	testStruct, ok := v.(*TestStruct)
	if !ok {
		return errors.New("Unsupported type for ProtoBuf")
	}

	testStruct.Words = protoMessage.Words
	list := make([]int, len(protoMessage.List))
	for i := 0; i < len(protoMessage.List); i++ {
		list[i] = int(protoMessage.List[i])
	}
	testStruct.List = list

	dict := make([]KeyValue, len(protoMessage.Dict))
	for i := 0; i < len(protoMessage.Dict); i++ {
		dict[i] = KeyValue{
			Key:   protoMessage.Dict[i].Key,
			Value: protoMessage.Dict[i].Value,
		}
	}
	testStruct.Dict = dict
	testStruct.Int = protoMessage.Int
	testStruct.Float = protoMessage.Float

	return nil
}

// Apache Avro

var (
	avroSchemaJSON = `
		{
		  "type": "record",
		  "name": "TypeStruct",
		  "doc:": "Schema for encoding/decoding sample message",
		  "namespace": "com.example",
		  "fields": [
		    {
		      "name": "words",
		      "type": "string"
		    },
		    {
		      "name": "list",
		      "type": "array",
              "items": "int"
		    },
		    {
		      "name": "dict",
		      "type": "map",
              "values" : "string"
		    },
		    {
		      "name": "int",
		      "type": "int"
		    },
		    {
		      "name": "float",
		      "type": "double"
		    }
		  ]
		}
`
)

func NewAvroSerializer() *AvroSerializer {
	codec, err := goavro.NewCodec(avroSchemaJSON)
	if err != nil {
		panic(err)
	}
	return &AvroSerializer{codec: codec}
}

type AvroSerializer struct {
	codec *goavro.Codec
}

func (s AvroSerializer) Marshal(v interface{}) ([]byte, error) {
	// Также можно использовать a.codec.TextualFromNative для текстового формата
	return s.codec.BinaryFromNative(nil, toAvroMap(v))
}

func toAvroMap(v interface{}) map[string]interface{} {
	object := v.(TestStruct)
	dictMap := make(map[string]string)
	for _, keyVal := range object.Dict {
		dictMap[keyVal.Key] = keyVal.Value
	}
	return map[string]interface{}{
		"words": object.Words,
		"list":  object.List,
		"dict":  dictMap,
		"int":   object.Int,
		"float": object.Float,
	}
}

func (s AvroSerializer) Unmarshal(data []byte, v interface{}) error {
	result, _, err := s.codec.NativeFromBinary(data)
	if err != nil {
		return err
	}
	return fromAvro(result.(map[string]interface{}), v)
}

func fromAvro(m map[string]interface{}, v interface{}) error {
	object := v.(*TestStruct)
	object.Words = m["words"].(string)
	list := make([]int, 0)
	for _, val := range m["list"].([]interface{}) {
		list = append(list, int(val.(int32)))
	}
	object.List = list
	dict := make([]KeyValue, 0)
	for key, val := range m["dict"].(map[string]interface{}) {
		dict = append(dict, KeyValue{Key: key,
			Value: val.(string)})
	}
	sort.SliceStable(dict, func(i, j int) bool {
		// Сотрирую по ключам инач не будут равны - не лучшее решение
		return dict[i].Key < dict[j].Key
	})
	object.Dict = dict
	object.Int = m["int"].(int32)
	object.Float = m["float"].(float64)
	return nil
}

func (s AvroSerializer) Name() string {
	return "Apache Avro"
}

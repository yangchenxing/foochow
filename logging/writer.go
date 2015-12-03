package logging

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/yangchenxing/foochow/structs"
)

func init() {
	// 注册Writer工厂
	structs.RegisterFactory(&WriterFactory{})
}

type WriterFactory struct{}

func (factory *WriterFactory) GetInstanceType() reflect.Type {
	return reflect.TypeOf((*Writer)(nil)).Elem()
}

func (factory *WriterFactory) Create(data map[string]interface{}) (interface{}, error) {
	if typeName, ok := data["type"].(string); ok {
		switch typeName {
		case "stderr":
			return os.Stderr, nil
		case "timerotate":
			writer := new(TimeRotateWriter)
			if err := structs.UnmarshalMap(writer, data); err != nil {
				return nil, err
			} else if err := writer.initialize(); err != nil {
				return nil, err
			}
			return writer, nil
		case "email":
			writer := new(EMailWriter)
			if err := structs.UnmarshalMap(writer, data); err != nil {
				return nil, err
			}
			writer.initialize()
			return writer, nil
		default:
			return nil, fmt.Errorf("未知日志Writer类型: %q", typeName)
		}
	}
	return nil, errors.New("缺少\"type\"字段")
}

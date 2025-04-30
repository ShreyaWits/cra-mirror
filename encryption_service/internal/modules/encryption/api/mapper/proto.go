package mapper

import (
	"encoding/base64"
	pb "encryption_microservice/internal/common/proto_gen"
	"unicode/utf8"
)

// Convert []map[string]interface{} to []*structpb.Struct
func ConvertToStructPB(data []map[string]string) []*pb.Data {
	result := make([]*pb.Data, 0, len(data))
	for _, item := range data {
		cleanedFields := make(map[string]string)
		for k, v := range item {
			if !utf8.ValidString(v) {
				cleanedFields[k] = base64.StdEncoding.EncodeToString([]byte(v))
			} else {
				cleanedFields[k] = v
			}
		}
		result = append(result, &pb.Data{Fields: cleanedFields})
	}
	return result
}

func ConvertFromStructPB(data []*pb.Data) []map[string]string {
	result := make([]map[string]string, 0, len(data))
	for _, s := range data {

		result = append(result, s.Fields)
	}
	return result
}

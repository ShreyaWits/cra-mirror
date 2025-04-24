package mapper

import (
	pb "encryption_microservice/internal/common/proto_gen"
)

// Convert []map[string]interface{} to []*structpb.Struct
func ConvertToStructPB(data []map[string]string) []*pb.Data {
	result := make([]*pb.Data, 0, len(data))
	for _, item := range data {

		data := &pb.Data{
			Fields: item,
		}
		result = append(result, data)
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

package etcd

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/registry"
)

func marshal(si *registry.ServiceInstance) (string, error) {
	return si.Endpoints[0], nil
}

func unmarshal(data []byte) (si *registry.ServiceInstance, err error) {
	err = json.Unmarshal(data, &si)
	return
}

//
//func marshal(endpoint string) (string, error) {
//	uri, err := url.Parse(endpoint)
//	if err != nil {
//		return "", err
//	}
//	return uri.Hostname() + ":" + uri.Port(), nil
//}
//
//func unmarshal(data []byte) (string, error) {
//	return "grpc://" + string(data), nil
//}

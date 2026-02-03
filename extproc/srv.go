package extproc

import (
	"fmt"
	"log"

	_ "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	continueResponse = &extprocv3.CommonResponse{
		Status: extprocv3.CommonResponse_CONTINUE,
	}
)

type ExtprocV3Server struct {
	extprocv3.UnimplementedExternalProcessorServer
}

func NewExtprocV3Server() *ExtprocV3Server {
	return &ExtprocV3Server{}
}

func (s *ExtprocV3Server) Process(srv extprocv3.ExternalProcessor_ProcessServer) error {

	for {
		req, err := srv.Recv()
		if err != nil {
			log.Printf("error occurred while receiving message %v\n", err)
			return nil
		}

		json, _ := protojson.Marshal(req)
		log.Printf("Received message %s\n", string(json))

		var res *extprocv3.ProcessingResponse
		switch req.GetRequest().(type) {
		case *extprocv3.ProcessingRequest_RequestHeaders:
			fmt.Println("here in request headers phase")
			res = &extprocv3.ProcessingResponse{
				Response: &extprocv3.ProcessingResponse_RequestHeaders{
					RequestHeaders: &extprocv3.HeadersResponse{
						Response: continueResponse,
					},
				},
			}
		case *extprocv3.ProcessingRequest_RequestBody:
			fmt.Println("here in request body phase")
			res = &extprocv3.ProcessingResponse{
				Response: &extprocv3.ProcessingResponse_RequestBody{
					RequestBody: &extprocv3.BodyResponse{
						Response: continueResponse,
					},
				},
			}
		case *extprocv3.ProcessingRequest_RequestTrailers:
			fmt.Println("here in request trailers phase")
			res = &extprocv3.ProcessingResponse{
				Response: &extprocv3.ProcessingResponse_RequestTrailers{
					RequestTrailers: &extprocv3.TrailersResponse{},
				},
			}
		case *extprocv3.ProcessingRequest_ResponseHeaders:
			fmt.Println("here in response headers phase")
			res = &extprocv3.ProcessingResponse{
				Response: &extprocv3.ProcessingResponse_ResponseHeaders{
					ResponseHeaders: &extprocv3.HeadersResponse{
						Response: continueResponse,
					},
				},
			}
		case *extprocv3.ProcessingRequest_ResponseBody:
			fmt.Println("here in response body phase")
			res = &extprocv3.ProcessingResponse{
				Response: &extprocv3.ProcessingResponse_ResponseBody{
					ResponseBody: &extprocv3.BodyResponse{
						Response: continueResponse,
					},
				},
			}
		case *extprocv3.ProcessingRequest_ResponseTrailers:
			fmt.Println("here in response trailers phase")
			res = &extprocv3.ProcessingResponse{
				Response: &extprocv3.ProcessingResponse_ResponseTrailers{
					ResponseTrailers: &extprocv3.TrailersResponse{},
				},
			}
		}

		json, _ = protojson.Marshal(res)
		log.Printf("Sending message %s\n", string(json))
		err = srv.Send(res)
		if err != nil {
			log.Printf("error occurred while sending message %v\n", err)
			return nil
		}
	}
}

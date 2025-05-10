package grpc

import (
	gen "VK/internal/gateways/generated"
	"VK/internal/repository/inmemory"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	ErrCanceledSubscription = "client canceled subscription"
	ErrEmptyRequest         = "empty request"
	ErrConvertTypes         = "unable to convert"
	ErrFailedSubscribe      = "subscribe failed"
	ErrFailedPublish        = "publish failed"
	ErrFailedSend           = "send failed"
)

type SubPubServer struct {
	subPub                        *inmemory.SubPub
	gen.UnimplementedPubSubServer // I embed UnimplementedPubSubServer for forward compatibility
}

func NewSubPubServer(s *inmemory.SubPub) *SubPubServer {
	return &SubPubServer{
		subPub: s,
	}
}

func (s *SubPubServer) Subscribe(request *gen.SubscribeRequest, g grpc.ServerStreamingServer[gen.Event]) error {
	if g.Context().Err() != nil {
		return status.Error(codes.Canceled, ErrCanceledSubscription)
	}

	chPublished := make(chan interface{})
	key := request.GetKey()
	if key == "" {
		return status.Error(codes.InvalidArgument, ErrEmptyRequest)
	}

	sub, err := s.subPub.Subscribe(key, func(msg interface{}) {
		chPublished <- msg
	})
	if err != nil {
		return status.Error(codes.Internal, ErrFailedSubscribe+err.Error())
	}

	defer func() {
		sub.Unsubscribe()
		close(chPublished)
	}()

	for {
		select {
		case <-g.Context().Done():
			return status.Error(codes.Canceled, ErrCanceledSubscription)
		case msg, ok := <-chPublished:
			if !ok {
				return status.Error(codes.Canceled, ErrFailedPublish)
			}

			strMsg, ok := msg.(string)
			if !ok {
				return status.Error(codes.Internal, ErrConvertTypes)
			}

			if err := g.Send(&gen.Event{Data: strMsg}); err != nil {
				return status.Error(codes.Internal, ErrFailedSend+err.Error())
			}
		}
	}
}

func (s *SubPubServer) Publish(ctx context.Context, request *gen.PublishRequest) (*emptypb.Empty, error) {
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, ErrCanceledSubscription)
	default:
		key := request.GetKey()
		data := request.GetData()
		if key == "" || data == "" {
			return nil, status.Error(codes.InvalidArgument, ErrEmptyRequest)
		}

		if err := s.subPub.Publish(key, data); err != nil {
			return nil, status.Error(codes.Internal, ErrFailedPublish+err.Error())
		}

		return &emptypb.Empty{}, nil
	}
}

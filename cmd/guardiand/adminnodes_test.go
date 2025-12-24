package guardiand

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	publicrpcv1 "github.com/certusone/wormhole/node/pkg/proto/publicrpc/v1"
)

/*
GetLastHeartbeats(ctx context.Context, in *GetLastHeartbeatsRequest, opts ...grpc.CallOption) (*GetLastHeartbeatsResponse, error)

	GetSignedVAA(ctx context.Context, in *GetSignedVAARequest, opts ...grpc.CallOption) (*GetSignedVAAResponse, error)
	GetCurrentGuardianSet(ctx context.Context, in *GetCurrentGuardianSetRequest, opts ...grpc.CallOption) (*GetCurrentGuardianSetResponse, error)
	GovernorGetAvailableNotionalByChain(ctx context.Context, in *GovernorGetAvailableNotionalByChainRequest, opts ...grpc.CallOption) (*GovernorGetAvailableNotionalByChainResponse, error)
	GovernorGetEnqueuedVAAs(ctx context.Context, in *GovernorGetEnqueuedVAAsRequest, opts ...grpc.CallOption) (*GovernorGetEnqueuedVAAsResponse, error)
	GovernorIsVAAEnqueued(ctx context.Context, in *GovernorIsVAAEnqueuedRequest, opts ...grpc.CallOption) (*GovernorIsVAAEnqueuedResponse, error)
	GovernorGetTokenList(ctx context.Context, in *GovernorGetTokenListRequest, opts ...grpc.CallOption) (*GovernorGetTokenListResponse, er
*/
func TestRpc(t *testing.T) {
	ctx := context.Background()
	testclientSocketPath := "/home/lfm/dev/wormhole_tron/wormhole/node/node_data1/mysockettron"
	conn, c, err := getPublicRPCServiceClient(ctx, testclientSocketPath)
	if err != nil {
		fmt.Printf("failed to get publicrpc client: %v", err)
	}
	defer conn.Close()
	vaaRequest := &publicrpcv1.GetSignedVAARequest{}
	vaaRequest.MessageId = &publicrpcv1.MessageID{}
	vaaRequest.MessageId.EmitterAddress = "000000000000000000000000f87804a8dcee9ed039ff0e3315b4d8a1890f2801"
	vaaRequest.MessageId.EmitterChain = publicrpcv1.ChainID(222)
	vaaRequest.MessageId.Sequence = 1

	/*vaaRequest.MessageId.EmitterAddress = "000000000000000000000000d73f34428098b44a589f13ad15a0e3d2efe92dbd"
	vaaRequest.MessageId.EmitterChain = publicrpcv1.ChainID(2)
	vaaRequest.MessageId.Sequence = 0*/
	signedVaa, err := c.GetSignedVAA(ctx, vaaRequest)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%x", signedVaa.VaaBytes)
	fmt.Println("")
	fmt.Println("")
	vaaHex := "0x" + hex.EncodeToString(signedVaa.VaaBytes)
	fmt.Println(vaaHex)

	/*gs, err := c.GetCurrentGuardianSet(ctx, &publicrpcv1.GetCurrentGuardianSetRequest{})
	if err != nil {
		fmt.Printf("failed to list current guardian get: %v", err)
	}
	fmt.Println(gs)*/
}

/*

 */

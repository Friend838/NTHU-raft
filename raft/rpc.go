package raft

import (
	"context"
	"errors"

	"github.com/justin0u0/raft/pb"
)

type rpcResponse struct {
	resp interface{}
	err  error
}

type rpc struct {
	req    interface{}
	respCh chan<- *rpcResponse
}

var (
	errRPCTimeout           = errors.New("rpc timeout")
	errResponseTypeMismatch = errors.New("response type mismatch")
	errInvalidRPCType       = errors.New("invalid rpc type")
)

func (r *raft) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	rpcResp, err := r.dispatchRPCRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := rpcResp.(*pb.AppendEntriesResponse)
	if !ok {
		return nil, errResponseTypeMismatch
	}

	return resp, nil
}

func (r *raft) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	rpcResp, err := r.dispatchRPCRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	resp, ok := rpcResp.(*pb.RequestVoteResponse)
	if !ok {
		return nil, errResponseTypeMismatch
	}

	return resp, nil
}

func (r *raft) dispatchRPCRequest(ctx context.Context, req interface{}) (interface{}, error) {
	respCh := make(chan *rpcResponse, 1)
	r.rpcCh <- &rpc{req: req, respCh: respCh}

	select {
	case <-ctx.Done():
		return nil, errRPCTimeout
	case rpcResp := <-respCh:
		if err := rpcResp.err; err != nil {
			return nil, err
		}

		return rpcResp.resp, nil
	}
}

func (r *raft) handleRPCRequest(rpc *rpc) {
	switch req := rpc.req.(type) {
	case *pb.AppendEntriesRequest:
		r.appendEntries(req, rpc.respCh)
	case *pb.RequestVoteRequest:
		r.requestVote(req, rpc.respCh)
	default:
		rpc.respCh <- &rpcResponse{err: errInvalidRPCType}
	}
}
// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var linear_pb = require('./linear_pb.js');

function serialize_linear_GetIssuesRequest(arg) {
  if (!(arg instanceof linear_pb.GetIssuesRequest)) {
    throw new Error('Expected argument of type linear.GetIssuesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_linear_GetIssuesRequest(buffer_arg) {
  return linear_pb.GetIssuesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_linear_GetIssuesResponse(arg) {
  if (!(arg instanceof linear_pb.GetIssuesResponse)) {
    throw new Error('Expected argument of type linear.GetIssuesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_linear_GetIssuesResponse(buffer_arg) {
  return linear_pb.GetIssuesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var LinearService = exports.LinearService = {
  getIssues: {
    path: '/linear.Linear/GetIssues',
    requestStream: false,
    responseStream: false,
    requestType: linear_pb.GetIssuesRequest,
    responseType: linear_pb.GetIssuesResponse,
    requestSerialize: serialize_linear_GetIssuesRequest,
    requestDeserialize: deserialize_linear_GetIssuesRequest,
    responseSerialize: serialize_linear_GetIssuesResponse,
    responseDeserialize: deserialize_linear_GetIssuesResponse,
  },
};

exports.LinearClient = grpc.makeGenericClientConstructor(LinearService);

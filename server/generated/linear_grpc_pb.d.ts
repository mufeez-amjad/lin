// package: linear
// file: linear.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as linear_pb from "./linear_pb";

interface ILinearService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    getIssues: ILinearService_IGetIssues;
}

interface ILinearService_IGetIssues extends grpc.MethodDefinition<linear_pb.GetIssuesRequest, linear_pb.GetIssuesResponse> {
    path: "/linear.Linear/GetIssues";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<linear_pb.GetIssuesRequest>;
    requestDeserialize: grpc.deserialize<linear_pb.GetIssuesRequest>;
    responseSerialize: grpc.serialize<linear_pb.GetIssuesResponse>;
    responseDeserialize: grpc.deserialize<linear_pb.GetIssuesResponse>;
}

export const LinearService: ILinearService;

export interface ILinearServer extends grpc.UntypedServiceImplementation {
    getIssues: grpc.handleUnaryCall<linear_pb.GetIssuesRequest, linear_pb.GetIssuesResponse>;
}

export interface ILinearClient {
    getIssues(request: linear_pb.GetIssuesRequest, callback: (error: grpc.ServiceError | null, response: linear_pb.GetIssuesResponse) => void): grpc.ClientUnaryCall;
    getIssues(request: linear_pb.GetIssuesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: linear_pb.GetIssuesResponse) => void): grpc.ClientUnaryCall;
    getIssues(request: linear_pb.GetIssuesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: linear_pb.GetIssuesResponse) => void): grpc.ClientUnaryCall;
}

export class LinearClient extends grpc.Client implements ILinearClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public getIssues(request: linear_pb.GetIssuesRequest, callback: (error: grpc.ServiceError | null, response: linear_pb.GetIssuesResponse) => void): grpc.ClientUnaryCall;
    public getIssues(request: linear_pb.GetIssuesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: linear_pb.GetIssuesResponse) => void): grpc.ClientUnaryCall;
    public getIssues(request: linear_pb.GetIssuesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: linear_pb.GetIssuesResponse) => void): grpc.ClientUnaryCall;
}

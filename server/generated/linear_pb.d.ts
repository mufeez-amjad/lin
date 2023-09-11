// package: linear
// file: linear.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class Error extends jspb.Message { 
    getMessage(): string;
    setMessage(value: string): Error;

    hasCode(): boolean;
    clearCode(): void;
    getCode(): number | undefined;
    setCode(value: number): Error;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Error.AsObject;
    static toObject(includeInstance: boolean, msg: Error): Error.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Error, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Error;
    static deserializeBinaryFromReader(message: Error, reader: jspb.BinaryReader): Error;
}

export namespace Error {
    export type AsObject = {
        message: string,
        code?: number,
    }
}

export class Issue extends jspb.Message { 
    getId(): string;
    setId(value: string): Issue;
    getTitle(): string;
    setTitle(value: string): Issue;
    getIdentifier(): string;
    setIdentifier(value: string): Issue;
    getBranchname(): string;
    setBranchname(value: string): Issue;
    getUrl(): string;
    setUrl(value: string): Issue;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Issue.AsObject;
    static toObject(includeInstance: boolean, msg: Issue): Issue.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Issue, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Issue;
    static deserializeBinaryFromReader(message: Issue, reader: jspb.BinaryReader): Issue;
}

export namespace Issue {
    export type AsObject = {
        id: string,
        title: string,
        identifier: string,
        branchname: string,
        url: string,
    }
}

export class GetIssuesRequest extends jspb.Message { 
    getApiKey(): string;
    setApiKey(value: string): GetIssuesRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetIssuesRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetIssuesRequest): GetIssuesRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetIssuesRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetIssuesRequest;
    static deserializeBinaryFromReader(message: GetIssuesRequest, reader: jspb.BinaryReader): GetIssuesRequest;
}

export namespace GetIssuesRequest {
    export type AsObject = {
        apiKey: string,
    }
}

export class GetIssuesResponse extends jspb.Message { 
    clearIssuesList(): void;
    getIssuesList(): Array<Issue>;
    setIssuesList(value: Array<Issue>): GetIssuesResponse;
    addIssues(value?: Issue, index?: number): Issue;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetIssuesResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetIssuesResponse): GetIssuesResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetIssuesResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetIssuesResponse;
    static deserializeBinaryFromReader(message: GetIssuesResponse, reader: jspb.BinaryReader): GetIssuesResponse;
}

export namespace GetIssuesResponse {
    export type AsObject = {
        issuesList: Array<Issue.AsObject>,
    }
}

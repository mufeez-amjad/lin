import {
  Server,
  ServerCredentials,
  ServerUnaryCall,
  sendUnaryData,
} from '@grpc/grpc-js';
import { LinearClient, LinearDocument } from "@linear/sdk";
import {
  GetIssuesRequest,
  GetIssuesResponse
} from './generated/linear_pb';
import { LinearService } from './generated/linear_grpc_pb';

const lin = new LinearClient({
  apiKey: '123',
});

const getIssues = async (
  call: ServerUnaryCall<GetIssuesRequest, GetIssuesResponse>,
  callback: sendUnaryData<GetIssuesResponse>,
) => {
  const { getApiKey } = call.request;
  const apiKey = getApiKey();
  
  const user = await lin.viewer;
  const issues = await user.assignedIssues({
      first: 10,
      orderBy: LinearDocument.PaginationOrderBy.UpdatedAt,
  });
  const org = await lin.organization;
  org.gitBranchFormat || "default";
};

const server = new Server();

server.addService(LinearService, {
  getIssues,
});

server.bindAsync('0.0.0.0:50051', ServerCredentials.createInsecure(), () => {
  server.start();

  console.log('Server running at 0.0.0.0:50051');
});

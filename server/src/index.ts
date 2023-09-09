import {
  Server,
  ServerCredentials,
  ServerUnaryCall,
  sendUnaryData,
} from '@grpc/grpc-js';
import { LinearClient, LinearDocument } from "@linear/sdk";
import {
  GetIssuesRequest,
  GetIssuesResponse,
  Issue
} from '../generated/linear_pb';
import { LinearService } from '../generated/linear_grpc_pb';

const lin = new LinearClient({
  apiKey: 'lin_api_Sx3vd6bk5REQmH0vnHGGsP4jHPkox7ycH5VYv6do',
  apiUrl: 'http://localhost:8090/graphql',
});

const getIssues = async (
  call: ServerUnaryCall<GetIssuesRequest, GetIssuesResponse>,
  callback: sendUnaryData<GetIssuesResponse>,
) => {
  console.log('getIssues called');
  const { getApiKey } = call.request;
  const apiKey = getApiKey();
  
  const user = await lin.viewer;
  const issues = await user.assignedIssues({
      orderBy: LinearDocument.PaginationOrderBy.UpdatedAt,
  });

  const issueMessages: Issue[] = await Promise.all(issues.nodes.map(async (issue) => {
    const issueMessage = new Issue();
    issueMessage.setId(issue.id);
    issueMessage.setTitle(issue.title);
    issueMessage.setIdentifier(issue.identifier);
    issueMessage.setBranchname(issue.branchName);
    issueMessage.setUrl(issue.url);
    return issueMessage;
  }));

  const response = new GetIssuesResponse();
  response.setIssuesList(issueMessages);

  return callback(null, response);
};

const server = new Server();

server.addService(LinearService, {
  getIssues,
});

server.bindAsync('0.0.0.0:50051', ServerCredentials.createInsecure(), () => {
  server.start();

  console.log('Server running at 0.0.0.0:50051');
});

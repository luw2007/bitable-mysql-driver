# Security policy

Report vulnerabilities privately through GitHub's **Report a vulnerability** form. Do not include credentials, access tokens, app tokens, table IDs, or private document data in public issues.

The driver deliberately suppresses verbose Feishu SDK request and response logs because they may contain authentication headers, application secrets, or access tokens. BOE integration tests run only when all required environment variables are explicitly configured.

export type ApiMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export type ApiAuthConfig = {
  type: string
  header: string
  required: boolean
  description: string
}

export type ApiResponseField = {
  name: string
  type: string
  required?: boolean
  description: string
}

export type ApiResponse = {
  status: number
  label: string
  description: string
  fields: ApiResponseField[]
  example: string
}

export type ApiParameter = {
  name: string
  location: 'body' | 'query' | 'header'
  type: string
  required?: boolean
  description: string
}

export type ApiExamples = {
  curl: string
  node: string
  python: string
  go: string
  java: string
}

export type ApiEndpointConfig = {
  id: string
  title: string
  section: string
  description: string
  method: ApiMethod
  baseUrl: string
  path: string
  auth: ApiAuthConfig
  parameters: ApiParameter[]
  responses: ApiResponse[]
  examples: ApiExamples
}

export type ApiSidebarItem = {
  title: string
  method?: ApiMethod
  href: string
  active?: boolean
}

export type ApiSidebarGroup = {
  title: string
  items: ApiSidebarItem[]
}

const baseUrl = 'https://tokengate-production.up.railway.app'

export const tokenGateApiEndpoint: ApiEndpointConfig = {
  id: 'chat-completions',
  title: 'Create chat completion',
  section: 'OpenAI compatible',
  description:
    'Creates a model response through TokenGate using an OpenAI-compatible request shape. The user API key determines the allowed group, upstream account pool, model availability, rate limits, and balance settlement.',
  method: 'POST',
  baseUrl,
  path: '/v1/chat/completions',
  auth: {
    type: 'Bearer token',
    header: 'Authorization',
    required: true,
    description: 'Customer API key created from the TokenGate dashboard.',
  },
  parameters: [
    { name: 'model', location: 'body', type: 'string', required: true, description: 'Model id allowed by the API key group, such as gpt-5.4.' },
    { name: 'messages', location: 'body', type: 'array', required: true, description: 'Conversation messages using OpenAI chat format.' },
    { name: 'temperature', location: 'body', type: 'number', required: false, description: 'Sampling temperature forwarded to the upstream provider when supported.' },
    { name: 'max_tokens', location: 'body', type: 'number', required: false, description: 'Maximum output token budget when supported by the selected model.' },
    { name: 'stream', location: 'body', type: 'boolean', required: false, description: 'Whether to request a streaming response.' },
  ],
  responses: [
    {
      status: 200,
      label: 'OK',
      description: 'The upstream model returned a successful OpenAI-compatible chat completion.',
      fields: [
        { name: 'id', type: 'string', required: true, description: 'Provider response identifier.' },
        { name: 'object', type: 'string', required: true, description: 'Usually chat.completion.' },
        { name: 'created', type: 'number', required: true, description: 'Unix timestamp for response creation.' },
        { name: 'model', type: 'string', required: true, description: 'Resolved model used for the request.' },
        { name: 'choices[]', type: 'array', required: true, description: 'Assistant message choices returned by the model.' },
        { name: 'choices[].message.content', type: 'string', required: false, description: 'Generated text for non-streaming calls.' },
        { name: 'usage.prompt_tokens', type: 'number', required: false, description: 'Input token count when reported by the upstream provider.' },
        { name: 'usage.completion_tokens', type: 'number', required: false, description: 'Output token count when reported by the upstream provider.' },
        { name: 'usage.total_tokens', type: 'number', required: false, description: 'Total billable token count when available.' },
      ],
      example: JSON.stringify(
        {
          id: 'chatcmpl_tokengate_123',
          object: 'chat.completion',
          created: 1778956800,
          model: 'gpt-5.4',
          choices: [
            {
              index: 0,
              message: {
                role: 'assistant',
                content: 'TokenGate is ready.',
              },
              finish_reason: 'stop',
            },
          ],
          usage: {
            prompt_tokens: 12,
            completion_tokens: 5,
            total_tokens: 17,
          },
        },
        null,
        2,
      ),
    },
    {
      status: 401,
      label: 'Unauthorized',
      description: 'The request is missing a valid TokenGate API key.',
      fields: [
        { name: 'error.code', type: 'string', required: true, description: 'Usually UNAUTHORIZED.' },
        { name: 'error.message', type: 'string', required: true, description: 'Human-readable authentication failure.' },
        { name: 'request_id', type: 'string', required: false, description: 'Request identifier for debugging and support.' },
      ],
      example: JSON.stringify(
        {
          error: {
            code: 'UNAUTHORIZED',
            message: 'Missing or invalid API key.',
          },
          request_id: 'req_123',
        },
        null,
        2,
      ),
    },
    {
      status: 500,
      label: 'Internal error',
      description: 'TokenGate could not complete the request due to an internal or upstream failure.',
      fields: [
        { name: 'error.code', type: 'string', required: true, description: 'Usually INTERNAL_ERROR or UPSTREAM_ERROR.' },
        { name: 'error.message', type: 'string', required: true, description: 'Human-readable failure reason.' },
        { name: 'request_id', type: 'string', required: false, description: 'Request identifier for debugging and support.' },
      ],
      example: JSON.stringify(
        {
          error: {
            code: 'UPSTREAM_ERROR',
            message: 'Upstream provider request failed.',
          },
          request_id: 'req_123',
        },
        null,
        2,
      ),
    },
  ],
  examples: {
    curl: `curl "${baseUrl}/v1/chat/completions" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-5.4",
    "messages": [
      { "role": "user", "content": "Say hello from TokenGate" }
    ]
  }'`,
    node: `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.TOKENGATE_API_KEY,
  baseURL: "${baseUrl}/v1",
});

const completion = await client.chat.completions.create({
  model: "gpt-5.4",
  messages: [{ role: "user", content: "Say hello from TokenGate" }],
});

console.log(completion.choices[0]?.message?.content);`,
    python: `import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["TOKENGATE_API_KEY"],
    base_url="${baseUrl}/v1",
)

completion = client.chat.completions.create(
    model="gpt-5.4",
    messages=[{"role": "user", "content": "Say hello from TokenGate"}],
)

print(completion.choices[0].message.content)`,
    go: `client := openai.NewClient(
  option.WithAPIKey(os.Getenv("TOKENGATE_API_KEY")),
  option.WithBaseURL("${baseUrl}/v1"),
)

completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
  Model: "gpt-5.4",
  Messages: []openai.ChatCompletionMessageParamUnion{
    openai.UserMessage("Say hello from TokenGate"),
  },
})`,
    java: `OpenAIClient client = OpenAIOkHttpClient.builder()
    .apiKey(System.getenv("TOKENGATE_API_KEY"))
    .baseUrl("${baseUrl}/v1")
    .build();

ChatCompletionCreateParams params = ChatCompletionCreateParams.builder()
    .model("gpt-5.4")
    .addUserMessage("Say hello from TokenGate")
    .build();

ChatCompletion completion = client.chat().completions().create(params);`,
  },
}

const unauthorizedResponse = tokenGateApiEndpoint.responses[1]
const internalErrorResponse = tokenGateApiEndpoint.responses[2]

const noAuth: ApiAuthConfig = {
  type: 'None',
  header: 'n/a',
  required: false,
  description: 'No authentication is required for this public endpoint.',
}

const sessionAuth: ApiAuthConfig = {
  type: 'Bearer token',
  header: 'Authorization',
  required: true,
  description: 'User session access token returned by the TokenGate authentication flow.',
}

const adminSessionAuth: ApiAuthConfig = {
  type: 'Bearer token',
  header: 'Authorization',
  required: true,
  description: 'Admin session access token. These endpoints are intended for the TokenGate operator dashboard.',
}

const okEnvelopeResponse: ApiResponse = {
  status: 200,
  label: 'OK',
  description: 'The request completed successfully.',
  fields: [
    { name: 'code', type: 'number', required: false, description: 'Application-level status code when the endpoint uses the TokenGate envelope.' },
    { name: 'data', type: 'object | array', required: false, description: 'Endpoint-specific response payload.' },
    { name: 'message', type: 'string', required: false, description: 'Optional human-readable message.' },
  ],
  example: JSON.stringify({ code: 0, data: {}, message: 'success' }, null, 2),
}

type EndpointDraft = {
  id: string
  title: string
  section: string
  description: string
  method: ApiMethod
  path: string
  auth?: ApiAuthConfig
  parameters?: ApiParameter[]
  responses?: ApiResponse[]
  bodyExample?: string
}

function curlFor(method: ApiMethod, path: string, auth: ApiAuthConfig, bodyExample?: string) {
  const lines = [`curl "${baseUrl}${path}"`]
  if (method !== 'GET') {
    lines.push(`  -X ${method}`)
  }
  if (auth.required) {
    const tokenName = auth === adminSessionAuth || auth === sessionAuth ? '$TOKENGATE_SESSION_TOKEN' : '$TOKENGATE_API_KEY'
    lines.push(`  -H "Authorization: Bearer ${tokenName}"`)
  }
  if (bodyExample) {
    lines.push('  -H "Content-Type: application/json"')
    lines.push(`  -d '${bodyExample}'`)
  }
  return lines.join(' \\\n')
}

function examplesFor(method: ApiMethod, path: string, auth: ApiAuthConfig, bodyExample?: string): ApiExamples {
  const hasBody = Boolean(bodyExample)
  const tokenName = auth === adminSessionAuth || auth === sessionAuth ? 'TOKENGATE_SESSION_TOKEN' : 'TOKENGATE_API_KEY'
  return {
    curl: curlFor(method, path, auth, bodyExample),
    node: `const response = await fetch("${baseUrl}${path}", {
  method: "${method}",
  headers: {
${auth.required ? `    Authorization: \`Bearer \${process.env.${tokenName}}\`,\n` : ''}${hasBody ? '    "Content-Type": "application/json",\n' : ''}  },${hasBody ? `\n  body: JSON.stringify(${bodyExample}),` : ''}
});

const data = await response.json();`,
    python: `import os
import requests

response = requests.request(
    "${method}",
    "${baseUrl}${path}",
    headers={${auth.required ? `"Authorization": f"Bearer {os.environ['${tokenName}']}"` : ''}},
${hasBody ? `    json=${bodyExample},` : ''}    
)
data = response.json()`,
    go: `req, _ := http.NewRequest("${method}", "${baseUrl}${path}", ${hasBody ? 'bytes.NewBuffer(payload)' : 'nil'})
${auth.required ? `req.Header.Set("Authorization", "Bearer "+os.Getenv("${tokenName}"))\n` : ''}${hasBody ? 'req.Header.Set("Content-Type", "application/json")\n' : ''}resp, err := http.DefaultClient.Do(req)`,
    java: `HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("${baseUrl}${path}"))
${auth.required ? `    .header("Authorization", "Bearer " + System.getenv("${tokenName}"))\n` : ''}${hasBody ? '    .header("Content-Type", "application/json")\n' : ''}    .method("${method}", ${hasBody ? 'HttpRequest.BodyPublishers.ofString(payload)' : 'HttpRequest.BodyPublishers.noBody()'})
    .build();`,
  }
}

function endpointFromDraft(draft: EndpointDraft): ApiEndpointConfig {
  const auth = draft.auth ?? sessionAuth
  return {
    ...tokenGateApiEndpoint,
    id: draft.id,
    title: draft.title,
    section: draft.section,
    description: draft.description,
    method: draft.method,
    path: draft.path,
    auth,
    parameters: draft.parameters ?? [],
    responses: draft.responses ?? [okEnvelopeResponse, unauthorizedResponse, internalErrorResponse],
    examples: examplesFor(draft.method, draft.path, auth, draft.bodyExample),
  }
}

const additionalApiEndpoints: ApiEndpointConfig[] = [
  endpointFromDraft({
    id: 'health-check',
    title: 'Health check',
    section: 'Core',
    description: 'Checks whether the TokenGate backend is reachable.',
    method: 'GET',
    path: '/health',
    auth: noAuth,
    responses: [
      {
        status: 200,
        label: 'OK',
        description: 'The service is healthy.',
        fields: [{ name: 'status', type: 'string', required: true, description: 'Health status, usually ok.' }],
        example: JSON.stringify({ status: 'ok' }, null, 2),
      },
      unauthorizedResponse,
      internalErrorResponse,
    ],
  }),
  endpointFromDraft({
    id: 'gateway-usage',
    title: 'Get gateway usage',
    section: 'Core',
    description: 'Returns usage information for the API key on the model gateway.',
    method: 'GET',
    path: '/v1/usage',
    auth: tokenGateApiEndpoint.auth,
  }),
  endpointFromDraft({
    id: 'edit-image',
    title: 'Edit image',
    section: 'OpenAI compatible',
    description: 'Edits an image through the OpenAI-compatible Images API when the selected group supports image-capable upstream accounts.',
    method: 'POST',
    path: '/v1/images/edits',
    auth: tokenGateApiEndpoint.auth,
    parameters: [
      { name: 'model', location: 'body', type: 'string', required: true, description: 'Image editing model enabled for the group.' },
      { name: 'image', location: 'body', type: 'file | string', required: true, description: 'Source image payload depending on client format.' },
      { name: 'prompt', location: 'body', type: 'string', required: true, description: 'Editing instruction.' },
    ],
    bodyExample: `{
  "model": "gpt-image-1",
  "prompt": "Add a soft studio background"
}`,
  }),
  endpointFromDraft({
    id: 'gemini-list-models',
    title: 'List Gemini models',
    section: 'Gemini compatible',
    description: 'Lists models using the Gemini v1beta API shape for Gemini SDK and CLI compatibility.',
    method: 'GET',
    path: '/v1beta/models',
    auth: tokenGateApiEndpoint.auth,
  }),
  endpointFromDraft({
    id: 'gemini-get-model',
    title: 'Get Gemini model',
    section: 'Gemini compatible',
    description: 'Gets metadata for one Gemini-compatible model.',
    method: 'GET',
    path: '/v1beta/models/{model}',
    auth: tokenGateApiEndpoint.auth,
    parameters: [{ name: 'model', location: 'query', type: 'string', required: true, description: 'Model id in the path.' }],
  }),
  endpointFromDraft({
    id: 'gemini-generate-content',
    title: 'Generate content',
    section: 'Gemini compatible',
    description: 'Runs a Gemini-compatible generateContent request through TokenGate.',
    method: 'POST',
    path: '/v1beta/models/{model}:generateContent',
    auth: tokenGateApiEndpoint.auth,
    parameters: [
      { name: 'model', location: 'query', type: 'string', required: true, description: 'Model id in the path.' },
      { name: 'contents', location: 'body', type: 'array', required: true, description: 'Gemini content parts.' },
    ],
    bodyExample: `{
  "contents": [
    { "parts": [{ "text": "Say hello from TokenGate" }] }
  ]
}`,
  }),
  endpointFromDraft({
    id: 'antigravity-models',
    title: 'List Antigravity models',
    section: 'Antigravity compatible',
    description: 'Lists models from the Antigravity-only routing namespace.',
    method: 'GET',
    path: '/antigravity/v1/models',
    auth: tokenGateApiEndpoint.auth,
  }),
  endpointFromDraft({
    id: 'antigravity-messages',
    title: 'Create Antigravity message',
    section: 'Antigravity compatible',
    description: 'Sends an Anthropic-compatible message request through the Antigravity-only routing namespace.',
    method: 'POST',
    path: '/antigravity/v1/messages',
    auth: tokenGateApiEndpoint.auth,
    parameters: tokenGateApiEndpoint.parameters,
    bodyExample: `{
  "model": "claude-sonnet-4.6",
  "max_tokens": 256,
  "messages": [{ "role": "user", "content": "Say hello" }]
}`,
  }),
  endpointFromDraft({
    id: 'register',
    title: 'Register user',
    section: 'Auth',
    description: 'Creates a regular TokenGate user account when registration is enabled.',
    method: 'POST',
    path: '/api/v1/auth/register',
    auth: noAuth,
    parameters: [
      { name: 'email', location: 'body', type: 'string', required: true, description: 'User email address.' },
      { name: 'password', location: 'body', type: 'string', required: true, description: 'User password.' },
    ],
    bodyExample: `{
  "email": "user@example.com",
  "password": "strong-password"
}`,
  }),
  endpointFromDraft({
    id: 'login',
    title: 'Login',
    section: 'Auth',
    description: 'Authenticates a user and returns session tokens for dashboard/user APIs.',
    method: 'POST',
    path: '/api/v1/auth/login',
    auth: noAuth,
    bodyExample: `{
  "email": "user@example.com",
  "password": "strong-password"
}`,
  }),
  endpointFromDraft({
    id: 'refresh-token',
    title: 'Refresh token',
    section: 'Auth',
    description: 'Refreshes a user session access token.',
    method: 'POST',
    path: '/api/v1/auth/refresh',
    auth: noAuth,
  }),
  endpointFromDraft({
    id: 'current-user',
    title: 'Get current user',
    section: 'Auth',
    description: 'Returns the authenticated user for the current session.',
    method: 'GET',
    path: '/api/v1/auth/me',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'list-api-keys',
    title: 'List API keys',
    section: 'API keys',
    description: 'Lists API keys owned by the authenticated user.',
    method: 'GET',
    path: '/api/v1/keys',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'create-api-key',
    title: 'Create API key',
    section: 'API keys',
    description: 'Creates a customer API key and assigns it to an allowed group.',
    method: 'POST',
    path: '/api/v1/keys',
    auth: sessionAuth,
    parameters: [
      { name: 'name', location: 'body', type: 'string', required: true, description: 'Display name for the key.' },
      { name: 'group_id', location: 'body', type: 'number', required: false, description: 'Allowed group id to bind to the key.' },
    ],
    bodyExample: `{
  "name": "Production key",
  "group_id": 1
}`,
  }),
  endpointFromDraft({
    id: 'update-api-key',
    title: 'Update API key',
    section: 'API keys',
    description: 'Updates metadata or limits for one user-owned API key.',
    method: 'PUT',
    path: '/api/v1/keys/{id}',
    auth: sessionAuth,
    parameters: [{ name: 'id', location: 'query', type: 'number', required: true, description: 'API key id in the path.' }],
    bodyExample: `{
  "name": "Production key"
}`,
  }),
  endpointFromDraft({
    id: 'delete-api-key',
    title: 'Delete API key',
    section: 'API keys',
    description: 'Deletes one user-owned API key.',
    method: 'DELETE',
    path: '/api/v1/keys/{id}',
    auth: sessionAuth,
    parameters: [{ name: 'id', location: 'query', type: 'number', required: true, description: 'API key id in the path.' }],
  }),
  endpointFromDraft({
    id: 'user-profile',
    title: 'Get profile',
    section: 'User',
    description: 'Returns the profile, balance, and account metadata for the current user.',
    method: 'GET',
    path: '/api/v1/user/profile',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'available-groups',
    title: 'List available groups',
    section: 'User',
    description: 'Lists groups that the current user can assign to API keys.',
    method: 'GET',
    path: '/api/v1/groups/available',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'group-rates',
    title: 'List group rates',
    section: 'User',
    description: 'Returns model pricing multipliers and group billing details visible to the current user.',
    method: 'GET',
    path: '/api/v1/groups/rates',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'list-usage',
    title: 'List usage records',
    section: 'Usage',
    description: 'Lists request usage records for the authenticated user.',
    method: 'GET',
    path: '/api/v1/usage',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'usage-stats',
    title: 'Get usage stats',
    section: 'Usage',
    description: 'Returns summarized usage metrics for the authenticated user.',
    method: 'GET',
    path: '/api/v1/usage/stats',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'active-subscription',
    title: 'Get active subscription',
    section: 'Subscriptions',
    description: 'Returns the current active subscription for the authenticated user.',
    method: 'GET',
    path: '/api/v1/subscriptions/active',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'subscription-summary',
    title: 'Get subscription summary',
    section: 'Subscriptions',
    description: 'Returns quota, balance, and subscription summary information for the authenticated user.',
    method: 'GET',
    path: '/api/v1/subscriptions/summary',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'payment-plans',
    title: 'List payment plans',
    section: 'Payments',
    description: 'Lists purchase plans available to the authenticated user.',
    method: 'GET',
    path: '/api/v1/payment/plans',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'payment-checkout-info',
    title: 'Get checkout info',
    section: 'Payments',
    description: 'Returns checkout configuration needed before creating a payment order.',
    method: 'GET',
    path: '/api/v1/payment/checkout-info',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'create-order',
    title: 'Create order',
    section: 'Payments',
    description: 'Creates a payment order for Stripe, WeChat Pay, Alipay, or another configured provider.',
    method: 'POST',
    path: '/api/v1/payment/orders',
    auth: sessionAuth,
    parameters: [
      { name: 'plan_id', location: 'body', type: 'number', required: true, description: 'Payment plan id.' },
      { name: 'provider', location: 'body', type: 'string', required: true, description: 'Payment provider key, such as stripe, wxpay, or alipay.' },
    ],
    bodyExample: `{
  "plan_id": 1,
  "provider": "stripe"
}`,
  }),
  endpointFromDraft({
    id: 'my-orders',
    title: 'List my orders',
    section: 'Payments',
    description: 'Lists payment orders for the authenticated user.',
    method: 'GET',
    path: '/api/v1/payment/orders/my',
    auth: sessionAuth,
  }),
  endpointFromDraft({
    id: 'admin-dashboard-stats',
    title: 'Admin dashboard stats',
    section: 'Admin',
    description: 'Returns high-level platform statistics for the admin dashboard.',
    method: 'GET',
    path: '/api/v1/admin/dashboard/stats',
    auth: adminSessionAuth,
  }),
  endpointFromDraft({
    id: 'admin-list-users',
    title: 'Admin list users',
    section: 'Admin',
    description: 'Lists TokenGate users for operator workflows.',
    method: 'GET',
    path: '/api/v1/admin/users',
    auth: adminSessionAuth,
  }),
  endpointFromDraft({
    id: 'admin-list-groups',
    title: 'Admin list groups',
    section: 'Admin',
    description: 'Lists routing and billing groups configured by the operator.',
    method: 'GET',
    path: '/api/v1/admin/groups',
    auth: adminSessionAuth,
  }),
  endpointFromDraft({
    id: 'admin-list-accounts',
    title: 'Admin list accounts',
    section: 'Admin',
    description: 'Lists upstream provider accounts connected to TokenGate.',
    method: 'GET',
    path: '/api/v1/admin/accounts',
    auth: adminSessionAuth,
  }),
  endpointFromDraft({
    id: 'admin-test-account',
    title: 'Admin test account',
    section: 'Admin',
    description: 'Runs an account connection test from the operator dashboard.',
    method: 'POST',
    path: '/api/v1/admin/accounts/{id}/test',
    auth: adminSessionAuth,
    parameters: [{ name: 'id', location: 'query', type: 'number', required: true, description: 'Upstream account id in the path.' }],
  }),
]

export const tokenGateApiEndpoints: ApiEndpointConfig[] = [
  tokenGateApiEndpoint,
  {
    ...tokenGateApiEndpoint,
    id: 'list-models',
    title: 'List models',
    section: 'Core',
    description:
      'Returns the models visible to the API key. TokenGate filters this list by the key group, enabled upstream accounts, and model availability.',
    method: 'GET',
    path: '/v1/models',
    parameters: [],
    responses: [
      {
        status: 200,
        label: 'OK',
        description: 'The available model catalog for this key.',
        fields: [
          { name: 'object', type: 'string', required: true, description: 'Usually list.' },
          { name: 'data[]', type: 'array', required: true, description: 'Models the key can request.' },
          { name: 'data[].id', type: 'string', required: true, description: 'Model identifier to pass in generation requests.' },
          { name: 'data[].object', type: 'string', required: false, description: 'Usually model.' },
          { name: 'data[].owned_by', type: 'string', required: false, description: 'Provider or upstream owner when available.' },
        ],
        example: JSON.stringify(
          {
            object: 'list',
            data: [
              { id: 'gpt-5.4', object: 'model', owned_by: 'openai' },
              { id: 'claude-sonnet-4.6', object: 'model', owned_by: 'anthropic' },
            ],
          },
          null,
          2,
        ),
      },
      unauthorizedResponse,
      internalErrorResponse,
    ],
    examples: {
      ...tokenGateApiEndpoint.examples,
      curl: `curl "${baseUrl}/v1/models" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY"`,
      node: `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.TOKENGATE_API_KEY,
  baseURL: "${baseUrl}/v1",
});

const models = await client.models.list();
console.log(models.data.map((model) => model.id));`,
      python: `import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["TOKENGATE_API_KEY"],
    base_url="${baseUrl}/v1",
)

models = client.models.list()
print([model.id for model in models.data])`,
      go: `models, err := client.Models.List(ctx)
if err != nil {
  return err
}
fmt.Println(models.Data)`,
      java: `ModelListPage models = client.models().list();
models.data().forEach(model -> System.out.println(model.id()));`,
    },
  },
  {
    ...tokenGateApiEndpoint,
    id: 'responses-api',
    title: 'Create response',
    section: 'OpenAI compatible',
    description:
      'Creates a response through TokenGate using the OpenAI Responses API shape. Use this endpoint for newer OpenAI-compatible clients and multi-modal request bodies.',
    method: 'POST',
    path: '/v1/responses',
    parameters: [
      { name: 'model', location: 'body', type: 'string', required: true, description: 'Model id allowed by the API key group.' },
      { name: 'input', location: 'body', type: 'string | array', required: true, description: 'Input text or structured input items for the response.' },
      { name: 'instructions', location: 'body', type: 'string', required: false, description: 'System-level instructions forwarded to the model.' },
      { name: 'temperature', location: 'body', type: 'number', required: false, description: 'Sampling temperature when supported by the upstream provider.' },
      { name: 'stream', location: 'body', type: 'boolean', required: false, description: 'Whether to request a streaming response.' },
    ],
    responses: [
      {
        status: 200,
        label: 'OK',
        description: 'The upstream model returned an OpenAI-compatible response object.',
        fields: [
          { name: 'id', type: 'string', required: true, description: 'Provider response identifier.' },
          { name: 'object', type: 'string', required: true, description: 'Usually response.' },
          { name: 'model', type: 'string', required: true, description: 'Resolved model used for the request.' },
          { name: 'output[]', type: 'array', required: false, description: 'Output items returned by the model.' },
          { name: 'output_text', type: 'string', required: false, description: 'Convenience text field when provided by the client or provider.' },
          { name: 'usage', type: 'object', required: false, description: 'Token usage when reported by the upstream provider.' },
        ],
        example: JSON.stringify(
          {
            id: 'resp_tokengate_123',
            object: 'response',
            model: 'gpt-5.4',
            output_text: 'TokenGate Responses API is ready.',
            usage: { input_tokens: 11, output_tokens: 7, total_tokens: 18 },
          },
          null,
          2,
        ),
      },
      unauthorizedResponse,
      internalErrorResponse,
    ],
    examples: {
      ...tokenGateApiEndpoint.examples,
      curl: `curl "${baseUrl}/v1/responses" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-5.4",
    "input": "Say hello from TokenGate"
  }'`,
      node: `const response = await client.responses.create({
  model: "gpt-5.4",
  input: "Say hello from TokenGate",
});

console.log(response.output_text);`,
      python: `response = client.responses.create(
    model="gpt-5.4",
    input="Say hello from TokenGate",
)

print(response.output_text)`,
      go: `response, err := client.Responses.New(ctx, responses.ResponseNewParams{
  Model: "gpt-5.4",
  Input: responses.ResponseNewParamsInputUnion{
    OfString: openai.String("Say hello from TokenGate"),
  },
})`,
      java: `ResponseCreateParams params = ResponseCreateParams.builder()
    .model("gpt-5.4")
    .input("Say hello from TokenGate")
    .build();

Response response = client.responses().create(params);`,
    },
  },
  {
    ...tokenGateApiEndpoint,
    id: 'images-api',
    title: 'Generate image',
    section: 'OpenAI compatible',
    description:
      'Creates image output through an OpenAI-compatible image generation request. Availability depends on the group and upstream account capabilities.',
    method: 'POST',
    path: '/v1/images/generations',
    parameters: [
      { name: 'model', location: 'body', type: 'string', required: true, description: 'Image-capable model enabled for the group, such as gpt-image-1.' },
      { name: 'prompt', location: 'body', type: 'string', required: true, description: 'Text prompt describing the image to generate.' },
      { name: 'size', location: 'body', type: 'string', required: false, description: 'Requested image size, for example 1024x1024.' },
      { name: 'quality', location: 'body', type: 'string', required: false, description: 'Provider-specific quality setting when supported.' },
      { name: 'response_format', location: 'body', type: 'string', required: false, description: 'Image response format such as url or b64_json when supported.' },
    ],
    responses: [
      {
        status: 200,
        label: 'OK',
        description: 'The provider returned generated image data or URLs.',
        fields: [
          { name: 'created', type: 'number', required: false, description: 'Unix timestamp for response creation.' },
          { name: 'data[]', type: 'array', required: true, description: 'Generated image outputs.' },
          { name: 'data[].url', type: 'string', required: false, description: 'Image URL when URL output is requested.' },
          { name: 'data[].b64_json', type: 'string', required: false, description: 'Base64 image data when b64_json output is requested.' },
          { name: 'usage', type: 'object', required: false, description: 'Usage metadata when reported by the upstream provider.' },
        ],
        example: JSON.stringify(
          {
            created: 1778956800,
            data: [{ url: 'https://example.com/generated-image.png' }],
          },
          null,
          2,
        ),
      },
      unauthorizedResponse,
      internalErrorResponse,
    ],
    examples: {
      ...tokenGateApiEndpoint.examples,
      curl: `curl "${baseUrl}/v1/images/generations" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-image-1",
    "prompt": "A clean product hero image for TokenGate",
    "size": "1024x1024"
  }'`,
      node: `const image = await client.images.generate({
  model: "gpt-image-1",
  prompt: "A clean product hero image for TokenGate",
  size: "1024x1024",
});`,
      python: `image = client.images.generate(
    model="gpt-image-1",
    prompt="A clean product hero image for TokenGate",
    size="1024x1024",
)`,
      go: `image, err := client.Images.Generate(ctx, images.ImageGenerateParams{
  Model: "gpt-image-1",
  Prompt: "A clean product hero image for TokenGate",
})`,
      java: `ImageGenerateParams params = ImageGenerateParams.builder()
    .model("gpt-image-1")
    .prompt("A clean product hero image for TokenGate")
    .build();`,
    },
  },
  {
    ...tokenGateApiEndpoint,
    id: 'anthropic-messages',
    title: 'Create message',
    section: 'Anthropic compatible',
    description:
      'Creates a Claude message through TokenGate using the Anthropic Messages API request shape. Use this for Claude Code OAuth-backed groups.',
    method: 'POST',
    path: '/v1/messages',
    parameters: [
      { name: 'model', location: 'body', type: 'string', required: true, description: 'Claude-compatible model id allowed by the API key group.' },
      { name: 'messages', location: 'body', type: 'array', required: true, description: 'Conversation messages using Anthropic Messages format.' },
      { name: 'max_tokens', location: 'body', type: 'number', required: true, description: 'Maximum output token budget.' },
      { name: 'system', location: 'body', type: 'string | array', required: false, description: 'System prompt or structured system content.' },
      { name: 'anthropic-version', location: 'header', type: 'string', required: false, description: 'Anthropic API version header. TokenGate accepts and forwards it when provided.' },
    ],
    responses: [
      {
        status: 200,
        label: 'OK',
        description: 'The upstream Claude-compatible account returned a message response.',
        fields: [
          { name: 'id', type: 'string', required: true, description: 'Provider message identifier.' },
          { name: 'type', type: 'string', required: true, description: 'Usually message.' },
          { name: 'role', type: 'string', required: true, description: 'Usually assistant.' },
          { name: 'model', type: 'string', required: true, description: 'Resolved Claude model.' },
          { name: 'content[]', type: 'array', required: true, description: 'Content blocks returned by Claude.' },
          { name: 'usage.input_tokens', type: 'number', required: false, description: 'Input token count when available.' },
          { name: 'usage.output_tokens', type: 'number', required: false, description: 'Output token count when available.' },
        ],
        example: JSON.stringify(
          {
            id: 'msg_tokengate_123',
            type: 'message',
            role: 'assistant',
            model: 'claude-sonnet-4.6',
            content: [{ type: 'text', text: 'TokenGate Claude routing is ready.' }],
            usage: { input_tokens: 11, output_tokens: 8 },
          },
          null,
          2,
        ),
      },
      unauthorizedResponse,
      internalErrorResponse,
    ],
    examples: {
      ...tokenGateApiEndpoint.examples,
      curl: `curl "${baseUrl}/v1/messages" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -H "anthropic-version: 2023-06-01" \\
  -d '{
    "model": "claude-sonnet-4.6",
    "max_tokens": 256,
    "messages": [
      { "role": "user", "content": "Say hello from TokenGate" }
    ]
  }'`,
      node: `const message = await fetch("${baseUrl}/v1/messages", {
  method: "POST",
  headers: {
    Authorization: \`Bearer \${process.env.TOKENGATE_API_KEY}\`,
    "Content-Type": "application/json",
    "anthropic-version": "2023-06-01",
  },
  body: JSON.stringify({
    model: "claude-sonnet-4.6",
    max_tokens: 256,
    messages: [{ role: "user", content: "Say hello from TokenGate" }],
  }),
});`,
      python: `import os
import requests

response = requests.post(
    "${baseUrl}/v1/messages",
    headers={
        "Authorization": f"Bearer {os.environ['TOKENGATE_API_KEY']}",
        "Content-Type": "application/json",
        "anthropic-version": "2023-06-01",
    },
    json={
        "model": "claude-sonnet-4.6",
        "max_tokens": 256,
        "messages": [{"role": "user", "content": "Say hello from TokenGate"}],
    },
)`,
      go: `reqBody := strings.NewReader(\`{
  "model": "claude-sonnet-4.6",
  "max_tokens": 256,
  "messages": [{"role": "user", "content": "Say hello from TokenGate"}]
}\`)
req, _ := http.NewRequest("POST", "${baseUrl}/v1/messages", reqBody)`,
      java: `HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("${baseUrl}/v1/messages"))
    .header("Authorization", "Bearer " + System.getenv("TOKENGATE_API_KEY"))
    .header("Content-Type", "application/json")
    .header("anthropic-version", "2023-06-01")
    .POST(HttpRequest.BodyPublishers.ofString(payload))
    .build();`,
    },
  },
  {
    ...tokenGateApiEndpoint,
    id: 'count-tokens',
    title: 'Count message tokens',
    section: 'Anthropic compatible',
    description:
      'Counts tokens for an Anthropic-compatible message request when the selected upstream account supports token counting.',
    method: 'POST',
    path: '/v1/messages/count_tokens',
    parameters: [
      { name: 'model', location: 'body', type: 'string', required: true, description: 'Claude-compatible model id to use for token counting.' },
      { name: 'messages', location: 'body', type: 'array', required: true, description: 'Messages to estimate.' },
      { name: 'system', location: 'body', type: 'string | array', required: false, description: 'Optional system prompt included in the token estimate.' },
      { name: 'anthropic-version', location: 'header', type: 'string', required: false, description: 'Anthropic API version header when your client sends one.' },
    ],
    responses: [
      {
        status: 200,
        label: 'OK',
        description: 'The provider returned the estimated input token count.',
        fields: [
          { name: 'input_tokens', type: 'number', required: true, description: 'Estimated number of input tokens.' },
        ],
        example: JSON.stringify({ input_tokens: 14 }, null, 2),
      },
      unauthorizedResponse,
      internalErrorResponse,
    ],
    examples: {
      ...tokenGateApiEndpoint.examples,
      curl: `curl "${baseUrl}/v1/messages/count_tokens" \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -H "anthropic-version: 2023-06-01" \\
  -d '{
    "model": "claude-sonnet-4.6",
    "messages": [
      { "role": "user", "content": "Count these tokens" }
    ]
  }'`,
      node: `const result = await fetch("${baseUrl}/v1/messages/count_tokens", {
  method: "POST",
  headers: {
    Authorization: \`Bearer \${process.env.TOKENGATE_API_KEY}\`,
    "Content-Type": "application/json",
    "anthropic-version": "2023-06-01",
  },
  body: JSON.stringify({
    model: "claude-sonnet-4.6",
    messages: [{ role: "user", content: "Count these tokens" }],
  }),
});`,
      python: `response = requests.post(
    "${baseUrl}/v1/messages/count_tokens",
    headers={"Authorization": f"Bearer {os.environ['TOKENGATE_API_KEY']}"},
    json={
        "model": "claude-sonnet-4.6",
        "messages": [{"role": "user", "content": "Count these tokens"}],
    },
)`,
      go: `req, _ := http.NewRequest("POST", "${baseUrl}/v1/messages/count_tokens", reqBody)
req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKENGATE_API_KEY"))`,
      java: `HttpRequest request = HttpRequest.newBuilder()
    .uri(URI.create("${baseUrl}/v1/messages/count_tokens"))
    .header("Authorization", "Bearer " + System.getenv("TOKENGATE_API_KEY"))
    .POST(HttpRequest.BodyPublishers.ofString(payload))
    .build();`,
    },
  },
  ...additionalApiEndpoints,
]

export const tokenGateApiSidebarGroups: ApiSidebarGroup[] = [
  {
    title: 'Core',
    items: [
      { title: 'Health check', method: 'GET', href: '#health-check' },
      { title: 'List models', method: 'GET', href: '#list-models' },
      { title: 'Gateway usage', method: 'GET', href: '#gateway-usage' },
    ],
  },
  {
    title: 'OpenAI compatible',
    items: [
      { title: 'Chat completions', method: 'POST', href: '#chat-completions', active: true },
      { title: 'Responses', method: 'POST', href: '#responses-api' },
      { title: 'Image generation', method: 'POST', href: '#images-api' },
      { title: 'Image edits', method: 'POST', href: '#edit-image' },
    ],
  },
  {
    title: 'Anthropic compatible',
    items: [
      { title: 'Messages', method: 'POST', href: '#anthropic-messages' },
      { title: 'Count tokens', method: 'POST', href: '#count-tokens' },
    ],
  },
  {
    title: 'Gemini compatible',
    items: [
      { title: 'List models', method: 'GET', href: '#gemini-list-models' },
      { title: 'Get model', method: 'GET', href: '#gemini-get-model' },
      { title: 'Generate content', method: 'POST', href: '#gemini-generate-content' },
    ],
  },
  {
    title: 'Antigravity compatible',
    items: [
      { title: 'List models', method: 'GET', href: '#antigravity-models' },
      { title: 'Messages', method: 'POST', href: '#antigravity-messages' },
    ],
  },
  {
    title: 'Auth',
    items: [
      { title: 'Register user', method: 'POST', href: '#register' },
      { title: 'Login', method: 'POST', href: '#login' },
      { title: 'Refresh token', method: 'POST', href: '#refresh-token' },
      { title: 'Current user', method: 'GET', href: '#current-user' },
    ],
  },
  {
    title: 'API keys',
    items: [
      { title: 'List keys', method: 'GET', href: '#list-api-keys' },
      { title: 'Create key', method: 'POST', href: '#create-api-key' },
      { title: 'Update key', method: 'PUT', href: '#update-api-key' },
      { title: 'Delete key', method: 'DELETE', href: '#delete-api-key' },
    ],
  },
  {
    title: 'User',
    items: [
      { title: 'Profile', method: 'GET', href: '#user-profile' },
      { title: 'Available groups', method: 'GET', href: '#available-groups' },
      { title: 'Group rates', method: 'GET', href: '#group-rates' },
    ],
  },
  {
    title: 'Usage',
    items: [
      { title: 'List usage', method: 'GET', href: '#list-usage' },
      { title: 'Usage stats', method: 'GET', href: '#usage-stats' },
    ],
  },
  {
    title: 'Subscriptions',
    items: [
      { title: 'Active subscription', method: 'GET', href: '#active-subscription' },
      { title: 'Summary', method: 'GET', href: '#subscription-summary' },
    ],
  },
  {
    title: 'Payments',
    items: [
      { title: 'List plans', method: 'GET', href: '#payment-plans' },
      { title: 'Checkout info', method: 'GET', href: '#payment-checkout-info' },
      { title: 'Create order', method: 'POST', href: '#create-order' },
      { title: 'My orders', method: 'GET', href: '#my-orders' },
    ],
  },
  {
    title: 'Admin',
    items: [
      { title: 'Dashboard stats', method: 'GET', href: '#admin-dashboard-stats' },
      { title: 'List users', method: 'GET', href: '#admin-list-users' },
      { title: 'List groups', method: 'GET', href: '#admin-list-groups' },
      { title: 'List accounts', method: 'GET', href: '#admin-list-accounts' },
      { title: 'Test account', method: 'POST', href: '#admin-test-account' },
    ],
  },
]

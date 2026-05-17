export type ApiMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE'

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

export type ApiExamples = {
  curl: string
  node: string
  python: string
  go: string
  java: string
}

export type ApiEndpointConfig = {
  title: string
  section: string
  description: string
  method: ApiMethod
  baseUrl: string
  path: string
  auth: ApiAuthConfig
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
          model: 'gpt-5.1-mini',
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
    curl: `curl ${baseUrl}/v1/chat/completions \\
  -H "Authorization: Bearer $TOKENGATE_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-5.1-mini",
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
  model: "gpt-5.1-mini",
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
    model="gpt-5.1-mini",
    messages=[{"role": "user", "content": "Say hello from TokenGate"}],
)

print(completion.choices[0].message.content)`,
    go: `client := openai.NewClient(
  option.WithAPIKey(os.Getenv("TOKENGATE_API_KEY")),
  option.WithBaseURL("${baseUrl}/v1"),
)

completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
  Model: "gpt-5.1-mini",
  Messages: []openai.ChatCompletionMessageParamUnion{
    openai.UserMessage("Say hello from TokenGate"),
  },
})`,
    java: `OpenAIClient client = OpenAIOkHttpClient.builder()
    .apiKey(System.getenv("TOKENGATE_API_KEY"))
    .baseUrl("${baseUrl}/v1")
    .build();

ChatCompletionCreateParams params = ChatCompletionCreateParams.builder()
    .model("gpt-5.1-mini")
    .addUserMessage("Say hello from TokenGate")
    .build();

ChatCompletion completion = client.chat().completions().create(params);`,
  },
}

export const tokenGateApiSidebarGroups: ApiSidebarGroup[] = [
  {
    title: 'Core',
    items: [
      { title: 'List models', method: 'GET', href: '#list-models' },
      { title: 'Create chat completion', method: 'POST', href: '#endpoint', active: true },
      { title: 'Create response', method: 'POST', href: '#responses-api' },
    ],
  },
  {
    title: 'OpenAI compatible',
    items: [
      { title: 'Chat completions', method: 'POST', href: '#endpoint', active: true },
      { title: 'Responses', method: 'POST', href: '#responses-api' },
      { title: 'Image generation', method: 'POST', href: '#images-api' },
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
    title: 'Accounts',
    items: [
      { title: 'Provider accounts', href: '#accounts' },
      { title: 'Groups and routing', href: '#groups' },
    ],
  },
  {
    title: 'Users',
    items: [
      { title: 'Customer balance', href: '#balance' },
      { title: 'Usage records', href: '#usage' },
    ],
  },
  {
    title: 'API keys',
    items: [
      { title: 'Create key', href: '#api-keys' },
      { title: 'Assign group', href: '#groups' },
    ],
  },
  {
    title: 'Billing',
    items: [
      { title: 'Plans', href: '#billing' },
      { title: 'Top-ups', href: '#billing' },
    ],
  },
  {
    title: 'Developer Webhooks',
    items: [
      { title: 'Roadmap', href: '#webhooks' },
    ],
  },
]

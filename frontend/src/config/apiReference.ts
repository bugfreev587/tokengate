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
    { name: 'model', location: 'body', type: 'string', required: true, description: 'Model id allowed by the API key group, such as gpt-5.1-mini.' },
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
    curl: `curl "${baseUrl}/v1/chat/completions" \\
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

const unauthorizedResponse = tokenGateApiEndpoint.responses[1]
const internalErrorResponse = tokenGateApiEndpoint.responses[2]

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
              { id: 'gpt-5.1-mini', object: 'model', owned_by: 'openai' },
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
            model: 'gpt-5.1-mini',
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
    "model": "gpt-5.1-mini",
    "input": "Say hello from TokenGate"
  }'`,
      node: `const response = await client.responses.create({
  model: "gpt-5.1-mini",
  input: "Say hello from TokenGate",
});

console.log(response.output_text);`,
      python: `response = client.responses.create(
    model="gpt-5.1-mini",
    input="Say hello from TokenGate",
)

print(response.output_text)`,
      go: `response, err := client.Responses.New(ctx, responses.ResponseNewParams{
  Model: "gpt-5.1-mini",
  Input: responses.ResponseNewParamsInputUnion{
    OfString: openai.String("Say hello from TokenGate"),
  },
})`,
      java: `ResponseCreateParams params = ResponseCreateParams.builder()
    .model("gpt-5.1-mini")
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
]

export const tokenGateApiSidebarGroups: ApiSidebarGroup[] = [
  {
    title: 'Core',
    items: [
      { title: 'List models', method: 'GET', href: '#list-models' },
    ],
  },
  {
    title: 'OpenAI compatible',
    items: [
      { title: 'Chat completions', method: 'POST', href: '#chat-completions', active: true },
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
]

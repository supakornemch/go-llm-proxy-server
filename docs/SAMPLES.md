# 💻 Sample Usage

The Go LLM Proxy Server is designed to be a drop-in replacement for OpenAI endpoints. Here are examples for various libraries.

## Python

### 1. OpenAI SDK
```python
from openai import OpenAI

client = OpenAI(
    api_key="your-virtual-key",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="gpt-4o", # or your gemini-alias
    messages=[{"role": "user", "content": "Explain quantum physics."}]
)
print(response.choices[0].message.content)
```

### 2. LangChain
```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="gemini-1.5-flash",
    api_key="your-virtual-key",
    base_url="http://localhost:8080/v1"
)
print(llm.invoke("Hello!").content)
```

## JavaScript / TypeScript

### 1. OpenAI SDK
```typescript
import OpenAI from 'openai';

const openai = new OpenAI({
  apiKey: 'your-virtual-key',
  baseURL: 'http://localhost:8080/v1',
});

async function main() {
  const completion = await openai.chat.completions.create({
    messages: [{ role: 'user', content: 'Say hello!' }],
    model: 'gpt-4o',
  });
  console.log(completion.choices[0]);
}
main();
```

## Go

```go
package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

func main() {
	config := openai.DefaultConfig("your-virtual-key")
	config.BaseURL = "http://localhost:8080/v1"

	client := openai.NewClientWithConfig(config)
	resp, _ := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4o",
			Messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "Hello!"},
			},
		},
	)
	fmt.Println(resp.Choices[0].Message.Content)
}
```

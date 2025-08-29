# Testing AgentRuntime with the Playground

## 🚀 New to AgentRuntime?

**First time using AgentRuntime?** Start with our [Quick Start Guide for Beginners](quickstart.md) - get up and running in just 5 minutes!

---

## 🔧 Advanced Guide (For Developers)

Want to build from source or customize further? Continue reading...

## What is the Playground?

The playground is a web-based testing environment that allows you to interact with your agentruntime agents through a user-friendly chat interface. It's designed to help both beginners and experienced developers:

- Test agent configurations quickly
- Debug agent behavior in real-time
- Experiment with different conversation flows
- Understand how agents respond to various inputs

Think of it as a local chat application where you can talk to your custom AI agents before deploying them to production.

## Prerequisites

Before you start, make sure you have:

1. **Go installed** (version 1.19 or later)
2. **Node.js and Yarn** for the web interface
3. **API Keys** for your chosen AI model provider (OpenAI, Anthropic, or xAI)

## Step 1: Set Up Your Environment

### 1.1 Clone and Build AgentRuntime

```bash
# Clone the repository
git clone https://github.com/habiliai/agentruntime
cd agentruntime

# Build the agentruntime binary
make build

# Or run directly with Go
go build -o agentruntime .
```

### 1.2 Set Up API Keys

Create environment variables for your AI provider:

```bash
# For OpenAI (recommended for beginners)
export OPENAI_API_KEY="your-openai-api-key-here"

# For Anthropic Claude
export ANTHROPIC_API_KEY="your-anthropic-api-key-here"

# For xAI Grok
export XAI_API_KEY="your-xai-api-key-here"
```

## Step 2: Create Your First Agent

### 2.1 Create an Agent Directory

```bash
# Create a directory for your agents
mkdir my-agents
cd my-agents
```

### 2.2 Create a Simple Agent Configuration

Create a file called `helper-agent.yaml`:

```yaml
name: 'helper-agent'
description: 'A friendly assistant that helps with general questions'
model: 'gpt-5-mini' # Cost-effective for testing
instructions: |
  You are a helpful and friendly assistant. 
  Answer questions clearly and concisely. 
  If you don't know something, admit it and suggest how the user might find the answer.
temperature: 0.7
```

### 2.3 Create a Specialist Agent (Optional)

Create another file called `code-reviewer.yaml`:

```yaml
name: 'code-reviewer'
description: 'An expert code reviewer specializing in best practices'
model: 'gpt-4o'
instructions: |
  You are an experienced software engineer focused on code quality.
  When reviewing code:
  1. Check for bugs and potential issues
  2. Suggest improvements for readability
  3. Recommend best practices
  4. Be constructive and educational in your feedback

  Always explain your reasoning and provide examples when possible.
temperature: 0.3
max_tokens: 1000
```

## Step 3: Start AgentRuntime Server

### 3.1 Run AgentRuntime with Your Agents

```bash
# Navigate back to the agentruntime directory
cd /path/to/agentruntime

# Start the server with your agents
./agentruntime /path/to/my-agents -p 3001

# Or using Go directly
go run . /path/to/my-agents -p 3001
```

You should see output like:

```
2024/01/15 10:30:00 Loading agents from: /path/to/my-agents
2024/01/15 10:30:00 Loaded agent: helper-agent
2024/01/15 10:30:00 Loaded agent: code-reviewer
2024/01/15 10:30:00 AgentRuntime server starting on port 3001
```

### 3.2 Verify Server is Running

Open another terminal and test:

```bash
curl http://localhost:3001/health
# Should return: {"status": "ok"}
```

## Step 4: Start the Playground

### 4.1 Install Dependencies

```bash
# Navigate to the playground directory
cd playground

# Install dependencies
yarn install
```

### 4.2 Start the Development Server

```bash
yarn dev
```

The playground will be available at [http://localhost:3000](http://localhost:3000).

## Step 5: Test Your Agents

### 5.1 Create Your First Thread

1. Open [http://localhost:3000](http://localhost:3000) in your browser
2. Click **"Create Thread"**
3. In the dialog:
   - **Instruction**: "Help me test my new agents" (optional)
   - **Participants**: Check the boxes for agents you want to include
4. Click **"Create Thread"**

### 5.2 Start Testing

Now you can:

1. **Send messages** to test how your agents respond
2. **Ask questions** to verify their knowledge and behavior
3. **Test edge cases** to see how they handle unusual inputs
4. **Compare responses** between different agents

### Example Test Conversations

**Testing the Helper Agent:**

```
You: What's the weather like today?
helper-agent: I don't have access to real-time weather data, but I can suggest a few ways you can check the weather:
1. Check a weather app on your phone
2. Visit weather.com or weather.gov
3. Ask a voice assistant like Siri or Google Assistant
Is there anything else I can help you with?
```

**Testing the Code Reviewer:**

````
You: Please review this Python function:
def calculate_average(numbers):
    return sum(numbers) / len(numbers)

code-reviewer: Here's my review of your `calculate_average` function:

**Potential Issues:**
1. **Division by zero**: If an empty list is passed, `len(numbers)` will be 0, causing a ZeroDivisionError.

**Suggested Improvement:**
```python
def calculate_average(numbers):
    if not numbers:
        raise ValueError("Cannot calculate average of empty list")
    return sum(numbers) / len(numbers)
````

**Alternative with default return:**

```python
def calculate_average(numbers):
    return sum(numbers) / len(numbers) if numbers else 0
```

The function is clean and readable, but adding error handling makes it more robust.

```

## Step 6: Advanced Testing

### 6.1 Test Multiple Agents in One Thread

1. Create a thread with multiple agents selected
2. Ask questions that might benefit from different perspectives
3. See how different agents approach the same problem

Example:
```

You: How should I structure a new web application?

helper-agent: For a new web application, consider starting with:

1. Choose your tech stack (React, Vue, or vanilla JS for frontend)
2. Decide on backend technology (Node.js, Python, etc.)
3. Plan your database structure
4. Set up version control with Git
   Would you like me to elaborate on any of these areas?

code-reviewer: From a code quality perspective, here's what I recommend for structuring a web app:

1. **Separation of concerns**: Keep frontend, backend, and database layers distinct
2. **Folder structure**: Use a consistent directory structure like:
   ```
   /src
     /components
     /services
     /utils
     /tests
   ```
3. **Configuration management**: Use environment variables for different deployment stages
4. **Testing strategy**: Plan for unit, integration, and e2e tests from the start
   This foundation will make your code more maintainable as it grows.

````

### 6.2 Test Agent Configurations

Experiment with different agent settings:

1. **Temperature**: Lower values (0.1-0.3) for consistent responses, higher (0.7-1.0) for creativity
2. **Model**: Try different models to balance cost and quality
3. **Instructions**: Refine prompts based on how agents respond
4. **Max tokens**: Adjust for longer or shorter responses

### 6.3 Test Error Handling

Try scenarios that might cause issues:
- Very long messages
- Messages in different languages
- Technical jargon or specialized terminology
- Ambiguous or unclear requests

## Troubleshooting

### Common Issues and Solutions

#### 1. "Connection Refused" Error
**Problem**: Can't connect to agentruntime server
**Solutions**:
- Check if agentruntime is running: `curl http://localhost:3001/health`
- Verify the port matches (default is 3001)
- Check for error messages in the agentruntime console

#### 2. "No Agents Available"
**Problem**: Playground shows no agents to select
**Solutions**:
- Ensure your YAML files are in the correct directory
- Check YAML syntax (indentation matters!)
- Verify agentruntime loaded your agents (check console output)

#### 3. Agent Not Responding
**Problem**: Agent appears but doesn't respond to messages
**Solutions**:
- Check API key is set correctly
- Verify the model name in your YAML (e.g., 'gpt-5-mini', not 'gpt-4-mini')
- Check agentruntime logs for API errors
- Try a simpler model first (like 'gpt-3.5-turbo')

#### 4. CORS Errors
**Problem**: Browser console shows CORS errors
**Solutions**:
- Ensure agentruntime server has CORS enabled
- Access playground from `http://localhost:3000` (not 127.0.0.1)
- Restart both agentruntime and playground servers

#### 5. YAML Configuration Errors
**Problem**: Agent fails to load due to YAML syntax
**Solutions**:
- Use spaces, not tabs for indentation
- Ensure strings with special characters are quoted
- Use a YAML validator online to check syntax
- Check the example configurations in this guide

### Debug Mode

For more detailed debugging, run agentruntime with verbose logging:

```bash
./agentruntime /path/to/my-agents -p 3001 -v
````

## Best Practices for Testing

### 1. Start Simple

- Begin with basic agents using simple instructions
- Test one agent at a time initially
- Use cost-effective models during development (gpt-5-mini, gpt-3.5-turbo)

### 2. Iterative Testing

- Test → Observe → Adjust → Repeat
- Keep notes on what works and what doesn't
- Version your agent configurations with Git

### 3. Test Edge Cases

- Empty messages
- Very long messages
- Messages in different languages
- Technical or domain-specific questions

### 4. Performance Testing

- Monitor response times
- Test with multiple concurrent conversations
- Check memory usage with many agents loaded

### 5. Documentation

- Document successful agent configurations
- Keep examples of good and bad responses
- Share learnings with your team

## Example Agent Templates

### Customer Support Agent

```yaml
name: 'support-agent'
description: 'Helpful customer support representative'
model: 'gpt-5-mini'
instructions: |
  You are a friendly and professional customer support representative.
  Always:
  1. Greet customers warmly
  2. Listen carefully to their concerns
  3. Provide clear, step-by-step solutions
  4. Ask follow-up questions if needed
  5. Offer additional help before ending the conversation

  Be patient, empathetic, and solution-focused.
temperature: 0.5
max_tokens: 500
```

### Technical Writer Agent

```yaml
name: 'tech-writer'
description: 'Technical documentation specialist'
model: 'gpt-4o'
instructions: |
  You are an experienced technical writer who creates clear, comprehensive documentation.

  When writing documentation:
  1. Use clear, concise language
  2. Structure information logically
  3. Include practical examples
  4. Anticipate user questions
  5. Use appropriate formatting (headers, lists, code blocks)

  Focus on making complex topics accessible to your target audience.
temperature: 0.3
max_tokens: 1500
```

### Creative Writing Assistant

```yaml
name: 'creative-writer'
description: 'Creative writing and storytelling assistant'
model: 'gpt-4o'
instructions: |
  You are a creative writing assistant who helps with storytelling, character development, and narrative structure.

  You excel at:
  1. Generating creative ideas and plot points
  2. Developing compelling characters with depth
  3. Suggesting improvements to dialogue and pacing
  4. Helping overcome writer's block
  5. Providing constructive feedback on creative work

  Be encouraging, imaginative, and supportive of the writer's vision.
temperature: 0.8
max_tokens: 1000
```

## Next Steps

Once you're comfortable with the playground:

1. **Deploy to Production**: Use your tested agents in real applications
2. **Advanced Features**: Explore tool calling and external integrations
3. **Monitoring**: Set up logging and monitoring for production agents
4. **Scaling**: Learn about running multiple agentruntime instances

## Getting Help

- Check the main [README.md](../README.md) for general setup
- Review agent logs for debugging information
- Test with simple configurations first
- Use the GitHub issues for bug reports and feature requests

Happy testing! 🚀

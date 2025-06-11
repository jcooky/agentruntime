# Agent Configuration Guide

This guide explains how to configure AI agents using the agentruntime framework. Agents are defined using a structured configuration that includes identity, capabilities, and behavior specifications.

## Agent Structure Overview

An agent in agentruntime is composed of several key components:

- **Agent Card**: Basic identity and metadata
- **Model Configuration**: AI model settings
- **Behavior Definition**: System prompts and role definitions
- **Skills**: Capabilities and tools the agent can use
- **Knowledge**: Information sources and context
- **Evaluation**: Testing and validation configuration

## Agent Properties Reference

The following table provides a quick reference for all agent configuration properties:

| Property                        | Type   | Required | Description                                                    |
| ------------------------------- | ------ | -------- | -------------------------------------------------------------- |
| **AgentCard Properties**        |
| `name`                          | string | ✅       | Human-readable name of the agent                               |
| `description`                   | string | ✅       | Clear description of the agent's purpose and capabilities      |
| `url`                           | string | ✅       | The address where the agent is hosted                          |
| `iconUrl`                       | string | ❌       | URL to an icon representing the agent                          |
| `version`                       | string | ✅       | Version identifier for the agent (e.g., "1.0.0")               |
| `documentationUrl`              | string | ❌       | Link to detailed documentation for the agent                   |
| `defaultInputModes`             | array  | ✅       | Supported input media types (e.g., "text", "image", "audio")   |
| `defaultOutputModes`            | array  | ✅       | Supported output media types                                   |
| `provider`                      | object | ❌       | Information about the service provider                         |
| `provider.organization`         | string | ❌       | Agent provider's organization name                             |
| `provider.url`                  | string | ❌       | Agent provider's URL                                           |
| **Model & Behavior Properties** |
| `model`                         | string | ❌       | Name/identifier of the AI model to use                         |
| `modelConfig`                   | object | ❌       | Model-specific configuration parameters                        |
| `system`                        | string | ❌       | System-level instructions defining personality and constraints |
| `role`                          | string | ❌       | The role or persona the agent should adopt                     |
| `prompt`                        | string | ❌       | Additional prompt instructions for specific tasks              |
| **Training & Examples**         |
| `messageExamples`               | array  | ❌       | Arrays of conversation examples for training                   |
| `messageExamples[].user`        | string | ❌       | User input example                                             |
| `messageExamples[].text`        | string | ❌       | Expected agent response                                        |
| `messageExamples[].actions`     | array  | ❌       | Actions the agent should consider                              |
| **Skills & Capabilities**       |
| `skills`                        | array  | ❌       | List of agent capabilities and tools                           |
| `skills[].type`                 | string | ✅       | Skill type: "llm", "mcp", or "nativeTool"                      |
| `skills[].name`                 | string | ❌       | Identifier for the skill                                       |
| `skills[].description`          | string | ❌       | Human-readable description of the skill                        |
| `skills[].instruction`          | string | ❌       | Instructions for LLM skills                                    |
| `skills[].command`              | string | ❌       | Command to run MCP server                                      |
| `skills[].args`                 | array  | ❌       | Arguments for MCP server                                       |
| `skills[].tools`                | array  | ❌       | List of MCP tool names                                         |
| `skills[].env`                  | object | ❌       | Environment variables or configuration                         |
| **Knowledge & Data**            |
| `knowledge`                     | array  | ❌       | Information sources and context data                           |
| **Evaluation & Testing**        |
| `evaluator`                     | object | ❌       | Testing and validation configuration                           |
| `evaluator.prompt`              | string | ❌       | Instructions for evaluating agent responses                    |
| `evaluator.numRetries`          | int    | ❌       | Number of retry attempts for evaluation                        |
| **Additional Configuration**    |
| `metadata`                      | object | ❌       | Additional configuration, tags, and custom properties          |

## Core Agent Properties

### AgentCard (Identity & Metadata)

The AgentCard contains essential information about your agent:

```yaml
name: Recipe Assistant
description: An AI agent that helps users with recipes and cooking advice
url: https://api.example.com/agents/recipe-assistant
iconUrl: https://example.com/icons/chef.png
version: '1.2.0'
documentationUrl: https://docs.example.com/recipe-assistant
defaultInputModes:
  - text
  - image
defaultOutputModes:
  - text
  - image
provider:
  organization: CookingCorp
  url: https://cookingcorp.com
```

**Properties:**

- `name` (string, required): Human-readable name of your agent
- `description` (string, required): Clear description of the agent's purpose and capabilities
- `url` (string, required): The address where your agent is hosted
- `iconUrl` (string, optional): URL to an icon representing your agent
- `version` (string, required): Version identifier for your agent (e.g., "1.0.0")
- `documentationUrl` (string, optional): Link to detailed documentation
- `defaultInputModes` (array): Supported input media types (e.g., "text", "image", "audio")
- `defaultOutputModes` (array): Supported output media types
- `provider` (object, optional): Information about the service provider

### Model Configuration

Configure the underlying AI model and its parameters:

```yaml
model: claude-3-sonnet-20240229
modelConfig:
  temperature: 0.7
  maxTokens: 4000
  topP: 0.9
```

**Properties:**

- `model` (string): Name/identifier of the AI model to use
- `modelConfig` (object): Model-specific configuration parameters

### Behavior Definition

Define how your agent behaves and responds:

```yaml
system: You are a helpful cooking assistant with expertise in international cuisine.
role: cooking_expert
prompt: Help users with recipes, cooking techniques, and meal planning.
```

**Properties:**

- `system` (string): System-level instructions that define the agent's personality and constraints
- `role` (string): The role or persona the agent should adopt
- `prompt` (string): Additional prompt instructions for specific tasks

### Message Examples

Provide training examples to guide the agent's responses:

```yaml
messageExamples:
  - - user: How do I make pasta?
      text: "To make pasta, you'll need flour, eggs, and salt. Here's a step-by-step guide:"
      actions:
        - provide_recipe
        - suggest_tools
```

**Properties:**

- `messageExamples` (array): Arrays of conversation examples
  - `user` (string): User input example
  - `text` (string): Expected agent response
  - `actions` (array): Actions the agent should consider

### Skills Configuration

Skills define what your agent can do. There are three types of skills:

#### 1. LLM Skills

Native language model capabilities:

```yaml
type: llm
name: recipe_generator
description: Generate custom recipes based on ingredients
instruction: Create detailed recipes with ingredients list and step-by-step instructions
```

#### 2. MCP (Model Context Protocol) Skills

External tools accessed via MCP:

```yaml
type: mcp
name: nutrition_calculator
command: nutrition-mcp-server
args:
  - --port
  - '3001'
tools:
  - calculate_nutrition
  - get_food_info
env:
  NUTRITION_API_KEY: your-api-key
  DATABASE_URL: nutrition-db-url
```

#### 3. Native Tools

Built-in system tools:

```yaml
type: nativeTool
name: file_manager
description: Manage recipe files and cooking notes
env:
  STORAGE_PATH: /recipes
  MAX_FILE_SIZE: 10MB
```

**Skill Properties:**

- `type` (string, required): "llm", "mcp", or "nativeTool"
- `name` (string): Identifier for the skill
- `description` (string): Human-readable description
- `instruction` (string): Instructions for LLM skills
- `command` (string): Command to run MCP server
- `args` (array): Arguments for MCP server
- `tools` (array): List of MCP tool names
- `env` (object): Environment variables or configuration

### Knowledge Sources

Provide information sources for your agent:

```yaml
knowledge:
  - type: database
    source: recipe_database
    connection: postgresql://localhost/recipes
  - type: file
    source: cooking_techniques.md
    format: markdown
```

### Evaluation Configuration

Set up testing and validation for your agent:

```yaml
evaluator:
  prompt: Evaluate if the recipe is accurate and safe to follow
  numRetries: 3
```

**Properties:**

- `prompt` (string): Instructions for evaluating agent responses
- `numRetries` (int): Number of retry attempts for evaluation

### Metadata

Store additional configuration and tags:

```yaml
metadata:
  category: cooking
  language: en
  audience: home_cooks
  difficulty: beginner-friendly
```

## Complete Example

Here's a complete agent configuration:

```yaml
name: Recipe Master
description: An expert cooking assistant that helps with recipes, techniques, and meal planning
url: https://api.cookingapp.com/recipe-master
iconUrl: https://cookingapp.com/icons/chef-hat.png
version: '2.1.0'
documentationUrl: https://docs.cookingapp.com/recipe-master
defaultInputModes:
  - text
  - image
defaultOutputModes:
  - text
  - image
provider:
  organization: CookingApp Inc
  url: https://cookingapp.com
model: claude-3-sonnet-20240229
modelConfig:
  temperature: 0.7
  maxTokens: 4000
system: You are Recipe Master, an expert cooking assistant with deep knowledge of international cuisine, cooking techniques, and nutrition.
role: culinary_expert
prompt: Help users create delicious meals by providing recipes, cooking advice, and meal planning assistance.
messageExamples:
  - - user: I want to make dinner with chicken and vegetables
      text: I'd be happy to help you create a delicious chicken and vegetable dinner! Let me suggest a few options based on your ingredients and preferences.
      actions:
        - suggest_recipes
        - ask_preferences
skills:
  - type: llm
    name: recipe_creator
    description: Create custom recipes based on available ingredients
    instruction: Generate detailed recipes with ingredient lists, step-by-step instructions, and cooking tips
  - type: mcp
    name: nutrition_analyzer
    command: nutrition-mcp-server
    args:
      - --database
      - usda
    tools:
      - analyze_nutrition
      - calculate_calories
    env:
      NUTRITION_DB_URL: https://api.nal.usda.gov/fdc/v1/
      API_KEY: your-usda-api-key
  - type: nativeTool
    name: image_analyzer
    description: Analyze food images to identify ingredients
    env:
      MODEL_PATH: /models/food-recognition
      CONFIDENCE_THRESHOLD: '0.8'
knowledge:
  - type: database
    source: recipe_collection
    connection: mongodb://localhost:27017/recipes
evaluator:
  prompt: Evaluate the recipe for accuracy, safety, and clarity of instructions
  numRetries: 2
metadata:
  category: cooking
  language: en
  expertise_level: expert
  cuisine_types:
    - international
    - fusion
    - traditional
```

## Best Practices

1. **Clear Descriptions**: Write clear, concise descriptions that help users understand what your agent can do
2. **Appropriate Skills**: Choose the right skill types for your agent's capabilities
3. **Proper Versioning**: Use semantic versioning to track agent updates
4. **Comprehensive Testing**: Use the evaluator to ensure your agent performs correctly
5. **Meaningful Metadata**: Add metadata that helps categorize and discover your agent
6. **Security**: Be careful with environment variables and sensitive configuration
7. **Documentation**: Provide thorough documentation for your agent's capabilities and usage

## Next Steps

- Review the [Skills Guide](skills.md) for detailed information about skill configuration
- Check out [Examples](../examples/) for more agent configurations
- See [Deployment Guide](deployment.md) for hosting your agent

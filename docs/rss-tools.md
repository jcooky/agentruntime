# RSS Tools Guide

This guide explains how to configure RSS tools in AI agents using the agentruntime framework. RSS tools allow AI agents to automatically search and read RSS feeds to gather real-time information from news sources, blogs, and other content feeds when users ask questions.

## Overview

RSS tools provide two main capabilities that AI agents can use automatically:

- **search_rss**: AI agent searches for specific content across multiple RSS feeds when users ask about topics
- **read_rss**: AI agent reads all recent items from a single RSS feed when users want general updates

These tools are particularly useful for agents that need to:

- Provide current news and events information to users
- Gather information from multiple sources automatically
- Track specific topics or keywords when requested
- Deliver real-time content updates in conversations

## How RSS Tools Work

Once configured, RSS tools work automatically:

1. **User asks a question** about current events or specific topics
2. **AI agent analyzes** the question and determines if RSS information would be helpful
3. **AI agent automatically chooses** the appropriate RSS tool:
   - Uses `search_rss` when looking for specific topics across multiple feeds
   - Uses `read_rss` when getting general updates from a particular source
4. **AI agent searches** the configured RSS feeds automatically
5. **AI agent provides** the information to the user with proper citations

## RSS Tool Configuration

RSS tools are configured as native tools in your agent's skills section. Here's the basic configuration structure:

```yaml
skills:
  - type: nativeTool
    name: rss
    description: Search and read RSS feeds for current information
    env:
      allowed_feed_urls:
        - url: https://feeds.npr.org/1001/rss.xml
          name: NPR Top Stories
          description: Latest news and stories from NPR
        - url: https://rss.cnn.com/rss/edition.rss
          name: CNN International
          description: Breaking news and international stories
        - url: https://feeds.bbci.co.uk/news/rss.xml
          name: BBC News
          description: Latest news from BBC
```

### Configuration Properties

- `type`: Must be set to `nativeTool`
- `name`: Identifier for the RSS tool skill
- `description`: Human-readable description of the tool's purpose
- `env.allowed_feed_urls`: Array of RSS feeds the agent can access

### Feed URL Configuration

Each RSS feed in the `allowed_feed_urls` array requires:

| Property      | Type   | Required | Description                                   |
| ------------- | ------ | -------- | --------------------------------------------- |
| `url`         | string | ✅       | The URL of the RSS feed                       |
| `name`        | string | ✅       | Human-readable name for the feed              |
| `description` | string | ✅       | Description of the feed's content and purpose |

## Available RSS Tools

### 1. search_rss

The AI agent automatically uses this tool when users ask about specific topics or keywords.

**When AI agent uses this tool:**

- User asks about specific topics (e.g., "What's happening with AI?")
- User wants to find articles about particular subjects
- User requests information about companies, people, or events

**How it works:**

- AI agent automatically selects relevant RSS feeds from your configured list
- AI agent searches for articles matching the user's query
- AI agent presents the results with source citations

**Example conversation:**

```
User: "Find recent articles about artificial intelligence"
Agent: "I'll search for recent AI articles across our news sources."
*Agent automatically uses search_rss to find relevant articles*
Agent: "Here are the latest AI articles I found from TechCrunch, Reuters, and BBC..."
```

### 2. read_rss

The AI agent automatically uses this tool when users want general updates from specific sources.

**When AI agent uses this tool:**

- User asks for general updates from a specific source (e.g., "What's new on BBC?")
- User wants to see all recent headlines from a particular feed
- User requests a general news summary

**How it works:**

- AI agent automatically selects the appropriate RSS feed
- AI agent reads all recent items from that feed
- AI agent provides a summary or list of recent articles

**Example conversation:**

```
User: "What are the latest headlines from BBC News?"
Agent: "Let me check the latest BBC headlines for you."
*Agent automatically uses read_rss with BBC News feed*
Agent: "Here are the latest BBC headlines from today..."
```

## Complete Agent Example

Here's a complete example of an agent configured with RSS tools:

```yaml
name: News Monitor
description: An AI agent that monitors news sources and provides current information
url: https://api.example.com/news-monitor
version: '1.0.0'
defaultInputModes:
  - text
defaultOutputModes:
  - text
model: anthropic/claude-3.5-haiku
modelConfig:
  temperature: 0.7
  maxTokens: 4000
system: |
  You are a news monitoring agent that helps users stay informed about current events. 
  You can search across multiple news sources and provide summaries of recent developments.
  Always cite your sources and provide context for the information you share.
role: news_analyst
prompt: |
  Help users find and understand current news and information from reliable sources. 
  When users ask about current events, automatically search the appropriate RSS feeds.
  Always cite your sources and provide context for the information you share.
skills:
  - type: nativeTool
    name: rss
    description: Monitor RSS feeds for current news and information
    env:
      allowed_feed_urls:
        - url: https://feeds.npr.org/1001/rss.xml
          name: NPR Top Stories
          description: Latest news and stories from NPR
        - url: https://rss.cnn.com/rss/edition.rss
          name: CNN International
          description: Breaking news and international coverage
        - url: https://feeds.bbci.co.uk/news/rss.xml
          name: BBC News
          description: Latest news from BBC
        - url: https://www.reuters.com/rssFeed/topNews
          name: Reuters Top News
          description: Breaking news and top stories from Reuters
        - url: https://feeds.feedburner.com/techcrunch
          name: TechCrunch
          description: Technology news and startup information
        - url: https://feeds.arstechnica.com/arstechnica/index
          name: Ars Technica
          description: Technology and science news
messageExamples:
  - - user: What's happening in technology news today?
      text: Let me search the latest technology news for you from our monitored sources.
      actions:
        - search_rss
  - - user: Give me a summary of today's top news
      text: I'll check the latest headlines from our news sources and provide you with a summary.
      actions:
        - read_rss
        - search_rss
```

## Agent Behavior Configuration

### System Prompts for RSS Agents

Configure your agent's system prompt to effectively use RSS tools:

```yaml
system: |
  You are a news monitoring agent with access to current news feeds.

  When users ask about current events or specific topics:
  1. Use search_rss to find relevant articles across multiple sources
  2. Use read_rss to get general updates from specific sources
  3. Always cite your sources and include publication dates
  4. Provide context and summaries, not just raw headlines
  5. If information is not available in your feeds, clearly state this

  Available news sources: NPR, BBC, Reuters, CNN, TechCrunch, Ars Technica
```

### Message Examples for Training

Train your agent with examples of how to use RSS tools:

```yaml
messageExamples:
  - - user: What's the latest news about climate change?
      text: I'll search for recent climate change news across our sources.
      actions: [search_rss]
  - - user: Show me today's top stories
      text: Let me check the latest headlines from our news feeds.
      actions: [read_rss]
  - - user: What's new on TechCrunch?
      text: I'll get the latest articles from TechCrunch for you.
      actions: [read_rss]
```

## Best Practices

### 1. Choose Quality Sources

Select RSS feeds from reputable sources that are relevant to your agent's purpose:

```yaml
# Good: Specific, reputable sources
- url: https://feeds.npr.org/1001/rss.xml
  name: NPR Top Stories
  description: Latest news and stories from NPR covering politics, world events, and analysis

# Avoid: Generic or unreliable sources
- url: https://example.com/random-news
  name: Random News
  description: News stuff
```

### 2. Provide Clear Descriptions

Write clear descriptions that help the agent understand what each feed contains:

```yaml
# Good: Descriptive and specific
- url: https://feeds.feedburner.com/techcrunch
  name: TechCrunch
  description: Technology news, startup funding, Silicon Valley updates, and tech industry analysis

# Avoid: Vague descriptions
- url: https://feeds.feedburner.com/techcrunch
  name: TechCrunch
  description: Tech stuff
```

### 3. Organize by Category

Group related feeds together for better organization:

```yaml
env:
  allowed_feed_urls:
    # General News
    - url: https://feeds.npr.org/1001/rss.xml
      name: NPR Top Stories
      description: Latest news and stories from NPR
    - url: https://rss.cnn.com/rss/edition.rss
      name: CNN International
      description: Breaking news and international coverage

    # Technology News
    - url: https://feeds.feedburner.com/techcrunch
      name: TechCrunch
      description: Technology news and startup information
    - url: https://feeds.arstechnica.com/arstechnica/index
      name: Ars Technica
      description: Technology and science news
```

### 4. Consider Feed Volume and Agent Behavior

Balance between comprehensive coverage and manageable volume:

```yaml
# High-volume feeds: Agent will use search_rss with specific queries
- url: https://rss.cnn.com/rss/edition.rss
  name: CNN International
  description: Breaking news and international coverage

# Lower-volume feeds: Agent can use read_rss to get all items
- url: https://feeds.example.com/weekly-digest
  name: Weekly Tech Digest
  description: Weekly technology industry analysis
```

### 5. Test Your Agent Configuration

Ensure your RSS-enabled agent works correctly:

- Test with different types of user questions
- Verify the agent chooses appropriate RSS tools
- Check that citations and sources are properly provided
- Ensure the agent handles cases where no relevant information is found

## Common Use Cases

### 1. News Monitoring Agent

An agent that provides current news and events:

```yaml
name: NewsBot
description: Stay updated with current events and breaking news
system: |
  You are a news monitoring assistant. When users ask about current events,
  automatically search relevant news sources and provide summaries with citations.
skills:
  - type: nativeTool
    name: rss
    env:
      allowed_feed_urls:
        - url: https://feeds.npr.org/1001/rss.xml
          name: NPR Top Stories
          description: Latest news and stories from NPR
        - url: https://rss.cnn.com/rss/edition.rss
          name: CNN International
          description: Breaking news and international coverage
```

### 2. Industry Research Agent

An agent specialized in tracking specific industry developments:

```yaml
name: TechTracker
description: Track technology industry news and developments
system: |
  You are a technology industry analyst. When users ask about tech news,
  startups, or industry developments, search relevant tech sources automatically.
skills:
  - type: nativeTool
    name: rss
    env:
      allowed_feed_urls:
        - url: https://feeds.feedburner.com/techcrunch
          name: TechCrunch
          description: Technology news and startup information
        - url: https://feeds.venturebeat.com/VentureBeat
          name: VentureBeat
          description: Technology and business news
```

### 3. Academic Research Agent

An agent for monitoring research publications and academic news:

```yaml
name: ResearchMonitor
description: Monitor academic research and scientific publications
system: |
  You are a research assistant. When users ask about scientific developments
  or academic research, search relevant academic sources automatically.
skills:
  - type: nativeTool
    name: rss
    env:
      allowed_feed_urls:
        - url: https://feeds.nature.com/nature/rss/current
          name: Nature
          description: Latest scientific research and discoveries
        - url: https://feeds.sciencedaily.com/sciencedaily/top_news
          name: Science Daily
          description: Science news and research updates
```

## Troubleshooting

### Common Issues

1. **Agent Not Using RSS Tools**

   - Check that the skill type is set to `nativeTool`
   - Verify the agent's system prompt includes instructions to use RSS tools
   - Ensure messageExamples demonstrate RSS tool usage

2. **Feed Not Accessible**

   - Verify the RSS feed URL is correct and accessible
   - Check if the feed requires authentication
   - Ensure the server hosting the feed is online

3. **Agent Provides Outdated Information**

   - Check that your RSS feeds contain recent content
   - Verify the feeds are actively updated
   - Consider adding more current news sources

4. **No Results from Search**
   - Ensure feed descriptions are clear and specific
   - Verify the feeds contain content matching user queries
   - Test with broader search terms

### Testing RSS Configuration

Test your RSS-enabled agent with these approaches:

1. **Create a test agent:**

```yaml
name: RSS Test Agent
description: Test agent for RSS configuration
system: |
  You are a test agent for RSS functionality. When users ask about news,
  automatically search your RSS feeds and provide results with citations.
skills:
  - type: nativeTool
    name: rss
    env:
      allowed_feed_urls:
        - url: https://feeds.npr.org/1001/rss.xml
          name: NPR Top Stories
          description: Latest news and stories from NPR
```

2. **Test with different question types:**
   - "What's the latest news?" (should use read_rss)
   - "Find articles about [specific topic]" (should use search_rss)
   - "What's happening on [specific source]?" (should use read_rss with that source)

## Security Considerations

1. **Feed Source Verification**: Only include RSS feeds from trusted sources
2. **Content Filtering**: Consider that RSS content is user-generated and may contain inappropriate material
3. **Rate Limiting**: RSS tools have built-in timeouts (30 seconds) to prevent abuse
4. **Privacy**: Be aware that accessing RSS feeds may leave logs on the source servers

## Next Steps

- Review the [Agent Configuration Guide](agent.md) for general agent setup
- Check out [Examples](../examples/) for more agent configurations with RSS tools
- See [Skills Guide](skills.md) for information about other available skills

---

_Note: RSS feeds may change their URLs or become unavailable over time. Regularly verify that your configured feeds are still accessible._

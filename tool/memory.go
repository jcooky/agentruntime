package tool

import (
	"strings"

	"github.com/habiliai/agentruntime/entity"
	"github.com/habiliai/agentruntime/memory"
	"github.com/mokiat/gog"
)

func (m *manager) registerMemorySKill(skill *entity.NativeAgentSkill) error {
	// Remember tool
	if err := registerNativeTool(
		m,
		"remember_memory",
		`Store important information for future reference and personalized responses.

**IMMEDIATE SAVE TRIGGERS** - when user mentions:
- Personal info: *"I'm John"*, *"I live in Seoul"*, *"I work at..."*
- Preferences: *"I like..."*, *"I prefer..."*, *"I don't like..."*  
- Goals/Plans: *"I want to..."*, *"My goal is..."*, *"I'm planning to..."*
- Decisions: *"I decided to..."*, *"We agreed on..."*, *"I chose..."*

**Key format**: category_subcategory_detail  
Examples: user_name_full, user_preference_coffee, project_requirements_2024

**IMPORTANT**: Always inform user when storing their information`,
		skill,
		func(ctx *Context, req struct {
			Key    string   `json:"key" jsonschema:"required,description=Unique identifier using format: category_subcategory_detail (e.g. user_name_full, user_preference_coffee, project_requirements_2024)"`
			Memory string   `json:"memory" jsonschema:"required,description=The information to store - be specific and descriptive (e.g. 'Prefers dark roast coffee with oat milk, no sugar')"`
			Tags   []string `json:"tags,omitempty" jsonschema:"description=Optional categorization tags (e.g. ['personal', 'preferences'], ['work', 'decisions'], ['goals'])"`
		}) (resp struct {
			Memory *memory.Memory `json:"memory,omitempty" jsonschema:"description=Successfully stored memory object with key, value, source, tags, and timestamp"`
			Error  *string        `json:"error,omitempty" jsonschema:"description=Error message if storage failed (e.g. invalid key format, duplicate key, storage error)"`
		}, globalErr error) {
			input := memory.RememberInput{
				Key:    req.Key,
				Value:  req.Memory,
				Source: memory.MemorySourceAgent,
				Tags:   req.Tags,
			}

			memory, err := m.memoryService.RememberMemory(ctx, input)
			if err != nil {
				resp.Error = gog.PtrOf(err.Error())
				return
			}
			resp.Memory = memory

			return
		},
	); err != nil {
		return err
	}

	// Search memory tool
	if err := registerNativeTool(
		m,
		"search_memory",
		`Find relevant memories using **semantic search** with natural language queries.

**Perfect for** when you:
- Start conversation → *find context about current topic*
- Don't know exact memory key → *but know general topic*
- Looking for patterns → *user preferences, past decisions*
- Need discovery → *related information across memories*

Uses **AI embeddings** to find conceptually related memories, not just keyword matches

**Good queries**: "coffee preferences", "fitness goals", "work projects", "user background"`,
		skill,
		func(ctx *Context, req struct {
			Query string `json:"query" jsonschema:"required,description=Natural language search query to find related memories (e.g. 'coffee preferences', 'fitness goals', 'work projects', 'user background')"`
			Limit *int   `json:"limit,omitempty" jsonschema:"description=Maximum number of memories to return (optional parameter, 1-100 range, default: 20, recommended: 10-20 for most conversations)"`
		}) (resp struct {
			Memories []memory.ScoredMemory `json:"memories" jsonschema:"description=Array of relevant memories ranked by similarity score (0-1, higher = more relevant)"`
			Error    *string               `json:"error,omitempty" jsonschema:"description=Error message if search failed (e.g. no memories found, query too vague, search service error)"`
		}, err error) {
			limit := 20
			if req.Limit != nil {
				limit = *req.Limit
			}

			memories, err := m.memoryService.SearchMemory(ctx, req.Query, limit)
			if err != nil {
				resp.Error = gog.PtrOf(err.Error())
				return
			}

			resp.Memories = memories
			return
		},
	); err != nil {
		return err
	}

	// List memories tool
	if err := registerNativeTool(
		m,
		"list_memories",
		`Get **complete overview** of all stored memories in the system.

**Use this when**:
- **First conversation** with user → *check what you know about them*
- **Important discussions** → *need complete context*
- **Memory audit** → *what information has been stored*
- **Search failed** → *no relevant results found*

**Returns**: All memories with keys, values, sources, tags, and timestamps

**CAUTION**: Returns ALL memories - use carefully if many memories exist`,
		skill,
		func(ctx *Context, req struct{}) (resp struct {
			Memories []*memory.Memory `json:"memories" jsonschema:"description=Complete list of all stored memories with full details (keys, values, sources, tags, timestamps)"`
			Error    *string          `json:"error,omitempty" jsonschema:"description=Error message if listing failed (e.g. no memories exist, storage access error)"`
		}, err error) {
			memories, err := m.memoryService.ListMemories(ctx)
			if err != nil {
				resp.Error = gog.PtrOf(err.Error())
				return
			}

			if memories == nil {
				memories = make([]*memory.Memory, 0, 1)
			}

			resp.Memories = memories
			return
		},
	); err != nil {
		return err
	}

	if err := registerNativeTool(
		m,
		"recall_memory",
		`Get **specific memory** using exact key identifier.

**Use when you know the EXACT key**:
- **Exact key known** → *user_name_full, user_preference_coffee*
- **Fast retrieval** → *specific information you've stored before*
- **Verification** → *confirm stored information*
- **Direct access** → *bypass search when key is certain*

**REQUIREMENT**: Key must be exact match (case-sensitive)

**If unsure of exact key** → use search_memory instead`,
		skill,
		func(ctx *Context, req struct {
			Key string `json:"key" jsonschema:"required,description=Exact memory key to retrieve - must match stored key exactly (e.g. user_name_full, user_preference_coffee). Use search_memory if unsure of exact key."`
		}) (resp struct {
			Memory *memory.Memory `json:"memory,omitempty" jsonschema:"description=The specific memory object retrieved by key (includes value, source, tags, timestamp)"`
			Error  *string        `json:"error,omitempty" jsonschema:"description=Error message if recall failed (e.g. key not found, invalid key format, access error)"`
		}, err error) {
			resp.Memory = &memory.Memory{
				Key:       req.Key,
				Value:     "",
				Tags:      []string{},
				Embedding: []float32{},
			}

			memory, err := m.memoryService.GetMemory(ctx, req.Key)
			if err != nil {
				resp.Error = gog.PtrOf(err.Error())
			} else if memory == nil {
				resp.Error = gog.PtrOf("memory not found")
			}

			if memory != nil {
				resp.Memory = memory
			}
			return
		},
	); err != nil {
		return err
	}

	// Update memory tool
	if err := registerNativeTool(
		m,
		"update_memory",
		`Update **existing memory** with new information.

**Use when information changes**:
- **Personal changes** → *new job, moved cities, changed preferences*
- **Corrections** → *fix incorrect information previously stored*
- **Additional details** → *add more context to existing memory*
- **Status updates** → *goal progress, project status changes*

**REQUIREMENT**: Memory key must already exist

**Examples**:
- User got promoted → update user_job_title
- Changed coffee preference → update user_preference_coffee
- Project completed → update project_status_current`,
		skill,
		func(ctx *Context, req struct {
			Key   string   `json:"key" jsonschema:"required,description=Exact memory key to update - must match existing stored key exactly (e.g. user_name_full, user_preference_coffee). Use search_memory if unsure of exact key."`
			Value *string  `json:"value,omitempty" jsonschema:"description=New value to replace existing memory content. If not provided, only tags will be updated."`
			Tags  []string `json:"tags,omitempty" jsonschema:"description=New tags to replace existing tags. If empty, tags will be cleared."`
		}) (resp struct {
			Memory *memory.Memory `json:"memory,omitempty" jsonschema:"description=Updated memory object with new values (includes updated value, tags, source, timestamp)"`
			Error  *string        `json:"error,omitempty" jsonschema:"description=Error message if update failed (e.g. key not found, invalid key format, update error)"`
		}, err error) {
			input := memory.UpdateMemoryInput{
				Value: req.Value,
				Tags:  req.Tags,
			}

			updatedMemory, err := m.memoryService.UpdateMemory(ctx, req.Key, input)
			if err != nil {
				resp.Error = gog.PtrOf(err.Error())
				return
			}
			resp.Memory = updatedMemory
			return
		},
	); err != nil {
		return err
	}

	// Delete memory tool
	if err := registerNativeTool(
		m,
		"delete_memory",
		`Delete **specific memory** by exact key identifier.

**Use with caution when**:
- **Outdated information** → *user changed preferences, old job title*
- **Incorrect data** → *wrong information was stored*
- **User requests deletion** → *privacy concerns, data cleanup*
- **Duplicate cleanup** → *after consolidating similar memories*

**REQUIREMENT**: Key must be exact match (case-sensitive)

**CAUTION**: This action is irreversible - memory will be permanently deleted`,
		skill,
		func(ctx *Context, req struct {
			Key string `json:"key" jsonschema:"required,description=Exact memory key to delete - must match stored key exactly (e.g. user_name_full, user_preference_coffee). Use search_memory if unsure of exact key."`
		}) (resp struct {
			Success bool    `json:"success" jsonschema:"description=True if memory was successfully deleted, false otherwise"`
			Error   *string `json:"error,omitempty" jsonschema:"description=Error message if deletion failed (e.g. key not found, invalid key format, access error)"`
		}, err error) {
			err = m.memoryService.DeleteMemory(ctx, req.Key)
			if err != nil {
				resp.Error = gog.PtrOf(err.Error())
				resp.Success = false
				return
			}
			resp.Success = true
			return
		},
	); err != nil {
		return err
	}

	m.usagePrompts[skill.Name] = strings.TrimSpace(`<tool:memory_instructions>
# AI Agent Memory System - Complete Usage Guide

## Essential Workflow

**Every conversation should start like this:**
- Use 'search_memory' with broad terms about the current topic
- OR use 'list_memories' if you haven't talked to this user before  
- Review retrieved memories BEFORE your first substantial response

**During conversation - IMMEDIATE SAVE triggers:**
Use 'remember_memory' RIGHT AWAY when user mentions:
- Personal info: "I'm John", "I live in Seoul", "I work at..."
- Preferences: "I like...", "I prefer...", "I don't like..."
- Goals/Plans: "I want to...", "My goal is...", "I'm planning to..."
- Decisions: "I decided to...", "We agreed on...", "I chose..."
- Context: "This project is about...", "We discussed..."
- Experiences: "Last time I...", "I tried...", "I learned..."

**IMMEDIATE UPDATE triggers:**
Use 'update_memory' RIGHT AWAY when user mentions changes:
- Status changes: "I got promoted", "I moved to Tokyo", "I finished the project"  
- Corrections: "Actually, I prefer tea", "I work at Apple now", "My name is Mike, not John"
- Progress updates: "I completed my goal", "The project status changed"

**Before responding:**
- If discussing specific topic → use 'search_memory' with relevant keywords
- If user references something specific → try 'recall_memory' with likely key

## Key Naming Rules

**Standard Format**: category_subcategory_detail

**Good examples:**
- user_name_full: "John Smith"
- user_location_city: "Seoul, South Korea"  
- user_job_title: "Software Engineer at Google"
- user_preference_coffee: "Dark roast with oat milk, no sugar"
- user_goal_fitness: "Run marathon in 6 months"
- project_name_current: "E-commerce platform redesign"
- decision_architecture_2024: "Chose microservices over monolith"

**Bad examples:**
- john (too vague)
- coffee_preference (missing user_ prefix)
- some_random_key (not descriptive)

**Category prefixes:**
- **user_**: Personal information about the user
- **project_**: Work/project related information  
- **decision_**: Important choices or agreements
- **conversation_**: Context from discussions

## Which Tool to Use?

**Do you know the EXACT key?**
- YES → use 'recall_memory' with exact key
- NO → use 'search_memory' with descriptive terms

**Information Status?**
- NEW info → use 'remember_memory' to create
- CHANGED info → use 'update_memory' to modify existing
- WRONG info → use 'update_memory' to correct or 'delete_memory' if invalid

**Examples:**
- "Get user's name" → recall_memory(key: "user_name_full")
- "Find coffee preferences" → search_memory(query: "coffee preferences")
- "What was that project?" → search_memory(query: "project discussion")
- "User said they moved" → update_memory(key: "user_location_city", value: "New City")
- "User corrected their job title" → update_memory(key: "user_job_title", value: "Senior Engineer")

## Avoid Duplicates & Handle Updates

**Before storing new memory:**
1. Search for existing related memories first
2. If similar memory exists:
   - **SAME key, NEW info** → use 'update_memory' to replace
   - **SAME info, DIFFERENT context** → create new key  
   - **IDENTICAL info** → skip saving
   - **COMPLEMENTARY info** → combine into existing memory

**Update vs New Memory:**
- User says: "I moved to Seoul" → update_memory(key: "user_location_city") ✅
- User says: "I also like tea" → remember_memory(key: "user_preference_tea") ✅ 
- User says: "Actually, I hate coffee" → update_memory(key: "user_preference_coffee") ✅

## How to Talk to Users

<communication_examples>
**When storing memories:**
- ✅ "I'll remember that you prefer dark roast coffee!"
- ✅ "Got it - I've noted that your goal is to run a marathon."
- ✅ "I'll save this project information for future reference."
- ❌ "Calling remember_memory tool..." (too technical)

**When updating memories:**
- ✅ "I've updated your job title to Senior Engineer!"
- ✅ "Got it - I've changed your location to Seoul."
- ✅ "I've noted that you completed your marathon goal!"
- ❌ "Calling update_memory tool..." (too technical)

**When recalling memories:**
- ✅ "I remember you mentioned you work at Google..."
- ✅ "Based on what you told me before about your coffee preferences..."
- ✅ "Didn't you say your goal was to run a marathon?"
</communication_examples>

## Quick Examples

**Automatic triggers in action:**

**New Information (remember_memory):**
- User: "I'm a software engineer" → remember_memory(key: "user_job_title", memory: "Software engineer")
- User: "I don't like spicy food" → remember_memory(key: "user_preference_food_spicy", memory: "Dislikes spicy food")
- User: "We decided to use React for this project" → remember_memory(key: "decision_framework_react", memory: "Chose React framework for current project")
- User: "My name is Sarah" → remember_memory(key: "user_name_full", memory: "Sarah")

**Changed Information (update_memory):**
- User: "I got promoted to Senior Engineer" → update_memory(key: "user_job_title", value: "Senior Software Engineer")
- User: "Actually, I love spicy food now" → update_memory(key: "user_preference_food_spicy", value: "Loves spicy food")
- User: "We switched to Vue.js" → update_memory(key: "decision_framework_react", value: "Switched from React to Vue.js framework for current project")
- User: "My name is actually Sara, not Sarah" → update_memory(key: "user_name_full", value: "Sara")

## Conversation Patterns

**New User (No memories):**
1. Start conversation normally
2. As user shares info → save immediately  
3. Build memory profile gradually

**Returning User:**
1. search_memory or list_memories FIRST
2. Acknowledge relevant memories in greeting
3. Continue building on existing context

**Topic Switch:**
When user changes subject → search_memory with new topic keywords → Surface relevant memories if found → Continue saving new information as usual

## Critical Success Factors

1. **Speed**: Save/update memories IMMEDIATELY when triggered (don't wait)
2. **Accuracy**: Use update_memory for changes, remember_memory for new info
3. **Consistency**: Use standard key naming format always
4. **Completeness**: Check for existing memories before deciding create vs update
5. **Transparency**: Tell user when you remember OR update something
6. **Context**: Start every conversation with memory check

**Remember**: Your goal is to build a comprehensive, organized memory system that makes every conversation feel personal and contextual!
</tool:memory_instructions>`)

	return nil
}

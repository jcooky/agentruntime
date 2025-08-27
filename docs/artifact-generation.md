# Artifact Generation System

## Overview

The Artifact Generation system enables AI agents to create interactive charts, graphs, tables, and **unlimited custom React components** within the Team0 application using XML tags. This provides a seamless way for users to visualize data and interact with dynamic components directly within the chat interface.

## Architecture

### XML-Based Universal Approach

AI agents include `<habili:artifact>` XML tags directly in their text responses. The frontend automatically detects, parses, and renders these tags as interactive React components. **This system works with all LLM models** (OpenAI, Anthropic, XAI, etc.), ensuring vendor independence.

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  AI Agent       │    │  XML Parser      │    │  React Component│
│  (Any LLM)      │───▶│  (Frontend)      │───▶│  (Native)       │
│  Text + XML     │    │  ArtifactXMLPars │    │  Direct Render  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Backend Integration (AgentRuntime)

#### Template-Based Instruction (`engine/data/instructions/chat.md.tmpl`)

All AI agents receive comprehensive artifact generation instructions through the chat template:

```xml
<artifact_instruction optional="false">
# ARTIFACT GENERATION:

## When to Use Artifacts
Create interactive artifacts when:
- **Data visualization needed**: Charts, graphs, visual representations
- **Interactive components**: Forms, dashboards, calculators, widgets
- **Complex UI**: Multi-step interfaces, tabs, accordions
- **Creative components**: Games, animations, unique experiences

## Two Generation Methods:

### Option 1: Pre-built Chart/Table (Simple)
### Option 2: Custom React Code (Unlimited)
</artifact_instruction>
```

#### Live Testing (`agentruntime_live_test.go`)

The system includes comprehensive live tests that validate:

- **Chart Generation**: JSON data visualization
- **Interactive Components**: React hooks and state management
- **Data Tables**: Complex tabular data presentation
- **Instruction Coverage**: AI understanding of the artifact system

**Verified Results**: 100% success rate across all test scenarios (39.461s execution time).

## Two Artifact Creation Methods

### Method 1: Pre-built Chart/Table (Quick & Simple)

For standard data visualization needs:

```xml
<habili:artifact
  type="chart"
  title="Monthly Revenue"
  description="Q1 2024 performance overview"
  data='{"labels": ["Jan", "Feb", "Mar"], "values": [45000, 52000, 61000], "label": "Revenue ($)"}'
  chartType="bar"
  colors='["#3B82F6", "#10B981"]'
  barThickness="35"
/>
```

### Method 2: Custom React Code (Unlimited Flexibility)

For any interactive component or complex UI:

```xml
<habili:artifact type="react" title="Interactive Dashboard">
<reactCode>
import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';

const InteractiveDashboard = () => {
  const [count, setCount] = useState(0);
  const [name, setName] = useState('');

  return (
    <div className="w-full max-w-2xl space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            Interactive Counter
            <Badge variant="secondary">{count}</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <Button
              onClick={() => setCount(c => c + 1)}
              className="bg-blue-600 hover:bg-blue-700"
            >
              Increment
            </Button>
            <Button
              onClick={() => setCount(c => c - 1)}
              variant="outline"
            >
              Decrement
            </Button>
            <Button
              onClick={() => setCount(0)}
              variant="destructive"
            >
              Reset
            </Button>
          </div>
          <div className="pt-4 border-t">
            <Input
              placeholder="Enter your name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="mb-2"
            />
            {name && (
              <p className="text-sm text-muted-foreground">
                Hello, {name}! You've clicked {count} times.
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default InteractiveDashboard;
</reactCode>
</habili:artifact>
```

## Frontend Components (Team0)

### 1. XML Parser (`components/ArtifactXMLParser.tsx`)

**Enhanced Features:**

- **Dual Parser**: Handles both chart/table artifacts and React code artifacts
- **ReactCode Support**: Extracts `<reactCode>` inner tags from `type="react"` artifacts
- **Safe Parsing**: Robust XML attribute parsing with error handling

```typescript
export interface ArtifactXMLData {
  type: string;
  title: string;
  description?: string;
  data: any;
  config: any;
  reactCode?: string; // Direct React code for unlimited components
  originalTag: string;
}
```

### 2. Message Integration (`app/team0/teams/AgentMessageBubble.tsx`)

**Automatic Detection & Rendering:**

- Scans AI messages for `<habili:artifact>` tags
- Renders artifacts inline with message text
- Supports multiple artifacts per message
- Dynamic library imports (Chart.js, shadcn/ui components)

### 3. ReactArtifactViewer (`components/ReactArtifactViewer.tsx`)

**Advanced React Execution:**

- **Safe Code Execution**: Uses `new Function()` for secure React code execution
- **Dynamic Imports**: Automatic imports for React, shadcn/ui, Chart.js libraries
- **Component Lifecycle**: Full React state and effect support
- **Error Handling**: Graceful error recovery with user-friendly messages

## XML Format Specifications

### Chart/Table Artifacts (Method 1)

```xml
<habili:artifact
  type="chart|table"
  title="Display Title"
  description="Optional subtitle"
  data='{"labels": [...], "values": [...]}'
  chartType="bar|line|pie|doughnut"
  colors='["#FF6B6B", "#4ECDC4"]'
  barThickness="30"
  fontSize="14"
/>
```

### React Code Artifacts (Method 2)

```xml
<habili:artifact type="react" title="Component Name" description="Optional">
<reactCode>
import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
// ... component code ...
const MyComponent = () => { /* ... */ };
export default MyComponent;
</reactCode>
</habili:artifact>
```

## Approved Libraries Only

**SECURITY POLICY**: Only the following 5 libraries are allowed in artifact code for security and performance reasons:

### 1. React (Core Library)

```javascript
// React core and hooks only
import React, { useState, useEffect, useMemo, useCallback } from 'react';
```

### 2. shadcn/ui Components (UI Library)

```javascript
// Cards and Layout
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

// Form Controls
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Switch } from '@/components/ui/switch';
import { Slider } from '@/components/ui/slider';

// Data Display
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';

// Feedback & Navigation
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
```

### 3. Tailwind CSS (Styling Only)

```javascript
// Use via className attribute only:
className = 'flex grid space-y-4 gap-2 w-full max-w-lg p-4 m-2 px-6 py-3';
className = 'bg-purple-600 text-purple-500 border-purple-200'; // Default purple theme
className =
  'text-lg font-bold text-center hover:bg-blue-700 disabled:opacity-50';
```

### 4. react-chartjs-2 (Chart Components)

```javascript
// React Chart.js wrapper (recommended)
import { Bar, Line, Pie, Doughnut } from 'react-chartjs-2';
```

### 5. chart.js (Advanced Chart Customization)

```javascript
// Chart.js direct imports for advanced features
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  LineElement,
  PointElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';

// Register components when needed
ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
);
```

## Color Customization Guidelines

### 🎨 **User Color Specification**

Users can specify colors in natural language, which AI should interpret flexibly:

```javascript
// Default: Purple theme (if user doesn't specify colors)
const defaultColors = ['#9333ea', '#a855f7', '#c084fc', '#d8b4fe', '#e9d5ff'];

// User color interpretation examples:
// "use blue colors" → ['#3b82f6', '#60a5fa', '#93c5fd']
// "make it green and red" → ['#22c55e', '#ef4444']
// "corporate/professional theme" → ['#1e40af', '#dc2626', '#059669']
// "bright/vibrant colors" → ['#f59e0b', '#10b981', '#8b5cf6']
// "warm colors" → ['#f59e0b', '#f97316', '#ef4444']
// "cool colors" → ['#06b6d4', '#3b82f6', '#8b5cf6']

// Tailwind color palette
const tailwindColors = {
  blue: ['#3b82f6', '#60a5fa', '#93c5fd'],
  green: ['#22c55e', '#4ade80', '#86efac'],
  red: ['#ef4444', '#f87171', '#fca5a5'],
  purple: ['#9333ea', '#a855f7', '#c084fc'], // DEFAULT
  orange: ['#f97316', '#fb923c', '#fdba74'],
  cyan: ['#06b6d4', '#22d3ee', '#67e8f9'],
};
```

### ❌ Forbidden Libraries

**The following are NOT allowed and will be blocked:**

- HTTP libraries (axios, fetch)
- Date libraries (moment.js, date-fns)
- Utility libraries (lodash, ramda)
- Third-party UI libraries (Material-UI, Ant Design)
- Animation libraries (framer-motion, lottie)
- Any other npm packages not listed above

## Live Test Results

**Comprehensive Validation** (from `agentruntime_live_test.go`):

✅ **Chart Generation Request** (5.97s)

- Perfect `<habili:artifact type="chart">` generation
- Accurate JSON data: `{"labels": ["January", "February", "March"], "values": [25000, 32000, 28000]}`
- Proper styling and configuration

✅ **Interactive Component Request** (9.69s)

- Complete React component with `<reactCode>` tags
- React hooks (`useState`) and event handlers
- shadcn/ui component integration

✅ **Data Table Request** (10.35s)

- **Dual artifact generation**: Bar chart + Interactive table
- Complex data manipulation with color-coded performance badges
- Professional UI with responsive design

✅ **Instruction Coverage Verification** (13.08s)

- AI demonstrates complete understanding of artifact system
- Provides examples and use case recommendations
- Interactive demonstrations within response

**Total Success Rate**: 100% across all scenarios

## Best Practices

### 1. Always Export Default Component

```javascript
const MyComponent = () => {
  /* ... */
};
export default MyComponent;
```

### 2. Use React Hooks for State

```javascript
import { useState, useEffect } from 'react';
const [state, setState] = useState(initialValue);
```

### 3. Apply Responsive Design

```javascript
<div className="w-full max-w-2xl mx-auto">
  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">{/* Content */}</div>
</div>
```

### 4. Handle User Interactions

```javascript
<Button onClick={() => handleClick()} className="w-full">
  Click Me
</Button>
```

## Real-World Usage Examples

### 1. Business Dashboard

```xml
<habili:artifact type="react" title="Sales Dashboard">
<reactCode>
import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

const SalesDashboard = () => {
  const [selectedPeriod, setSelectedPeriod] = useState('Q1');

  const salesData = {
    Q1: { revenue: 125000, deals: 45, growth: 12.5 },
    Q2: { revenue: 142000, deals: 52, growth: 13.6 },
    Q3: { revenue: 138000, deals: 48, growth: -2.8 },
    Q4: { revenue: 165000, deals: 61, growth: 19.6 }
  };

  const currentData = salesData[selectedPeriod];

  return (
    <div className="w-full max-w-4xl space-y-6">
      <div className="flex gap-2 mb-4">
        {Object.keys(salesData).map(period => (
          <Button
            key={period}
            variant={selectedPeriod === period ? "default" : "outline"}
            onClick={() => setSelectedPeriod(period)}
            className="px-4 py-2"
          >
            {period}
          </Button>
        ))}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Revenue</CardTitle>
            <Badge variant="secondary">💰</Badge>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${currentData.revenue.toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">
              {selectedPeriod} Performance
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Deals Closed</CardTitle>
            <Badge variant="secondary">🤝</Badge>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{currentData.deals}</div>
            <p className="text-xs text-muted-foreground">
              New customers acquired
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Growth Rate</CardTitle>
            <Badge
              variant={currentData.growth > 0 ? "default" : "destructive"}
            >
              {currentData.growth > 0 ? "📈" : "📉"}
            </Badge>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {currentData.growth > 0 ? '+' : ''}{currentData.growth}%
            </div>
            <p className="text-xs text-muted-foreground">
              Quarter over quarter
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default SalesDashboard;
</reactCode>
</habili:artifact>
```

### 2. Interactive Data Entry

```xml
<habili:artifact type="react" title="Customer Feedback Form">
<reactCode>
import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';

const FeedbackForm = () => {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    rating: 0,
    feedback: ''
  });

  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e) => {
    e.preventDefault();
    setSubmitted(true);
    // In real app: send to API
  };

  const handleRating = (rating) => {
    setFormData(prev => ({ ...prev, rating }));
  };

  if (submitted) {
    return (
      <Card className="w-full max-w-md mx-auto">
        <CardContent className="pt-6 text-center">
          <div className="text-6xl mb-4">🎉</div>
          <h3 className="text-lg font-semibold mb-2">Thank you!</h3>
          <p className="text-sm text-muted-foreground">
            Your feedback has been submitted successfully.
          </p>
          <Button
            onClick={() => setSubmitted(false)}
            className="mt-4"
            variant="outline"
          >
            Submit Another
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md mx-auto">
      <CardHeader>
        <CardTitle>Customer Feedback</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              value={formData.email}
              onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
              required
            />
          </div>

          <div className="space-y-2">
            <Label>Rating</Label>
            <div className="flex gap-1">
              {[1, 2, 3, 4, 5].map(star => (
                <button
                  key={star}
                  type="button"
                  onClick={() => handleRating(star)}
                  className={`text-2xl transition-colors ${
                    star <= formData.rating ? 'text-yellow-500' : 'text-gray-300'
                  }`}
                >
                  ⭐
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="feedback">Feedback</Label>
            <Textarea
              id="feedback"
              value={formData.feedback}
              onChange={(e) => setFormData(prev => ({ ...prev, feedback: e.target.value }))}
              placeholder="Tell us about your experience..."
              required
            />
          </div>

          <Button type="submit" className="w-full">
            Submit Feedback
          </Button>
        </form>
      </CardContent>
    </Card>
  );
};

export default FeedbackForm;
</reactCode>
</habili:artifact>
```

## Security & Performance

### Security Features

- **Safe Code Execution**: Uses `new Function()` instead of `eval()`
- **Controlled Environment**: Limited global scope access
- **Approved Libraries**: Only pre-approved imports allowed
- **React Sandbox**: Components execute within React's virtual DOM

### Performance Optimizations

- **Dynamic Imports**: Libraries loaded only when needed
- **Component Caching**: Parsed components cached for reuse
- **Efficient Parsing**: Optimized XML parsing with minimal overhead
- **Native Rendering**: Direct React rendering without iframe isolation

## Multi-LLM Support

### Vendor Independence

The XML-based approach works with **all LLM providers**:

```
✅ OpenAI (gpt-4o, gpt-4-turbo)
✅ Anthropic (claude-3.5-haiku, claude-4-sonnet)
✅ XAI (grok models)
✅ Future models (vendor-agnostic design)
```

### Template Integration

All models receive identical instructions through `chat.md.tmpl`:

- Comprehensive artifact generation guidelines
- Examples for both chart and React methods
- Best practices and design system integration
- Natural language to XML attribute mapping

## Benefits

- **🚀 Unlimited Creativity**: Direct React code enables any UI component
- **⚡ High Performance**: Native React rendering without sandboxing overhead
- **🔄 Vendor Independence**: Works with all LLM providers
- **🎨 Design System Integration**: Full shadcn/ui and Tailwind CSS support
- **✅ Production Ready**: Thoroughly tested with live validation
- **🔧 Easy Maintenance**: Clean architecture with clear separation of concerns
- **📱 Responsive**: Mobile-first design with Tailwind CSS
- **🛡️ Secure**: Safe code execution with controlled environment

This system provides unparalleled flexibility for AI-driven UI generation while maintaining security, performance, and design consistency across the Team0 application.

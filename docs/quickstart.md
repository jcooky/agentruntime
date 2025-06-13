# 🚀 Quick Start Guide for Beginners

**New to programming?** No problem! Follow these simple steps to test AI agents in just 5 minutes.

## What You Need

- A computer (Windows, Mac, or Linux)
- An API key from OpenAI, Anthropic, or xAI (we'll help you get one)
- 5-10 minutes of your time

## Step 1: Download AgentRuntime

1. Go to the [AgentRuntime Releases](https://github.com/habiliai/agentruntime/releases) page
2. Download the file for your operating system:
   - **Windows**: `agentruntime-windows.exe`
   - **Mac(ARM64)**: `agentruntime-macos-arm64`
   - **Mac(Intel)**: `agentruntime-macos-amd64`
   - **Linux**: `agentruntime`

## Step 2: Get Your API Key

**For OpenAI (Recommended for beginners):**

1. Visit [OpenAI API Keys](https://platform.openai.com/api-keys)
2. Sign up or log in
3. Click "Create new secret key"
4. Copy the key (starts with `sk-...`)
5. Keep it safe! You'll need it in Step 3.

**Cost**: Usually $5-10 for extensive testing

## Step 3: Create Your First Agent

1. Create a new folder on your desktop called `my-agents`
2. Inside this folder, create a file called `assistant.yaml`
3. Copy and paste this content:

```yaml
name: 'my-assistant'
description: 'A helpful AI assistant for general questions'
model: 'gpt-4o-mini'
instructions: |
  You are a friendly and helpful assistant. 
  Answer questions clearly and be polite.
  If you don't know something, just say so.
temperature: 0.7
```

## Step 4: Start AgentRuntime

1. Open your terminal/command prompt:

   - **Windows**: Press `Win + R`, type `cmd`, press Enter
   - **Mac**: Press `Cmd + Space`, type `terminal`, press Enter
   - **Linux**: Press `Ctrl + Alt + T`

2. Navigate to where you downloaded the agentruntime file:

   ```bash
   # Windows example
   cd Downloads

   # Mac/Linux example
   cd ~/Downloads
   ```

3. Run AgentRuntime with your API key:

   ```bash
   # Windows
   agentruntime-windows.exe C:\Users\YourName\Desktop\my-agents -p 3001

   # Mac/Linux
   ./agentruntime-macos-arm64 ~/Desktop/my-agents -p 3001
   ```

4. When prompted, enter your API key (the one from Step 2)

5. You should see:
   ```
   ✅ Loaded agent: my-assistant
   🚀 Server running on http://localhost:3001
   ```

## Step 5: Start the Playground

1. Open a new terminal window (same way as Step 4)
2. Navigate to the agentruntime folder you downloaded
3. Go to the playground folder:
   ```bash
   cd playground
   ```
4. Install and start the playground:

   ```bash
   # Install dependencies (only needed once)
   npm install

   # Start the playground
   npm run dev
   ```

## Step 6: Test Your Agent! 🎉

1. Open your web browser
2. Go to [http://localhost:3000](http://localhost:3000)
3. Click "Create Thread"
4. Select your agent (`my-assistant`)
5. Start chatting!

## ❓ Common Questions

**Q: What if I get a "command not found" error?**
A: Make sure you're in the right folder and the file name matches exactly.

**Q: What if my agent doesn't respond?**
A: Check that your API key is correct and you have credits in your OpenAI account.

**Q: Can I create multiple agents?**
A: Yes! Just create more `.yaml` files in your `my-agents` folder.

**Q: Is this safe?**
A: Yes, everything runs on your computer. Your conversations stay private.

## Next Steps

Once you've successfully tested your first agent:

1. **Try different agent personalities** - Modify the `instructions` in your YAML file
2. **Create specialized agents** - Make agents for specific tasks like writing, coding, or customer support
3. **Test multiple agents** - Create conversations with multiple agents in one thread
4. **Read the advanced guide** - Check out [playground.md](playground.md) for more features

## Getting Help

- **Having trouble?** Check the [troubleshooting guide](playground.md#troubleshooting) in the playground documentation
- **Want to learn more?** Read the full [playground guide](playground.md) for advanced features
- **Found a bug?** Report it on [GitHub Issues](https://github.com/habiliai/agentruntime/issues)

Happy testing! 🚀

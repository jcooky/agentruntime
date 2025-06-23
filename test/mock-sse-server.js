// Mock SSE MCP Server for testing Remote MCP support
// Run with: node test/mock-sse-server.js

const http = require('http');

const PORT = 8888;

// Simple MCP protocol implementation
const PROTOCOL_VERSION = "2024-11-05";

const server = http.createServer((req, res) => {
  console.log(`${req.method} ${req.url}`);
  
  // Enable CORS
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');
  
  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }
  
  // SSE endpoint
  if (req.url === '/sse' && req.method === 'GET') {
    res.writeHead(200, {
      'Content-Type': 'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection': 'keep-alive',
    });
    
    // Send initial connection message
    res.write(`data: ${JSON.stringify({
      jsonrpc: "2.0",
      method: "connection.ready"
    })}\n\n`);
    
    // Handle SSE messages
    let body = '';
    req.on('data', chunk => {
      body += chunk.toString();
    });
    
    // Keep connection alive
    const keepAlive = setInterval(() => {
      res.write(': keepalive\n\n');
    }, 30000);
    
    req.on('close', () => {
      clearInterval(keepAlive);
    });
    
    return;
  }
  
  // JSON-RPC endpoint
  if (req.url === '/' && req.method === 'POST') {
    let body = '';
    req.on('data', chunk => {
      body += chunk.toString();
    });
    
    req.on('end', () => {
      try {
        const request = JSON.parse(body);
        console.log('Received:', request);
        
        let response;
        
        switch (request.method) {
          case 'initialize':
            response = {
              jsonrpc: "2.0",
              id: request.id,
              result: {
                protocolVersion: PROTOCOL_VERSION,
                capabilities: {
                  tools: {}
                },
                serverInfo: {
                  name: "mock-sse-server",
                  version: "1.0.0"
                }
              }
            };
            break;
            
          case 'tools/list':
            response = {
              jsonrpc: "2.0",
              id: request.id,
              result: {
                tools: [
                  {
                    name: "echo",
                    description: "Echoes back the input message",
                    inputSchema: {
                      type: "object",
                      properties: {
                        message: {
                          type: "string",
                          description: "The message to echo"
                        }
                      },
                      required: ["message"]
                    }
                  },
                  {
                    name: "get_time",
                    description: "Returns the current server time",
                    inputSchema: {
                      type: "object",
                      properties: {}
                    }
                  }
                ]
              }
            };
            break;
            
          case 'tools/call':
            const toolName = request.params.name;
            const args = request.params.arguments;
            
            if (toolName === 'echo') {
              response = {
                jsonrpc: "2.0",
                id: request.id,
                result: {
                  content: [
                    {
                      type: "text",
                      text: `Echo: ${args.message}`
                    }
                  ]
                }
              };
            } else if (toolName === 'get_time') {
              response = {
                jsonrpc: "2.0",
                id: request.id,
                result: {
                  content: [
                    {
                      type: "text",
                      text: `Current time: ${new Date().toISOString()}`
                    }
                  ]
                }
              };
            } else {
              response = {
                jsonrpc: "2.0",
                id: request.id,
                error: {
                  code: -32601,
                  message: "Method not found"
                }
              };
            }
            break;
            
          default:
            response = {
              jsonrpc: "2.0",
              id: request.id,
              error: {
                code: -32601,
                message: "Method not found"
              }
            };
        }
        
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(response));
        console.log('Sent:', response);
        
      } catch (error) {
        console.error('Error:', error);
        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
          jsonrpc: "2.0",
          error: {
            code: -32700,
            message: "Parse error"
          }
        }));
      }
    });
    
    return;
  }
  
  // 404 for other routes
  res.writeHead(404);
  res.end('Not found');
});

server.listen(PORT, () => {
  console.log(`Mock SSE MCP Server running at http://localhost:${PORT}`);
  console.log('SSE endpoint: http://localhost:' + PORT + '/sse');
  console.log('Available tools: echo, get_time');
}); 
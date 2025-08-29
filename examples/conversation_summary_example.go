package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/habiliai/agentruntime"
	"github.com/habiliai/agentruntime/config"
	"github.com/habiliai/agentruntime/engine"
	"github.com/habiliai/agentruntime/entity"
)

func main() {
	ctx := context.Background()

	// Create an agent with a simple configuration
	agent := entity.Agent{
		Name:        "ConversationBot",
		Role:        "Assistant",
		Description: "A bot that demonstrates conversation summarization",
		Prompt:      "You are a helpful assistant that can handle long conversations efficiently.",
		ModelName:   "openai/gpt-5-mini",
		System:      "Be concise and helpful.",
	}

	// Configuration for conversation summarization
	summaryConfig := config.ConversationSummaryConfig{
		MaxTokens:                   5000, // 5k tokens limit
		SummaryTokens:               1000, // 1k tokens per summary
		MinConversationsToSummarize: 5,    // At least 5 conversations before summarizing
		ModelForSummary:             "openai/gpt-5-mini",
	}

	// Create AgentRuntime with conversation summarization enabled
	runtime, err := agentruntime.NewAgentRuntime(
		ctx,
		agentruntime.WithOpenAIAPIKey(os.Getenv("OPENAI_API_KEY")),
		agentruntime.WithAgent(agent),
		agentruntime.WithConversationSummary(summaryConfig),
		// Or use default settings:
		// agentruntime.WithDefaultConversationSummary(),
		// Or just set token limit:
		// agentruntime.WithConversationSummaryTokenLimit(5000),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer runtime.Close()

	// Create a long conversation history to demonstrate summarization
	longHistory := createLongConversationHistory()

	// Run the agent with the long history and an attached file
	req := engine.RunRequest{
		History: longHistory,
		ThreadInstruction: "Please help the user with their questions. " +
			"You may reference previous parts of our conversation.",
		Files: []engine.File{
			{
				ContentType: "image/png",
				Data:        createDummyImageBase64(), // This would be a real base64 image in practice
				Filename:    "ml_workflow_diagram.png",
			},
		},
	}

	fmt.Printf("Running agent with %d conversation messages (including media files)...\n", len(longHistory))
	fmt.Printf("This includes images and would normally exceed token limits, but summarization will help.\n\n")

	response, err := runtime.Run(ctx, req, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Agent Response: %s\n", response.Text())

	// The conversation history has been processed with summarization
	// if it exceeded the configured token limit
}

func createLongConversationHistory() []engine.Conversation {
	// Create a series of conversations that would exceed typical token limits
	conversations := make([]engine.Conversation, 0)

	// Simulate a long technical discussion
	topics := []struct {
		user     string
		question string
		response string
	}{
		{
			user:     "user",
			question: "What is machine learning and how does it work?",
			response: "Machine learning is a subset of artificial intelligence that enables systems to automatically learn and improve from experience without being explicitly programmed. It works by using algorithms to analyze data, identify patterns, and make predictions or decisions. The main types include supervised learning (learning with labeled examples), unsupervised learning (finding patterns in unlabeled data), and reinforcement learning (learning through interaction and rewards).",
		},
		{
			user:     "user",
			question: "Can you explain different types of neural networks?",
			response: "There are several types of neural networks: 1) Feedforward Neural Networks - the simplest type where information flows in one direction. 2) Convolutional Neural Networks (CNNs) - excellent for image processing and computer vision tasks. 3) Recurrent Neural Networks (RNNs) - good for sequential data like text and time series. 4) Long Short-Term Memory (LSTM) networks - a special type of RNN that can learn long-term dependencies. 5) Transformer networks - currently state-of-the-art for natural language processing tasks.",
		},
		{
			user:     "user",
			question: "How do I choose the right machine learning algorithm for my problem?",
			response: "Choosing the right algorithm depends on several factors: 1) Problem type - classification, regression, clustering, or reinforcement learning. 2) Data size - some algorithms work better with large datasets, others with small ones. 3) Data quality - how clean and complete is your data. 4) Interpretability requirements - do you need to explain the model's decisions. 5) Performance requirements - speed vs accuracy tradeoffs. 6) Available computational resources. Start with simple algorithms like linear regression or decision trees, then move to more complex ones like ensemble methods or neural networks if needed.",
		},
		{
			user:     "user",
			question: "What is the difference between AI, ML, and deep learning?",
			response: "AI (Artificial Intelligence) is the broadest term, referring to machines that can perform tasks that typically require human intelligence. ML (Machine Learning) is a subset of AI that focuses on algorithms that can learn from data without explicit programming. Deep Learning is a subset of ML that uses neural networks with multiple layers (hence 'deep') to learn complex patterns in data. Think of it as nested circles: AI contains ML, and ML contains Deep Learning. Each level gets more specific and typically more powerful for certain types of problems.",
		},
		{
			user:     "user",
			question: "What are some common challenges in machine learning projects?",
			response: "Common challenges include: 1) Data quality issues - missing, inconsistent, or biased data. 2) Overfitting - when models perform well on training data but poorly on new data. 3) Feature selection and engineering - choosing the right input variables. 4) Model interpretability - understanding how models make decisions. 5) Scalability - handling large datasets and real-time predictions. 6) Evaluation metrics - choosing the right way to measure success. 7) Deployment and maintenance - moving models to production and keeping them updated. 8) Ethical considerations - ensuring fairness and avoiding bias.",
		},
		{
			user:     "user",
			question: "How important is data preprocessing in machine learning?",
			response: "Data preprocessing is extremely important - it's often said that 80% of a data scientist's time is spent on data cleaning and preparation. Key preprocessing steps include: 1) Data cleaning - handling missing values, outliers, and errors. 2) Data transformation - scaling, normalization, and encoding categorical variables. 3) Feature engineering - creating new features from existing ones. 4) Data splitting - dividing data into training, validation, and test sets. 5) Handling imbalanced datasets - ensuring all classes are adequately represented. Poor preprocessing can lead to inaccurate models, while good preprocessing can significantly improve model performance.",
		},
	}

	for i, topic := range topics {
		// Add user question
		conversations = append(conversations, engine.Conversation{
			User: topic.user,
			Text: topic.question,
		})

		// Add bot response
		conversations = append(conversations, engine.Conversation{
			User: "assistant",
			Text: topic.response,
		})

		// Add some follow-up questions to make it longer
		if i < len(topics)-1 {
			conversations = append(conversations, engine.Conversation{
				User: "user",
				Text: "That's very helpful, thank you! Can you give me a practical example?",
			})

			conversations = append(conversations, engine.Conversation{
				User: "assistant",
				Text: "Certainly! For example, if you're building a recommendation system for an e-commerce site, you might use collaborative filtering to analyze user behavior patterns and suggest products based on what similar users have purchased. This involves collecting data on user purchases, preprocessing it to handle missing values, training a model to identify similar users or items, and then using that model to make real-time recommendations.",
			})
		}
	}

	// Note: Files are attached at the RunRequest level, not individual conversations
	conversations = append(conversations, engine.Conversation{
		User: "user",
		Text: "Here's a diagram I found that illustrates the machine learning workflow. I've attached it to this request.",
	})

	conversations = append(conversations, engine.Conversation{
		User: "assistant",
		Text: "Thank you for sharing the diagram! That's a great visual representation of the machine learning workflow. The diagram clearly shows the iterative process from data collection through model deployment and monitoring.",
	})

	// Add a final question that references the entire conversation
	conversations = append(conversations, engine.Conversation{
		User: "user",
		Text: "Based on everything we've discussed, what would you say is the most important thing for someone starting their first machine learning project?",
	})

	return conversations
}

// createDummyImageBase64 creates a small dummy image for demonstration
// In a real application, this would be actual image data
func createDummyImageBase64() string {
	// Create a small dummy PNG image (1x1 pixel)
	// This is just for demonstration - in practice you'd have real image data
	dummyPNG := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // Width=1, Height=1
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, // Bit depth=8, Color type=6
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, // IDAT chunk header
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, // Compressed data
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, // More compressed data
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, // IEND chunk
		0x42, 0x60, 0x82, // End
	}

	return base64.StdEncoding.EncodeToString(dummyPNG)
}

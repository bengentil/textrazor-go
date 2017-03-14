TextRazor Go SDK
====================

GO SDK for the TextRazor Text Analytics API.

TextRazor offers state-of-the-art natural language processing tools through a simple API, allowing you to build semantic technology into your applications in minutes.  

Hundreds of applications rely on TextRazor to understand unstructured text across a range of verticals, with use cases including social media monitoring, enterprise search, recommendation systems and ad targetting.  

Getting Started
===============

- Get a free API key from [https://www.textrazor.com](https://www.textrazor.com).

- Install the TextRazor Go SDK

	```bash
	go get -u github.com/bengentil/textrazor-go
	```

- Create an instance of the TextRazor object and start analyzing your text.

	```go
	package main

	import "github.com/bengentil/textrazor-go"

	const text="Barclays misled shareholders and the public about one of the biggest investments in the bank's history, a BBC Panorama investigation has found."

	func main() {
	  client := textrazor.NewClient(YOUR_API_KEY_HERE)
	  params := textrazor.Params{"extractors": {"entities", "entailments"}}
	  analysis, _ := client.AnalyzeText(text, params)

	  for _, entity := range analysis.Entities {
	    println(entity.EntityID, entity.ConfidenceScore)
	  }
	}
	```

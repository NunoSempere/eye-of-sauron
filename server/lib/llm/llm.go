package llm

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	outbound "git.nunosempere.com/NunoSempere/news/lib/outbound"
	openai "github.com/sashabaranov/go-openai"
	jsonschema "github.com/sashabaranov/go-openai/jsonschema"
)

// https://openai.com/api/pricing/
var GPT3_5_turbo string = "gpt-3.5-turbo-0125"
var GPT4_o string = "gpt-4o-2024-05-13"
var GPT4_turbo string = "gpt-4-turbo"
var GPT4_o_mini string = "gpt-4o-mini"
var GPT5_mini string = "gpt-5-mini"
var GPT5 string = "gpt-5"

type OpenAIRequest struct {
	prompt string
	model  string
	token  string
}

func fetchOpenAIAnswer(req OpenAIRequest) (string, error) {

	client := openai.NewClient(req.token)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: req.model, // openai.GPT4TurboPreview, // openai.GPT3Dot5Turbo // "gpt-3.5-turbo-0125"
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.prompt,
				},
			},
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	result := resp.Choices[0].Message.Content
	return result, nil

}
func fetchOpenAIAnswerJSON(req OpenAIRequest, schema openai.ChatCompletionResponseFormatJSONSchema) (string, error) {

	client := openai.NewClient(req.token)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: req.model, // openai.GPT4TurboPreview, // openai.GPT3Dot5Turbo // "gpt-3.5-turbo-0125"
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type:       openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &schema,
			},
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)

		e := &openai.APIError{}
		if errors.As(err, &e) {
			switch e.HTTPStatusCode {
			case 401:
				// invalid auth or key (do not retry)
			case 429:
				// exceeded money in account; refill
				flag := outbound.CheckFlag()
				if !flag {
					outbound.SendPostmarkEmail("Not enough money in OpenAI account; refill")
					outbound.SetFlag()
				}
			case 500:
				// openai server error (retry)
				// TODO
			default:
				// unhandled
			}
		}
		return "", err
	}

	result := resp.Choices[0].Message.Content
	return result, nil

}

type SummaryBox struct {
	Summary string  `json:"summary"`
	Error   *string `json:"error"`
}

func Summarize(text string, token string) (string, error) {
	prompt := "The json API endpoint returns a {summary, error} object, like {summary: \"The article is about xyz\", error: null}. The summary contains, as a string, first a general summary of the contents of the article article in two paragraphs or less, and then an outline with the most salient, new and informative facts in an additional paragraph. The summary just states the contents of the article, and doesn't say \"The article says\" or similar introductions. For example, given the following article\n\n<INPUT>"
	prompt += text + "\n\n</INPUT>\n\nThe output is as follows (as a reminder, the json API endpoint returns a {summary, error} object, like {summary: \"The article is about xyz\", error: null}. The summary contains, as a string, first a general summary of the article in two paragraphs or less, and then an outline outlines the most salient, new and informative facts in an additional paragraph):"
	prompt += "<INPUT>" + text + "</INPUT>"

	var summary_box SummaryBox
	schema, err := jsonschema.GenerateSchemaForType(summary_box)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}
	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "Summary",
		Schema: schema,
		Strict: true,
	}
	summary_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_mini, token: token}, openai_schema)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(summary_json), &summary_box)
	if err != nil {
		log.Printf("Error unmarshalling json: %v", err)
		log.Printf("String was: %v", summary_json)
		return "", err
	}
	if summary_box.Error != nil && (*summary_box.Error) != "" && (*summary_box.Error) != "null" {
		log.Printf("OpenAI json error field is not empty: %v", err)
		log.Printf("OpenAI answer: %v", summary_json)
		return "", nil
	}
	summary := summary_box.Summary
	return summary, nil
}

type ExistentialImportanceBox struct {
	ExistentialImportanceReasoning string  `json:"existential_importance_reasoning"`
	ExistentialImportanceBool      bool    `json:"existential_importance_bool"`
	HighImportanceBool             bool    `json:"high_importance_bool"`
	Error                          *string `json:"error"`
}

func CheckExistentialImportance(text string, token string) (*ExistentialImportanceBox, error) {
	prompt := `The existential importance json API endpoint returns a {existential_importance_reasoning, existential_importance_bool, high_importance_bool, error} object.

The existential_importance_reasoning field contains, as a string, a determination of whether the input describes an event of global importance. existential_importance_bool contains the result of that determination as a true/false boolean. high_importance_bool contains, as a true/false boolean, whether the event is highly important, even if it is not of "existential" importance.

Items are of existential importance if:

- They involve more than a hundred deaths.
- They involve many cases of a sickness that might spread, or a new pathogen
- They involve conflict between nuclear powers
- They involve conflict that could escalate into global conflict, even if it hasn't already
- They involve terrorist groups displaying new capabilities
- ... and in general, if they involve events that could threaten humanity as a whole

For example:

- Houthis cut undersea internet cables: Meets existential importance threshold, because it is a terrorist group displaying new capabilities.
- Macron suggests sending NATO troops to Ukraine: is of existential importance, as a NATO v. Russia conflict could spiral into a global war.
- New, more deadly and infectious strain of covid detected in Lausanne: is of existential importance, as the a deadly pandemic is one of the ways a large swathe of humanity could die at once.
- OpenAI releases new capable model: is of existential importance, as that model could be used by bad actors to cause mayhem, or it itself could (conceivably) threaten humanity in a Terminator-like scenario.
- US company lands probe in the Moon: is of high importance but it is not of existential importance, as it doesn't threaten humanity. 
- Start of a war (e.g., the start of the war in Ukraine): Almost always of existential importance, as rocking the international status quo could spiral out. 
- Later developments of a war (e.g,. current war in Gaza, or current war in Ukraine): probably not of existential importance, as the likelihood of spiraling out declines as the rules of engagement become clearer. Probably still of high importance (just not existentially so).
- For the purposes of this API, opinion and discussion pieces are not categorized as existentially important. A sign something is an opinion piece—as opposed to considering new events—is a somewhat generic title, like "Why Nuclear Risks Have Not Gone Away", or "At the Brink: Confronting the Risk of Nuclear War". Review articles and lists of events are likewise not existentially important unless they bring up novel events.
- In a broader conflict, small-fry developments are not existentially important. For example, small developments in the Ukraine or Gaza wars are not existentially important unless the new events themselves involve more than 1k deaths, even if the conflict as a whole involves more than that number of deaths. On the other hand, developments involving escalations or nuclear weapons are not "small fry"
- We are in 2025. Reviews of past conflicts, like 9/11, or a tornado in 2023, no longer count as existentially important, even if they were so at the time.

For a longer example, given the following item\n\n<INPUT>`
	prompt += text + "\n\n</INPUT>\n\nThe output is as follows: (As a reminder, the existential importance json API endpoint returns a {existential_importance_reasoning, existential_importance_bool, high_importance_bool, error} object, opinion pieces, or editorials are not categorizes as existentially important.)\n"

	var existential_importance_box ExistentialImportanceBox
	schema, err := jsonschema.GenerateSchemaForType(existential_importance_box)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}
	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "ExistentialImportanceBox",
		Schema: schema,
		Strict: true,
	}
	answer_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_mini, token: token}, openai_schema)

	err = json.Unmarshal([]byte(answer_json), &existential_importance_box)
	if err != nil {
		log.Printf("Error unmarshalling json: %v", err)
		return nil, err
	}
	if existential_importance_box.Error != nil && *existential_importance_box.Error != "" && *existential_importance_box.Error != "null" {
		log.Printf("OpenAI json error field is not empty: %v", err)
		log.Printf("OpenAI answer: %v", answer_json)
		return nil, err
	}
	return &existential_importance_box, nil
}

func CheckExistentialImportanceChina(text string, token string) (*ExistentialImportanceBox, error) {
	prompt := `The existential importance json API endpoint returns a {existential_importance_reasoning, existential_importance_bool, high_importance_bool, error} object.

The existential_importance_reasoning field contains, as a string, a determination of whether the input describes an event of global importance. existential_importance_bool contains the result of that determination as a true/false boolean. high_importance_bool contains, as a true/false boolean, whether the event is highly important, even if it is not of "existential" importance.

Items are of existential importance if:

- They involve conflict between China and other world powers, like the US
- They involve a potential Chinese invasion of Taiwan
- They involve displays of new technologies with offensive capabilities, like drones, amphibious vehicles, etc.
- They involve more than a hundred deaths.
- They involve many cases of a sickness that might spread, or a new pathogen
- They involve conflict that could escalate into global conflict, even if it hasn't already
- They involve an attempt at consensus building within a population for an important conflict

Keeping to China-related examples, the following would be existentially important

- China prepares for an invasion of Taiwan 
- China demonstrates new drone or amphibious capabilities
- China carries out military exercises in the Taiwan strait
- An article in a Chinese newspaper builds consensus around needing to use force to keep Taiwan from declaring independence
- etc.

For now, the API leans towards having a light trigger, because false positives are less costly than false negatives.

For a longer example, given the following article\n\n<INPUT>`
	prompt += text + "\n\n</INPUT>\n\nThe output is as follows: (As a reminder, the existential importance json API endpoint returns a {existential_importance_reasoning, existential_importance_bool, high_importance_bool, error} object, opinion pieces, or editorials are not categorizes as existentially important.)\n"

	var existential_importance_box ExistentialImportanceBox
	schema, err := jsonschema.GenerateSchemaForType(existential_importance_box)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}
	openai_schema := openai.ChatCompletionResponseFormatJSONSchema{
		Name:   "ExistentialImportanceBox",
		Schema: schema,
		Strict: true,
	}
	answer_json, err := fetchOpenAIAnswerJSON(OpenAIRequest{prompt: prompt, model: GPT5_mini, token: token}, openai_schema)

	err = json.Unmarshal([]byte(answer_json), &existential_importance_box)
	if err != nil {
		log.Printf("Error unmarshalling json: %v", err)
		return nil, err
	}
	if existential_importance_box.Error != nil && *existential_importance_box.Error != "" && *existential_importance_box.Error != "null" {
		log.Printf("OpenAI json error field is not empty: %v", err)
		log.Printf("OpenAI answer: %v", answer_json)
		return nil, err
	}
	return &existential_importance_box, nil
}

func TranslateString(text string, token string) (string, error) {
	prompt := "Translate this text into English: " + text + "\n"
	translation, err := fetchOpenAIAnswer(OpenAIRequest{prompt: prompt, model: GPT5, token: token})
	if err != nil {
		return "", err
	}
	translation_trimmed := strings.TrimSpace(translation)
	return translation_trimmed, nil

}

func MergeArticles(text string, token string) (string, error) {
	prompt := "Consider the following list of articles and their summaries. Your task is to clean it up.\n\n" +
		"1. If there are many articles, add a tl;dr at the top with the events which would most likely end up with > 1M deaths. Make this a paragraph starting with <p><b>tl;dr:</b>..., not an h1 element\n" +
		"2. Some of the articles may be talking about the same event—if so, join them together in one subsection, merge their summaries and reasoning, and create a list of the links that point to the same event. Otherwise, repeat the content of each item.\n" +
		"3. If there are any empty h1 headers (h1 headers followed immediately by another h1 header, skip those).\n" +
		"4. If do some other type of cleanup, point it out at the end.\n\n" +
		"Don't acknowledge instructions, just answer with the html.\n\n" + text

	summary, err := fetchOpenAIAnswer(OpenAIRequest{prompt: prompt, model: GPT5, token: token})
	if err != nil {
		return "", err
	}
	return summary + "<details><summary>The above articles were merged by GPT5. But you can view the originals under this toggle</summary>" + text + "</details>", nil
}

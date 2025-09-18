package filters

import (
	"log"

	"git.nunosempere.com/NunoSempere/news/lib/llm"
	"git.nunosempere.com/NunoSempere/news/lib/readability"
	"git.nunosempere.com/NunoSempere/news/lib/types"
)

// ApplyFilters applies a slice of filters sequentially, stopping at the first failure
func ApplyFilters(source types.ExpandedSource, filters []types.Filter) (types.ExpandedSource, bool) {
	for _, f := range filters {
		var ok bool
		source, ok = f(source)
		if !ok {
			return source, false
		}
	}
	return source, true
}

// StandardFilterPipeline creates a standard set of filters used by most sources
func DeprecatedStandardFilterPipeline(database_url string) []types.Filter {
	return []types.Filter{
		IsFreshFilter(),
		IsDupeFilter(database_url),
		IsGoodHostFilter(),
		CleanTitleFilter(),
	}
}

// ExtractContentAndSummarize extracts article content and generates summary using LLM
func ExtractContentAndSummarize(source types.ExpandedSource, openai_key string) (types.ExpandedSource, bool) {
	content, err := readability.GetArticleContent(source.Link)
	if err != nil {
		log.Printf("Filtered because: Error getting article content: %v", err)
		return source, false
	}

	summary, err := llm.Summarize(content, openai_key)
	if err != nil {
		log.Printf("Filtered because: Error summarizing: %v", err)
		return source, false
	}
	source.Summary = summary
	return source, true
}

// CheckImportance performs existential importance check using LLM
func CheckImportance(source types.ExpandedSource, openai_key string) (types.ExpandedSource, bool) {
	existential_importance_snippet := "# " + source.Title + "\n\n" + source.Summary
	existential_importance_box, err := llm.CheckExistentialImportance(existential_importance_snippet, openai_key)
	if err != nil || existential_importance_box == nil {
		log.Printf("Filtered because: is not important")
		return source, false
	}
	source.ImportanceBool = existential_importance_box.ExistentialImportanceBool
	source.ImportanceReasoning = existential_importance_box.ExistentialImportanceReasoning

	log.Printf("importance bool: %t", source.ImportanceBool)
	return source, source.ImportanceBool
}

// StandardProcessingPipeline processes source through standard filters, content extraction, and importance check
func DeprecatedStandardProcessingPipeline(source types.Source, openai_key string, database_url string) (types.ExpandedSource, bool) {
	// Initialize expanded source
	es := types.ExpandedSource{
		Title:  source.Title,
		Link:   source.Link,
		Date:   source.Date,
		Origin: source.Origin,
	}

	// Apply standard filters
	filters := DeprecatedStandardFilterPipeline(database_url)
	es, ok := ApplyFilters(es, filters)
	if !ok {
		return es, false
	}

	// Extract content and summarize
	es, ok = ExtractContentAndSummarize(es, openai_key)
	if !ok {
		return es, false
	}

	// Check importance
	return CheckImportance(es, openai_key)
}

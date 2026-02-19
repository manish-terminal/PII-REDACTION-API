package detector

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/asoasis/pii-redaction-api/internal/model"
)

type Pipeline struct {
	regex   *RegexDetector
	ner     *NERDetector
	context *ContextAnalyzer
}

func NewPipeline(defaultLocale string) *Pipeline {
	return &Pipeline{
		regex:   NewRegexDetector(defaultLocale),
		ner:     NewNERDetector(),
		context: NewContextAnalyzer(),
	}
}

func (p *Pipeline) Detect(ctx context.Context, req model.DetectionRequest) ([]model.Detection, error) {
	var (
		regexResults []model.Detection
		nerResults   []model.Detection
		regexErr     error
		nerErr       error
		wg           sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				regexErr = fmt.Errorf("regex detector panicked: %v", r)
			}
		}()
		regexResults, regexErr = p.regex.Detect(ctx, req.Text, req.Locale)
	}()

	/*
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					nerErr = fmt.Errorf("ner detector panicked: %v", r)
				}
			}()
			nerResults, nerErr = p.ner.Detect(ctx, req.Text)
		}()
	*/
	wg.Done() // Manually done for the commented out goroutine

	wg.Wait()

	if regexErr != nil {
		return nil, regexErr
	}
	if nerErr != nil {
		return nil, nerErr
	}

	// Merge
	merged := mergeDetections(regexResults, nerResults)

	// Refine
	refined := p.context.Refine(ctx, req.Text, merged)

	// Filter by entity types if requested
	if len(req.EntityTypes) > 0 {
		refined = filterByEntityTypes(refined, req.EntityTypes)
	}

	// Filter by confidence threshold
	threshold := req.ConfidenceThreshold
	if threshold == 0 {
		threshold = 0.60
	}
	refined = filterByConfidence(refined, threshold)

	// Sort by start position
	sort.Slice(refined, func(i, j int) bool {
		return refined[i].Start < refined[j].Start
	})

	return refined, nil
}

func mergeDetections(sets ...[]model.Detection) []model.Detection {
	var all []model.Detection
	for _, set := range sets {
		all = append(all, set...)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Start == all[j].Start {
			return all[i].End > all[j].End
		}
		return all[i].Start < all[j].Start
	})

	var merged []model.Detection
	for _, det := range all {
		if len(merged) == 0 || det.Start >= merged[len(merged)-1].End {
			merged = append(merged, det)
		} else if det.Confidence > merged[len(merged)-1].Confidence {
			merged[len(merged)-1] = det
		}
	}
	return merged
}

func filterByEntityTypes(detections []model.Detection, types []string) []model.Detection {
	typeMap := make(map[string]bool)
	for _, t := range types {
		typeMap[t] = true
	}

	var filtered []model.Detection
	for _, det := range detections {
		if typeMap[det.EntityType] {
			filtered = append(filtered, det)
		}
	}
	return filtered
}

func filterByConfidence(detections []model.Detection, threshold float64) []model.Detection {
	var filtered []model.Detection
	for _, det := range detections {
		if det.Confidence >= threshold {
			filtered = append(filtered, det)
		}
	}
	return filtered
}

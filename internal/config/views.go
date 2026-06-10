package config

import "maps"

const reviewNeededViewName = "review-needed"

var builtInViews = map[string]View{
	reviewNeededViewName: {
		Description: "Ready team PRs that do not have GitHub approval",
		Filters: map[string]interface{}{
			"author":          "team",
			"draft":           false,
			"bot":             false,
			"review_decision": "not-approved",
		},
	},
}

// RegisterBuiltInView registers a built-in View. User-defined config Views
// with the same name override built-ins when resolving available Views.
func RegisterBuiltInView(name string, view View) {
	builtInViews[name] = view
}

// RegisteredViews returns built-in Views plus user-defined Views. User-defined
// config Views win when names collide.
func RegisteredViews(cfg *Config) map[string]View {
	views := maps.Clone(builtInViews)
	if views == nil {
		views = map[string]View{}
	}
	if cfg == nil {
		return views
	}
	for name, view := range cfg.Views {
		views[name] = view
	}
	return views
}

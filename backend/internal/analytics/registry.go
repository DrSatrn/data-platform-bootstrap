// This file defines the curated metric registry returned to the frontend so
// the reporting UI can browse semantic metrics without hardcoding them.
package analytics

// MetricDefinition describes one repo-managed semantic metric.
type MetricDefinition struct {
	ID                   string   `json:"id" yaml:"id"`
	Name                 string   `json:"name" yaml:"name"`
	Description          string   `json:"description" yaml:"description"`
	Owner                string   `json:"owner" yaml:"owner"`
	DatasetRef           string   `json:"dataset_ref" yaml:"dataset_ref"`
	TimeDimension        string   `json:"time_dimension" yaml:"time_dimension"`
	Dimensions           []string `json:"dimensions" yaml:"dimensions"`
	Measures             []string `json:"measures" yaml:"measures"`
	DefaultVisualization string   `json:"default_visualization" yaml:"default_visualization"`
}

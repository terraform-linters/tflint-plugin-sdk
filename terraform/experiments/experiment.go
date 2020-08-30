package experiments

// Experiment is an alternative representation of experiments.Experiment.
// https://github.com/hashicorp/terraform/blob/v0.13.1/experiments/experiment.go#L5
type Experiment string

// Set is an alternative representation of experiments.Set.
// https://github.com/hashicorp/terraform/blob/v0.13.1/experiments/set.go#L5
type Set map[Experiment]struct{}

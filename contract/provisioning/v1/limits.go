package provisioning

import planlimitsv1 "github.com/7K-Inari/inari-api/gen/go/inari/planlimits/v1"

// LimitsFromProto validates and clones a wire document.
func LimitsFromProto(in *planlimitsv1.PlanLimits) (*Limits, error) {
	if in == nil {
		return nil, ValidateLimits(nil)
	}
	out := &Limits{
		MaxClusters:       cloneInt64(in.MaxClusters),
		MaxDeveloperSeats: cloneInt64(in.MaxDeveloperSeats),
		AllFeatures:       in.AllFeatures,
		FeatureGates:      cloneGates(in.FeatureGates),
	}
	if err := ValidateLimits(out); err != nil {
		return nil, err
	}
	return out, nil
}

// LimitsToProto validates and clones a wire document.
func LimitsToProto(in *Limits) (*planlimitsv1.PlanLimits, error) {
	if err := ValidateLimits(in); err != nil {
		return nil, err
	}
	return &planlimitsv1.PlanLimits{
		MaxClusters:       cloneInt64(in.MaxClusters),
		MaxDeveloperSeats: cloneInt64(in.MaxDeveloperSeats),
		AllFeatures:       in.AllFeatures,
		FeatureGates:      cloneGates(in.FeatureGates),
	}, nil
}

func cloneInt64(in *int64) *int64 {
	if in == nil {
		return nil
	}
	value := *in
	return &value
}

func cloneGates(in map[string]bool) map[string]bool {
	if in == nil {
		return nil
	}
	out := make(map[string]bool, len(in))
	for name, enabled := range in {
		out[name] = enabled
	}
	return out
}

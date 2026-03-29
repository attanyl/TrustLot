package recon

// ReconConfig holds tolerances and thresholds for reconciliation rules.
type ReconConfig struct {
	PriceTolerance    float64
	QuantityTolerance float64
}

// DefaultConfig returns hardcoded defaults suitable for MVP.
func DefaultConfig() ReconConfig {
	return ReconConfig{
		PriceTolerance:    0.01,
		QuantityTolerance: 0.0,
	}
}

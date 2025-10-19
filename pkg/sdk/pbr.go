package sdk

import (
	"context"
	"encoding/json"
)

// GetPBRPolicies retrieves all Policy-Based Routing (PBR) policies from the OpenWRT device.
func (o *OpenWRT) GetPBRPolicies(ctx context.Context) (map[string]PBR, error) {
	result, err := o.lucirpc.Uci(ctx, "get_all", []string{"pbr"})
	if err != nil {
		return nil, err
	}

	var cfgs map[string]PBR
	err = json.Unmarshal([]byte(result), &cfgs)
	if err != nil {
		return nil, err
	}

	return cfgs, nil
}

// EnablePBRPolicy enables or disables a specific PBR policy by its name.
func (o *OpenWRT) EnablePBRPolicy(ctx context.Context, policyName string, enabled bool) error {
	currentPolicies, err := o.GetPBRPolicies(ctx)
	if err != nil {
		return err
	}

	for cfg, currentPolicy := range currentPolicies {
		if currentPolicy.Name != policyName {
			continue
		}

		enableValue := "0"
		if enabled {
			enableValue = "1"
		}

		_, err := o.lucirpc.Uci(ctx, "set", []string{"pbr", cfg, "enabled", enableValue})
		if err != nil {
			return err
		}

		_, err = o.lucirpc.Uci(ctx, "commit", []string{"pbr"})
		if err != nil {
			return err
		}
	}

	return nil
}

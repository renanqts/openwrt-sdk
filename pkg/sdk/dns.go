package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GetDNSRecords retrieves all DNS records from the OpenWRT device.
func (o *OpenWRT) GetDNSRecords(ctx context.Context) (map[string]DNSRecord, error) {
	result, err := o.lucirpc.Uci(ctx, "get_all", []string{"dhcp"})
	if err != nil {
		return nil, err
	}

	var records map[string]DNSRecord
	err = json.Unmarshal([]byte(result), &records)
	if err != nil {
		return nil, err
	}

	for key, record := range records {
		switch record.Type {
		case "domain":
			records[key] = DNSRecord{
				Type: "A",
				IP:   record.IP,
				Name: record.Name,
			}
		case "cname":
			records[key] = DNSRecord{
				Type:   "CNAME",
				CName:  record.CName,
				Target: record.Target,
			}
		default:
			// it does not care about other types
			delete(records, key)
		}
	}

	return records, nil
}

// SetDNSRecords adds new DNS records to the OpenWRT device.
func (o *OpenWRT) SetDNSRecords(ctx context.Context, records []DNSRecord) error {
	for _, record := range records {
		switch strings.ToUpper(record.Type) {
		case "A":
			if err := o.addA(ctx, record); err != nil {
				return err
			}
		case "CNAME":
			if err := o.addCName(ctx, record); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid record type: %s", record.Type)
		}
	}

	if _, err := o.lucirpc.Uci(ctx, "commit", []string{"dhcp"}); err != nil {
		return err
	}

	return nil
}

// UpdateDNSRecords updates existing DNS records on the OpenWRT device.
func (o *OpenWRT) UpdateDNSRecords(ctx context.Context, updateRecords []DNSRecord) error {
	currentRecords, err := o.GetDNSRecords(ctx)
	if err != nil {
		return err
	}

	for cfg, currentRecord := range currentRecords {
		for index, updateRecord := range updateRecords {
			if updateRecord.Type == "A" && updateRecord.Name == currentRecord.Name {
				_, err := o.lucirpc.Uci(ctx, "delete", []string{"dhcp", cfg})
				if err != nil {
					return err
				}

				if err := o.addA(ctx, updateRecord); err != nil {
					return err
				}
				updateRecords = append(updateRecords[:index], updateRecords[index+1:]...)
			}

			if updateRecord.Type == "CNAME" && updateRecord.CName == currentRecord.CName {
				_, err := o.lucirpc.Uci(ctx, "delete", []string{"dhcp", cfg})
				if err != nil {
					return err
				}

				if err := o.addCName(ctx, updateRecord); err != nil {
					return err
				}
				updateRecords = append(updateRecords[:index], updateRecords[index+1:]...)
			}
		}
	}

	if len(updateRecords) > 0 {
		return fmt.Errorf("records not found: %v", updateRecords)
	}

	if _, err := o.lucirpc.Uci(ctx, "commit", []string{"dhcp"}); err != nil {
		return err
	}

	return nil
}

// DeleteDNSRecords deletes DNS records from the OpenWRT device.
func (o *OpenWRT) DeleteDNSRecords(ctx context.Context, deleteRecords []DNSRecord) error {
	currentRecords, err := o.GetDNSRecords(ctx)
	if err != nil {
		return err
	}

	for cfg, currentRecord := range currentRecords {
		for index, deleteRecord := range deleteRecords {
			if (deleteRecord.Type == "A" && deleteRecord.Name == currentRecord.Name) ||
				(deleteRecord.Type == "CNAME" && deleteRecord.CName == currentRecord.CName) {
				_, err := o.lucirpc.Uci(ctx, "delete", []string{"dhcp", cfg})
				if err != nil {
					return err
				}
				deleteRecords = append(deleteRecords[:index], deleteRecords[index+1:]...)
			}
		}
	}

	if len(deleteRecords) > 0 {
		return fmt.Errorf("records not found: %v", deleteRecords)
	}

	// should we remove even when records not found?
	if _, err := o.lucirpc.Uci(ctx, "commit", []string{"dhcp"}); err != nil {
		return err
	}

	return nil
}

func (o *OpenWRT) addA(ctx context.Context, record DNSRecord) error {
	if record.Type != "a" && record.Type != "A" {
		return fmt.Errorf("invalid record type: %s", record.Type)
	}

	if record.Name == "" {
		return fmt.Errorf("name is required")
	}

	if record.IP == "" {
		return fmt.Errorf("ip is required")
	}

	cfg, err := o.lucirpc.Uci(ctx, "add", []string{"dhcp", "domain"})
	if err != nil {
		return err
	}

	if _, err := o.lucirpc.Uci(ctx, "set", []string{"dhcp", cfg, "name", record.Name}); err != nil {
		return err
	}

	if _, err := o.lucirpc.Uci(ctx, "set", []string{"dhcp", cfg, "ip", record.IP}); err != nil {
		return err
	}

	return nil
}

func (o *OpenWRT) addCName(ctx context.Context, record DNSRecord) error {
	if record.Type != "cname" && record.Type != "CNAME" {
		return fmt.Errorf("invalid record type: %s", record.Type)
	}

	if record.CName == "" {
		return fmt.Errorf("cname is required")
	}

	if record.Target == "" {
		return fmt.Errorf("target is required")
	}

	cfg, err := o.lucirpc.Uci(ctx, "add", []string{"dhcp", "cname"})
	if err != nil {
		return err
	}

	if _, err := o.lucirpc.Uci(ctx, "set", []string{"dhcp", cfg, "cname", record.CName}); err != nil {
		return err
	}

	if _, err := o.lucirpc.Uci(ctx, "set", []string{"dhcp", cfg, "target", record.Target}); err != nil {
		return err
	}

	return nil
}

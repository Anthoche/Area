package areas

import (
	"context"
	"encoding/json"
	"fmt"

	"area/src/database"

	"gorm.io/gorm"
)

// List returns the service catalog stored in the database.
func List(ctx context.Context) ([]Service, error) {
	if database.Db == nil {
		return nil, fmt.Errorf("catalog database not initialized")
	}
	services, err := gorm.G[database.AreaService](database.Db).Order("id").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	caps, err := gorm.G[database.AreaCapability](database.Db).Order("id").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("list capabilities: %w", err)
	}
	fields, err := gorm.G[database.AreaField](database.Db).Order("id").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	fieldsByCap := map[string][]Field{}
	for _, f := range fields {
		key := f.ServiceID + ":" + f.CapabilityID
		fieldsByCap[key] = append(fieldsByCap[key], areaFieldFromModel(f))
	}

	out := make([]Service, 0, len(services))
	serviceByID := map[string]*Service{}
	for _, svc := range services {
		converted := Service{
			ID:         svc.ID,
			Name:       svc.Name,
			Enabled:    svc.Enabled,
			MoreInfo:   svc.MoreInfo,
			OAuthScope: decodeStringSlice(svc.OAuthScopes),
			Hidden:     svc.ID == "core",
		}
		out = append(out, converted)
		serviceByID[svc.ID] = &out[len(out)-1]
	}

	for _, cap := range caps {
		svc := serviceByID[cap.ServiceID]
		if svc == nil {
			continue
		}
		converted := Capability{
			ID:             cap.ID,
			Name:           cap.Name,
			Description:    cap.Description,
			ActionURL:      cap.ActionURL,
			DefaultPayload: decodeMap(cap.DefaultPayload),
			Fields:         fieldsByCap[svc.ID+":"+cap.ID],
		}
		if cap.Kind == "trigger" {
			svc.Triggers = append(svc.Triggers, converted)
		} else {
			svc.Reactions = append(svc.Reactions, converted)
		}
	}

	return out, nil
}

// Get returns a single service by ID from the database.
func Get(ctx context.Context, id string) (*Service, error) {
	if id == "" {
		return nil, fmt.Errorf("service id is required")
	}
	if database.Db == nil {
		return nil, fmt.Errorf("catalog database not initialized")
	}
	svc, err := gorm.G[database.AreaService](database.Db).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("get service: %w", err)
	}
	caps, err := gorm.G[database.AreaCapability](database.Db).Where("service_id = ?", id).Order("id").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("list capabilities: %w", err)
	}
	capIDs := make([]string, 0, len(caps))
	for _, cap := range caps {
		capIDs = append(capIDs, cap.ID)
	}
	fieldsByCap := map[string][]Field{}
	if len(capIDs) > 0 {
		fields, err := gorm.G[database.AreaField](database.Db).Where("service_id = ? AND capability_id IN ?", id, capIDs).Order("id").Find(ctx)
		if err != nil {
			return nil, fmt.Errorf("list fields: %w", err)
		}
		for _, f := range fields {
			key := f.ServiceID + ":" + f.CapabilityID
			fieldsByCap[key] = append(fieldsByCap[key], areaFieldFromModel(f))
		}
	}
	out := Service{
		ID:         svc.ID,
		Name:       svc.Name,
		Enabled:    svc.Enabled,
		MoreInfo:   svc.MoreInfo,
		OAuthScope: decodeStringSlice(svc.OAuthScopes),
		Hidden:     svc.ID == "core",
	}
	for _, cap := range caps {
		converted := Capability{
			ID:             cap.ID,
			Name:           cap.Name,
			Description:    cap.Description,
			ActionURL:      cap.ActionURL,
			DefaultPayload: decodeMap(cap.DefaultPayload),
			Fields:         fieldsByCap[id+":"+cap.ID],
		}
		if cap.Kind == "trigger" {
			out.Triggers = append(out.Triggers, converted)
		} else {
			out.Reactions = append(out.Reactions, converted)
		}
	}
	return &out, nil
}

func areaFieldFromModel(f database.AreaField) Field {
	return Field{
		Key:         f.Key,
		Type:        f.Type,
		Required:    f.Required,
		Description: f.Description,
		Example:     decodeAny(f.Example),
	}
}

func decodeMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func decodeStringSlice(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func decodeAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

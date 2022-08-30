package grpc_api

import (
	"context"
	"github.com/webitel/engine/model"
	"github.com/webitel/protos/engine"
	"strings"
)

type trigger struct {
	*API
	engine.UnsafeTriggerServiceServer
}

func NewTriggerApi(api *API) *trigger {
	return &trigger{API: api}
}

func (api *trigger) CreateTrigger(ctx context.Context, in *engine.CreateTriggerRequest) (*engine.Trigger, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.Trigger{
		Name:        in.Name,
		Enabled:     in.Enabled,
		Type:        model.TriggerTypeCron,
		Schema:      GetLookup(in.Schema),
		Variables:   in.GetVariables(),
		Description: in.Description,
		Expression:  in.Expression,
		Timezone:    GetLookup(in.Timezone),
		Timeout:     in.Timeout,
	}
	var tr *model.Trigger
	tr, err = api.ctrl.CreateTrigger(session, req)
	if err != nil {
		return nil, err
	}

	return toEngineTrigger(tr), nil
}

func (api *trigger) SearchTrigger(ctx context.Context, in *engine.SearchTriggerRequest) (*engine.ListTrigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var list []*model.Trigger
	var endList bool
	req := &model.SearchTrigger{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    int(in.GetPage()),
			PerPage: int(in.GetSize()),
			Fields:  in.Fields,
			Sort:    in.Sort,
		},
		Ids: in.Id,
	}

	list, endList, err = api.ctrl.SearchTrigger(session, req)

	items := make([]*engine.Trigger, 0, len(list))
	for _, v := range list {
		items = append(items, toEngineTrigger(v))
	}
	return &engine.ListTrigger{
		Next:  !endList,
		Items: items,
	}, nil
}

func (api *trigger) ReadTrigger(ctx context.Context, in *engine.ReadTriggerRequest) (*engine.Trigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var tr *model.Trigger

	tr, err = api.ctrl.ReadTrigger(session, in.Id)

	if err != nil {
		return nil, err
	}

	return toEngineTrigger(tr), nil
}

func (api *trigger) UpdateTrigger(ctx context.Context, in *engine.UpdateTriggerRequest) (*engine.Trigger, error) {
	session, err := api.ctrl.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	req := &model.Trigger{
		Id:          in.Id,
		Name:        in.Name,
		Enabled:     in.Enabled,
		Type:        model.TriggerTypeCron,
		Schema:      GetLookup(in.Schema),
		Variables:   in.GetVariables(),
		Description: in.Description,
		Expression:  in.Expression,
		Timezone:    GetLookup(in.Timezone),
		Timeout:     in.Timeout,
	}
	var tr *model.Trigger
	tr, err = api.ctrl.UpdateTrigger(session, req)
	if err != nil {
		return nil, err
	}

	return toEngineTrigger(tr), nil
}

func (api *trigger) PatchTrigger(ctx context.Context, in *engine.PatchTriggerRequest) (*engine.Trigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var tr *model.Trigger
	patch := &model.TriggerPatch{}

	//TODO
	for _, v := range in.Fields {
		switch v {
		case "name":
			patch.Name = &in.Name
		case "enabled":
			patch.Enabled = &in.Enabled
		case "schema.id":
			patch.Schema = GetLookup(in.Schema)
		case "description":
			patch.Description = &in.Description
		case "expression":
			patch.Expression = &in.Expression
		case "timezone.id":
			patch.Timezone = GetLookup(in.Timezone)
		default:
			if patch.Variables == nil && strings.HasPrefix(v, "variables.") {
				patch.Variables = in.Variables
			}
		}
	}

	tr, err = api.ctrl.PatchTrigger(session, in.Id, patch)
	if err != nil {
		return nil, err
	}

	return toEngineTrigger(tr), nil
}

func (api *trigger) DeleteTrigger(ctx context.Context, in *engine.DeleteTriggerRequest) (*engine.Trigger, error) {
	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var tr *model.Trigger
	tr, err = api.ctrl.RemoveTrigger(session, in.Id)
	if err != nil {
		return nil, err
	}

	return toEngineTrigger(tr), nil
}

func toEngineTrigger(src *model.Trigger) *engine.Trigger {

	return &engine.Trigger{
		Id:          src.Id,
		Name:        src.Name,
		Enabled:     src.Enabled,
		Type:        engine.TriggerType_cron, // TODO
		Schema:      GetProtoLookup(src.Schema),
		Variables:   src.Variables,
		Description: src.Description,
		Expression:  src.Expression,
		Timezone:    GetProtoLookup(src.Timezone),
		Timeout:     src.Timeout,
	}
}

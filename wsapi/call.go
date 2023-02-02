package wsapi

import (
	"fmt"
	"github.com/webitel/engine/app"
	"github.com/webitel/engine/model"
	"strings"
	"time"
)

func (api *API) InitCall() {
	api.Router.Handle("subscribe_call", api.ApiWebSocketHandler(api.subscribeSelfCalls))
	api.Router.Handle("un_subscribe_call", api.ApiWebSocketHandler(api.unSubscribeSelfCalls))

	api.Router.Handle("call_invite", api.ApiAsyncWebSocketHandler(api.callInvite))
	api.Router.Handle("call_eavesdrop", api.ApiAsyncWebSocketHandler(api.callEavesdrop))
	api.Router.Handle("call_eavesdrop_state", api.ApiAsyncWebSocketHandler(api.callEavesdropState))
	api.Router.Handle("call_user", api.ApiAsyncWebSocketHandler(api.callToUser))
	api.Router.Handle("call_hangup", api.ApiWebSocketHandler(api.callHangup))
	api.Router.Handle("call_hold", api.ApiWebSocketHandler(api.callHold))
	api.Router.Handle("call_unhold", api.ApiWebSocketHandler(api.callUnHold))
	api.Router.Handle("call_dtmf", api.ApiWebSocketHandler(api.callDTMF))
	api.Router.Handle("call_mute", api.ApiWebSocketHandler(api.callMute))
	api.Router.Handle("call_blind_transfer", api.ApiWebSocketHandler(api.callBlindTransfer))
	api.Router.Handle("call_bridge", api.ApiWebSocketHandler(api.callBridge))
	api.Router.Handle("call_recordings", api.ApiWebSocketHandler(api.callRecording))

	api.Router.Handle("call_by_user", api.ApiAsyncWebSocketHandler(api.callByUser))
	api.Router.Handle("test", api.ApiAsyncWebSocketHandler(api.test))

	api.Router.Handle("sip_proxy", api.ApiWebSocketHandler(api.sipProxy))
}

func (api *API) test(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	time.Sleep(time.Second * 10)
	return nil, nil
}

func (api *API) callByUser(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	calls, err := api.ctrl.UserActiveCall(conn.GetSession())

	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["items"] = calls
	return res, nil
}

func (api *API) sipProxy(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var data string
	if data, ok = req.Data["data"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "data")
	}
	//conn.Sip.Send([]byte(data))
	if data != "" {
	}
	return nil, nil
}

func (api *API) subscribeSelfCalls(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	h, e := api.App.GetHubById(req.Session.Domain(0))
	if e != nil {
		return nil, e
	}

	e = h.SubscribeSessionCalls(conn)
	if e != nil {
		return nil, e
	}

	return api.callByUser(conn, req)
}

func (api *API) unSubscribeSelfCalls(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	h, e := api.App.GetHubById(req.Session.Domain(0))
	if e != nil {
		return nil, e
	}

	return nil, h.UnSubscribeCalls(conn)
}

func (api *API) callEavesdrop(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	reqEa := &model.EavesdropCall{}

	if reqEa.Id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	reqEa.Dtmf, _ = req.Data["control"].(bool)
	reqEa.ALeg, _ = req.Data["listenA"].(bool)
	reqEa.BLeg, _ = req.Data["listenB"].(bool)
	reqEa.WhisperALeg, _ = req.Data["whisperA"].(bool)
	reqEa.WhisperBLeg, _ = req.Data["whisperB"].(bool)

	reqEa.State, _ = req.Data["state"].(string)

	vars := make(map[string]string)
	vars[model.CALL_VARIABLE_SOCK_ID] = conn.Id()

	callId, err := api.ctrl.EavesdropCall(conn.GetSession(), req.Session.DomainId, reqEa, vars)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = callId

	return res, nil
}

func (api *API) callEavesdropState(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	reqEa := &model.EavesdropCall{}

	if reqEa.Id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if reqEa.State, ok = req.Data["state"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "state")
	}

	err := api.ctrl.EavesdropStateCall(conn.GetSession(), req.Session.DomainId, reqEa)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	return res, nil
}

func (api *API) callHangup(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var id, nodeId string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	var cause = req.GetFieldString("cause")

	cr := model.HangupCall{
		UserCallRequest: model.UserCallRequest{
			Id:    id,
			AppId: &nodeId,
		},
		Cause: nil,
	}

	if cause != "" {
		cr.Cause = &cause
	}

	err := api.App.HangupCall(conn.GetSession().DomainId, &cr)

	return nil, err
}

func (api *API) callBlindTransfer(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var id, destination string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if destination, ok = req.Data["destination"].(string); !ok || len(destination) < 1 {
		return nil, NewInvalidWebSocketParamError(req.Action, "destination")
	}

	err := api.ctrl.BlindTransferCall(conn.GetSession(), conn.DomainId, &model.BlindTransferCall{
		UserCallRequest: model.UserCallRequest{
			Id: id,
		},
		Destination: destination,
	})

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (api *API) callHold(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var id, nodeId string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	if cli, err := api.App.CallManager().CallClientById(nodeId); err != nil {
		return nil, err
	} else {
		err = cli.Hold(id)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (api *API) callDTMF(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var id, nodeId string
	var key string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}
	if key, ok = req.Data["dtmf"].(string); !ok || len(key) < 1 {
		return nil, NewInvalidWebSocketParamError(req.Action, "dtmf")
	}

	if cli, err := api.App.CallManager().CallClientById(nodeId); err != nil {
		return nil, err
	} else {
		err = cli.DTMF(id, []rune(key)[0])
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (api *API) callUnHold(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var id, nodeId string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}

	if cli, err := api.App.CallManager().CallClientById(nodeId); err != nil {
		return nil, err
	} else {
		err = cli.UnHold(id)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (api *API) callInvite(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var callReq = &model.OutboundCallRequest{}
	var ok bool
	var props map[string]interface{}
	if callReq.Destination, ok = req.Data["destination"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "destination")
	}

	if props, ok = req.Data["params"].(map[string]interface{}); ok {
		var variables map[string]interface{}

		callReq.Params.Timeout, _ = props["timeout"].(int)
		callReq.Params.Video, _ = props["video"].(bool)
		callReq.Params.Screen, _ = props["screen"].(bool)
		callReq.Params.Record, _ = props["record"].(bool)
		callReq.Params.DisableAutoAnswer, _ = props["disableAutoAnswer"].(bool)
		callReq.Params.Display, _ = props["display"].(string)

		if variables, ok = props["variables"].(map[string]interface{}); ok {
			callReq.Params.Variables = make(map[string]string)
			for k, v := range variables {
				switch v.(type) {
				case string:
					callReq.Params.Variables[k] = v.(string)
				case interface{}:
					callReq.Params.Variables[k] = fmt.Sprintf("%v", v)
				}
			}
		}
	}

	vars := make(map[string]string)
	vars[model.CALL_VARIABLE_SOCK_ID] = conn.Id()

	if id, err := api.ctrl.CreateCall(conn.GetSession(), callReq, vars); err != nil {
		return nil, err
	} else {
		data := make(map[string]interface{})
		data["id"] = id
		return data, nil
	}
}

func (api *API) callToUser(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok, useVideo, useScreen bool
	var callId, callToId, parentCallId, sendToCallId string
	var toUserId float64
	var variables map[string]interface{}

	if toUserId, ok = req.Data["toUserId"].(float64); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "toUserId")
	}
	sendToCallId, _ = req.Data["sendToCallId"].(string)

	callId = model.NewUuid()
	callToId = model.NewUuid()

	info, err := api.App.GetUserCallInfo(conn.UserId, conn.DomainId)
	if err != nil {
		return nil, err
	}

	infoTo, err := api.App.GetUserCallInfo(int64(toUserId), conn.DomainId)
	if err != nil {
		return nil, err
	}

	invite := &model.CallRequest{
		Endpoints: info.GetCallEndpoints(),
		Variables: map[string]string{
			model.CALL_VARIABLE_ID:                callId,
			model.CALL_VARIABLE_DIRECTION:         model.CALL_DIRECTION_INTERNAL,
			model.CALL_VARIABLE_DISPLAY_DIRECTION: model.CALL_DIRECTION_OUTBOUND,
			model.CALL_VARIABLE_USER_ID:           fmt.Sprintf("%v", conn.UserId),
			model.CALL_VARIABLE_DOMAIN_ID:         fmt.Sprintf("%v", conn.DomainId),
			model.CALL_VARIABLE_SOCK_ID:           conn.Id(),

			"sip_h_X-Webitel-Destination": infoTo.Extension,

			"origination_uuid": callId,
			//"media_webrtc":     "true",
			//"absolute_codec_string": "VP8",

			"hangup_after_bridge":        "true",
			"hold_music":                 "silence",
			"effective_caller_id_number": info.Extension,
			"effective_caller_id_name":   info.Name,
			"effective_callee_id_name":   infoTo.Name,
			"effective_callee_id_number": infoTo.Extension,

			"origination_caller_id_name":   infoTo.Name,
			"origination_caller_id_number": infoTo.Extension,
			"origination_callee_id_name":   info.Name,
			"origination_callee_id_number": info.Extension,
		},
		Timeout:      0,
		CallerName:   infoTo.Name,
		CallerNumber: infoTo.Extension,
		Applications: []*model.CallRequestApplication{
			{
				AppName: "bridge",
				Args: fmt.Sprintf("{sip_route_uri=%s,request_parent_call_id=%s,origination_uuid=%s,sip_h_X-Webitel-Uuid=%s, sip_h_X-Webitel-User-Id=%d}%s", api.App.CallManager().SipRouteUri(),
					sendToCallId, callToId, callToId, int64(toUserId), strings.Join(infoTo.GetCallEndpoints(), ",")),
			},
		},
	}

	if variables, ok = req.Data["variables"].(map[string]interface{}); ok {
		for k, v := range variables {
			switch v.(type) {
			case string:
				invite.AddUserVariable(k, v.(string))
			case interface{}:
				invite.AddUserVariable(k, fmt.Sprintf("%v", v))
			}
		}
	}

	if useVideo, ok = req.Data["useVideo"].(bool); ok && useVideo {
		invite.AddVariable("video_request", "true")
	}

	if useScreen, ok = req.Data["useScreen"].(bool); ok && useScreen {
		invite.AddVariable("screen_request", "true")
	}

	if parentCallId, ok = req.Data["parentCallId"].(string); ok && parentCallId != "" {
		invite.AddVariable("request_parent_call_id", parentCallId)
	}

	_, err = api.App.CallManager().MakeOutboundCall(invite)

	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	data["call_id"] = callId
	return data, nil
}

func (api *API) callMute(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok, mute bool
	var id, nodeId string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}
	if mute, ok = req.Data["mute"].(bool); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "mute")
	}

	if cli, err := api.App.CallManager().CallClientById(nodeId); err != nil {
		return nil, err
	} else {
		err = cli.Mute(id, mute)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (api *API) callBridge(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var fromId, toId string

	if fromId, ok = req.Data["from_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "from_id")
	}

	if toId, ok = req.Data["to_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "to_id")
	}
	res := make(map[string]interface{})
	err := api.App.BridgeCall(conn.DomainId, fromId, toId)
	//FIXME set result
	return res, err
}

func (api *API) callRecording(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var id string
	var ok bool

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}

	fileId, err := api.App.GetLastCallFile(conn.DomainId, id)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res["file_id"] = fileId
	return res, nil
}

func (api *API) callSendVideo(conn *app.WebConn, req *model.WebSocketRequest) (map[string]interface{}, *model.AppError) {
	var ok bool
	var id, nodeId, id2, nodeId2 string

	if id, ok = req.Data["id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "id")
	}
	if nodeId, ok = req.Data["app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "app_id")
	}
	if id2, ok = req.Data["parent_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "parent_id")
	}
	if nodeId2, ok = req.Data["parent_app_id"].(string); !ok {
		return nil, NewInvalidWebSocketParamError(req.Action, "parent_app_id")
	}

	api.App.CallManager().Bridge(id, nodeId, id2, nodeId2)

	return nil, nil
}

package model

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	CALL_DIRECTION_INTERNAL = "internal"
	CALL_DIRECTION_INBOUND  = "inbound"
	CALL_DIRECTION_OUTBOUND = "outbound"
)

const (
	CALL_VARIABLE_DIRECTION         = "sip_h_X-Webitel-Direction"
	CALL_VARIABLE_DISPLAY_DIRECTION = "sip_h_X-Webitel-Display-Direction"
	CALL_VARIABLE_USER_ID           = "sip_h_X-Webitel-User-Id"
	CALL_VARIABLE_DOMAIN_ID         = "sip_h_X-Webitel-Domain-Id"
	CALL_VARIABLE_SOCK_ID           = "sip_h_X-Webitel-Sock-Id"
	CALL_VARIABLE_ID                = "sip_h_X-Webitel-Uuid"
	CALL_VARIABLE_USE_VIDEO         = "wbt_video"
	CALL_VARIABLE_USE_SCREEN        = "wbt_screen"
	CALL_VARIABLE_SIP_AUTO_ANSWER   = "sip_auto_answer"
)
const (
	CALL_STRATEGY_DEFAULT = iota
	CALL_STRATEGY_FAILOVER
	CALL_STRATEGY_MULTIPLE
)

type CallRequestApplication struct {
	AppName string
	Args    string
}

const (
	EndpointTypeUser        = "user"
	EndpointTypeDestination = "destination"
)

type Endpoint struct {
	Type   string `json:"type"`
	Number string `json:"number"`
	Id     string `json:"id"`
	Name   string `json:"name"`
}

type EndpointRequest struct {
	AppId       *string
	UserId      *int64
	Extension   *string
	SchemaId    *int
	Destination *string
}

type CallRequest struct {
	Endpoints    []string
	Strategy     uint8
	Destination  string
	Variables    map[string]string
	Timeout      uint16
	CallerName   string
	CallerNumber string
	Dialplan     string
	Context      string
	Applications []*CallRequestApplication
}

type OutboundCallRequest struct {
	CreatedAt   int64            `json:"created_at"`
	CreatedById int64            `json:"created_by_id"`
	From        *EndpointRequest `json:"from"`
	To          *EndpointRequest `json:"to"`
	Destination string           `json:"destination"`
	Params      CallParameters   `json:"params"`
}

type UserCallRequest struct {
	Id    string  `json:"id"`
	AppId *string `json:"app_id"`
}

type HangupCall struct {
	UserCallRequest
	Cause *string `json:"cause"`
}

type DtmfCall struct {
	UserCallRequest
	Digit rune
}

type BlindTransferCall struct {
	UserCallRequest
	Destination string
}

type BridgeCall struct {
	FromId string `json:"from_id" db:"from_id"`
	ToId   string `json:"to_id" db:"to_id"`
	AppId  string `json:"app_id" db:"app_id"`
}

type EavesdropCall struct {
	UserCallRequest
	//Group       string //TODO https://freeswitch.org/confluence/display/FREESWITCH/mod_dptools%3A+eavesdrop
	Dtmf        bool
	ALeg        bool
	BLeg        bool
	WhisperALeg bool
	WhisperBLeg bool
}

type CallParameters struct {
	Timeout int
	Audio   bool
	Video   bool
	Screen  bool

	Record           bool
	AutoAnswer       bool
	Display          string
	Variables        map[string]string
	DisableStun      bool
	CancelDistribute bool
	IsOnline         bool
}

func (r *OutboundCallRequest) IsValid() *AppError {
	return nil
}

type CallInstance struct {
	Id        string  `json:"id" db:"id"`
	AppId     *string `json:"app_id" db:"app_id"`
	State     string  `json:"state" db:"state"`
	Timestamp int64   `json:"timestamp" db:"timestamp"`
}

//
type CallHold struct {
	Start  int64   `json:"start"`
	Finish int64   `json:"finish"`
	Sec    float32 `json:"sec"`
}

type Call struct {
	Id          string            `json:"id" db:"id"`
	AppId       string            `json:"app_id" db:"app_id"`
	State       string            `json:"state" db:"state"`
	Timestamp   *time.Time        `json:"timestamp" db:"timestamp"`
	Type        string            `json:"type" db:"type"`
	ParentId    *string           `json:"parent_id" db:"parent_id"`
	User        *Lookup           `json:"user" db:"user"`
	Extension   *string           `json:"extension" db:"extension"`
	Gateway     *Lookup           `json:"gateway" db:"gateway"`
	Direction   string            `json:"direction" db:"direction"`
	Destination string            `json:"destination" db:"destination"`
	From        *Endpoint         `json:"from" db:"from"`
	To          *Endpoint         `json:"to" db:"to"`
	Variables   map[string]string `json:"variables" db:"variables"`

	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	AnsweredAt *time.Time `json:"answered_at" db:"answered_at"`
	BridgedAt  *time.Time `json:"bridged_at" db:"bridged_at"`
	HangupAt   *time.Time `json:"hangup_at" db:"hangup_at"`

	Duration int `json:"duration" db:"duration"`
	HoldSec  int `json:"hold_sec" db:"hold_sec"`
	WaitSec  int `json:"wait_sec" db:"wait_sec"`
	BillSec  int `json:"bill_sec" db:"bill_sec"`

	Queue  *Lookup `json:"queue" db:"queue"`
	Member *Lookup `json:"member" db:"member"`
	Team   *Lookup `json:"team" db:"team"`
	Agent  *Lookup `json:"agent" db:"agent"`

	JoinedAt         *time.Time `json:"joined_at" db:"joined_at"`
	LeavingAt        *time.Time `json:"leaving_at" db:"leaving_at"`
	ReportingAt      *time.Time `json:"reporting_at" db:"reporting_at"`
	QueueBridgedAt   *time.Time `json:"queue_bridged_at" db:"queue_bridged_at"`
	QueueWaitSec     *int       `json:"queue_wait_sec" db:"queue_wait_sec"`
	QueueDurationSec *int       `json:"queue_duration_sec" db:"queue_duration_sec"`
	ReportingSec     *int       `json:"reporting_sec" db:"reporting_sec"`
	Display          *string    `json:"display" db:"display"`
	Supervisor       []*Lookup  `json:"supervisor" db:"supervisor"`

	Task          *CCTask     `json:"task"`
	Hold          []*CallHold `json:"hold" db:"hold"`
	BlindTransfer *string     `json:"blind_transfer" db:"blind_transfer"`
}

type CCTask struct {
	AppId                string               `json:"app_id" db:"app_id"`
	HasReporting         bool                 `json:"has_reporting" db:"has_reporting"`
	AttemptId            int64                `json:"attempt_id" db:"attempt_id"`
	Channel              string               `json:"channel" db:"channel"`
	QueueId              *int                 `json:"queue_id" db:"queue_id"`
	QueueName            *string              `json:"queue_name" db:"queue_name"`
	MemberId             *int64               `json:"member_id" db:"member_id"`
	MemberName           *string              `json:"member_name,omitempty" db:"member_name"`
	MemberCallId         *string              `json:"member_channel_id" db:"member_channel_id"`
	AgentCallId          *string              `json:"agent_channel_id" db:"agent_channel_id"`
	Destination          *MemberCommunication `json:"communication" db:"destination"`
	LeavingAt            *int64               `json:"leaving_at" db:"leaving_at"`
	ProcessingTimeoutAt  *int64               `json:"processing_timeout_at" db:"processing_timeout_at"`
	ProcessingSec        *int                 `json:"processing_sec" db:"processing_sec"`
	ProcessingRenewalSec *int                 `json:"processing_renewal_sec" db:"processing_renewal_sec"`
	Form                 *json.RawMessage     `json:"form" db:"form"`
	HasForm              bool                 `json:"has_form,omitempty" db:"has_form,omitempty"`
	State                *string              `json:"state,omitempty" db:"state"`
	Variables            *Variables           `json:"variables,omitempty" db:"variables"`
	BridgedAt            *int64               `json:"bridged_at,omitempty" db:"bridged_at"`
	AgentId              *int                 `json:"agent_id,omitempty" db:"agent_id"`
}

type CallAnnotation struct {
	Id        int64     `json:"id" db:"id"`
	CallId    string    `json:"call_id,omitempty" db:"call_id"`
	CreatedBy *Lookup   `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedBy *Lookup   `json:"updated_by" db:"updated_by"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Note      string    `json:"note" db:"note"`
	StartSec  int32     `json:"start_sec" db:"start_sec"`
	EndSec    int32     `json:"end_sec" db:"end_sec"`
}

func (c *Call) MarshalJSON() ([]byte, error) {
	type Alias Call
	return json.Marshal(&struct {
		*Alias
		CreatedAt  int64 `json:"created_at" db:"created_at"`
		AnsweredAt int64 `json:"answered_at" db:"answered_at"`
		BridgedAt  int64 `json:"bridged_at" db:"bridged_at"`
		HangupAt   int64 `json:"hangup_at" db:"hangup_at"`

		JoinedAt       int64 `json:"joined_at" db:"joined_at"`
		LeavingAt      int64 `json:"leaving_at" db:"leaving_at"`
		ReportingAt    int64 `json:"reporting_at" db:"reporting_at"`
		QueueBridgedAt int64 `json:"queue_bridged_at" db:"queue_bridged_at"`
	}{
		Alias:      (*Alias)(c),
		CreatedAt:  TimeToInt64(&c.CreatedAt),
		AnsweredAt: TimeToInt64(c.AnsweredAt),
		BridgedAt:  TimeToInt64(c.BridgedAt),
		HangupAt:   TimeToInt64(c.HangupAt),

		JoinedAt:       TimeToInt64(c.JoinedAt),
		LeavingAt:      TimeToInt64(c.LeavingAt),
		ReportingAt:    TimeToInt64(c.ReportingAt),
		QueueBridgedAt: TimeToInt64(c.QueueBridgedAt),
	})
}

func (c Call) AllowFields() []string {
	return c.DefaultFields()
}

func (c Call) DefaultOrder() string {
	return "-created_at"
}

func (c Call) DefaultFields() []string {
	return []string{"id", "app_id", "state", "timestamp", "parent_id", "user", "extension", "gateway", "direction", "destination", "from", "to", "variables",
		"created_at", "answered_at", "bridged_at", "hangup_at", "duration", "hold_sec", "wait_sec", "bill_sec",
		"queue", "member", "team", "agent", "joined_at", "leaving_at", "reporting_at", "queue_bridged_at",
		"queue_wait_sec", "queue_duration_sec", "reporting_sec", "display", "supervisor", "blind_transfer",
	}
}

func (c Call) EntityName() string {
	return "cc_call_active_list"
}

//todo
func (a *CallAnnotation) IsValid() *AppError {
	return nil
}

type CallFileTranscriptLookup struct {
	Id     int64   `json:"id"`
	Locale string  `json:"locale"`
	FileId int64   `json:"file_id"`
	File   *Lookup `json:"file"`
}

type CallFile struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	StartAt  int64  `json:"start_at"`
	StopAt   int64  `json:"stop_at"`
}

func TimeToInt64(t *time.Time) int64 {
	if t == nil {
		return 0
	}

	return t.UnixNano() / int64(time.Millisecond)
}

func Int64ToTime(i int64) *time.Time {
	if i == 0 {
		return nil
	}

	t := time.Unix(0, i*int64(time.Millisecond))
	return &t
}

type Variables map[string]interface{}

type HistoryFileJob struct {
	Id        int64   `json:"id"`
	FileId    int64   `json:"file_id"`
	CreatedAt int64   `json:"created_at"`
	Action    string  `json:"action"`
	State     int32   `json:"state"`
	Error     *string `json:"error"`
}

type HistoryCall struct {
	Id          string     `json:"id" db:"id"`
	AppId       string     `json:"app_id" db:"app_id"`
	Type        string     `json:"type" db:"type"`
	ParentId    *string    `json:"parent_id" db:"parent_id"`
	User        *Lookup    `json:"user" db:"user"`
	Extension   *string    `json:"extension" db:"extension"`
	Gateway     *Lookup    `json:"gateway" db:"gateway"`
	Direction   string     `json:"direction" db:"direction"`
	Destination string     `json:"destination" db:"destination"`
	From        *Endpoint  `json:"from" db:"from"`
	To          *Endpoint  `json:"to" db:"to"`
	Variables   *Variables `json:"variables" db:"variables"`

	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	AnsweredAt *time.Time `json:"answered_at" db:"answered_at"`
	BridgedAt  *time.Time `json:"bridged_at" db:"bridged_at"`
	HangupAt   *time.Time `json:"hangup_at" db:"hangup_at"`
	StoredAt   *time.Time `json:"stored_at" db:"stored_at"`
	HangupBy   string     `json:"hangup_by" db:"hangup_by"`
	Cause      string     `json:"cause" db:"cause"`

	Duration int `json:"duration" db:"duration"`
	HoldSec  int `json:"hold_sec" db:"hold_sec"`
	WaitSec  int `json:"wait_sec" db:"wait_sec"`
	BillSec  int `json:"bill_sec" db:"bill_sec"`

	SipCode     *int                        `json:"sip_code" db:"sip_code"`
	Files       []*CallFile                 `json:"files" db:"files"`
	Transcripts []*CallFileTranscriptLookup `json:"transcripts" db:"transcripts"`

	Queue  *Lookup `json:"queue" db:"queue"`
	Member *Lookup `json:"member" db:"member"`
	Team   *Lookup `json:"team" db:"team"`
	Agent  *Lookup `json:"agent" db:"agent"`

	JoinedAt         *time.Time  `json:"joined_at" db:"joined_at"`
	LeavingAt        *time.Time  `json:"leaving_at" db:"leaving_at"`
	ReportingAt      *time.Time  `json:"reporting_at" db:"reporting_at"`
	QueueBridgedAt   *time.Time  `json:"queue_bridged_at" db:"queue_bridged_at"`
	QueueWaitSec     *int        `json:"queue_wait_sec" db:"queue_wait_sec"`
	QueueDurationSec *int        `json:"queue_duration_sec" db:"queue_duration_sec"`
	ReportingSec     *int        `json:"reporting_sec" db:"reporting_sec"`
	Result           *string     `json:"result" db:"result"`
	Tags             StringArray `json:"tags" db:"tags"`
	Display          *string     `json:"display" db:"display"`
	TransferFrom     *string     `json:"transfer_from" db:"transfer_from"`
	TransferTo       *string     `json:"transfer_to" db:"transfer_to"`
	HasChildren      bool        `json:"exists_parent" db:"has_children"`
	AgentDescription *string     `json:"agent_description" db:"agent_description"`

	Hold              []*CallHold       `json:"hold" db:"hold"`
	Annotations       []*CallAnnotation `json:"annotations" db:"annotations"`
	AmdResult         *string           `json:"amd_result" db:"amd_result"`
	AmdDurationSec    uint32            `json:"amd_duration" db:"amd_duration"`
	HangupDisposition *string           `json:"hangup_disposition" db:"hangup_disposition"`
	BlindTransfer     *string           `json:"blind_transfer" db:"blind_transfer"`

	FilesJob []*HistoryFileJob `json:"files_job" db:"files_job"`
	TalkSec  int32             `json:"talk_sec" db:"talk_sec"`
}

func (c HistoryCall) DefaultOrder() string {
	return "-created_at"
}

func (c HistoryCall) AllowFields() []string {
	return []string{"id", "app_id", "parent_id", "user", "extension", "gateway", "direction", "destination", "from", "to", "variables",
		"created_at", "answered_at", "bridged_at", "hangup_at", "stored_at", "hangup_by", "cause", "duration", "hold_sec", "wait_sec", "bill_sec",
		"sip_code", "files", "queue", "member", "team", "agent", "joined_at", "leaving_at", "reporting_at", "queue_bridged_at",
		"queue_wait_sec", "queue_duration_sec", "result", "reporting_sec", "tags", "display", "transfer_from", "transfer_to", "has_children",
		"agent_description", "hold", "annotations", "amd_result", "amd_duration", "hangup_disposition", "blind_transfer", "files_job", "transcripts", "talk_sec",
	}
}

func (c HistoryCall) DefaultFields() []string {
	return []string{"id", "app_id", "parent_id", "user", "extension", "gateway", "direction", "destination", "from", "to", "variables",
		"created_at", "answered_at", "bridged_at", "hangup_at", "stored_at", "hangup_by", "cause", "duration", "hold_sec", "wait_sec", "bill_sec",
		"sip_code", "files", "queue", "member", "team", "agent", "joined_at", "leaving_at", "reporting_at", "queue_bridged_at",
		"queue_wait_sec", "queue_duration_sec", "result", "reporting_sec", "tags", "display", "agent_description", "amd_result",
	}
}

func (c HistoryCall) EntityName() string {
	return "cc_calls_history_list"
}

func (c *HistoryCall) GetResult() string {
	if c.Result != nil {
		return *c.Result
	}

	return ""
}

type SearchCall struct {
	ListRequest
	CreatedAt     *FilterBetween
	Duration      *FilterBetween
	AnsweredAt    *FilterBetween
	Number        *string
	ParentId      *string
	Direction     []string
	Missed        *bool
	SkipParent    bool
	HasFile       bool
	UserIds       []int64
	QueueIds      []int64
	TeamIds       []int64
	AgentIds      []int64
	MemberIds     []int64
	GatewayIds    []int64
	SupervisorIds []int64
	State         []string
}

type SearchHistoryCall struct {
	ListRequest
	CreatedAt        *FilterBetween
	Duration         *FilterBetween
	AnsweredAt       *FilterBetween
	StoredAt         *FilterBetween
	Number           string
	ParentId         *string
	Cause            *string
	CauseArr         []string // fixme
	Direction        *string
	Directions       []string //fixme
	Missed           *bool
	SkipParent       bool
	UserIds          []int64
	QueueIds         []int64
	TeamIds          []int64
	AgentIds         []int64
	MemberIds        []int64
	GatewayIds       []int64
	Ids              []string
	TransferFromIds  []string
	TransferToIds    []string
	DependencyIds    []string
	Tags             []string
	Variables        StringMap
	AmdResult        []string
	HasFile          *bool
	HasTranscript    *bool
	Fts              *string
	AgentDescription string
	OwnerIds         []int64
}

type CallEventInfo struct {
	Id        string  `json:"id" db:"id"`
	Event     string  `json:"event" db:"-"`
	Timestamp float64 `json:"timestamp,string" db:"timestamp"`
	DomainId  string  `json:"domain_id" db:"domain_id"`
	UserId    string  `json:"user_id,omitempty" db:"user_id"`
	AppId     string  `json:"app_id,omitempty" db:"app_id"`
}

type CallEvent struct {
	CallEventInfo
	//CCAppId   string      `json:"cc_app_id,omitempty"`
	Body CallPayload `json:"data,string,omitempty" db:"-"`
}

type CallServiceHangup struct {
	Subclass string `json:"Event-Subclass" db:"-"`
	CCAppId  string `json:"cc_app_id,omitempty" db:"cc_app_id"`
	CallEventInfo
	Data string `json:"data" db:"-"`
}

type AggregateGroup struct {
	Id       string
	Interval string // sec

	Aggregate string
	Field     string
	Top       int32
	Desc      bool
}

type AggregateMetrics struct {
	Min   []string `json:"min"`
	Max   []string `json:"max"`
	Avg   []string `json:"avg"`
	Sum   []string `json:"sum"`
	Count []string `json:"count"`
}

type Aggregate struct {
	Name     string           `json:"name"`
	Relative bool             `json:"relative"` // %
	Group    []AggregateGroup `json:"group"`
	AggregateMetrics
	Limit int32    `json:"limit"`
	Sort  []string `json:"sort"`
}

type CallAggregate struct {
	SearchHistoryCall
	Aggs []Aggregate
}

type AggregateData []byte

type AggregateResult struct {
	Name string `json:"name" db:"name"`
	Data []byte `json:"data" db:"data"`
}

type CallPayload map[string]interface{}

func (cp CallPayload) MarshalJSON() ([]byte, error) {
	return json.Marshal((*map[string]interface{})(&cp))
}

func (e *CallServiceHangup) MarshalJSON() []byte {
	d, _ := json.Marshal(e)
	return d
}

func (cp *CallPayload) UnmarshalText(b []byte) error {
	return json.Unmarshal(b, (*map[string]interface{})(cp))
}

func (cr *CallRequest) AddUserVariable(name, value string) {
	cr.AddVariable(fmt.Sprintf("usr_%s", name), value)
}

func (cr *CallRequest) AddVariable(name, value string) {
	if cr.Variables == nil {
		cr.Variables = make(map[string]string)
	}
	cr.Variables[name] = value
}

func NewWebSocketCallEvent(call *CallEvent) *WebSocketEvent {
	e := NewWebSocketEvent(WEBSOCKET_EVENT_CALL)
	e.Add("call", call)

	return e
}

package b2bua

import (
	"errors"
	"fmt"
	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
	"github.com/webitel/engine/b2bua/account"
	"github.com/webitel/engine/b2bua/call"
	"github.com/webitel/engine/b2bua/session"
	"github.com/webitel/engine/b2bua/stack"
	"github.com/webitel/engine/b2bua/ua"
	"github.com/webitel/engine/b2bua/utils"
	"strconv"
	"sync"
)

var (
	logger log.Logger
)

const (
	transport = "udp"
)

type B2B struct {
	transport string
	host      string

	stack *stack.SipStack
	ua    *ua.UserAgent

	accounts map[int]*Account
	sync.RWMutex
	calls []*call.Call

	cb OnCallback
}

func init() {
	logger = utils.NewLogrusLogger(log.ErrorLevel, "Register", nil)
}

type SdpDescription struct {
	Type string `json:"type"`
	Sdp  string `json:"sdp"`
}

type OnCallback interface {
	OnB2B(sipId string, sdp SdpDescription)
}

func (b2b *B2B) AddAccount(id int, acc *Account) {
	b2b.Lock()
	b2b.accounts[id] = acc // TODO
	b2b.Unlock()
}

func (b2b *B2B) GetAccount(id int) (*Account, bool) {
	b2b.RLock()
	acc, ok := b2b.accounts[id]
	b2b.RUnlock()

	return acc, ok
}

func New(cb OnCallback) *B2B {
	st := stack.NewSipStack(&stack.SipStackConfig{
		UserAgent:  "webitel-webrtc",
		Extensions: []string{"replaces", "outbound"},
		Host:       "10.9.8.111",
		//Dns:        "8.8.8.8",
	})
	//utils.SetLogLevel("transport.Layer", 3)
	//utils.SetLogLevel("transaction.Layer", 3)

	if err := st.Listen(transport, "0.0.0.0:5067"); err != nil {
		logger.Panic(err)
	}

	ua := ua.NewUserAgent(&ua.UserAgentConfig{
		SipStack: st,
	})

	b2b := &B2B{
		transport: transport,
		host:      "10.9.8.111",
		stack:     st,
		ua:        ua,
		accounts:  make(map[int]*Account),
		cb:        cb,
	}

	//utils.SetLogLevel("UserAgent", 3)

	ua.RegisterStateHandler = func(state account.RegisterState) {
		logger.Infof("RegisterStateHandler: user => %s%s, state => %v, expires => %v, reason => %v", state.Account.AuthInfo.AuthUser,
			state.Account.AuthInfo.Realm, state.StatusCode, state.Expiration, state.Reason)
	}

	ua.InviteStateHandler = b2b.inviteStateHandler

	_ = st.OnRequest(sip.OPTIONS, b2b.handleOptions)

	return b2b
}

/*
AuthInfo{
		DisplayName: "Igor",
		Expires:     120,
		AuthInfo: account.AuthInfo{
			AuthUser: "1008",
			Password: "tWKtntyTB1UV",
			Realm:    "demo.webitel.com",
		},
	}
*/

func (b2b *B2B) Register(userId int, conf AuthInfo) error {

	if _, ok := b2b.GetAccount(userId); ok {
		return errors.New("is registered")
	}

	acc, err := b2b.NewAccount(conf)
	if err != nil {
		return err
	}

	err = acc.Register()
	if err != nil {
		return err
	}

	b2b.AddAccount(userId, acc)
	return nil
}

func (b2b *B2B) inviteStateHandler(sess *session.Session, req *sip.Request, resp *sip.Response, state session.Status) {
	logger.Infof("InviteStateHandler: state => %v, type => %s", state, sess.Direction())

	switch state {
	case session.InviteReceived:
		userId := findCustomHeaderValue("X-Webitel-User-Id", *req)
		wId := findCustomHeaderValue("X-Webitel-Uuid", *req)
		call := &call.Call{
			Src:     sess,
			Dest:    nil,
			WCallId: wId,
			Req:     req,
		}
		call.UserId, _ = strconv.Atoi(userId)
		b2b.appendCall(call)

		sess.Provisional(100, "Trying")
		sess.Provisional(180, "Ringing")

		//sess.Reject(403, "TODO")

	case session.InviteSent:
		(*req).AppendHeader(&sip.GenericHeader{
			HeaderName: "X-Webitel-Test",
			Contents:   "true",
		})
		fmt.Println("FIXME")
		//b2b.cb.OnB2B(sess.CallID().Value(), SdpDescription{
		//	Type: "answer",
		//	Sdp:  sess.RemoteSdp(),
		//})
	case session.Confirmed:
		b2b.maybeSendSdp(sess)
		//TODO: Add support for forked calls
		//c := b2b.findCall(sess)
		//if c != nil && c.Dest == sess {
		//	answer := c.Dest.RemoteSdp()
		//	c.Src.ProvideAnswer(answer)
		//	c.Src.Accept(200)
		//}

	case session.ReInviteReceived:
		logger.Infof("re-INVITE")
		switch sess.Direction() {
		case session.Incoming:
			sess.Accept(200)
		case session.Outgoing:
			//TODO: Need to provide correct answer.
		}

	// Handle 4XX+
	case session.Canceled:
		fallthrough
	case session.Failure:
		fallthrough
	case session.Terminated:
		c := b2b.findCall(sess)
		if c != nil {
			if c.Src != nil {
				c.Src.End()
			}
			if c.Dest != nil {
				c.Dest.End()
			}
		}
		b2b.removeCall(sess)
	case session.EarlyMedia:
		b2b.maybeSendSdp(sess)
	default:
		fmt.Println("call state ", state, " sdp ", sess.RemoteSdp() != "")
	}
}

func (b2b *B2B) Dial(userId int, sdp string, destination string) (string, error) {

	acc, ok := b2b.GetAccount(userId)
	if !ok {
		return "", errors.New("not found account")
	}

	to, err := parser.ParseSipUri(fmt.Sprintf("sip:%s@%s", destination, b2b.host))
	if err != nil {
		return "", err
	}

	c := &call.Call{
		UserId: userId,
	}

	b2b.appendCall(c)

	if c.Src, err = b2b.ua.Invite(acc.profile, &to, to, &sdp); err != nil {
		return "", err
	}

	return c.Src.CallID().Value(), nil
}

func (b2b *B2B) Recovery(userId int, sipId string, sdp string) (string, error) {
	acc, ok := b2b.GetAccount(userId)
	if !ok {
		return "", errors.New("not found account")
	}

	for _, v := range b2b.calls {
		if v.Src != nil && v.Src.CallID().Value() == sipId {
			err := b2b.ua.Recovery(acc.profile, acc.recipient, v.Src, &sdp)
			if err != nil {
				return "", err
			}

			return v.Src.RemoteSdp(), nil
		}

		if v.Dest != nil && v.Dest.CallID().Value() == sipId {
			err := b2b.ua.Recovery(acc.profile, acc.recipient, v.Dest, &sdp)
			if err != nil {
				return "", err
			}
			return v.Dest.RemoteSdp(), nil
		}
	}

	return "", nil
}

func (b2b *B2B) Answer(userId int, wid string, sdp string) (string, error) {
	//acc, ok := b2b.GetAccount(userId)
	//if !ok {
	//	return "", errors.New("not found account")
	//}

	var call = b2b.findCallByWId(wid)
	if call == nil {
		return "", errors.New("not found call")
	}
	//req := call.Req
	//to, _ := (*req).To()
	//from, _ := (*req).From()
	//caller := from.Address
	//called := to.Address
	//
	//displayName := ""
	//if from.DisplayName != nil {
	//	displayName = from.DisplayName.String()
	//}
	//// Create a temporary profile. In the future, it will support reading profiles from files or data
	//// For example: use a specific ip or sip account as outbound trunk
	//profile := account.NewProfile(caller, displayName, nil, 0, b2b.stack)
	//
	//fmt.Println(acc.profile.Contact())
	////s := "sip:" + called.User().String() + "@" + "10.9.8.111" + ";transport=" + "udp"
	//recipient, err2 := parser.ParseSipUri(profile.URI.String())
	//if err2 != nil {
	//	logger.Error(err2)
	//}
	//
	////offer := call.Src.RemoteSdp()
	//dest, err := b2b.ua.Invite(profile, called, recipient, &sdp)
	//if err != nil {
	//	return "", err
	//}
	//
	//call.Dest = dest

	call.Src.ProvideAnswer(sdp)
	call.Src.Provisional(200, "OK")

	return call.Src.RemoteSdp(), nil
}

func (b2b *B2B) RemoteSdp(userId int, wid string) (SdpDescription, error) {
	_, ok := b2b.GetAccount(userId)
	if !ok {
		return SdpDescription{}, errors.New("not found account")
	}

	var call = b2b.findCallByWId(wid)
	if call == nil {
		return SdpDescription{}, errors.New("not found call")
	}

	return SdpDescription{
		Type: "offer",
		Sdp:  call.Src.RemoteSdp(),
	}, nil

}

func (b2b *B2B) maybeSendSdp(sess *session.Session) {
	if sess.Direction() == session.Outgoing {
		c := b2b.findCall(sess)
		if c != nil {
			b2b.cb.OnB2B(sess.CallID().Value(), SdpDescription{
				Type: "answer",
				Sdp:  sess.RemoteSdp(),
			})
		}

	}
}

func (b2b *B2B) handleOptions(req sip.Request, tx sip.ServerTransaction) {
	res := sip.NewResponseFromRequest("", req, 200, "I See You", "")
	if _, err := b2b.stack.Respond(res); err != nil {
		logger.Errorf("respond '200 I See You' failed: %s", err)
	}
	return
}

func (b2b *B2B) appendCall(c *call.Call) {
	b2b.Lock()
	b2b.calls = append(b2b.calls, c)
	b2b.Unlock()
}

func (b2b *B2B) removeCall(sess *session.Session) {
	b2b.Lock()
	defer b2b.Unlock()

	for idx, call := range b2b.calls {
		if call.Src == sess || call.Dest == sess {
			b2b.calls = append(b2b.calls[:idx], b2b.calls[idx+1:]...)
			return
		}
	}
}

func (b2b *B2B) findCall(sess *session.Session) *call.Call {
	for _, c := range b2b.calls {
		cid := sess.CallID().String()
		if c.Src != nil && c.Src.CallID().String() == cid {
			return c
		}
		if c.Dest != nil && c.Dest.CallID().String() == cid {
			return c
		}
	}
	return nil
}

func (b2b *B2B) findCallByWId(wid string) *call.Call {
	for _, c := range b2b.calls {
		if c.WCallId == wid {
			return c
		}
	}
	return nil
}

func (b2b *B2B) Stop() {
	b2b.ua.Shutdown()
}

func findCustomHeaderValue(name string, req sip.Request) string {
	for _, h := range req.Headers() {
		switch v := h.(type) {
		case *sip.GenericHeader:
			if v.HeaderName == name {
				return v.Value()
			}
		}
	}

	return ""
}

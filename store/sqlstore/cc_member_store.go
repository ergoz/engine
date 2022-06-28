package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	"github.com/lib/pq"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
	"net/http"
	"strings"
)

type SqlMemberStore struct {
	SqlStore
}

func NewSqlMemberStore(sqlStore SqlStore) store.MemberStore {
	us := &SqlMemberStore{sqlStore}
	return us
}

func (s SqlMemberStore) Create(domainId int64, member *model.Member) (*model.Member, *model.AppError) {
	var out *model.Member
	if err := s.GetMaster().SelectOne(&out, `with m as (
			insert into call_center.cc_member (queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, domain_id, agent_id, skill_id)
			values (:QueueId, :Priority, :ExpireAt, :Variables, :Name, :TimezoneId, :Communications, :BucketId, :MinOfferingAt, :DomainId, :AgentId, :SkillId)
			returning *
		)
		select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(agn.id, agn.name::text) as agent, call_center.cc_get_lookup(cs.id, cs.name::text) as skill
		from m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_skill cs on m.skill_id = cs.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id`,
		map[string]interface{}{
			"DomainId":       domainId,
			"QueueId":        member.QueueId,
			"Priority":       member.Priority,
			"ExpireAt":       member.ExpireAt,
			"Variables":      member.Variables.ToJson(),
			"Name":           member.Name,
			"TimezoneId":     member.Timezone.Id,
			"Communications": member.ToJsonCommunications(),
			"BucketId":       member.Bucket.GetSafeId(),
			"MinOfferingAt":  member.MinOfferingAt,
			"AgentId":        member.Agent.GetSafeId(),
			"SkillId":        member.Skill.GetSafeId(),
		}); nil != err {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.save.app_error", nil,
			fmt.Sprintf("name=%v, %v", member.Name, err.Error()), extractCodeFromErr(err))
	} else {
		return out, nil
	}
}

func (s SqlMemberStore) BulkCreate(domainId, queueId int64, members []*model.Member) ([]int64, *model.AppError) {
	var err error
	var stmp *sql.Stmt
	var tx *gorp.Transaction
	tx, err = s.GetMaster().Begin()
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	_, err = tx.Exec("CREATE temp table cc_member_tmp ON COMMIT DROP as table call_center.cc_member with no data")
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	stmp, err = tx.Prepare(pq.CopyIn("cc_member_tmp", "id", "queue_id", "priority", "expire_at", "variables", "name",
		"timezone_id", "communications", "bucket_id", "ready_at", "agent_id", "skill_id"))
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	defer stmp.Close()
	result := make([]int64, 0, len(members))
	for k, v := range members {
		_, err = stmp.Exec(k, queueId, v.Priority, v.ExpireAt, v.Variables.ToJson(), v.Name, v.Timezone.Id, v.ToJsonCommunications(),
			v.Bucket.GetSafeId(), v.MinOfferingAt, v.Agent.GetSafeId(), v.Skill.GetSafeId())
		if err != nil {
			goto _error
		}
	}

	_, err = stmp.Exec()
	if err != nil {
		goto _error
	} else {

		_, err = tx.Select(&result, `with i as (
			insert into call_center.cc_member(queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, domain_id, agent_id, skill_id)
			select queue_id, priority, expire_at, variables, name, timezone_id, communications, bucket_id, ready_at, :DomainId, agent_id, skill_id
			from cc_member_tmp
			returning id
		)
		select id from i`, map[string]interface{}{
			"DomainId": domainId,
		})
		if err != nil {
			goto _error
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), extractCodeFromErr(err))
	}

	return result, nil

_error:
	tx.Rollback()
	return nil, model.NewAppError("SqlMemberStore.Save", "store.sql_member.bulk_save.app_error", nil, err.Error(), extractCodeFromErr(err))
}

// todo fix deprecated fields

func (s SqlMemberStore) SearchMembers(domainId int64, search *model.SearchMemberRequest) ([]*model.Member, *model.AppError) {
	var members []*model.Member

	order := GetOrderBy("cc_member", model.MemberDeprecatedField(search.Sort))
	if order == "" {
		order = "order by id desc"
	}

	fields := GetFields(model.MemberDeprecatedFields(search.Fields), model.Member{})

	if _, err := s.GetReplica().Select(&members,
		`with comm as (select c.id, json_build_object('id', c.id, 'name', c.name)::jsonb j
              from call_center.cc_communication c
              where c.domain_id = :Domain)
   , resources as (select r.id, json_build_object('id', r.id, 'name', r.name)::jsonb j
                   from call_center.cc_outbound_resource r
                   where r.domain_id = :Domain)
   , result as (select m.id
                from call_center.cc_member m
                where m.domain_id = :Domain
                  and (:Ids::int8[] isnull or m.id = any (:Ids::int8[]))
                  and (:QueueIds::int4[] isnull or m.queue_id = any (:QueueIds::int4[]))
                  and (:BucketIds::int4[] isnull or m.bucket_id = any (:BucketIds::int4[]))
                  and (:Destination::varchar isnull or
                       m.search_destinations && array [:Destination::varchar]::varchar[])

                  and (:CreatedFrom::timestamptz isnull or m.created_at >= :CreatedFrom::timestamptz)
                  and (:CreatedTo::timestamptz isnull or created_at <= :CreatedTo::timestamptz)

                  and (:OfferingFrom::timestamptz isnull or m.ready_at >= :OfferingFrom::timestamptz)
                  and (:OfferingTo::timestamptz isnull or m.ready_at <= :OfferingTo::timestamptz)

                  and (:PriorityFrom::int isnull or m.priority >= :PriorityFrom::int)
                  and (:PriorityTo::int isnull or m.priority <= :PriorityTo::int)
                  and (:AttemptsFrom::int isnull or m.attempts >= :AttemptsFrom::int)
                  and (:AttemptsTo::int isnull or m.attempts <= :AttemptsTo::int)

				  and (:AgentIds::int4[] isnull or m.agent_id = any(:AgentIds::int4[]))

                  and (:StopCauses::varchar[] isnull or m.stop_cause = any (:StopCauses::varchar[]))
                  and (:Name::varchar isnull or m.name ilike :Name::varchar)
                  and (:Q::varchar isnull or
                       (m.name ~~ :Q::varchar or m.search_destinations && array [rtrim(:Q::varchar, '%')]::varchar[]))
				`+order+`
                limit :Limit offset :Offset)
	, list as (
		select m.id,
			   call_center.cc_member_destination_views_to_json(array(select (xid::int2, x ->> 'destination',
																			 resources.j,
																			 comm.j,
																			 (x -> 'priority')::int,
																			 (x -> 'state')::int,
																			 x -> 'description',
																			 (x -> 'last_activity_at')::int8,
																			 (x -> 'attempts')::int,
																			 x ->> 'last_cause',
																			 x ->>
																			 'display')::call_center.cc_member_destination_view
																	 from jsonb_array_elements(m.communications) with ordinality as x (x, xid)
																			  left join comm on comm.id = (x -> 'type' -> 'id')::int
																			  left join resources on resources.id = (x -> 'resource' -> 'id')::int)) communications,
			   call_center.cc_get_lookup(cq.id, cq.name::varchar)                                                                                    queue,
			   m.priority,
			   m.expire_at,
			   m.created_at,
			   m.variables,
			   m.name,
			   call_center.cc_get_lookup(m.timezone_id::bigint,
										 ct.name::varchar)                                                                                           "timezone",
			   call_center.cc_get_lookup(m.bucket_id, cb.name::varchar)                                                                              bucket,
			   m.ready_at as ready_at,
			   m.stop_cause,
			   m.stop_at,
			   m.last_hangup_at as last_hangup_at,
			   m.attempts,
			   call_center.cc_get_lookup(agn.id, agn.name::varchar)                                                                                  agent,
			   call_center.cc_get_lookup(cs.id, cs.name::varchar)                                                                                    skill,
			   exists(select 1 from call_center.cc_member_attempt a where a.member_id = m.id) as                                                     reserved
		from call_center.cc_member m
				 inner join result on m.id = result.id
				 inner join call_center.cc_queue cq on m.queue_id = cq.id
				 left join flow.calendar_timezones ct on ct.id = m.timezone_id
				 left join call_center.cc_agent_list agn on m.agent_id = agn.id
				 left join call_center.cc_bucket cb on m.bucket_id = cb.id
				 left join call_center.cc_skill cs on m.skill_id = cs.id
	)
	select `+strings.Join(fields, " ,")+` from list`, map[string]interface{}{
			"Domain": domainId,
			"Limit":  search.GetLimit(),
			"Offset": search.GetOffset(),
			"Q":      search.GetRegExpQ(),

			"Ids":         pq.Array(search.Ids),
			"QueueIds":    pq.Array(search.QueueIds),
			"BucketIds":   pq.Array(search.BucketIds),
			"AgentIds":    pq.Array(search.AgentIds),
			"Destination": search.Destination,

			"CreatedFrom":  model.GetBetweenFromTime(search.CreatedAt),
			"CreatedTo":    model.GetBetweenToTime(search.CreatedAt),
			"OfferingFrom": model.GetBetweenFromTime(search.OfferingAt),
			"OfferingTo":   model.GetBetweenToTime(search.OfferingAt),

			"PriorityFrom": model.GetBetweenFrom(search.Priority),
			"PriorityTo":   model.GetBetweenTo(search.Priority),
			"AttemptsFrom": model.GetBetweenFrom(search.Attempts),
			"AttemptsTo":   model.GetBetweenTo(search.Attempts),

			"StopCauses": pq.Array(search.StopCauses),
			"Name":       search.Name,
		}); err != nil {
		return nil, model.NewAppError("SqlMemberStore.GetAllPage", "store.sql_member.get_all.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return members, nil
	}
}

func (s SqlMemberStore) Get(domainId, queueId, id int64) (*model.Member, *model.AppError) {
	var member *model.Member
	if err := s.GetReplica().SelectOne(&member, `select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(cs.id, cs.name::text) as skill, call_center.cc_get_lookup(agn.id, agn.name::varchar) agent
		from call_center.cc_member m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id
		    left join call_center.cc_skill cs on m.skill_id = cs.id
	where m.id = :Id and m.queue_id = :QueueId and exists(select 1 from call_center.cc_queue q where q.id = :QueueId and q.domain_id = :DomainId)`, map[string]interface{}{
		"Id":       id,
		"DomainId": domainId,
		"QueueId":  queueId,
	}); err != nil {
		return nil, model.NewAppError("SqlMemberStore.Get", "store.sql_member.get.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	} else {
		return member, nil
	}
}

func (s SqlMemberStore) Update(domainId int64, member *model.Member) (*model.Member, *model.AppError) {
	err := s.GetMaster().SelectOne(&member, `with m as (
    update call_center.cc_member m1
        set priority = :Priority,
            expire_at = :ExpireAt,
            variables = :Variables,
            name = :Name,
            timezone_id = :TimezoneId,
            communications = :Communications,
            bucket_id = :BucketId,
			ready_at = :MinOfferingAt,
			stop_cause = :StopCause::varchar,
			agent_id = :AgentId,
			skill_id = :SkillId,
			stop_at = case when :StopCause::varchar notnull then now() else stop_at end
    where m1.id = :Id and m1.queue_id = :QueueId
    returning *
)
select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(cs.id, cs.name::text) as skill, call_center.cc_get_lookup(agn.id, agn.name::varchar) agent
		from m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_skill cs on m.skill_id = cs.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id`, map[string]interface{}{
		"Priority":       member.Priority,
		"ExpireAt":       member.ExpireAt,
		"Variables":      member.Variables.ToJson(),
		"Name":           member.Name,
		"TimezoneId":     member.Timezone.Id,
		"Communications": member.ToJsonCommunications(),
		"BucketId":       member.Bucket.GetSafeId(),
		"Id":             member.Id,
		"QueueId":        member.QueueId,
		"DomainId":       domainId,
		"MinOfferingAt":  member.MinOfferingAt,
		"StopCause":      member.StopCause,
		"AgentId":        member.Agent.GetSafeId(),
		"SkillId":        member.Skill.GetSafeId(),
	})
	if err != nil {
		code := extractCodeFromErr(err)
		if code == http.StatusNotFound { //todo
			return nil, model.NewAppError("SqlMemberStore.Update", "store.sql_member.update.lock", nil,
				fmt.Sprintf("Id=%v, %s", member.Id, err.Error()), http.StatusBadRequest)
		}

		return nil, model.NewAppError("SqlMemberStore.Update", "store.sql_member.update.app_error", nil,
			fmt.Sprintf("Id=%v, %s", member.Id, err.Error()), code)
	}
	return member, nil
}

//TODO add force
func (s SqlMemberStore) Delete(queueId, id int64) *model.AppError {
	var cnt int64
	res, err := s.GetMaster().Exec(`delete
from call_center.cc_member c
where c.id = :Id
  and c.queue_id = :QueueId
  and not exists(select 1 from call_center.cc_member_attempt a where a.member_id = c.id and a.state != 'leaving' for update)`,
		map[string]interface{}{"Id": id, "QueueId": queueId})

	if err != nil {
		return model.NewAppError("SqlMemberStore.Delete", "store.sql_member.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	cnt, err = res.RowsAffected()
	if err != nil {
		return model.NewAppError("SqlMemberStore.Delete", "store.sql_member.delete.app_error", nil,
			fmt.Sprintf("Id=%v, %s", id, err.Error()), extractCodeFromErr(err))
	}

	if cnt == 0 {
		return model.NewAppError("SqlMemberStore.Delete", "store.sql_member.delete.app_error", nil,
			fmt.Sprintf("Id=%v, not found", id), http.StatusNotFound)
	}

	return nil
}

func (s SqlMemberStore) MultiDelete(search *model.MultiDeleteMembers) ([]*model.Member, *model.AppError) {
	var res []*model.Member

	_, err := s.GetMaster().Select(&res, `with m as (
    delete from call_center.cc_member m
    where (:Ids::int8[] isnull or m.id = any (:Ids::int8[]))
		  and (:QueueIds::int4[] isnull or m.queue_id = any (:QueueIds::int4[]))
		  and (:BucketIds::int4[] isnull or m.bucket_id = any (:BucketIds::int4[]))
		  and (:Destination::varchar isnull or
			   m.search_destinations && array [:Destination::varchar]::varchar[])

		  and (:CreatedFrom::timestamptz isnull or m.created_at >= :CreatedFrom::timestamptz)
		  and (:CreatedTo::timestamptz isnull or created_at <= :CreatedTo::timestamptz)

		  and (:OfferingFrom::timestamptz isnull or m.ready_at >= :OfferingFrom::timestamptz)
		  and (:OfferingTo::timestamptz isnull or m.ready_at <= :OfferingTo::timestamptz)

		  and (:PriorityFrom::int isnull or m.priority >= :PriorityFrom::int)
		  and (:PriorityTo::int isnull or m.priority <= :PriorityTo::int)
		  and (:AttemptsFrom::int isnull or m.attempts >= :AttemptsFrom::int)
		  and (:AttemptsTo::int isnull or m.attempts <= :AttemptsTo::int)

		  and (:StopCauses::varchar[] isnull or m.stop_cause = any (:StopCauses::varchar[]))
		  and (:Name::varchar isnull or m.name ilike :Name::varchar)
		  and (:Q::varchar isnull or
			   (m.name ~~ :Q::varchar or m.search_destinations && array [rtrim(:Q::varchar, '%')]::varchar[]))

		and (:Numbers::varchar[] isnull or search_destinations && :Numbers::varchar[])		
		and (:Variables::jsonb isnull or variables @> :Variables::jsonb)
		and (:AgentIds::int4[] isnull or m.agent_id = any(:AgentIds::int4[]))
		and not exists(select 1 from call_center.cc_member_attempt a where a.member_id = m.id and a.state != 'leaving' for update)
    returning *
)
select m.id,  m.stop_at, m.stop_cause, m.attempts, m.last_hangup_at, m.created_at, m.queue_id, m.priority, m.expire_at, m.variables, m.name, call_center.cc_get_lookup(ct.id, ct.name) as "timezone",
			   call_center.cc_member_communications(m.communications) as communications,  call_center.cc_get_lookup(qb.id, qb.name::text) as bucket, ready_at,
               call_center.cc_get_lookup(cs.id, cs.name::text) as skill, call_center.cc_get_lookup(agn.id, agn.name::varchar) agent
		from m
			left join flow.calendar_timezones ct on m.timezone_id = ct.id
			left join call_center.cc_bucket qb on m.bucket_id = qb.id
			left join call_center.cc_skill cs on m.skill_id = cs.id
			left join call_center.cc_agent_list agn on m.agent_id = agn.id`, map[string]interface{}{
		"Q":           search.GetQ(),
		"QueueIds":    pq.Array(search.QueueIds),
		"Ids":         pq.Array(search.Ids),
		"BucketIds":   pq.Array(search.BucketIds),
		"Destination": search.Destination,

		"CreatedFrom":  model.GetBetweenFromTime(search.CreatedAt),
		"CreatedTo":    model.GetBetweenToTime(search.CreatedAt),
		"OfferingFrom": model.GetBetweenFromTime(search.OfferingAt),
		"OfferingTo":   model.GetBetweenToTime(search.OfferingAt),

		"PriorityFrom": model.GetBetweenFrom(search.Priority),
		"PriorityTo":   model.GetBetweenTo(search.Priority),
		"AttemptsFrom": model.GetBetweenFrom(search.Attempts),
		"AttemptsTo":   model.GetBetweenTo(search.Attempts),

		"StopCauses": pq.Array(search.StopCauses),
		"Name":       search.Name,

		"AgentIds":  pq.Array(search.AgentIds),
		"Numbers":   pq.Array(search.Numbers),
		"Variables": search.Variables.ToSafeJson(),
	})

	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.MultiDelete", "store.sql_member.multi_delete.app_error", nil,
			fmt.Sprintf("Ids=%v, %s", search.Ids, err.Error()), extractCodeFromErr(err))
	}

	return res, nil
}

func (s SqlMemberStore) ResetMembers(domainId int64, req *model.ResetMembers) (int64, *model.AppError) {
	cnt, err := s.GetMaster().SelectInt(`with upd as (
    update call_center.cc_member m
    set stop_cause = null,
        stop_at = null,
        attempts = 0
    where m.domain_id = :DomainId
        and m.queue_id = :QueueId
        and (stop_at notnull and not stop_cause in ('success', 'cancel', 'terminate', 'no_communications') )
        and (:Ids::int8[] isnull or m.id = any(:Ids::int8[]))
        and (:Numbers::varchar[] isnull or search_destinations && :Numbers::varchar[])
        and (:Variables::jsonb isnull or variables @> :Variables::jsonb)
        and (:Buckets::int8[] isnull or m.bucket_id = any(:Buckets::int8[]))
        and (:AgentIds::int4[] isnull or m.agent_id = any(:AgentIds::int4[]))
returning m.id
)
select count(*) cnt
from upd`, map[string]interface{}{
		"DomainId": domainId,
		"Ids":      pq.Array(req.Ids),
		"Buckets":  pq.Array(req.Buckets),
		//"Cause":     pq.Array(req.Causes),
		"AgentIds":  pq.Array(req.AgentIds),
		"Numbers":   pq.Array(req.Numbers),
		"Variables": req.Variables.ToSafeJson(),
		"QueueId":   req.QueueId,
	})

	if err != nil {
		return 0, model.NewAppError("SqlMemberStore.ResetMembers", "store.sql_member.reset.app_error", nil,
			fmt.Sprintf("QueueId=%v, %s", req.QueueId, err.Error()), extractCodeFromErr(err))
	}

	return cnt, nil
}

func (s SqlMemberStore) AttemptsList(memberId int64) ([]*model.MemberAttempt, *model.AppError) {
	var attempts []*model.MemberAttempt
	//FIXME
	if _, err := s.GetReplica().Select(&attempts, `with active as (
    select a.id,
           --a.member_id,
           (extract(EPOCH from a.created_at) * 1000)::int8 as created_at,
           'TODO' as destination,
           a.weight,
           a.originate_at,
           a.answered_at,
           a.bridged_at,
           a.hangup_at,
           call_center.cc_get_lookup(cor.id, cor.name) as resource,
           leg_a_id,
           leg_b_id,
           node_id as node,
           result,
           call_center.cc_get_lookup(u.id, u.name) as agent,
           call_center.cc_get_lookup(cb.id::int8, cb.name::varchar) as bucket,
           logs,
           false as active
    from call_center.cc_member_attempt a
        left join call_center.cc_outbound_resource cor on a.resource_id = cor.id
        left join call_center.cc_agent ca on a.agent_id = ca.id
        left join directory.wbt_user u on u.id = ca.user_id
        left join call_center.cc_bucket cb on a.bucket_id = cb.id
    where a.member_id = :MemberId
    order by a.created_at
), log as (
    select a.id,
          -- a.member_id,
           (extract(EPOCH from a.created_at) * 1000)::int8 as created_at,
           'TODO' as destination,
           a.weight,
           a.originate_at,
           a.answered_at,
           a.bridged_at,
           a.hangup_at,
           call_center.cc_get_lookup(cor.id, cor.name) as resource,
           leg_a_id,
           leg_b_id,
           node_id as node,
           result,
           call_center.cc_get_lookup(u.id, u.name) as agent,
           call_center.cc_get_lookup(cb.id::int8, cb.name::varchar) as bucket,
           logs,
           false as active
    from call_center.cc_member_attempt_log a
        left join call_center.cc_outbound_resource cor on a.resource_id = cor.id
        left join call_center.cc_agent ca on a.agent_id = ca.id
        left join directory.wbt_user u on u.id = ca.user_id
        left join call_center.cc_bucket cb on a.bucket_id = cb.id
    where a.member_id = :MemberId
    order by a.created_at
)
select *
from active a
union all
select *
from log a`, map[string]interface{}{"MemberId": memberId}); err != nil {
		return nil, model.NewAppError("SqlMemberStore.AttemptsList", "store.sql_member.get_attempts_all.app_error", nil,
			fmt.Sprintf("MemberId=%v, %s", memberId, err.Error()), extractCodeFromErr(err))
	}

	return attempts, nil
}

func (s SqlMemberStore) SearchAttemptsHistory(domainId int64, search *model.SearchAttempts) ([]*model.AttemptHistory, *model.AppError) {
	var att []*model.AttemptHistory

	f := map[string]interface{}{
		"Domain":    domainId,
		"Limit":     search.GetLimit(),
		"Offset":    search.GetOffset(),
		"From":      model.GetBetweenFromTime(&search.JoinedAt),
		"To":        model.GetBetweenToTime(&search.JoinedAt),
		"Ids":       pq.Array(search.Ids),
		"QueueIds":  pq.Array(search.QueueIds),
		"BucketIds": pq.Array(search.BucketIds),
		"MemberIds": pq.Array(search.MemberIds),
		"AgentIds":  pq.Array(search.AgentIds),
		"Result":    search.Result,
	}

	err := s.ListQuery(&att, search.ListRequest,
		`domain_id = :Domain
	and joined_at between :From::timestamptz and :To::timestamptz
	and (:Ids::int8[] isnull or id = any(:Ids))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:BucketIds::int8[] isnull or bucket_id = any(:Ids))
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
	and (:Result::varchar isnull or result = :Result )`,
		model.AttemptHistory{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.SearchAttemptsHistory", "store.sql_member.attempts_history.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

func (s SqlMemberStore) SearchAttempts(domainId int64, search *model.SearchAttempts) ([]*model.Attempt, *model.AppError) {
	var att []*model.Attempt

	f := map[string]interface{}{
		"Domain":    domainId,
		"Limit":     search.GetLimit(),
		"Offset":    search.GetOffset(),
		"From":      search.JoinedAt.From,
		"To":        search.JoinedAt.To,
		"Ids":       pq.Array(search.Ids),
		"QueueIds":  pq.Array(search.QueueIds),
		"BucketIds": pq.Array(search.BucketIds),
		"MemberIds": pq.Array(search.MemberIds),
		"AgentIds":  pq.Array(search.AgentIds),
		"Result":    search.Result,
	}

	err := s.ListQuery(&att, search.ListRequest,
		`domain_id = :Domain
	and joined_at_timestamp between to_timestamp( (:From::int8 / 1000)::int8 ) and to_timestamp( (:To::int8 / 1000)::int8 )
	and (:Ids::int8[] isnull or id = any(:Ids))
	and (:QueueIds::int[] isnull or queue_id = any(:QueueIds) )
	and (:BucketIds::int8[] isnull or bucket_id = any(:Ids))
	and (:MemberIds::int8[] isnull or member_id = any(:MemberIds) )
	and (:AgentIds::int[] isnull or agent_id = any(:AgentIds) )
	and (:Result::varchar isnull or result = :Result )`,
		model.Attempt{}, f)
	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.SearchAttempts", "store.sql_member.attempts.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

func (s SqlMemberStore) ListOfflineQueueForAgent(domainId int64, search *model.SearchOfflineQueueMembers) ([]*model.OfflineMember, *model.AppError) {
	var att []*model.OfflineMember
	_, err := s.GetReplica().Select(&att, `with comm as (
    select c.id, json_build_object('id', c.id, 'name',  c.name)::jsonb j
    from call_center.cc_communication c
    where c.domain_id = :Domain
)

,resources as (
    select r.id, json_build_object('id', r.id, 'name',  r.name)::jsonb j
    from call_center.cc_outbound_resource r
    where r.domain_id = :Domain
)
, result as (
	select m.id
    from call_center.cc_member m
        inner join call_center.cc_queue cq2 on m.queue_id = cq2.id
        inner join call_center.cc_agent a on a.id = :AgentId
    where m.domain_id = :Domain and cq2.type = 0 and cq2.enabled and (:Q::varchar isnull or m.name ilike :Q)
        and not exists (select 1 from call_center.cc_member_attempt a where a.member_id = m.id)
        and m.stop_at isnull
        and (cq2.team_id isnull or a.team_id = cq2.team_id)
		and m.agent_id isnull
		and m.skill_id isnull
		and (m.expire_at isnull or m.expire_at > now())
		and (m.ready_at isnull or m.ready_at < now())
        and m.queue_id in (
            select distinct cqs.queue_id
            from call_center.cc_skill_in_agent sa
                inner join call_center.cc_queue_skill cqs on cqs.skill_id = sa.skill_id and sa.capacity between cqs.min_capacity and cqs.max_capacity
            where  sa.enabled and cqs.enabled and sa.agent_id = :AgentId
        )
    order by cq2.priority desc , m.priority desc, m.ready_at, m.created_at
    limit :Limit
    offset :Offset
)
select m.id, call_center.cc_member_destination_views_to_json(array(select ( xid::int2, x ->> 'destination',
								resources.j,
                                comm.j,
                                (x -> 'priority')::int ,
                                (x -> 'state')::int  ,
                                x -> 'description'  ,
                                (x -> 'last_activity_at')::int8,
                                (x -> 'attempts')::int,
                                x ->> 'last_cause',
                                x ->> 'display'    )::call_center.cc_member_destination_view
                         from jsonb_array_elements(m.communications) with ordinality as x (x, xid)
                            left join comm on comm.id = (x -> 'type' -> 'id')::int
                            left join resources on resources.id = (x -> 'resource' -> 'id')::int)) communications,
       call_center.cc_get_lookup(cq.id, cq.name::varchar) queue, call_center.cc_view_timestamp(m.expire_at) expire_at, call_center.cc_view_timestamp(m.created_at) created_at,
			m.variables, m.name
from call_center.cc_member m
    inner join result on m.id = result.id
    inner join call_center.cc_queue cq on m.queue_id = cq.id`, map[string]interface{}{
		"Domain":  domainId,
		"Limit":   search.GetLimit(),
		"Offset":  search.GetOffset(),
		"Q":       search.GetQ(),
		"AgentId": search.AgentId,
	})

	if err != nil {
		return nil, model.NewAppError("SqlMemberStore.ListOfflineQueueForAgent", "store.sql_member.list_offline_queue.app_error", nil,
			err.Error(), extractCodeFromErr(err))
	}

	return att, nil
}

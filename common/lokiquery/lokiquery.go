package lokiquery

import (
	"fmt"
	"sort"
	"time"

	"github.com/guonaihong/gout"
	"github.com/talkincode/toughradius/common"
)

func LokiQuery(q *LokiQueryForm) ([]Lokilog, error) {
	queryStr := q.QueryString()
	queryReq := gout.H{"limit": q.Limit, "start": q.Start, "end": q.End}

	if queryStr != "" {
		queryReq["query"] = queryStr + " | json"
	}

	var resp LokiQueryResult
	err := gout.
		GET(common.UrlJoin2(q.LokiApi, "/loki/api/v1/query_range")).
		Debug(q.Debug).
		SetTimeout(time.Second*60).
		SetBasicAuth(q.LokiUser, q.LokiPwd).
		SetQuery(queryReq).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}

	var result []Lokilog
	for _, v := range resp.Data.Result {
		var log Lokilog
		log.Job = v.Stream.Job
		log.Level = v.Stream.Level
		log.Caller = v.Stream.Caller
		log.Msg = v.Stream.Msg
		log.Timestamp = v.Stream.Ts.Format(time.RFC3339)
		log.Namespace = v.Stream.Namespace
		log.Result = v.Stream.Result
		log.Username = v.Stream.Username
		log.Metrics = v.Stream.Metrics
		log.Nasip = v.Stream.Nasip
		log.Error = v.Stream.Error
		log.ShortMessage = ""
		if v.Stream.Namespace != "" {
			log.ShortMessage += "Namespace=" + v.Stream.Namespace + " "
		}

		if v.Stream.Username != "" {
			log.ShortMessage += "Username=" + v.Stream.Username + " "
		}

		log.ShortMessage += "Msg=" + v.Stream.Msg + " "

		if v.Stream.Error != "" {
			log.ShortMessage += "Error=" + v.Stream.Error + " "
		}

		result = append(result, log)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp < result[j].Timestamp
	})
	return result, nil
}

func LokiCountOverTime(q *LokiQueryForm) (*LokiMetricResult, error) {
	queryStr := q.QueryString()
	queryReq := gout.H{"limit": q.Limit, "start": q.Start, "end": q.End}

	if queryStr != "" {
		queryStr = fmt.Sprintf("count_over_time(%s [%s])", queryStr, q.Step)
		queryReq["query"] = queryStr
	}
	var resp LokiMetricResult
	err := gout.
		GET(common.UrlJoin2(q.LokiApi, "/loki/api/v1/query_range")).
		Debug(q.Debug).
		SetTimeout(time.Second*60).
		SetBasicAuth(q.LokiUser, q.LokiPwd).
		SetQuery(queryReq).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func LokiSumRate(q *LokiQueryForm) (*LokiMetricResult, error) {
	queryStr := q.QueryString()
	queryReq := gout.H{"limit": q.Limit, "start": q.Start, "end": q.End}

	if queryStr != "" {
		// sum(rate({job="teamsacs_master"} [5m]))
		queryStr = fmt.Sprintf("sum(rate(%s [%s]))", queryStr, q.Step)
		queryReq["query"] = queryStr
	}
	var resp LokiMetricResult
	err := gout.
		GET(common.UrlJoin2(q.LokiApi, "/loki/api/v1/query_range")).
		Debug(q.Debug).
		SetTimeout(time.Second*60).
		SetBasicAuth(q.LokiUser, q.LokiPwd).
		SetQuery(queryReq).
		BindJSON(&resp).
		Do()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

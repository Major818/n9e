package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Major818/n9e/v4/src/models"
	"github.com/Major818/n9e/v4/src/modules/server/cache"
	"github.com/Major818/n9e/v4/src/modules/server/collector"
	"github.com/Major818/n9e/v4/src/modules/server/config"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/logger"
)

type CollectRecv struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func collectRulePost(c *gin.Context) {
	var recv []CollectRecv
	errors.Dangerous(c.ShouldBind(&recv))

	buf := &bytes.Buffer{}
	creator := loginUsername(c)
	for _, obj := range recv {
		cl, err := collector.GetCollector(obj.Type)
		errors.Dangerous(err)

		if err := cl.Create([]byte(obj.Data), creator); err != nil {
			if _, ok := err.(collector.DryRun); ok {
				fmt.Fprintf(buf, "%s\n", err)
			} else {
				errors.Bomb("%s add rule err %s", obj.Type, err)
			}
		}
	}

	buf.WriteString("ok")
	renderData(c, buf.String(), nil)
}

func collectRulesGetByLocalEndpoint(c *gin.Context) {
	collect := cache.CollectCache.GetBy(urlParamStr(c, "endpoint"))
	renderData(c, collect, nil)
}

func collectRuleGet(c *gin.Context) {
	t := queryStr(c, "type")
	id := queryInt64(c, "id")

	cl, err := collector.GetCollector(t)
	errors.Dangerous(err)

	ret, err := cl.Get(id)
	renderData(c, ret, err)
}

func collectRulesGet(c *gin.Context) {
	nid := queryInt64(c, "nid", -1)
	tp := queryStr(c, "type", "")
	var resp []interface{}
	var types []string

	if tp == "" {
		types = []string{"port", "proc", "log", "plugin"}
	} else {
		types = []string{tp}
	}

	nids := []int64{nid}
	for _, t := range types {
		cl, err := collector.GetCollector(t)
		if err != nil {
			logger.Warning(t, err)
			continue
		}

		ret, err := cl.Gets(nids)
		if err != nil {
			logger.Warning(t, err)
			continue
		}
		resp = append(resp, ret...)
	}

	renderData(c, resp, nil)
}

func collectRulesGetV2(c *gin.Context) {
	nid := queryInt64(c, "nid", 0)
	limit := queryInt(c, "limit", 20)
	typ := queryStr(c, "type", "")

	total, list, err := models.GetCollectRules(typ, nid, limit, offset(c, limit))

	renderData(c, map[string]interface{}{
		"total": total,
		"list":  list,
	}, err)
}

func collectRulePut(c *gin.Context) {
	var recv CollectRecv
	errors.Dangerous(c.ShouldBind(&recv))

	cl, err := collector.GetCollector(recv.Type)
	errors.Dangerous(err)

	buf := &bytes.Buffer{}
	creator := loginUsername(c)
	if err := cl.Update([]byte(recv.Data), creator); err != nil {
		if _, ok := err.(collector.DryRun); ok {
			fmt.Fprintf(buf, "%s\n", err)
		} else {
			errors.Bomb("%s update rule err %s", recv.Type, err)
		}
	}
	buf.WriteString("ok")
	renderData(c, buf.String(), nil)
}

type CollectsDelRev struct {
	Type string  `json:"type"`
	Ids  []int64 `json:"ids"`
}

func collectsRuleDel(c *gin.Context) {
	var recv []CollectsDelRev
	errors.Dangerous(c.ShouldBind(&recv))

	username := loginUsername(c)
	for _, obj := range recv {
		for i := 0; i < len(obj.Ids); i++ {
			cl, err := collector.GetCollector(obj.Type)
			errors.Dangerous(err)

			if err := cl.Delete(obj.Ids[i], username); err != nil {
				errors.Dangerous(err)
			}
		}
	}

	renderData(c, "ok", nil)
}

func collectRuleTypesGet(c *gin.Context) {
	category := queryStr(c, "category")
	switch category {
	case "remote":
		renderData(c, collector.GetRemoteCollectors(), nil)
	case "local":
		renderData(c, collector.GetLocalCollectors(), nil)
	default:
		renderData(c, nil, nil)
	}
}

func collectRuleTemplateGet(c *gin.Context) {
	t := urlParamStr(c, "type")
	collector, err := collector.GetCollector(t)
	errors.Dangerous(err)

	tpl, err := collector.Template()
	renderData(c, tpl, err)
}

type RegExpCheckDto struct {
	Success bool                `json:"success"`
	Data    []map[string]string `json:"tags"`
}

var RegExpExcludePatition string = "```EXCLUDE```"

func regExpCheck(c *gin.Context) {
	param := make(map[string]string, 0)
	errors.Dangerous(c.ShouldBind(&param))

	ret := &RegExpCheckDto{
		Success: true,
		Data:    make([]map[string]string, 0),
	}

	// ??????????????????
	if t, ok := param["time"]; !ok || t == "" {
		tmp := map[string]string{"time": "time????????????????????????"}
		ret.Data = append(ret.Data, tmp)
	} else {
		timePat, _ := GetPatAndTimeFormat(param["time"])
		if timePat == "" {
			tmp := map[string]string{"time": genErrMsg("????????????")}
			ret.Data = append(ret.Data, tmp)
		} else {
			suc, tRes, _ := checkRegPat(timePat, param["log"], true)
			if !suc {
				ret.Success = false
				tRes = genErrMsg("????????????")
			}
			tmp := map[string]string{"time": tRes}
			ret.Data = append(ret.Data, tmp)
		}
	}

	// ????????????
	calc_method, _ := param["calc_method"]

	// ???????????????(with exclude)
	if re, ok := param["re"]; !ok || re == "" {
		tmp := map[string]string{"re": "re????????????????????????"}
		ret.Data = append(ret.Data, tmp)
	} else {
		// ??????exclude?????????
		exclude := ""
		if strings.Contains(re, RegExpExcludePatition) {
			l := strings.Split(re, RegExpExcludePatition)
			if len(l) >= 2 {
				param["re"] = l[0]
				exclude = l[1]
			}
		}

		// ???????????????
		suc, reRes, isSub := checkRegPat(param["re"], param["log"], false)
		if !suc {
			ret.Success = false
			reRes = genErrMsg("?????????")
		}
		if calc_method != "" && calc_method != "cnt" && !isSub {
			ret.Success = false
			reRes = genSubErrMsg("?????????")
		}
		tmp := map[string]string{"?????????": reRes}
		ret.Data = append(ret.Data, tmp)

		// ??????exclude, ?????????????????????
		if exclude != "" {
			suc, exRes, _ := checkRegPat(exclude, param["log"], false)
			if !suc {
				//ret.Success = false
				exRes = "?????????????????????,???????????????????????????"
			}
			tmp := map[string]string{"?????????": exRes}
			ret.Data = append(ret.Data, tmp)
		}
	}

	// ??????tags
	var nonTagKey = map[string]bool{
		"re":          true,
		"log":         true,
		"time":        true,
		"calc_method": true,
	}

	for tagk, pat := range param {
		// ????????????tag??????????????????
		if _, ok := nonTagKey[tagk]; ok {
			continue
		}
		suc, tagRes, isSub := checkRegPat(pat, param["log"], false)
		if !suc {
			// ????????????
			ret.Success = false
			tagRes = genErrMsg(tagk)
		} else if !isSub {
			// ??????????????????
			ret.Success = false
			tagRes = genSubErrMsg(tagk)
		} else if includeIllegalChar(tagRes) || includeIllegalChar(tagk) {
			// ???????????????
			ret.Success = false
			tagRes = genIllegalCharErrMsg()
		}

		tmp := map[string]string{tagk: tagRes}
		ret.Data = append(ret.Data, tmp)
	}

	renderData(c, ret, nil)
}

//?????????????????????????????????????????????????????????pattern???time?????????????????????
func GetPatAndTimeFormat(tf string) (string, string) {
	var pat, timeFormat string
	switch tf {
	case "dd/mmm/yyyy:HH:MM:SS":
		pat = `([012][0-9]|3[01])/[JFMASOND][a-z]{2}/(2[0-9]{3}):([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "02/Jan/2006:15:04:05"
	case "dd/mmm/yyyy HH:MM:SS":
		pat = `([012][0-9]|3[01])/[JFMASOND][a-z]{2}/(2[0-9]{3})\s([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "02/Jan/2006 15:04:05"
	case "yyyy-mm-ddTHH:MM:SS":
		pat = `(2[0-9]{3})-(0[1-9]|1[012])-([012][0-9]|3[01])T([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "2006-01-02T15:04:05"
	case "dd-mmm-yyyy HH:MM:SS":
		pat = `([012][0-9]|3[01])-[JFMASOND][a-z]{2}-(2[0-9]{3})\s([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "02-Jan-2006 15:04:05"
	case "yyyy-mm-dd HH:MM:SS":
		pat = `(2[0-9]{3})-(0[1-9]|1[012])-([012][0-9]|3[01])\s([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "2006-01-02 15:04:05"
	case "yyyy/mm/dd HH:MM:SS":
		pat = `(2[0-9]{3})/(0[1-9]|1[012])/([012][0-9]|3[01])\s([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "2006/01/02 15:04:05"
	case "yyyymmdd HH:MM:SS":
		pat = `(2[0-9]{3})(0[1-9]|1[012])([012][0-9]|3[01])\s([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "20060102 15:04:05"
	case "mmm dd HH:MM:SS":
		pat = `[JFMASOND][a-z]{2}\s+([1-9]|[1-2][0-9]|3[01])\s([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "Jan 2 15:04:05"
	case "mmdd HH:MM:SS":
		pat = `(0[1-9]|1[012])([012][0-9]|3[01])\s([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "0102 15:04:05"
	case "dd mmm yyyy HH:MM:SS":
		pat = `([012][0-9]|3[01])\s+[JFMASOND][a-z]{2}\s+(2[0-9]{3})\s([01][0-9]|2[0-4])(:[012345][0-9]){2}`
		timeFormat = "02 Jan 2006 15:04:05"
	default:
		logger.Errorf("match time pac failed : [timeFormat:%s]", tf)
		return "", ""
	}
	return pat, timeFormat
}

// ????????????????????????body???
func checkRegPat(pat string, log string, origin bool) (succ bool, result string, isSub bool) {
	if pat == "" {
		return false, "", false
	}

	reg, err := regexp.Compile(pat)
	if err != nil {
		return false, "", false
	}

	res := reg.FindStringSubmatch(log)
	switch len(res) {
	// ?????????
	case 0:
		return false, "", false
	// ????????????????????????????????????????????????
	case 1:
		return true, res[0], false
	// ?????????????????????????????????
	default:
		var msg string
		if origin {
			msg = res[0]
			isSub = false
		} else {
			msg = res[1]
			isSub = true
		}
		return true, msg, isSub
	}
}

func includeIllegalChar(s string) bool {
	illegalChars := ":,=\r\n\t"
	return strings.ContainsAny(s, illegalChars)
}

// ????????????????????????
func genErrMsg(sign string) string {
	return fmt.Sprintf("???????????????????????????????????????[%s]?????????", sign)
}

// ??????????????????????????????
func genSubErrMsg(sign string) string {
	return fmt.Sprintf("?????????????????????????????????????????????????????????()?????????????????????????????????[%s]?????????", sign)
}

// ??????????????????????????????
func genIllegalCharErrMsg() string {
	return fmt.Sprintf(`???????????????????????????tag???key??????value??????????????????:[:,/=\r\n\t], ???????????????`)
}

func collectRulesGetByRemoteEndpoint(c *gin.Context) {
	rules := cache.CollectRuleCache.GetBy(urlParamStr(c, "endpoint"))
	renderData(c, rules, nil)

}

func apiCollectsGet(c *gin.Context) {
	node := queryStr(c, "node")
	region := queryStr(c, "region")
	key := region + "-" + node
	collects := cache.ApiCollectCache.GetBy(key)
	renderData(c, collects, nil)
}

func snmpCollectsGet(c *gin.Context) {
	node := queryStr(c, "node")
	region := queryStr(c, "region")
	key := region + "-" + node
	collects := cache.SnmpCollectCache.GetBy(key)
	renderData(c, collects, nil)
}

func snmpHWsGet(c *gin.Context) {
	node := queryStr(c, "node")
	region := queryStr(c, "region")
	key := region + "-" + node
	hws := cache.SnmpHWCache.GetBy(key)
	renderData(c, hws, nil)
}

func apiRegionGet(c *gin.Context) {
	renderData(c, config.Config.Monapi.Region, nil)
}

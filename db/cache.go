package db

import (
	"sync"

	"github.com/astaxie/beego"
)

type Subscrible2Users struct {
	M map[int][]*User
}

type PatternCache struct {
	sync.RWMutex
	M map[int]*Pattern
}

type SubscribleItem struct {
	Pattern *Pattern
	Users   []*User
}

type SafeSubcribleCache struct {
	sync.RWMutex
	M  map[int][]*SubscribleItem
	M2 map[int]int //通过 expression_id + pattern_id 反查 subscrible_id
}

type SafeExpressionCache struct {
	sync.RWMutex
	M map[string]int
}

var subscrible2users = &Subscrible2Users{M: make(map[int][]*User)}
var Patterns = &PatternCache{M: make(map[int]*Pattern)}
var Subscribles = &SafeSubcribleCache{M: make(map[int][]*SubscribleItem), M2: make(map[int]int)}
var Expressions = &SafeExpressionCache{M: make(map[string]int)}

func (this *SafeExpressionCache) Get(fp string) int {
	this.RLock()
	defer this.RUnlock()
	id, exists := this.M[fp]
	if !exists {
		return 0
	}
	return id
}

func (this *SafeExpressionCache) Set(fp string, id int) {
	this.Lock()
	defer this.Unlock()
	this.M[fp] = id
}

func (this *PatternCache) Get(id int) *Pattern {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[id]
	if !exists {
		return nil
	}
	return val
}

func (this *PatternCache) Set(id int, pattern *Pattern) {
	this.Lock()
	defer this.Unlock()
	this.M[id] = pattern
}

func (this *SafeSubcribleCache) GetSubscrible(expression_id int) []*SubscribleItem {
	this.RLock()
	defer this.RUnlock()
	sub_items, _ := this.M[expression_id]
	return sub_items
}

func (this *SafeSubcribleCache) GetSubscribleUsers(expression_id, pattern_id int) []*User {
	for _, item := range this.GetSubscrible(expression_id) {
		if item.Pattern.Id == pattern_id {
			return item.Users
		}
	}
	return nil
}

//判断subscrible状态，true为可用，false为不可用
//不可用一般是被pause了
func (this *SafeSubcribleCache) GetSubscribleStatus(expression_id, pattern_id int) bool {
	for _, item := range this.GetSubscrible(expression_id) {
		if item.Pattern.Id == pattern_id {
			return true
		}
	}
	return false
}

//通过 expression_id 和 paatern_id得到一个key
//TODO 改造映射函数..
func WrapExpressionPattern(expression_id, pattern_id int) int {
	return expression_id*1000 + pattern_id*10
}

func (this *SafeSubcribleCache) GetSubscribleId(expression_id, pattern_id int) int {
	this.RLock()
	defer this.RUnlock()
	id, _ := this.M2[WrapExpressionPattern(expression_id, pattern_id)]
	return id
}

func QuerySubscribleUsers() (ret map[int][]*User, err error) {
	sql := "select subscrible_id, user_id from subscrible_user"
	rows, err := PortalDB.Query(sql)
	if err != nil {
		beego.Debug("ERROR:", err)
		return ret, err
	}
	ret = make(map[int][]*User)
	defer rows.Close()
	for rows.Next() {
		var subscrible_id int
		var user_id int
		err = rows.Scan(
			&subscrible_id,
			&user_id,
		)

		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}
		ret[subscrible_id] = append(ret[subscrible_id], Users.M[user_id])
	}

	return ret, nil
}

func buildExpressionCache() {
	tmp_expressions, err := QueryExpressions()
	if err == nil {
		Expressions.M = tmp_expressions
	}
}

func buildSubscribleCache() {

	tmp_patterns, err := QueryPattern()
	if err == nil {
		Patterns.M = tmp_patterns
	}

	tmp_subscribleuser, err := QuerySubscribleUsers()
	if err == nil {
		subscrible2users.M = tmp_subscribleuser
	}

	sql := "select id, expression_id, pattern_id ,pause from subscrible"
	rows, err := PortalDB.Query(sql)
	if err != nil {
		beego.Debug("ERROR:", err)
		return
	}

	tmp_cache := make(map[int][]*SubscribleItem)
	tmp_cache2 := make(map[int]int)

	defer rows.Close()
	for rows.Next() {
		var subscrible_id int
		var expression_id int
		var pattern_id int
		var pause int
		err = rows.Scan(
			&subscrible_id,
			&expression_id,
			&pattern_id,
			&pause,
		)
		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}

		sub_item := &SubscribleItem{
			Pattern: Patterns.M[pattern_id],
			Users:   subscrible2users.M[subscrible_id],
		}

		if pause == 0 {
			tmp_cache[expression_id] = append(tmp_cache[expression_id], sub_item)
		}
		tmp_cache2[WrapExpressionPattern(expression_id, pattern_id)] = subscrible_id

	}

	Subscribles.Lock()
	defer Subscribles.Unlock()

	Subscribles.M = tmp_cache
	Subscribles.M2 = tmp_cache2

}

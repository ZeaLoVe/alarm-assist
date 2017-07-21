package cache

import (
	"sync"

	"github.com/astaxie/beego"
)

type Subscrible2Users struct {
	sync.RWMutex
	M map[int][]*User
}

type PatternCache struct {
	sync.RWMutex
	M map[int]*Pattern
}

type SafeSubcribleCache struct {
	sync.RWMutex
	M  map[int]*Subscrible   //subscrible_id 对应subscrible
	M2 map[int]int           //通过 expression_id + pattern_id 反查 subscrible_id
	M3 map[int][]*Subscrible //expression_id 对应 subscrible数组
}

type SafeExpressionCache struct {
	sync.RWMutex
	M  map[string]int
	M2 map[int]*Expression //id -> *Expression
}

var subscrible2users = &Subscrible2Users{M: make(map[int][]*User)}
var Patterns = &PatternCache{M: make(map[int]*Pattern)}
var Subscribles = &SafeSubcribleCache{M: make(map[int]*Subscrible), M2: make(map[int]int), M3: make(map[int][]*Subscrible)}
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

func (this *SafeExpressionCache) GetById(id int) *Expression {
	this.RLock()
	defer this.RUnlock()
	expression, exists := this.M2[id]
	if !exists {
		return nil
	}
	return expression
}

func (this *SafeExpressionCache) Set(fp string, id int, exp *Expression) {
	this.Lock()
	defer this.Unlock()
	this.M[fp] = id
	this.M2[id] = exp
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

func (this *PatternCache) Set(pattern *Pattern) {
	this.Lock()
	defer this.Unlock()
	this.M[pattern.Id] = pattern
}

func (this *Subscrible2Users) GetSubscribleUsers(subscrible_id int) []*User {
	this.RLock()
	defer this.RUnlock()
	if users, exist := this.M[subscrible_id]; exist {
		return users
	}
	return nil
}

func (this *Subscrible2Users) SetSubscribleUsers(subscrible_id int, user_list []*User) {
	this.Lock()
	defer this.Unlock()
	this.M[subscrible_id] = user_list
}

//通过 expression_id 和 paatern_id得到一个key
//TODO 改造映射函数..
func WrapExpressionPattern(expression_id, pattern_id int) int {
	return expression_id*1000 + pattern_id*10
}

func (this *SafeSubcribleCache) GetSubscribles(expression_id int) []*Subscrible {
	this.RLock()
	defer this.RUnlock()
	sub_items, _ := this.M3[expression_id]
	return sub_items
}

func (this *SafeSubcribleCache) GetSubscribleId(expression_id, pattern_id int) int {
	this.RLock()
	defer this.RUnlock()
	id, _ := this.M2[WrapExpressionPattern(expression_id, pattern_id)]
	return id
}

func (this *SafeSubcribleCache) SetSubscrible(sub *Subscrible) {
	this.Lock()
	defer this.Unlock()

	if _, exist := this.M[sub.Id]; exist { //已经存在，修改cache
		this.M2[WrapExpressionPattern(sub.Expression.Id, sub.Pattern.Id)] = sub.Id
		for i, old_sub := range this.M3[sub.Expression.Id] {
			if old_sub.Id == sub.Id {
				this.M3[sub.Expression.Id][i] = sub
				break
			}
		}
	} else { //不存在，添加进cache
		this.M[sub.Id] = sub
		this.M2[WrapExpressionPattern(sub.Expression.Id, sub.Pattern.Id)] = sub.Id
		this.M3[sub.Expression.Id] = append(this.M3[sub.Expression.Id], sub)
	}
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
	tmp_expressions, tmp_expressions2, err := QueryExpressions()
	if err == nil {
		Expressions.M = tmp_expressions
		Expressions.M2 = tmp_expressions2
	}
}

func buildSubscribleCache() {

	tmp_patterns, err := QueryPattern()
	if err == nil {
		Patterns.M = tmp_patterns
	}

	tmp_subscribleuser, err := QuerySubscribleUsers()
	if err == nil {
		subscrible2users.Lock()
		subscrible2users.M = tmp_subscribleuser
		subscrible2users.Unlock()
	}

	sql := "select id, name, creator, note, expression_id, pattern_id, pause from subscrible"
	rows, err := PortalDB.Query(sql)
	if err != nil {
		beego.Debug("ERROR:", err)
		return
	}

	tmp_cache := make(map[int]*Subscrible)
	tmp_cache2 := make(map[int]int)
	tmp_cache3 := make(map[int][]*Subscrible)

	defer rows.Close()
	for rows.Next() {
		s := Subscrible{}
		err = rows.Scan(
			&s.Id,
			&s.Name,
			&s.Creator,
			&s.Note,
			&s.Expression.Id,
			&s.Pattern.Id,
			&s.Pause,
		)

		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}
		if exp := Expressions.GetById(s.Expression.Id); exp != nil {
			s.Expression = *exp
		}

		if pat := Patterns.Get(s.Pattern.Id); pat != nil {
			s.Pattern = *pat
		}

		subscrible2users.RLock()
		s.Users = subscrible2users.M[s.Id]
		subscrible2users.RUnlock()

		tmp_cache[s.Id] = &s
		tmp_cache2[WrapExpressionPattern(s.Expression.Id, s.Pattern.Id)] = s.Id

		//expression to []subscrible 仅用在报警上，不会用于CURD操作，所以暂停的时候，就不需要加载该订阅到cache
		if s.Pause == 0 {
			tmp_cache3[s.Expression.Id] = append(tmp_cache3[s.Expression.Id], &s)
		}

	}

	Subscribles.Lock()
	defer Subscribles.Unlock()

	Subscribles.M = tmp_cache
	Subscribles.M2 = tmp_cache2
	Subscribles.M3 = tmp_cache3

}

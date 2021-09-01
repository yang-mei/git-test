package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"gorm.io/gorm"

	_ "github.com/mitchellh/mapstructure"
)

type aa struct {
	Name string `json:"name"`
}

type Host struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Rets struct {
	Date     string  `json:"date"`
	UID      int     `json:"uid"`
	Nickname string  `json:"nickname"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Count    int     `json:"count"`
}

func main() {
	fmt.Println(strconv.FormatFloat(float64(99999999999)/float64(100), 'f', -1, 64))
	print(11)
}

func test11(a int) {
	panic(nil)
}

func Go(ctx context.Context, l *log.Helper, g func(), msg interface{}) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				l.WithContext(ctx).Errorf("go %v err:%v", msg, err)
				// logrus.Errorf("msg:%v, err:%s", msg, err)
				print(err)
			}
		}()
		g()
	}()
}

func VerifyMobileFormat(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

type AgentLog struct {
	gorm.Model
	Assigner int
	KolID    int
	UserID   int
	DeptID   int
	Type     int
}

func (AgentLog) TableName() string {
	return "agent_log"
}

func GetQuarterDates(date time.Time) []time.Time {
	dates := make([]time.Time, 0)
	for i := 0; i <= 6; i++ {
		if i == 0 {
			bnmonth := int(date.Month())
			if date.Month()%3 == 1 {
				bnmonth = int(date.Month())
			} else if date.Month()%3 == 2 {
				bnmonth = int(date.Month() - 1)
			} else {
				bnmonth = int(date.Month() - 2)
			}
			dates = append(dates, date)
			y, _, _ := date.Date()
			date = time.Date(y, time.Month(bnmonth), 1, 0, 0, 0, 0, time.Local)
			continue
		}
		date = date.AddDate(0, 0, -1)
		dates = append(dates, date)
		bnmonth := int(date.Month())
		if date.Month()%3 == 1 {
			bnmonth = int(date.Month())
		} else if date.Month()%3 == 2 {
			bnmonth = int(date.Month() - 1)
		} else {
			bnmonth = int(date.Month() - 2)
		}
		y, _, _ := date.Date()
		date = time.Date(y, time.Month(bnmonth), 1, 0, 0, 0, 0, time.Local)
	}
	return dates
}

func GetDates(t time.Time) []time.Time {
	y, m, _ := t.Date()
	endFirstDate := time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
	dates := make([]time.Time, 0)
	for i := 0; i <= 6; i++ {
		if i == 0 {
			dates = append(dates, t)
			continue
		}
		endFirstDate = endFirstDate.AddDate(0, 0, -1)
		dates = append(dates, endFirstDate)
		ly, lm, _ := endFirstDate.Date()
		endFirstDate = time.Date(ly, lm, 1, 0, 0, 0, 0, time.Local)
	}
	return dates
}

func LastMonthDate2(t time.Time) time.Time {
	deltaDay := -0
	for {
		lastMonth := t.AddDate(0, -3, deltaDay)
		if lastMonth.Month() == t.Month() {
			deltaDay--
			continue
		}
		return lastMonth
	}
}

func GetMonthRange(startTime, endTime time.Time) []time.Time {
	dates := make([]time.Time, 0)
	startY, startM, _ := startTime.Date()
	_, endM, _ := endTime.Date()
	subM := int(endM - startM)
	if subM == 3 {
		dates = append(dates, time.Date(startY, startM+1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1),
			time.Date(startY, startM+2, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1),
			time.Date(startY, startM+3, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1),
		)
	}
	if subM < 3 {
		for i := 0; i < subM+1; i++ {
			date := time.Date(startY, startM+time.Month(i+1), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
			if subM == i {
				date = endTime.AddDate(0, 0, -1)
			}
			dates = append(dates, date)
		}
	}
	return dates
}

func GetWeekDate(startTime, endTime time.Time) []time.Time {
	dates := make([]time.Time, 0)
	days := int(endTime.Sub(startTime).Hours() / float64(24))
	for i := 0; i < days; i++ {
		date := startTime.AddDate(0, 0, i)
		if date.Weekday() == 0 {
			dates = append(dates, date)
		}
		if i == days-1 {
			if date.Weekday() > 0 {
				dates = append(dates, date)
			}
		}
	}
	return dates
}

func GegLastDayByYearAndMonth(year int, month int) (days int) {
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30

		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	return days
}

const (
	RangeTypeDay        = 1 // 昨日/更多周期（天）
	RangeTypeWeek       = 2 // 上周/更多周期（周）
	RangeTypeMonth      = 3 // 上月/更多周期（月）
	RangeTypeToday      = 4 // 今日
	RangeTypeThisWeek   = 5 // 本周
	RangeTypeThisMonth  = 6 // 本月
	RangeTypeLast7Days  = 7 // 近7天
	RangeTypeLast30Days = 8 // 近30天
)

func LastMonthDate(t time.Time) time.Time {
	deltaDay := 0
	for {
		lastMonth := t.AddDate(0, -1, deltaDay)
		if lastMonth.Month() == t.Month() {
			deltaDay--
			continue
		}
		return lastMonth
	}
}

func PrevDate(date time.Time, range_ int) time.Time {
	switch range_ {
	case RangeTypeDay:
		return date.AddDate(0, 0, -1)
	case RangeTypeThisWeek, RangeTypeWeek:
		return date.AddDate(0, 0, -7)
	case RangeTypeThisMonth, RangeTypeMonth:
		return LastMonthDate(date)
	case RangeTypeLast7Days:
		return date.AddDate(0, 0, -7)
	case RangeTypeLast30Days:
		return date.AddDate(0, 0, -30)
	}
	return date.AddDate(0, 0, -1)
}

//func main() {
//	http.HandleFunc("/", Handler)
//	http.ListenAndServe("127.0.0.0:8000", nil)
//}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world")
}

//需要传递函数
func callback(i int) {
	fmt.Println("i am callBack")
	fmt.Println(i)
}

//main中调用的函数
func one(i int, f func(int)) {
	two(i, fun(f))
}

//one()中调用的函数
func two(i int, c Call) {
	c.call(i)
}

//定义的type函数
type fun func(int)

//fun实现的Call接口的call()函数
func (f fun) call(i int) {
	f(i)
}

//接口
type Call interface {
	call(int)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(string, string)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w string, r string) {
	f(w, r)
}

func ParseKPIDate(date string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02", date, time.Local)
	return t
}

func LastMonth(t time.Time) time.Time {
	deltaDay := 0
	for {
		lastMonth := t.AddDate(0, -1, deltaDay)
		if lastMonth.Month() == t.Month() {
			deltaDay--
			continue
		}
		return lastMonth
	}
	return time.Time{}
}

func ReserveDecimal(d float64, n int) float64 {
	return float64(int64(d*math.Pow10(n))) / math.Pow10(n)
}

var sheets = []string{"斗鱼IP组", "斗鱼网游组", "斗鱼手游组"}

//func GetExe() {
//	xlFile, err := xlsx.OpenFile("")
//	if err != nil {
//		panic(err)
//	}
//	for _, sheet := range xlFile.Sheets {
//		if sheet.Name == sheets[0] || sheet.Name == sheets[1] || sheet.Name == sheets[2] {
//			groupsRows := xlsx.GetRows("groups")
//		}
//	}
//}

type person struct {
	Name string
	Age  int
}

type personSlice []*person

func (s personSlice) Len() int           { return len(s) }
func (s personSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s personSlice) Less(i, j int) bool { return s[i].Age < s[j].Age }

func OrderSlice() {
	a := personSlice{
		{
			Name: "AAA",
			Age:  0,
		},
		{
			Name: "BBB",
			Age:  0,
		},
		{
			Name: "CCC",
			Age:  0,
		},
		{
			Name: "DDD",
			Age:  11,
		},
		{
			Name: "EEE",
			Age:  0,
		},
	}
	sort.Stable(a)
	fmt.Println(a)
}

//func SortIncomes(incomes map[int]float64) {
//	type kv struct {
//		Key   int
//		Value float64
//	}
//	var ss []kv
//	for k, v := range incomes {
//		ss = append(ss, kv{k, v})
//	}
//	sort.Stable(ss, func(i, j int) bool {
//		return ss[i].Value > ss[j].Value // 降序
//	})
//
//	fmt.Println(ss)
//}

//func initPostgres() *gorm.DB {
//	host := "pgm-bp15ro9yj01b46stjo.pg.rds.aliyuncs.com"
//	port := 3433
//	username := "postgres"
//	dbname := "xdashboard"
//	password := "TWs8BsE8WGKPYo"
//	url := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", host, port, username, dbname, password)
//	db, err := gorm.Open("postgres", url)
//	if err != nil {
//		logrus.Panicf("postgres db connection error: %v", err)
//	}
//	logrus.Infof("connect to postgres successful!")
//	db.DB().SetMaxOpenConns(100)
//	db.SingularTable(true)
//	return db
//}

func GetLastWeek() string {
	thisWeekMonday := GetFirstDateOfWeek()
	TimeMonday, _ := time.Parse("2006-01-02", thisWeekMonday)
	lastWeekMonday := TimeMonday.AddDate(0, 0, -14)
	lastWeekSunday := TimeMonday.AddDate(0, 0, -8)
	fmt.Println(lastWeekSunday)
	weekMonday := lastWeekMonday.Format("20060102")
	return weekMonday
}

func GetFirstDateOfWeek() (weekMonday string) {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	weekMonday = weekStartDate.Format("2006-01-02")
	return
}

func ExampleClient() {
	client := redis.NewClient(&redis.Options{
		Addr:     "47.98.205.143:16372",
		Password: "Abcd1234*", // no password set
		DB:       5,           // use default DB
	})
	//字符串
	val2, err := client.IncrBy("test-hexists", 2).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(val2)
	exists, err := client.Exists("test-hexists").Result()
	fmt.Println(exists)
	hExists, err := client.HExists("test-hexists", "a").Result()
	fmt.Println(hExists)

	//ret := client.SAdd("acKey", 2)
	//val2, err := ret.Result()
	////err := client.Set("acKey", 20, 60*time.Second).Err()
	//if err != nil {
	//	panic(err)
	//}

	//val, err := client.SMembers("acKey").Result()
	//if err == redis.Nil {
	//	fmt.Println("val")
	//}
	//fmt.Printf("val:%v", val)
	//client.Del("ymtest")
	////client.HMSet("ymtest", map[string]interface{}{"1": 1})
	//task := make(map[string]int, 0)
	//a, err := client.G("ymtest", task).Result()
	if err != nil {
		panic(err)
	}
	//fmt.Println(a)
}

type loginSinaSt struct {
	Username     string
	password     string
	savestate    int
	entry        string
	mainpageflag int
}

func loginSina() {
	login_url := `https://passport.weibo.cn/sso/login`

	a := &loginSinaSt{
		Username:     "15805265032",
		password:     "Ym1209030111",
		savestate:    1,
		entry:        "mweibo",
		mainpageflag: 1}
	bodyByte, err := json.Marshal(a)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", login_url, bytes.NewBuffer(bodyByte))
	req.Header.Set("user-agent", "Mozilla/5.0")
	req.Header.Set("Referer", `https://passport.weibo.cn/signin/login?entry=mweibo&res=wel&wm=3349&r=https%3A%2F%2Fm.weibo.cn%2F`)
	if err != nil {
		fmt.Print(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	defer resp.Body.Close()
	bs, _ := ioutil.ReadAll(resp.Body)
	fmt.Sprint(string(bs))
}

// def login_sina():
// """
// 登录新浪
// :return:
// """
// # 登录URL
// login_url = 'https://passport.weibo.cn/sso/login'
// # 请求头
// headers = {'user-agent': 'Mozilla/5.0',
// 'Referer': 'https://passport.weibo.cn/signin/login?entry=mweibo&res=wel&wm=3349&r=https%3A%2F%2Fm.weibo.cn%2F'}
// # 传递用户名和密码
// data = {'username': '15805265032',
// 'password': 'Ym1209030111',
// 'savestate': 1,
// 'entry': 'mweibo',
// 'mainpageflag': 1}
// try:
// r = s.post(login_url, headers=headers, data=data)
// r.raise_for_status()
// except Exception as e:
// print(e)
// return 0
// # 打印请求结果
// print(json.loads(r.text)['msg'])
// return 1

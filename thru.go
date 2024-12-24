package thru

import (
	"database/sql/driver"
	"math"
	"time"
)

var (
	// 每个月的最大天数
	maxDays = [13]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
)

type Time struct {
	time time.Time
}

// ---- 创建时间 ----

func New(t time.Time) Time {
	return Time{time: t}
}

func Now() Time {
	return Time{time: time.Now()}
}

func Date[Month ~int](
	year int, month Month, day, hour,
	min, sec, nsec int, loc ...*time.Location,
) Time {
	if loc == nil {
		loc = append(loc, time.Local)
	}
	m := time.Month(month)
	return Time{time: time.Date(year, m, day, hour, min, sec, nsec, loc[0])}
}

// Unix 返回给定时间戳的本地时间。secs 可以是秒、毫秒或纳秒级时间戳。
func Unix(secs int64) Time {
	if secs <= 9999999999 { // 10 位及以下，视为秒级时间戳
		return Time{time: time.Unix(secs, 0)}
	}
	return Time{time: time.Unix(0, secs)}
}

// ---- 添加时间 ----

// AddYear 添加年月日，指定 d 参数时年、月会溢出。默认添加一年。
func (t Time) AddYear(ymd ...int) Time {
	y, m := 1, 0

	if i := len(ymd); i > 0 {
		if y = ymd[0]; i > 1 {
			m = ymd[1]
		}
		if i > 2 {
			return Time{time: t.time.AddDate(y, m, ymd[2])}
		}
	}

	months := t.Month() + m            // 计算总月数
	y += t.Year() + (months-1)/12      // 计算新的年份
	if m = (months-1)%12 + 1; m <= 0 { // 计算剩余月数并处理负数情况
		m += 12 // 将月份调整到正确范围 (1-12)
		y--     // 由于月份向前偏移了一年，所以年份需要减 1。
	}

	d := t.Day()                            // 获取原始日期
	if maxDay := DaysIn(y, m); d > maxDay { // 处理溢出天数
		d = maxDay
	}

	return Date(
		y, m, d,
		t.Hour(), t.Minute(), t.Second(), t.Second(9),
		t.Location(),
	)
}

// AddMonth 添加月日，默认添加一月。
func (t Time) AddMonth(md ...int) Time {
	m := 1
	if i := len(md); i <= 1 { // 不传天数以避免溢出
		if i == 1 {
			m = md[0]
		}
		return t.AddYear(0, m)
	}
	return t.AddYear(0, m, md[1])
}

// AddDay 添加天数，默认添加一天。
func (t Time) AddDay(d ...int) Time {
	day := 1
	if d != nil {
		day = d[0]
	}
	return t.AddYear(0, 0, day)
}

// Add 返回 t + d 时间
func (t Time) Add(d time.Duration) Time {
	return Time{time: t.time.Add(d)}
}

// ---- 选择时间 ----

// Go 偏移 ±y 年并选择 m 月 d 日，如果 m, d 为负数，则从最后的月、日开始偏移。
func (t Time) Go(y int, md ...int) Time {
	year, month, day := t.Year()+y, t.Month(), t.Day()

	if i := len(md); i > 0 {
		if m := float64(md[0]); m > 0 {
			month = int(math.Min(m, 12))
		} else if m < 0 {
			month = int(math.Max(13+m, 1))
		}
		if i > 1 {
			day = md[1]
		}
	}

	maxDay := float64(DaysIn(year, month))
	if d := float64(day); d > 0 {
		day = int(math.Min(d, maxDay))
	} else {
		day = int(math.Max(maxDay+d+1, 1))
	}

	return Date(
		year, month, day,
		t.Hour(), t.Minute(), t.Second(), t.Second(9),
		t.Location(),
	)
}

// GoYear 和 Go() 一样，但 y 指定为确切年份而非偏移。
func (t Time) GoYear(y int, md ...int) Time {
	return t.Go(y-t.Year(), md...)
}

func (t Time) GoMonth(m int, d ...int) Time {
	var day int
	if d != nil {
		day = d[0]
	}
	return t.Go(0, m, day)
}

func (t Time) GoDay(d int) Time {
	return t.Go(0, 0, d)
}

// ---- 开始时间 ----

// Start 它和 `AddYear(ymd ...int)` 类似，但是返回开始时间。
func (t Time) Start(ymd ...int) Time {
	y, m, d := t.Year(), 1, 1

	if i := len(ymd); i == 1 { // 传年
		y += ymd[0]
	} else if i == 2 { // 传年月
		y, m = y+ymd[0], t.Month()+ymd[1]
	} else if i > 2 { // 传年月日
		y, m, d = y+ymd[0], t.Month()+ymd[1], t.Day()+ymd[2]
	}

	return Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// StartMonth 返回 m 月 d 日的开始时间
func (t Time) StartMonth(md ...int) Time {
	m, d := 0, 0
	if i := len(md); i > 0 {
		if m = md[0]; i > 1 {
			d = md[1]
		}
	}
	return t.Start(0, m, d)
}

// StartDay 返回 t 月 d 日开始时间
func (t Time) StartDay(d ...int) Time {
	var day int
	if d != nil {
		day = d[0]
	}
	return t.Start(0, 0, day)
}

// StartWeek 偏移至 ±n 周开始时间，默认返回本周开始时间。
func (t Time) StartWeek(n ...int) Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}

	day := t.Day() - int(weekday) + 1
	if n != nil {
		day += n[0] * 7
	}

	return Date(t.Year(), t.Month(), day, 0, 0, 0, 0, t.Location())
}

// ---- 结束时间 ----

// End 它和 `AddYear(ymd ...int)` 类似，但是返回结束时间。
func (t Time) End(ymd ...int) Time {
	y, m, d := t.Year(), t.Month(), t.Day()

	if i := len(ymd); i == 1 { // 传年
		y += ymd[0]
	} else if i == 2 { // 传年月
		y, m = y+ymd[0], m+ymd[1]
	} else if i > 2 { // 传年月日
		y, m, d = y+ymd[0], m+ymd[1], d+ymd[2]
	}

	return Date(y, m, d, 23, 59, 59, 999999999, t.Location())
}

// EndMonth 时间 t 的 m 月 d 日的结束时间
func (t Time) EndMonth(md ...int) Time {
	m, d := 0, 0
	if i := len(md); i > 0 {
		if m = md[0]; i > 1 {
			d = md[1]
		}
	}
	return t.End(0, m, d)
}

// EndDay 时间 t 的 d 日的结束时间
func (t Time) EndDay(d ...int) Time {
	var day int
	if d != nil {
		day = d[0]
	}
	return t.End(0, 0, day)
}

// EndWeek 偏移至 ±n 周结束时间，默认返回本周结束时间。
func (t Time) EndWeek(n ...int) Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}

	day := t.Day() + (7 - int(weekday))
	if n != nil {
		day += n[0] * 7
	}

	return Date(t.Year(), t.Month(), day, 23, 59, 59, 999999999, t.Location())
}

// ---- 获取时间 ----

// Year 返回 t 的年份
func (t Time) Year() int {
	return t.time.Year()
}

// Month 返回 t 的月份
func (t Time) Month() int {
	return int(t.time.Month())
}

// Day 返回 t 的天数
func (t Time) Day() int {
	return t.time.Day()
}

// Hour 返回小时，范围 [0, 23]
func (t Time) Hour() int {
	return t.time.Hour()
}

// Minute 返回分钟，范围 [0, 59]
func (t Time) Minute() int {
	return t.time.Minute()
}

// Second 返回时间的秒数或指定纳秒精度的小数部分
//
// 参数 n (可选) 指定返回的精度：
//   - 不提供或 0: 返回整秒数 (0-59)
//   - 1-9: 返回纳秒精度的小数部分，n 表示小数位数
func (t Time) Second(n ...int) int {
	if n == nil || n[0] == 0 {
		return t.time.Second()
	}
	divisor := int(math.Pow10(9 - Clamp(n[0], 1, 9)))
	return t.time.Nanosecond() / divisor
}

// Clock 返回一天中的小时、分钟和秒
func (t Time) Clock() (hour, min, sec int) {
	return t.time.Clock()
}

// Weekday 返回星期几
func (t Time) Weekday() time.Weekday {
	return t.time.Weekday()
}

// YearDay 返回年份中的日期，非闰年的范围为 [1,365]，闰年的范围为 [1,366]。
func (t Time) YearDay() int {
	return t.time.YearDay()
}

// Days 返回本月份的最大天数
func (t Time) Days() int {
	return DaysIn(t.Year(), t.Month())
}

// Unix 返回时间戳，可选择指定精度。
//
// 参数 n (可选) 指定返回的时间戳精度：
//   - 不提供或 0: 秒级 (10位)
//   - 3: 毫秒级 (13位)
//   - 6: 微秒级 (16位)
//   - 9: 纳秒级 (19位)
//   - 其他值: 对应位数的时间戳
func (t Time) Unix(n ...int) int64 {
	if n == nil || n[0] == 0 {
		return t.time.Unix()
	}
	precision := Clamp(n[0]+10, 1, 19)
	divisor := int64(math.Pow10(19 - precision))
	return t.time.UnixNano() / divisor
}

// UTC 返回 UTC 时间
func (t Time) UTC() Time {
	return Time{time: t.time.UTC()}
}

// Local 返回本地时间
func (t Time) Local() Time {
	return Time{time: t.time.Local()}
}

// In 返回指定的 loc 时间
func (t Time) In(loc *time.Location) Time {
	return Time{time: t.time.In(loc)}
}

// Round 返回距离当前时间最近的 "跃点"。
//
// 这个函数有点不好理解，让我来尝试解释一下：
// 想象在当前时间之外存在着另一个时间循环，时间轴每次移动 d，而每个 d 就是一个 "跃点"。
// 该函数将返回距离当前时间最近的那个 "跃点"，如果当前时间正好位于两个 "跃点" 中间，返回指向未来的那个 "跃点"。
//
// 示例：假设当前时间是 2021-07-21 14:35:29.650
//
//   - 舍入到秒：t.Round(time.Second)
//
//     根据定义，时间 14:35:29.650 距离下一个跃点 14:35:30.000 最近 (只需要 0.35ns)，
//     而距离上一个跃点 14:35:29.000 较远 (需要 65ns)，故返回下一个跃点 14:35:30.000。
//
//   - 舍入分钟：t.Round(time.Minute)
//
//     时间 14:35:29.650 距离上一个跃点 14:35:00 最近（只需要 29.650s），
//     而距离下一个跃点 14:36:00 较远 (需要 30.350s)，故返回上一个跃点 14:35:00.000。
//
//   - 舍入 15 分钟：t.Round(15 * time.Minute)
//
//     跃点：--- 14:00:00 --- 14:15:00 --- 14:30:00 -- t --- 14:45:00 ---
//
//     时间 14:35:29.650 处在 14:30 (上一个跃点) 和 14:45 (下一个跃点) 之间，
//     距离上一个跃点最近，故返回上一个跃点时间：14:30:00。
func (t Time) Round(d time.Duration) Time {
	return Time{time: t.time.Round(d)}
}

// Truncate 与 Round 类似，但是 Truncate 永远返回指向过去时间的 "跃点"，不会进行四舍五入。
//
// 可视化理解：
//
//	---|-------|-------|-------|---- 时间线
//	  d1       t      d2      d3
//
// Truncate 将永远返回 d1，如果时间 t 正好处于某个 "跃点" 位置，返回 t。
func (t Time) Truncate(d time.Duration) Time {
	return Time{time: t.time.Truncate(d)}
}

// Time 返回 time.Time
func (t Time) Time() time.Time {
	return t.time
}

// Location 返回时区信息
func (t Time) Location() *time.Location {
	return t.time.Location()
}

// ---- 比较时间 ----

// DiffIn 返回 t 和 u 的时间差。
//
// 参数 unit 指定返回差异的单位：
//   - "y": 年
//   - "M": 月
//   - "d": 日
//   - "h": 小时
//   - "m": 分钟
//   - "s": 秒
func (t Time) DiffIn(u Time, unit string) int {
	switch unit {
	case "y":
		return t.Year() - u.Year()
	case "M":
		return (t.Year()-u.Year())*12 + t.Month() - u.Month()
	case "d":
		return int(t.Sub(u).Hours() / 24)
	case "h":
		return int(t.Sub(u).Hours())
	case "m":
		return int(t.Sub(u).Minutes())
	case "s":
		return int(t.Sub(u).Seconds())
	}
	return 0
}

func (t Time) DiffAbsIn(u Time, unit string) int {
	return int(math.Abs(float64(t.DiffIn(u, unit))))
}

// Sub 返回 t - u 的时间差
func (t Time) Sub(u Time) time.Duration {
	return t.time.Sub(u.time)
}

// Before 返回 t 是否在 u 之前 (t < u)
func (t Time) Before(u Time) bool {
	return t.time.Before(u.time)
}

// After 返回 t 是否在 u 之后 (t > u)
func (t Time) After(u Time) bool {
	return t.time.After(u.time)
}

// Equal 返回 t == u
func (t Time) Equal(u Time) bool {
	return t.time.Equal(u.time)
}

// Compare 比较 t 和 u。
// 如果 t 小于 u，返回 -1；大于返回 1；等于返回 0。
func (t Time) Compare(u Time) int {
	return t.time.Compare(u.time)
}

// Since 返回自 t 以来经过的时间。它是 Now().Sub(t) 的简写。
func Since(t Time) time.Duration {
	return time.Since(t.time)
}

// Until 返回直到 t 的持续时间。它是 t.Sub(Now()) 的简写。
func Until(t Time) time.Duration {
	return time.Until(t.time)
}

// ---- 序列化时间 ----

// Scan 由 DB 转到 Go 时调用
func (t *Time) Scan(value any) error {
	if v, ok := value.(time.Time); ok {
		*t = Time{time: v}
	}
	return nil
}

// Value 由 Go 转到 DB 时调用
func (t Time) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}

// MarshalJSON 将 t 转为 JSON 字符串时调用
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Format(DateTime) + `"`), nil
}

// UnmarshalJSON 将 JSON 字符串转为 t 时调用
func (t *Time) UnmarshalJSON(b []byte) (err error) {
	*t, err = ParseE(string(b))
	return
}

// ---- 其他 ----

// IsZero 返回 t 是否零时，即 0001-01-01 00:00:00 UTC。
func (t Time) IsZero() bool {
	return t.time.IsZero()
}

func (t Time) ZeroOr(u Time) Time {
	if t.IsZero() {
		return u
	}
	return t
}

// IsDST 返回时间是否夏令时
func (t Time) IsDST() bool {
	return t.time.IsDST()
}

// IsLeap 返回 year 是否闰年
func IsLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysIn 返回 y 年 m 月的最大天数
//
// 1, 3, 5, 7, 8, 10, 12 月有 31 天；4, 6, 9, 11 月有 30 天；
// 平年 2 月有 28 天，闰年 29 天。
func DaysIn(y, m int) int {
	if m == 2 && IsLeap(y) {
		return 29
	}
	return maxDays[m]
}

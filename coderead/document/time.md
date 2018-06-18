### time

#### type Time

A Time represents an instant in time with nanosecond precision.
纳秒精度
Programs using times should typically store and pass them as values, not pointers.
That is, time variables and struct fields should be of type time.Time, not *time.Time.
使用time.Time 需要传递值而非指针
A Time value can be used by multiple goroutines simultaneously except that the methods GobDecode, UnmarshalBinary, UnmarshalJSON and UnmarshalText are not concurrency-safe.
一个time.Time可以被多个goroutine同时使用，除了一下几个方法
Time instants can be compared using the Before, After, and Equal methods.
time.Time 可以使用 Before, After, Equal 方法来进行比较
The Sub method subtracts two instants, producing a Duration.
Sub方法可以相见两个Time, 并产生一个Duration
The Add method adds a Time and a Duration, producing a Time.
Add方法可以使 一个Time和一个Duration 相加，产生一个Time.
The zero value of type Time is January 1, year 1, 00:00:00.000000000 UTC.
Time的零值是 January 1, year 1, 00:00:00.000000000 UTC
As this time is unlikely to come up in practice, the IsZero method gives a simple way of detecting a time that has not been initialized explicitly.
Time提供了IsZero方法来判定零值
Each Time has associated with it a Location, consulted when computing the presentation form of the time, such as in the Format, Hour, and Year methods. The methods Local, UTC, and In return a Time with a specific location. Changing the location in this way changes only the presentation; it does not change the instant in time being denoted and therefore does not affect the computations described in earlier paragraphs.

In addition to the required “wall clock” reading, a Time may contain an optional reading of the current process's monotonic clock, to provide additional precision for comparison or subtraction. See the “Monotonic Clocks” section in the package documentation for details.

Note that the Go == operator compares not just the time instant but also the Location and the monotonic clock reading.
使用 == 来判断时间是否相等, 不仅仅判定时间，还是有位置和 单调变化钟?
Therefore, Time values should not be used as map or database keys without first guaranteeing that the identical Location has been set for all values, which can be achieved through use of the UTC or Local method, and that the monotonic clock reading has been stripped by setting t = t.Round(0). In general, prefer t.Equal(u) to t == u, since t.Equal uses the most accurate comparison available and correctly handles the case when only one of its arguments has a monotonic clock reading.
Time是不可哈希的, 不能用作map 或者 database 的key

#### Wall Time & Monotonic Time

+ CLOCK_MONOTONIC: monotonic time 单调时间: 系统启动以后流逝的时间.用户不能修改这个时间，但是当系统进入休眠（suspend）时，CLOCK_MONOTONIC是不会增加的。
+ CLOCK_REALTIME: wall time 挂钟时间: 实际时间, 可调整和修改。可以被系统命令修改, 也可以被NTP(Network Time Protocol: 使计算机时间同步化的一种协议)修改。当系统休眠（suspend）时，仍然会运行的（系统恢复时，kernel去作补偿）


### type Timer

The Timer type represents a single event. When the Timer expires, the current time will be sent on C, unless the Timer was created by AfterFunc.
Timer类型代表一个单一的时间. 当定时器过期时, timer会像chan中发送当前时间戳
除非当前的timer是由AfterFunc创建的
A Timer must be created with NewTimer or AfterFunc.
timer必须由NewTimer或者 AfterFunc来创建


    func AfterFunc
    func AfterFunc(d Duration, f func()) *Timer
    AfterFunc waits for the duration to elapse and then calls f in its own goroutine.
    It returns a Timer that can be used to cancel the call using its Stop method.
    After会等待 duration时间间隔结束  然后在他自己的goroutine中调用 f 函数
    返回的timer指针 用于 stop这个方法

    func NewTimer
    func NewTimer(d Duration) *Timer
    NewTimer creates a new Timer that will send the current time on its channel after at least duration d.


# Signature
## Publisher
- **Eulogist/Bunker**, the standard auth service of NEMC robot field.
- **ToolDelta**, the common robot provider.
- **Neo-Omega**, the core of a robot.





## Author
Eternal Crystal





## Document Info
- API Version: **v1.0**
- Documents Version: **v1.1**







# Catalogue
- [Signature](#signature)
  - [Publisher](#publisher)
  - [Author](#author)
  - [Document Info](#document-info)
- [Catalogue](#catalogue)
- [New concept: Vitality API](#new-concept-vitality-api)
  - [Abstract](#abstract)
  - [Prior data collection](#prior-data-collection)
    - [Summary](#summary)
    - [Findings](#findings)
    - [Preliminary conclusion](#preliminary-conclusion)
    - [Limitations](#limitations)
- [Introduction to Vitality API](#introduction-to-vitality-api)
  - [Background](#background)
  - [G79 user transaction](#g79-user-transaction)
  - [Session](#session)
    - [Data structure](#data-structure)
    - [Description](#description)
    - [Session type](#session-type)
  - [Vitality API](#vitality-api)
    - [Register Active G79 User](#register-active-g79-user)
    - [Request Session Info](#request-session-info)
    - [Clean Up Session](#clean-up-session)
    - [Request Daily Growth](#request-daily-growth)
    - [Get Currency Online](#get-currency-online)
    - [Keep G79 User Alive](#keep-g79-user-alive)
  - [Changes in rental server login response](#changes-in-rental-server-login-response)
  - [Implementation of Session Maintainer](#implementation-of-session-maintainer)
  - [Adaptation](#adaptation)







# New concept: Vitality API
## Abstract
**Vitality API** is a new API that used to solve the issue of robots being banned during their use.<br/>
In addition, this API is also used to make the robot can get xp again like daily check-in.

**Vitality API** nowadays are still in testing and is in early stage.<br/>
We can not make sure that under this API, the robot will not get banned by NetEase.





## Prior data collection
### Summary
Based on the data collection from a large amount of one-way tests, 
it is currently indicated that multiple logins will lead to a ban.

Based on a large amount of existing statistical evidence, 
we have determined that the lockdown usually occurs between 6 a.m. and 7 a.m.



### Findings
The following data can shows the end time of the ban.<br/>
Note that these data were tallied by nv1 auth service.

```
Simple A

2025-10-26 08:12:18
2025-11-18 08:12:26
2025-10-26 08:12:27
2025-10-26 08:12:29
2025-11-18 08:12:22
2025-10-26 08:12:28
2025-10-25 07:12:13
2025-10-26 08:12:18
2025-10-26 08:12:25
2025-10-26 08:12:29
2025-10-26 08:12:24
2025-10-26 08:12:24
2025-10-26 08:12:24
2025-10-26 08:12:25
2025-10-26 08:12:25
2025-10-25 23:18:04
2025-10-26 08:12:26
2025-11-18 08:12:26
2025-10-22 15:15:09
2025-10-24 07:41:31
2025-10-25 07:12:14
2025-10-20 11:08:33
2025-11-17 22:17:03
2025-10-26 08:12:23
2025-10-26 09:13:02
2025-10-26 08:12:29
2025-10-26 08:12:30
2025-10-26 08:12:22
2025-10-26 09:38:03
2025-10-26 08:12:26
2025-10-26 09:48:02
2025-11-18 08:12:31
2025-10-26 08:12:22
2025-10-26 08:12:22
2025-10-26 08:12:23
2025-10-26 08:12:29
2025-10-26 08:12:24
2025-10-26 11:10:05
2025-10-26 08:12:24
2025-10-26 08:12:28
2025-10-26 08:12:28
2025-10-26 12:15:02
2025-11-14 15:15:12
2025-10-26 08:12:24
2025-10-26 13:41:12
2025-11-18 08:12:21
2025-10-26 08:12:24
2025-10-26 08:12:28
2025-10-26 08:12:26
2025-10-26 08:12:24
2025-10-26 08:12:25
2025-10-26 12:03:05
2025-11-01 07:40:25
2025-10-26 08:12:31
2025-10-26 08:12:31
2025-10-26 17:57:04
2025-10-26 08:12:26
2025-10-26 16:29:04
```

```
Sample B

2025-10-27 07:34:29
2025-10-27 07:38:33
2025-11-19 07:39:35
2025-10-27 07:39:27
2025-10-27 07:35:47
2025-10-27 07:39:12
2025-10-27 07:43:53
2025-10-27 07:37:51
2025-10-27 07:37:54
2025-11-19 07:44:17
2025-10-27 07:39:24
2025-10-27 07:36:59
2025-11-19 07:37:50
2025-10-27 07:25:07
2025-10-27 07:39:24
2025-11-19 07:41:23
2025-10-27 07:38:47
2025-10-27 07:36:58
2025-10-27 07:38:47
2025-10-27 07:39:32
2025-11-19 07:41:23
2025-10-27 07:11:21
2025-10-27 07:11:21
2025-10-27 07:11:23
2025-10-27 07:11:23
2025-10-27 07:11:21
2025-10-27 07:11:23
2025-11-15 08:21:16
2025-11-19 07:33:03
2025-10-27 07:39:24
2025-10-27 07:11:21
2025-10-27 07:11:21
2025-10-27 07:11:23
2025-10-27 07:11:21
2025-10-27 07:11:23
2025-10-27 07:11:21
2025-10-25 08:43:02
2025-11-19 07:41:23
2025-10-27 19:32:02
2025-11-11 11:18:00
2025-10-27 07:11:21
2025-10-27 07:11:21
2025-11-19 07:41:23
2025-10-27 07:11:21
2025-10-27 07:11:23
2025-10-27 07:11:23
2025-10-27 07:11:23
2025-10-27 07:44:14
2025-11-19 07:33:24
2025-10-26 08:12:24
2025-10-27 23:18:50
```

```
Sample C

2025-10-27 07:11:23
2025-10-27 07:11:21
2025-10-27 07:11:23
2025-10-27 07:11:21
2025-10-28 07:48:11
2025-11-20 07:48:27
2025-10-28 07:46:55
2025-11-20 07:48:54
2025-10-28 07:53:37
2025-10-28 07:53:48
2025-11-20 07:53:58
2025-10-28 07:46:55
2025-11-19 07:39:58
2025-10-28 07:50:11
2025-11-19 07:44:17
2025-10-28 07:53:22
2025-11-20 07:48:06
2025-11-19 07:33:03
2025-10-28 07:16:39
2025-10-27 07:44:11
2025-10-27 07:11:21
2025-10-27 07:11:23
2025-10-27 07:11:21
2025-11-20 07:54:01
2025-10-27 07:35:47
2025-11-20 07:52:40
2025-11-18 08:12:22
2025-11-19 07:38:33
2025-10-27 07:34:29
2025-10-28 07:48:11
2025-10-25 07:12:12
2025-11-16 23:09:02
2025-11-18 08:12:28
2025-11-18 08:12:19
2025-11-19 07:36:44
2025-10-26 12:15:02
2025-10-26 13:41:12
2025-10-27 07:38:33
2025-10-28 07:53:37
2025-10-28 07:47:52
2025-10-28 17:34:02
2025-10-28 08:34:50
2025-10-27 07:11:21
2025-10-27 07:11:23
2025-11-19 07:41:23
2025-10-25 08:43:02
2025-10-28 07:16:39
2025-11-20 07:48:27
2025-10-26 08:12:22
```

Through the following field visits and investigations, 
the research supports the possibility that multiple logins may lead to bans.

- Personal experience (Based on Eulogist/Bunker auth service; Both use personal mc account and sa auth account that sent by @Yeah114)
    - Data collection from rental server `Free City` and `落花`.
        - Due to `Free City` always get stuck, and we find that it likely to get stuck each 30 to 60 minutes (in a range). Therefore, the robot always relogin and access the same rental server each 30 to 60 minutes.
        - `落花` work as expected like most of the normal rental server. It provides some kind of comparison.

    - Debug `nemc-tan-lobby-solver` (Create lobby room multiple times in a day)
        - Few weeks ago, we are working on the codes that implements the feature to make the other people can login to the tan lobby room that start by a robot. So, to debug the codes, these days we use the same account to login to NetEase multiple times. This is not some kind of automatic testing tool, but because we start to manually test usability after completing some functions. The login frequency varies from 5 minutes to one hour, and it will occur less than 50 times (or 30 times) in a day.

    - Login each 60 minutes per day and lead banned in the next morning
        - This is a solution that used to solve the token refresh problem that related to STUN and TRUN auth server's token which sent from signaling server (We now have a better solution, so it is no longer in use). Under this solution, we login each 60 minutes and just login (no other request being made), and we observed that a ban occurred in the morning of the next day.

- Collections from other users (Based on nv1 auth service)
    - 飛飛 (2679159026; Use their personal mc account)
        - 2025/10/14
            - Use nv1's api for `1,216` times
            - Banned to **2025-11-14 15:15:12** (Happened on **2025-10-15 15:15:12**)
        - 2025/10/16
            - Use nv1's api for `954` times
            - Banned to **2025-10-24 07:39:04** (Happened on **2025-10-17 07:39:04**)
        - 2025/10/18
            - Use nv1's api for `548` times
            - Banned to **2025-10-26 08:12:22** (Happened on **2025-10-19 08:12:22**)
        - 2025/10/19
            - Use nv1's api for `14` times
                - 10/19 10:10:54 (panic: disconnectionScreen.serverFull)
                - 10/19 10:11:13 (panic: disconnectionScreen.serverFull)
                - 10/19 10:11:31 (panic: disconnectionScreen.serverFull)
                - 10/19 10:11:49 (panic: disconnectionScreen.serverFull)
                - 10/19 11:56:02 (panic: conn dead: netConn closed)
                - 10/19 12:29:53 (Terminated by user)
                - 10/19 12:30:34 (Terminated by user)
                - 10/19 13:08:56 (panic: conn dead: netConn closed)
                - 10/19 13:09:08 (panic: 验证服务器返回: 未搜索到租赁服)
                - 10/19 13:09:18 (panic: 验证服务器返回: 未搜索到租赁服)
                - 10/19 13:09:29 (panic: 验证服务器返回: 未搜索到租赁服)
                - 10/19 13:09:39 (panic: 验证服务器返回: 未搜索到租赁服)
                - 10/19 13:09:49 (panic: 验证服务器返回: 未搜索到租赁服)
                - 10/19 13:10:00 (panic: 验证服务器返回: 未搜索到租赁服)
            - Banned to **2025-11-19 07:39:58** (Happened on **2025-10-20 07:39:58**)
        - 2025/10/22
            - Use nv1's api for `6` times
            - Not banned
        - 2025/10/23 (up to 17:11:02)
            - Use nv1's api for `3` times
            - Not banned (currently unknown)
    - Other users
        - Some users only user nv1 for query the rental server information, and also get banned in the next day in the morning.

- Collections from Neo-Omega/YoRHa users
    - Only a few users get banned



### Preliminary conclusion
During the data and what going on that said above, we have a preliminary conclusion.

1. Login multiple times (the same mc account) one day will lead a high possibility that made the bot being banned in the next day in the morning.
2. Can not ensure the only use auth service but not used for access rental server will get a higher possibility that being banned than use service and also login to rental server.
3. Based on the available data, statistically speaking, bans always occur between 6 a.m. and 9 a.m.
4. Both nv1 and Eulogist/bunker can lead to banned, and this means that the IP address may not be the main reason for the ban.



### Limitations
1. This research can not completely prove whether the IP of auth service will lead a high possibility to made the bot begin banned.
2. The sample size may not be sufficient to summarize or reflect the overall situation.
3. Due to the complexity of the actual situation, the research cannot control the singularity of all possible variables, which may cause some potential problems.







# Introduction to Vitality API
## Background
Based on research that mentioned above, we have initially learned
that using the same account for multiple logins will result in a ban.

Therefore, we hope to avoid multiple logins as a priority through the following methods.
- Cache the g79 user, so if a user request API multiple times in a short time, the auth service will not request new g79 login, but use the cached one and this can solve the problem of multiple logins within a short period of time.
- Consider that login each 60 minutes will also result in being banned, so we hope keep the g79 user alive when users' robot are running.  Therefore, when the robot disconnected from rental server, and restart and request auth service API, the auth service will not request new g79 login, but use the cached one.
- Due to the robot can request auth service API to keep g79 user alive, so it can also request auth service to made the mc account get xp from online time (and this maybe can make the bot looks like a 'normal' player, but not a robot).

That is to say, the core of the solution is to suppress multiple logins by caching sessions.





## G79 user transaction
In most programming, we need to use concurrency. This means that race problems occur frequently.

Multiple users may request an API simultaneously, and each request independently carries a different token.<br/>
Therefore, you might say that requests from different users can be processed without locks.<br/>
However, that's not the case. The same user may simultaneously call the usual rental service login API and then call some other apis, which may lead to data race issues.

Specifically, we hope that the user's robot can maintain the online status of g79 users by actively requesting apis, and in this process, acquire online experience.<br/>
The implementation of this client can be regarded as a session maintainer.<br/>
In fact, there should not be multiple session maintainers, as we only need one maintainer to handle the session maintenance and online experience acquisition for one g79 user.<br/>
However, if a user uses one robot for multiple rental services, this means that the user needs to use a separate session maintainer instead of each robot using a separate session maintainer (because this is the same mc account).<br/>
Therefore, there is a possibility that the session maintainer requested the API, and at the same time, multiple other robots of this user (using the same mc account) requested the rental server login API.

So, we introduce a new concept that named **G79 User Transaction**.<br/>
**G79 User Transaction** holds a locker for each mc account (or the token). So that in the same time, either the vitality API is operating a g79 user, or another api is using a g79 user. And by do that, we can ensure the consistency of the data view.

Consider user use `FBToken` to generate request to auth service, so we can use `FBToken` as the identifier for each g79 user transaction.<br/>
Here is our implement for g79 user transaction.

```go
var activeGuTranMutex = new(sync.Mutex)
var activeGuTranMapping = make(map[string]*g79Transaction)

// g79Transaction ..
type g79Transaction struct {
	locker *sync.Mutex
	holder int
}

// newG79Transaction ..
func newG79Transaction() *g79Transaction {
	return &g79Transaction{
		locker: new(sync.Mutex),
		holder: 0,
	}
}

// LockG79Transaction locks the g79 user transaction that
// corresponding to helperToken. If target transaction not
// exist, then it will creates a new one.
func LockG79Transaction(helperToken string) {
	var g79UserLocker *g79Transaction
	var ok bool

	func() {
		activeGuTranMutex.Lock()
		defer activeGuTranMutex.Unlock()

		g79UserLocker, ok = activeGuTranMapping[helperToken]
		if !ok {
			g79UserLocker = newG79Transaction()
			activeGuTranMapping[helperToken] = g79UserLocker
		}
		g79UserLocker.holder++
	}()

	g79UserLocker.locker.Lock()
}

// UnlockG79Transaction unlocks the g79 user transaction
// that corresponding to helperToken. If target transaction
// not exist, or the underlying locker is already unlocked,
// then this func will be panic.
func UnlockG79Transaction(helperToken string) {
	activeGuTranMutex.Lock()
	defer activeGuTranMutex.Unlock()

	g79UserLocker, ok := activeGuTranMapping[helperToken]
	if !ok {
		// We should panic here because when here is not ok,
		// it means somewhere may have some internal error,
		// and the lock states is not completely.
		panic(fmt.Sprintf("UnlockG79Transaction: Transaction %#v not found", helperToken))
	}

	g79UserLocker.holder--
	if g79UserLocker.holder == 0 {
		delete(activeGuTranMapping, helperToken)
	}
	g79UserLocker.locker.Unlock()
}
```

You can see that we identify each transaction through `helperToken`.

By assuming that multiple acquiters who want to acquire the lock will eventually all request to unlock the mutex lock, the number of threads whose lock is acquired can be recorded to check whether the transaction can be deleted from the underlying mapping.<br/>
That means, once a lock is required, add 1 to the holder count of this lock. And once this thread unlock this lock, then minus 1 to the holder count. Once the holder count is equal to 0, then the lock can delete from the underlying map so that we avoid the leak of memory.<br/>
Simply put, if multiple requests from a user have been completed, the underlying mapping should remove the relevant locks to free up memory. And the above-mentioned code fulfills this function.

This means that each user is assigned an independent lock, which means that requests from different users are concurrent, and blocking is only for each user themselves. Thus, we have taken into account a consistent view of the data while ensuring performance.





## Session
### Data structure
**Session** keeps an active g79 user, and also holds the following fields.
- Session ID (UUID; string)
- Session Type (constant enumerate; uint8)
- Session Start Time (unix time; int64)
- Session Expire Time (unix time; int64)



### Description
One session is specific to one `FBToken` rather than the user themselves.<br/>
This means that if a user has multiple auth server helpers, each helper can hold a session.

Due to one g79 user's life time only have 40 minutes, therefore, we define the lifecycle of a single session as `30` minutes.<br/>
Client can use **Vitality API** to control the session. Also, when client request rental server login (or do other things like request Open API), the internal implements should first try load session that corresponding to the auth server helper.<br/>
If not found or session is expired, create a new session and associate this session with the g79 user. Then, use this g79 user to finish left things, like login to rental server or query rental server status.

That means, you always use g79 user that come from **Session**, but not do g79 login for each request that made by the users.<br/>
Note that you should save the session and active g79 user to the underlying **database** rather than save them in the memory.<br/>
This means that even if the auth server crashes and restarts, the session and active g79 user will not be revoked but will continue.

User can extend the life time of one session by **Vitality API**. By doing this, the internal implements refresh the g79 user's token,
to keep the g79 user alive, so that the corresponding session can have more time to alive.<br/>
Please note that we stipulate that when user request extend the life time of his session, the time between request time and `Session Start Time` must not exceed `12` hours.



### Session type
Session type is used to store the type of this session. The constant enumerate are as follows.

| Name                | Data Type | Value |
| ------------------- | --------- | ----- |
| SessionTypeMpayUser | uint8     | 0     |
| SessionTypePeAuth   | uint8     | 1     |
| SessionTypeSaAuth   | uint8     | 2     |

That means, if this session is used `Pe Auth` login, then `Session Type` is `SessionTypePeAuth(1)`.<br/>
Or, if this session is under `Sa Auth`, then then `Session Type` is `SessionTypeSaAuth(2)`.<br/>
Otherwise, `Session Type` is `SessionTypeMpayUser(0)`.





## Vitality API
### Register Active G79 User
Summary
> **Register Active G79 User** registers a session that associated with a specific g79 user login.


Basic Info
> | Entry       | Value                            |
> | ----------- | -------------------------------- |
> | Method      | POST                             |
> | URL         | /vitality_api/registry_active_gu |
> | ContentType | application/json                 |
> | Response    | JSON                             |


Client Request
> | Key                   | Data Type |
> | --------------------- | --------- |
> | token                 | string    |
> | override_session      | bool      |
> | provided_pe_auth_data | string    |
> | provided_sa_auth_data | string    |


Server Response
> | Key                 | Data Type |
> | ------------------- | --------- |
> | error_info          | string    |
> | success             | bool      |
> | session_id          | string    |
> | session_type        | uint8     |
> | session_expire_time | int64     |


Description
> When client send this request, if `override_session` is `false`, and this auth server helper alread have a **Session** that is not expired, then response `session_id, session_type` and `session_expire_time` of the corresponding **Session**.
> 
> Otherwise, if `provided_pe_auth_data` and `provided_sa_auth_data` both is empty, then use mpay user info from database (that corresponding to this auth server helper) to do g79 user login.<br/>
> After login is finished, creates a session and associate this g79 user with the newly created session.
> 
> Otherwise, the user provide non-empty `provided_pe_auth_data` or `provided_sa_auth_data`.<br/>
> You use one of these strings to do g79 user login, and creates a session and associate this g79 user with the newly created session.
> 
> Note that
> - If `override_session` is `false`, please **not** change the session expire (unix) time.
> - If `override_session` is `true`, you should always do g79 user login, creates new session and store related information.
> - You should always returns `session_id, session_type` and `session_expire_time` if you meet none error.
> - Ensure when `provided_pe_auth_data` or `provided_sa_auth_data` is not empty, only one of them is non-empty.
>
> For the meaning of `session_id, session_type` and `session_expire_time`, please refer to [Session](#session).<br/>
> If meet error, then `success` will be `false`, and `error_info` will carry out the specific error information.<br/>
> Otherwise, `success` is `true` and `error_info` is empty.



### Request Session Info
Summary
> **Request Session Info** returns the corresponding session info of the user's auth server helper.


Basic Info
> | Entry       | Value                              |
> | ----------- | ---------------------------------- |
> | Method      | POST                               |
> | URL         | /vitality_api/request_session_info |
> | ContentType | application/json                   |
> | Response    | JSON                               |


Client Request
> | Key   | Data Type |
> | ----- | --------- |
> | token | string    |


Server Response
> | Key                 | Data Type |
> | ------------------- | --------- |
> | error_info          | string    |
> | success             | bool      |
> | response_type       | string    |
> | session_id          | string    |
> | session_type        | uint8     |
> | session_expire_time | int64     |


Constant Enumerate
> | Name                    | Data Type | Value |
> | ----------------------- | --------- | ----- |
> | ResponseTypeFindSession | uint8     | 0     |
> | ResponseTypeNoSession   | uint8     | 1     |


Description
> If current auth server helper can find its corresponding **Session** and this session is not expired, then response corresponding `session_id, session_type, session_expire_time` with `response_type` which is `ResponseTypeFindSession(0)`.<br/>
> Otherwise, response a `response_type` which is `ResponseTypeNoSession(1)`.
>
> For the meaning of `session_id, session_type` and `session_expire_time`, please refer to [Session](#session).<br/>
> If meet error, then `success` will be `false`, and `error_info` will carry out the specific error information.<br/>
> Otherwise, `success` is `true` and `error_info` is empty.



### Clean Up Session
Summary
> **Clean Up Session** deletes corresponding **Session** of the user's auth server helper from underlying database.<br/>
> If target session is expired, or the session is not found, then result in **NOP** (no operation).


Basic Info
> | Entry       | Value                          |
> | ----------- | ------------------------------ |
> | Method      | POST                           |
> | URL         | /vitality_api/clean_up_session |
> | ContentType | application/json               |
> | Response    | JSON                           |


Client Request
> | Key        | Data Type |
> | ---------- | --------- |
> | token      | string    |
> | session_id | string    |


Server Response
> | Key        | Data Type |
> | ---------- | --------- |
> | error_info | string    |
> | success    | bool      |


Description
> If current auth server helper already have a session and is not expired, then delete this session from underlying database.
>
> If provided `session_id` not matched the one that recorded in database of the auth server, response an error.<br/>
> If the **Session** is expired, also response an error.
>
> For the meaning of `session_id`, please refer to [Session](#session).<br/>
> If meet error, then `success` will be `false`, and `error_info` will carry out the specific error information.<br/>
> Otherwise, `success` is `true` and `error_info` is empty.



### Request Daily Growth
Summary
> **Request Daily Growth** requests NetEase server, and returns the XP gained today through staying online and recharging NetEase.


Basic Info
> | Entry       | Value                              |
> | ----------- | ---------------------------------- |
> | Method      | POST                               |
> | URL         | /vitality_api/request_daily_growth |
> | ContentType | application/json                   |
> | Response    | JSON                               |


Client Request
> | Key        | Data Type |
> | ---------- | --------- |
> | token      | string    |
> | session_id | string    |


Server Response
> | Key              | Data Type |
> | ---------------- | --------- |
> | error_info       | string    |
> | success          | bool      |
> | xp_from_online   | int       |
> | xp_from_recharge | int       |


Description
> The auth server first load the **Session** that corresponding to the auth server helper.<br/>
> Then load the corresponding active g79 user that corresponding to this **Session**.
> 
> If provided `session_id` not matched the one that recorded in database of the auth server, response an error.<br/>
> If the **Session** is expired, also response an error.<br/>
> Otherwise, the session is still alive, so please use its active g79 user to request daily growth.<br/>
> 
> A go implements is like follows.
> ```go
> activeGu.RecordG79UserData.CreateHttpClient().
>     SetMethod(http.MethodPost).
>  	SetUrl(gameinfo.G79ServerList.ApiGatewayUrl + "/pe-get-daily-growth-info").
> 	SetTokenMode(g79.TOKEN_MODE_NORMAL).
> 	Do()
> ```
>
> The g79 server response is as follows (this is not a completely one, just show what data you needed).
> ```json5
> {
> 	"entity": {
> 		"1": 0, // xp_from_online
> 		"2": 0, // xp_from_recharge
> 	}
> }
> ```
>
> For the meaning of `session_id`, please refer to [Session](#session).<br/>
> If meet error, then `success` will be `false`, and `error_info` will carry out the specific error information.<br/>
> Otherwise, `success` is `true` and `error_info` is empty.



### Get Currency Online
Summary
> **Get Currency Online** requests NetEase server to make the mc account online.<br/>
> Or in other words, let the server count the online time of this mc account.


Basic Info
> | Entry       | Value                             |
> | ----------- | --------------------------------- |
> | Method      | POST                              |
> | URL         | /vitality_api/get_currency_online |
> | ContentType | application/json                  |
> | Response    | JSON                              |


Client Request
> | Key        | Data Type |
> | ---------- | --------- |
> | token      | string    |
> | session_id | string    |


Server Response
> | Key                | Data Type |
> | ------------------ | --------- |
> | error_info         | string    |
> | success            | bool      |
> | rest_currency_time | int       |
> | format_date_string | str       |


Description
> The auth server first load the **Session** that corresponding to the auth server helper.<br/>
> Then load the corresponding active g79 user that corresponding to this **Session**.
> 
> If provided `session_id` not matched the one that recorded in database of the auth server, response an error.<br/>
> If the **Session** is expired, also response an error.<br/>
> Otherwise, the session is still alive, so please use its active g79 user to get currency online.<br/>
> 
> A go implements is like follows.
> ```go
> activeGu.RecordG79UserData.CreateHttpClient().
> 	SetMethod(http.MethodPost).
> 	SetUrl(gameinfo.G79ServerList.ApiGatewayUrl + "/get-currency-online").
> 	SetRawBody([]byte("{}")).
> 	SetTokenMode(g79.TOKEN_MODE_NORMAL).
> 	Do()
> ```
>
> The g79 server response is as follows (this is not a completely one, just show what data you needed).
> ```json5
> {
> 	"entity": {
> 		"rest_currency_time": 0, 	  // rest_currency_time
> 		"date": "2000-01-01 00:00:00" // format_date_string
> 	}
> }
> ```
>
> For the meaning of `session_id`, please refer to [Session](#session).<br/>
> If meet error, then `success` will be `false`, and `error_info` will carry out the specific error information.<br/>
> Otherwise, `success` is `true` and `error_info` is empty.



### Keep G79 User Alive
Summary
> **Keep G79 User Alive** extends the life time for the corresponding **Session** by refresh the g79 user's token.


Basic Info
> | Entry       | Value                       |
> | ----------- | --------------------------- |
> | Method      | POST                        |
> | URL         | /vitality_api/keep_gu_alive |
> | ContentType | application/json            |
> | Response    | JSON                        |


Client Request
> | Key        | Data Type |
> | ---------- | --------- |
> | token      | string    |
> | session_id | string    |


Server Response
> | Key                 | Data Type |
> | ------------------- | --------- |
> | error_type          | uint8     |
> | error_info          | string    |
> | success             | bool      |
> | session_expire_time | int64     |


Constant Enumerate
> | Name                      | Data Type | Value |
> | ------------------------- | --------- | ----- |
> | KeepGuAliveErrorMeetError | uint8     | 0     |
> | KeepGuAliveErrorLifeLimit | uint8     | 1     |


Description
> The auth server first load the **Session** that corresponding to the auth server helper.<br/>
> Then load the corresponding active g79 user that corresponding to this **Session**.
> 
> If provided `session_id` not matched the one that recorded in database of the auth server, response an error.<br/>
> If the **Session** is expired, also response an error.<br/>
> 
> If the time between request time and `Session Expire Time` exceed `12` hours, then response empty `error_info` with `error_type` which is `KeepGuAliveErrorLifeLimit(1)`.<br/>
> Otherwise, the session is still alive, so please use its active g79 user to keep g79 user alive.<br/>
> 
> A go implements is like follows.
> ```go
> gu.CreateHttpClient().
> 	SetMethod(http.MethodPost).
> 	SetUrl(gameinfo.G79ServerList.CoreServerUrl + "/authentication/update").
> 	SetTokenMode(TOKEN_MODE_NORMAL).
> 	SetEncryptSuffix(0x4).
> 	Do()
> ```
>
> The g79 server response is as follows (this is not a completely one, just show what data you needed).
> ```json5
> {
> 	"entity": {
> 		"token": "" // New token of this g79 user
> 	}
> }
> ```
>
> After you extend the life time of the g79 user, please extend the corresponding **Session** for `30` minutes.<br/>
> Then, update the info of active g79 user and active **Session** to underlying database.
> 
> If meet any error but is not `KeepGuAliveErrorLifeLimit`, set `error_type` to `KeepGuAliveErrorMeetError(0)`, then specific a non empty `error_info` and set `success` to `false`.<br/>
> Or everything is work as expected, the life time is successfully to extend, then response empty `error_info` and set `success` to `true`.





## Changes in rental server login response
Most of the users use their robot just in one rental server. So, its very convenient to make the robot itself to start and running a **Session Maintainer** that used to maintain a **Session**.

However, some users use a robot for multiple rental servers. And due to multiple **Session Maintainer** is not allowed (due to this may lead to greater bandwidth consumption and an overall abnormal request frequency), therefore they possibly use an independent maintainer.

To align with these two groups of users, you can allow the user to set whether to running **Session Maintainer** in client side automatically when they use the robot.<br/>
You can show this settings in the website of your auth service, or anywhere that easy for user to access and manage.

Therefore, we add a new field to `/phoenix/login` (the response of auth server) that shows as follows.
Note that this field is used to notify current rental server login need running a maintainer or not.

| Key             | Data Type |
| --------------- | --------- |
| enable_vitality | bool      |

It's necessary to say is that no matter what value you set for `enable_vitality`, the users can always use running a maintainer that under **Vitality API** protocol.<br/>
This field (enable_vitality) is just to notify whether the robot (or the access point) should running a maintainer automatically or not.





## Implementation of Session Maintainer
**Session Maintainer** first use [Register Active G79 User](#register-active-g79-user) to register a new session (or used a session that is not expired).

Then, start `2` threads.
- For the first thread, it refresh the session each `20` minutes by use [Keep G79 User Alive](#keep-g79-user-alive). Note that it should compute the sleep time due to the response of [Register Active G79 User](#register-active-g79-user) and [Keep G79 User Alive](#keep-g79-user-alive) will told the expire unix time of current session.
- For the second thread, it calls [Get Currency Online](#get-currency-online) each `5` minutes.

If any of a thread get error from auth server, then it means that this session is possibly expired or deleted.<br/>
Then, the program should stop these `2` threads, and register a new session and start another `2` threads.

If get error when register new session, then the program should panic and stop to running anymore.<br/>
In this case, the corresponding g79 user may get banned from NetEase, or the `FBToken` is expired.

For a possible implements, see [cmd/main.go](/cmd/main.go) for more information.





## Adaptation
**Eulogist/Bunker** and **Eulogist Community** nowadays are completed the adaptation to **Vitality API**.<br/>
For **YoRHa/Bunker** and **neomega-core**, we plan to complete the adaptation in the near future.

The implementation of other access points, other robots, some building importers, and some auth servers, should follow this documents to finish all adaptation if needed.<br/>
In addition, **Vitality API** is an optional feature, and it is compatible with the auth service protocol of version `v1`.

Note that **Vitality API** is still in early stage.<br/>
We can not ensure that everything will keep the same in the future.

One more thing, if you are considered to adapt to this new API, please be vigilant about the implementation of **G79 User Transaction**.<br/>
Please implement the mutex lock carefully, eliminate all possible conditional races, and try your best to avoid the occurrence of race state problems.
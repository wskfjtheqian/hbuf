# HBUF

库地址 [https://github.com/wskfjtheqian/hbuf](https://github.com/wskfjtheqian/hbuf)  
Idea 开发插件源码  [https://github.com/wskfjtheqian/hbuf_idea](https://github.com/wskfjtheqian/hbuf_idea)

### 功能表

#### 一、为不同语言生成统一的数据结构体、序列化和返回序列方案

| 语言   | golang | dart | java | typescript | C | C# |
|------|--------|------|------|------------|---|----|
| 结构   | 完成     | 完成   | 完成   | -          | - | -  |
| JSON | 完成     | 完成   | 完成   | -          | - | -  |
| 二进制  | -      | -    | -    | -          | - | -  |

#### 二、为不同语言生成统一的RPC接口，支持函数和广播

| 语言 | golang | dart | java | javascript | C | C# |
|----|--------|------|------|------------|---|----|
| 函数 | 函数（服务） | 函数   | -    | 函数         | - | -  |
| 广播 | -      | -    | -    | -          | - | -  |

#### 三、生成Sql语句和调用函数

| 语言      | golang | 功能说明                    |
|---------|--------|-------------------------|
| Insert  | 完成     | 插入单条数据（null数据不插入）       |
| Inserts | 完成     | 插入多条数据（null数据也将插入）      |
| Update  | 完成     | 更新数据（null数据不更新）         |
| Set     | 完成     | 更新数据（null数据也将更新）        |
| Del     | 完成     | 删除数据（伪删除将delete_time设置为非空） |
| Remove  | 完成     | 真实删除                    |
| Get     | 完成     | 获得单条数据                  |
| List    | 完成     | 获得多条数据                  |
| Map     | 完成     | 获得多条数据并生成MAP数据          |
| Count   | 完成     | 统计数据                    |

#### 四、生成正则表达式表单验证

| 语言 | golang | dart | 
|----|--------|------| 
| -  | 完成     | 完成   |

#### 五、生成表格和表单UI

| 语言 | dart(flutter) |
|----|---------------| 
| 表格 | 完成            | 
| 表单 | 完成部分          | 

#### 多语言文本支持

| 语言      | 说明 |
|---------|----| 
| Flutter | 完成 | 
| Golang  | -  | 

### 对应语言库

| 语言         | 库地址                                                                                          | 说明                  |
|------------|----------------------------------------------------------------------------------------------|---------------------|
| golang     | [https://github.com/wskfjtheqian/hbuf_golang](https://github.com/wskfjtheqian/hbuf_golang)   | Golang 服务开发库        |
| dart       | [https://github.com/wskfjtheqian/hbuf_dart](https://github.com/wskfjtheqian/hbuf_dart)       | Dart RPC 接口调用库      |
| java       | [https://github.com/wskfjtheqian/hbuf_java](https://github.com/wskfjtheqian/hbuf_java)       | JAVA RPC 接口调用库      |
| flutter    | [https://github.com/wskfjtheqian/hbuf_flutter](https://github.com/wskfjtheqian/hbuf_flutter) | Flutter GM客户端开发库    |
| typescript | [https://github.com/wskfjtheqian/hbuf_ts](https://github.com/wskfjtheqian/hbuf_ts)           | TypeSscript RPC端开发库 |

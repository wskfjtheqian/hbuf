# HBUF 语言说明

### 功能表说明

#### 一、包名

package 语言 = "包名"

|   语言    | 说明  |
|:-------:|:---:|
| Flutter | 不支持 |
| Golang  | 支持  |
|  Java   | 支持  |

```hbuf
package go = "admin"
package java = "com.hbuf.admon"
```

#### 二、引入

import "文件名.hbuf"

```hbuf
import "public.hbuf"
```

#### 三、枚举

enum 枚举名 {  
&nbsp; &nbsp; 字段名 = ID  
}

```hbuf
enum Status {
    player = 0

    stop = 1
}
```

#### 四、数据

data 数据名:父数据 = ID {  
&nbsp; &nbsp; int32 字段名 = ID  
}

```hbuf
data Base = 0 {
    int64 id = 0

    string name = 1

    int32 age = 2
}

data Class = 1 {
    string? class_name = 0

    int32? class_no = 1
}

data Student:Base, Class = 2 {
    int32 no = 0
}
```

#### 五、服务

data 服务名:父服务 = ID {  
&nbsp; &nbsp; 返回数据 函数名（请求数据 参数名）= ID  
&nbsp; &nbsp; 方法名（请求数据 参数名）= ID  
}

```hbuf
data GetBaseReps = 0 {
    Base info = 0
}

data GetBaseReq = 1 {
    int64 id = 0
}

server BaseServer = 0 {
    GetBaseReps GetBase(GetBaseReq req) = 0
}

data GetClassReps = 2 {
    Class info = 0
}

data GetClassReq = 3 {
    int64 id = 0
}

data MessageReq = 4 {
    int64 id = 0
    
    string message = 1
}

server StudentServer:BaseServer = 0 {
    GetClassReps GetClass(GetClassReq req) = 0
    
    SendMessage(MessageReq req) = 1
}

```

#### 六、注解



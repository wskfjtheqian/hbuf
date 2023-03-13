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
package go="admin"
package java="com.hbuf.admon"
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






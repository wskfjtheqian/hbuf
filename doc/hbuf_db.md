# HBUF 语言说明

### 功能表说明

#### 一、包名

package 语言 = "包名"

|    语言     |     功能      |         说明          |        示例         |
|:---------:|:-----------:|:-------------------:|:-----------------:|
|   index   |     int     |                     |                   |
|   name    |   表名或字段名    |                     |                   |
|    key    |     主键      |                     |                   |
|   force   |    强制更新     |                     |                   |
|    typ    |   string    |                     |                   |
|  insert   | 生成插入单条数据函数  |                     |                   |
|  inserts  | 生成插入多表数据函数  |                     |                   |
|  update   |   生成更新函数    |                     |                   |
|    set    |   生成设置值函数   |                     |                   |
|    del    |   生成删除数据    |                     |                   |
|    get    | 生成获得但单条数据函数 |                     |                   |
|   list    |  生成列表列表函数   |                     |                   |
|    map    |   生成集合函数    |                     |                   |
|   count   |   生成统计函数    |                     |   count="true"    |
|   table   |    关联表结构    |                     |                   |
|    rm     |  生成永久删除函数   |                     |     rm="true"     |
|   where   |     条件      |                     |                   |
|  offset   |     偏移量     |                     |                   |
|   limit   |     数量      |                     |                   |
|   order   |     排序      |                     | order="id $ DESC" |
|   group   |     分组      |                     |  group="id, age"  |
| converter |     转换器     | 数据类型和Golang 类型的相互转换 | converter="json"  |

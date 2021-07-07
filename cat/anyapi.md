
Api:  
/pub/stock/any_apis 
详见swag 文件档：
http://62.234.169.68:8001/docs/index.html


### jsonPath
jsonPath　用于指定从json路径中提取数据的规则，仅限提取单一字段值．

#### demo
如以下json 数据
```json
{"data":{
  "name": {"first": "Tom", "last": "Anderson"},
  "age":37,
  "children": ["Sara","Alex","Jack"],
  "fav.movie": "Deer Hunter",
  "friends": [
    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
  ]}
}
```

提取数据的jsonPath及可以获取到的值列表如下
```bash
"data.name.last"          >> "Anderson"
"data.age"                >> 37
"data.children"           >> ["Sara","Alex","Jack"]
"data.children.#"         >> 3
"data.children.1"         >> "Alex"
"data.child*.2"           >> "Jack"
"data.c?ildren.0"         >> "Sara"
"data.fav\.movie"         >> "Deer Hunter"
"data.friends.#.first"    >> ["Dale","Roger","Jane"]
"data.friends.1.last"     >> "Craig"
```


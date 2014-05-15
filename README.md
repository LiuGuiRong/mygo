# mygo #

##mygo 是什么？

一个数据库单向导入工具。mysql -> mongo

## 如何使用

* 1. 配置 ./conf/db.ini, 主要包含目标数据库的配置，如账户，密码，使用端口等。

* 2. 启动mygo.bat， 便可按照提示进行数据库导入。

##注意事项

* 1. 您可以输入多个或者单个mysql表明进行导入，多个导入时，请以逗号分割，如: bufftable,skilltable,    

* 2. mygo会默认检查mysql需要导入表的格式，如果有没有初始化的字段，则会警告。此警告开关默认开启。

* 3. 便捷的全部mysql表导入待添加。

##有问题反馈

在使用中有任何问题，欢迎反馈给我，可以用以下联系方式跟我交流

##关于作者

```

    author: liuguirong
    email:hroomkeeper@gmail.com
    qq: 16839242
    web: oxrush.com

```

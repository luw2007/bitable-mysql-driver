# bitable://cli_a14eda43cb7ad013:l5zyi***********************16Y0@open.feishu.cn/bascnQIrLs6MrhIvftGsdYJgRFd?log_level=trace
use bascnQIrLs6MrhIvftGsdYJgRFd;

# table1 tblHTHTelIVwIeqN
CREATE TABLE `单向关联`
(
    ID text
) COMMENT '视图';
# table2 tblmoAekspwLGzEt
CREATE TABLE `双向关联`
(
    ID text
) COMMENT '视图';
# table3 tblZAuXVNjvK3PE6
CREATE TABLE `测试读取-请勿修改`
(
    `ID`     text,
    `数字`     int COMMENT '{"formatter":"0"}',
    `单选`     varchar(3) COMMENT '{"options":[{"name":"是"},{"name":"否"}]}',
    `多选`     varchar(4) COMMENT '{"options":[{"name":"选项一"},{"name":"选项二"},{"name":"选项三"}]}',
    `日期`     varchar(5) COMMENT '{"auto_fill":true,"date_formatter":"yyyy-MM-dd"}',
    `复选框`    varchar(7),
    `人员`     varchar(11) COMMENT '{"multiple":true}',
    `超链接`    varchar(15),
    `附件`     varchar(17),
    `公式`     varchar(20) COMMENT '{"formatter":""}',
    `创建时间`   varchar(1001),
    `最后更新时间` varchar(1002),
    `创建人`    varchar(1003),
    `修改人`    varchar(1004)
) COMMENT '总表';

ALTER TABLE `<table1_id>`
    ADD COLUMN `单向关联` varchar(18) COMMENT '{"table_id":"<table1_id>"}';
ALTER TABLE `<table1_1>`
    ADD COLUMN `双向关联` varchar(21) COMMENT '{"multiple":true,"table_id":"<table2_id>","back_field_name":"双向关联"}';


# <table1_1> 执行完语句后可以查询得到
# 如：
# ALTER TABLE `tblZAuXVNjvK3PE6` ADD COLUMN `单向关联` varchar(18) COMMENT '{"table_id":"tblHTHTelIVwIeqN"}';
# ALTER TABLE `tblZAuXVNjvK3PE6` ADD COLUMN `双向关联` varchar(21) COMMENT '{"multiple":true,"table_id":"tblmoAekspwLGzEt","back_field_name":"双向关联"}';

# 引用查找 不知道如何添加


# 插入数据
# <user1_open_id> 用户开放平台id，可以通过添加一个表单，然后 select * from table 查看。
# INSERT INTO `tblZAuXVNjvK3PE6` (`ID`,persons.`人员`,`单选`,options.`多选`,`数字`,`日期`,url.`超链接`)
#   VALUES ('F1','[{"id":"<user1_open_id>"}]','是','["选项一"]',3,1640268982000,'{"link": "https://bilibili.com","text":"bilibili"}');

INSERT INTO `<table1_id>` (`ID`, persons.`人员`, `公式`, `单选`, options.`多选`, `数字`, `日期`, url.`超链接`)
VALUES ('F1', '[{"id":"ou_ca10ae5af227e61f3a58b02e2695877c"},{"id":"ou_40eff67041d3f4902782d867e3edcc5d"}]', '0.0', '是',
        '["选项一"]', 1, 1640268982000, '{"link": "https://bilibili.com","text":"bilibili"}'),
       ('F2', '[]', '0.0', '是', '["选项二"]', 3, 1650268982000, '{"link": "https://z.cn","text":"z"}'),
       ('F3', '[]', '0.0', '是', '["选项三"]', 9, 1660270450000, '{"link": "https://bili.com","text":"bili"}');

INSERT INTO `<table1_id>` (`ID`)
VALUES ('F4');
# table4 tbl9GhbLx2AyQTTd
CREATE TABLE `测试写入-请勿修改`
(
    `多行文本`   text,
    `数字`     int COMMENT '{"formatter":"0"}',
    `单选`     varchar(3) COMMENT '{"options":[{"name":"是"},{"name":"否"}]}',
    `多选`     varchar(4) COMMENT '{"options":[{"name":"选项一"},{"name":"选项二"},{"name":"选项三"}]}',
    `日期`     varchar(5) COMMENT '{"auto_fill":true,"date_formatter":"yyyy-MM-dd"}',
    `复选框`    varchar(7),
    `人员`     varchar(11) COMMENT '{"multiple":true}',
    `超链接`    varchar(15),
    `附件`     varchar(17),
    `公式`     varchar(20) COMMENT '{"formatter":""}',
    `创建时间`   varchar(1001),
    `最后更新时间` varchar(1002),
    `创建人`    varchar(1003),
    `修改人`    varchar(1004)
) COMMENT '总表';






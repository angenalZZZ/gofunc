//通用扩展方法
Date.prototype.Add = function (seconds) { var t = new Date(); t.setTime(this.getTime() + seconds * 1000); return t; };
Date.prototype.AddDate = Date.prototype.AddDays = function (days) { var t = new Date(); t.setTime(this.getTime() + days * 24 * 3600 * 1000); return t; };
Date.prototype.Date = function () { return this.toISOString().split("T")[0]; };
Date.prototype.Time = function () { return this.toISOString().split("T")[1].split(".")[0]; };
Date.prototype.DateTime = function () { return this.toISOString().replace("T", " ").split(".")[0]; };
String.prototype.Date = function () { return this.replace("T", " ").split(" ")[0]; };
String.prototype.Time = function () { return this.replace("T", " ").split(" ")[1].split(".")[0]; };
String.prototype.DateTime = function () { return this.replace("T", " ").split(".")[0]; };

//配置订阅任务
subscribe = [
    {
        name: "001", // 订阅名(注："全局订阅前缀"参考yaml配置文件;如果未设置特殊符号,"完整订阅名"=name)
        spec: "+", // 特殊符号(注："+"表示"完整订阅名"="全局订阅前缀"+name)
        func: function (records) {
            if (!records || records.constructor.name != "Array") return "";
            var items = records.filter(function (item) { return item.constructor.name == "Object" && item.hasOwnProperty("Code"); });
            if (items.length == 0) return "";
            return "insert into logtest(Code,Type,Message,Account,CreateTime) values"
                + items.map(function (item) {
                    return "("
                        + "'" + item.Code.replace("'", "") + "',"
                        + item.Type + ","
                        + "'" + item.Message.replace("'", "''") + "',"
                        + "'" + item.Account + "',"
                        + "'" + item.CreateTime.replace("T", " ").split(".")[0] + "'"
                        + ")";
                }).join(",") + ";";
        }
    },
];

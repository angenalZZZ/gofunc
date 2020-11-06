//通用扩展方法
Date.prototype.Add = function (seconds) { var t = new Date(); t.setTime(this.getTime() + seconds * 1000); return t; };
Date.prototype.AddDate = Date.prototype.AddDays = function (days) { var t = new Date(); t.setTime(this.getTime() + days * 24 * 3600 * 1000); return t; };
Date.prototype.Date = function () { return this.toISOString().split("T")[0]; };
Date.prototype.Time = function () { return this.toISOString().split("T")[1].split(".")[0]; };
Date.prototype.DateTime = function () { return this.toISOString().replace("T", " ").split(".")[0]; };
String.prototype.Date = function () { return this.replace("T", " ").split(" ")[0]; };
String.prototype.Time = function () { return this.replace("T", " ").split(" ")[1].split(".")[0]; };
String.prototype.DateTime = function () { return this.replace("T", " ").split(".")[0]; };

//配置计划任务
cron = [
    {
        name: "001",
        spec: "* * * * *", // every minutes
        func: function () {
            $.trace = true; // 反代接口 URL调试
            var bSubject = nats.subject + "-"; // 反代消息 注意后缀
            var item = { Subject: bSubject, Time: new Date().Time() };
            item.ActionName = '反代接口';
            var res = $.q("post", "https://postman-echo.com/post", item, "url");
            // dump(res);
        }
    },
];

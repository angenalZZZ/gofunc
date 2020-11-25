//扩展方法
// Date.prototype.Add = function (seconds) { var t = new Date(); t.setTime(this.getTime() + seconds * 1000); return t; };
// Date.prototype.AddDate = Date.prototype.AddDays = function (days) { var t = new Date(); t.setTime(this.getTime() + days * 24 * 3600 * 1000); return t; };
// Date.prototype.Date = function () { return this.toISOString().split("T")[0]; };
// Date.prototype.Time = function () { return this.toISOString().split("T")[1].split(".")[0]; };
// Date.prototype.DateTime = function () { return this.toISOString().replace("T", " ").split(".")[0]; };
// String.prototype.Date = function () { return this.replace("T", " ").split(" ")[0]; };
// String.prototype.Time = function () { return this.replace("T", " ").split(" ")[1].split(".")[0]; };
String.prototype.DateTime = function () { return this.replace("T", " ").split(".")[0]; };
function col(s) { if (!s) { return "NULL"; } return "'" + s.replace("'", "''") + "'"; };

//订阅处理
function sql(records) {
    if (!records || records.constructor.name != "Array") return;
    var items = records.filter(function (item) { return item.constructor.name == "Object" && item.hasOwnProperty("Code"); });
    if (items.length == 0) return;
    var sqlString = "insert into logtest(Type,Code,Message,Account,CreateTime,CreateUser) values"
        + items.map(function (item) {
            return "("
                + item.Type + ","
                + col(item.Code) + ","
                + col(item.Message) + ","
                + col(item.Account) + ","
                + col(item.CreateTime.DateTime()) + ","
                + col(item.CreateUser) + ")";
        }).join(",") + ";";
    // console.log(sqlString);
    return sqlString;
}

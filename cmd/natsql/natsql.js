function sql(records) {
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

//配置订阅任务
subscribe = [
    {
        name: "001", // 订阅名(注："全局订阅前缀"参考yaml配置文件;如果未设置特殊符号,"完整订阅名"=name)
        spec: "+", // 特殊符号(注："+"表示"完整订阅名"="全局订阅前缀"+name 例如："test-001")
        func: function () { return this.name; } // function(records)订阅处理/所在目录(默认=name)
    },
];

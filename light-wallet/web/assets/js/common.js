var Common = {
    host: 'http://127.0.0.1:2345',
    app:{},

    baseDecimal:new BigNumber(10).pow(18),

    init: function () {
        var that = this;
        that.app.init()
    },


    post: function (_method, _biz, _page, callback) {
        var that = this;

        var result = new Object();
        var timestamp = 1234567;
        var sign = "67ff54447b89f06fe4408b89902e585167abad291ec41118167017925e24e320";
        var data = {
            base: {
                timestamp: timestamp,
                sign: sign,
            },
            biz: _biz,
            page: _page,
        }

        $.ajax({
            url: that.host + '/' + _method,
            type: 'post',
            dataType: 'json',
            async: false,
            data: JSON.stringify(data),
            beforeSend: function () {
            },
            success: function (res) {
                if (callback) {
                    callback(res)
                }
            }
        })

        return result;
    },


}

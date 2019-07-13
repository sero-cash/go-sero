var Index = {

    init:function () {
        var that = this;
        that.getAccountList();
        that.getBlockHeight();
        setInterval(function () {
            that.getAccountList();
        },10000);

        setInterval(function () {
            that.getBlockHeight();
        },10000);
    },

    loadProperties: function (type) {
        jQuery.i18n.properties({
            name: 'lang', // 资源文件名称
            path: 'assets/i18n/', // 资源文件所在目录路径
            mode: 'map', // 模式：变量或 Map
            language: type, // 对应的语言
            cache: false,
            encoding: 'UTF-8',
            callback: function () { // 回调方法
                $('._back').text($.i18n.prop('back'));
                $('.h_title').text($.i18n.prop('backpack'));
                $('.pack_list ul:eq(0) li:eq(0)').text($.i18n.prop('item'));
                $('.pack_list ul:eq(0) li:eq(1)').text($.i18n.prop('amount'));
                $('.pack_list ul:eq(0) li:eq(2)').text($.i18n.prop('operate'));
                $('.nav_sub li:eq(0)').text($.i18n.prop('home'));
                $('.nav_sub li:eq(1)').text($.i18n.prop('battle'));
                $('.nav_sub li:eq(2)').text($.i18n.prop('backpack'));
                $('.nav_sub li:eq(3)').text($.i18n.prop('rank'));
                $('.nav_sub li:eq(4)').text($.i18n.prop('assets_record'));
                $('.parters li:eq(0)').text($.i18n.prop('partner'));
                $('.parters li:eq(1)').text($.i18n.prop('sero_site'));
                $('.parters li:eq(2)').text($.i18n.prop('change'));
                $('.parters li:eq(3)').text($.i18n.prop('sero_wallet'));
                $('footer h2').text($.i18n.prop('about_this_game'));
                $('footer h3').text($.i18n.prop('about_detail'));
                $('.pack_list2 span:eq(0)').text($.i18n.prop('gold'));
                $('.pack_list2 span:eq(1)').text($.i18n.prop('power_gems'));
                $('.pack_list2 span:eq(2)').text($.i18n.prop('soul_gems'));
                $('.pack_list2 span:eq(3)').text($.i18n.prop('time_gems'));
                $('.pack_list2 span:eq(4)').text($.i18n.prop('real_gems'));
                $('.pack_list2 span:eq(5)').text($.i18n.prop('space_gems'));
                $('.pack_list2 span:eq(6)').text($.i18n.prop('mind_gems'));
                $('.pack_list2 span:eq(7)').text($.i18n.prop('gloves'));
                $('.pack_list2 span:eq(8)').text($.i18n.prop('infinite_gloves'));
                $('.pack_list2 bdo:eq(0)').text($.i18n.prop('withdraw'));
                $('.pack_list2 bdo:eq(1)').text($.i18n.prop('deposit'));
                $('.layer_grap4 .layer_msg p:eq(0)').text($.i18n.prop('recharge_coin'));
                $('.layer_grap4 .layer_msg p:eq(1)').text($.i18n.prop('payment'));
            }
        });
    },

    getAccountList: function () {
        var biz = {}
        $('.pkrs').empty();
        Common.post("account/list", biz, {}, function (res) {

            if (res.base.code === 'SUCCESS') {

                if (res.biz) {

                    var dataArray = res.biz;
                    var balance = new BigNumber(0);

                    for (var i = 0; i < dataArray.length; i++) {
                        var data = dataArray[i];
                        var _balance = new BigNumber(0);
                        if(data.Balance.SERO){
                            _balance = new BigNumber(data.Balance.SERO);
                            _balance = _balance.dividedBy(Common.baseDecimal);
                            balance = balance.plus(_balance)
                        }
                        $('.pkrs').append(`
                            
                            <div class="col-lg-12 mb-4">
                                <div class="card text-white bg-primary shadow">
                                    <div class="card-body">
                                         <a style="text-decoration: none;color: white;" href="account-detail.html?pk=${data.PK}">
                                            <p class="m-0">Account${i+1} (<small>${data.PK.substring(0,8) + " ... " + data.PK.substring(data.PK.length-8,data.PK.length)}</small>)</p>
                                            <p class="text-white-50 small m-0 pkr">
                                            ${data.PkrBase58[data.PkrBase58.length-1]}&nbsp;&nbsp &nbsp;</p>
                                            <p class="text-right text-warning m-0"><strong>${_balance.toFixed(6)}</strong> SERO</p>
                                         </a>
                                    </div>
                                </div>
                            </div>
                           
                        `);
                    }

                    $('.dashboard span:eq(0)').text(balance.toFixed(6));
                }
            }
        })
    },

    getBlockHeight: function () {

        var data = {
            biz:{}
        }

        $.ajax({
            url: 'https://explorer.api.sero.cash/block/count',
            type: 'post',
            dataType: 'json',
            async: false,
            data: JSON.stringify(data),
            beforeSend: function () {
            },
            success: function (res) {
                $('.dashboard span:eq(1)').text( res.biz.blockCount);
            }
        })

    },

    getPeddingTx: function () {
        var biz = {};
        var page = {
            page_no:1,
            page_size:10,
        }
        Common.post("tx/list",biz,page,function (res) {
            if (res.base.code === "SUCCESS"){

            }
        })
    }


};
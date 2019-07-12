var Index = {

    init:function () {
        var that = this;
        that.getAccountList();
        that.getBlockHeight();
        setInterval(function () {
            that.getAccountList();
        },1000);

        setInterval(function () {
            that.getBlockHeight();
        },10000);
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
                                            ${data.PkrBase58[data.PkrBase58.length-1]}&nbsp;&nbsp;<i class="fas fa-copy"></i>&nbsp; &nbsp;</p>
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
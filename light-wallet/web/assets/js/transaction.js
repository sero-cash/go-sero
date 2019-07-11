var Transaction = {


    init: function () {
        var that = this;
        that.getAccountlist();

        $('.address').bind('change', function () {
            that.changeAccount();
        });

        $('.currency').bind('change', function () {
            that.changeCurrency();
        });


        $('#amount').bind('change',function () {
            that.calculate();
        });

        $('#gasprice').bind('change',function () {
            that.calculate();
        });


    },

    getAccountlist: function () {
        var biz = {}
        Common.post("account/list", biz, {}, function (res) {

            if (res.base.code === 'SUCCESS') {

                if (res.biz) {
                    var dataArray = res.biz;

                    var hasSetCurrency = false;
                    for (var i = 0; i < dataArray.length; i++) {
                        var data = dataArray[i];

                        var balance = new BigNumber(0).toFixed(6);

                        if (data.Balance && data.Balance.SERO) {
                            if (i === 0) {
                                var balanceObj = data.Balance;
                                for (var currency of Object.keys(balanceObj)) {
                                    $('.currency').append(`<option value="${currency}" ${i === 0 ? 'selected' : ''}>${currency }</option>`);
                                    if (!hasSetCurrency) {
                                        $('.currencyp').text(new BigNumber(balanceObj[currency]).dividedBy(Common.baseDecimal).toFixed(6) + ' ' + currency);
                                        hasSetCurrency = true;
                                    }
                                }
                            }
                            if (data.Balance.SERO) {
                                balance = new BigNumber(data.Balance.SERO).dividedBy(Common.baseDecimal).toFixed(6);
                            }
                        } else {
                            if (!hasSetCurrency) {
                                $('.currency').append(`<option value="SERO" selected}>SERO</option>`);
                                $('.currencyp').text("0.000000 SERO");
                                hasSetCurrency = true;
                            }
                        }
                        $('.address').append(`<option value="${data.PK}" ${i === 0 ? 'selected' : ''}>${data.PK.substring(0, 8) + ' ... ' + data.PK.substring(data.PK.length - 8) }   ${ balance + ' SERO'}</option>`);
                    }
                }
            }
        })
    },

    changeAccount: function () {
        var pk = $(".address").val();

        var biz = {
            PK: pk,
        }
        Common.post("account/detail", biz, {}, function (res) {
            $('.currency').empty();
            if (res.base.code === 'SUCCESS') {

                if (res.biz) {
                    var data = res.biz;
                    if (data.Balance && data.Balance.SERO) {
                        var balanceObj = data.Balance;
                        var i = 0;
                        for (var currency of Object.keys(balanceObj)) {
                            $('.currency').append(`<option value="${balanceObj[currency]}" ${i === 0 ? 'selected' : ''}>${currency }</option>`);
                            $('.currencyp').text(new BigNumber(balanceObj[currency]).dividedBy(Common.baseDecimal).toFixed(6) + ' ' + currency);
                            i++;
                        }
                    } else {
                        $('.currency').append(`<option value="SERO" selected}>SERO</option>`);
                        $('.currencyp').text("0.000000 SERO");
                    }
                }
            }
        })
    },

    changeCurrency: function () {
        var balance = $('.currency').val();
        var currency = $('.currency').find("option:selected").text();

        balance = new BigNumber(balance).dividedBy(Common.baseDecimal).toFixed(6);

        $('.currencyp').text(`${balance} ${currency}`);
    },


    calculate: function () {

        var amount = $("#amount").val();
        var gasprice = $("#gasprice").val();
        var currency = $(".currency").val();

        if (amount > 0 && gasprice > 0) {
            amount = new BigNumber(amount).multipliedBy(Common.baseDecimal);
            gasprice = new BigNumber(gasprice).multipliedBy(new BigNumber(10).pow(9));
            var fee = gasprice.multipliedBy(250000).dividedBy(Common.baseDecimal);
            var total = fee.plus(amount).dividedBy(Common.baseDecimal);

            $('.calculate span:eq(0)').text(fee.toFixed(6) +' ' + currency);
            $('.calculate span:eq(1)').text(total.toFixed(6) +' ' + currency);
        }else{
            $('.calculate span:eq(0)').text('0.000000 ' + currency);
            $('.calculate span:eq(1)').text('0.000000 ' + currency);
        }
    },


    subTx: function () {

        var from = $(".address").val();
        var currency = $(".currency").val();
        var amount = $("#amount").val();
        var to = $("#address").val();
        var gasprice = $("#gasprice").val();

        amount = new BigNumber(amount).multipliedBy(Common.baseDecimal);
        gasprice = new BigNumber(gasprice).multipliedBy(new BigNumber(10).pow(9));
        var fee = gasprice.multipliedBy(250000);
        var total = fee.plus(amount);

        $(".modal-body ul li:eq(0) div div:eq(1)").text(from);
        $(".modal-body ul li:eq(1) div div:eq(1)").text(to);
        $(".modal-body ul li:eq(2) div div:eq(1)").text(amount.toFixed(6));
        $(".modal-body ul li:eq(3) div div:eq(1)").text(0);
        $(".modal-body ul li:eq(4) div div:eq(1)").text(0);

        $('#myModal').modal('show');

        $('.modal-footer button:eq(0)').bind('click', function () {
            $('#sub1').attr('disabled', false);
        });

        $('.modal-footer button:eq(1)').bind('click', function () {
            var biz = {
                From: from,
                To: to,
                Currency: currency,
                Amount: amount.toString(10),
                GasPrice: gasprice.toString(10),
            }
            Common.post('tx/transfer', biz, {}, function (res) {
                if (res.base.code === 'SUCCESS') {

                } else {
                    alert(res.base.desc);
                }
            })
        });

    }


}
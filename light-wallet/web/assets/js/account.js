var Account = {


    init: function () {


        $('.close').bind('click', function () {
            $('.modal').hide();
        });

    },

    newAccount: function () {

        var pwd = $("#pwd").val();
        var confirmPwd = $("#confirmPwd").val();

        if (pwd !== confirmPwd) {
            $('.modal-title').text('Tips');
            $('.modal-body').text('Inconsistent password entered twice');
            $('#myModal').modal('toggle');

            $('.modal-footer button').bind('click', function () {
                $("#sub1").attr('disabled', false);
                $('#myModal').modal('hide');
            });
        } else {
            var biz = {
                passphrase: pwd,
            }
            Common.post("account/create", biz, {}, function (res) {
                if (res.base.code === "SUCCESS") {
                    $('.modal-title').text('Write Down Your Mnemonic phrases!');
                    $('#myModal').modal({backdrop: 'static', keyboard: false})
                    $('.modal-footer button:eq(0)').bind('click', function () {
                        $("#sub1").attr('disabled', false);
                        $('#myModal').modal('hide');
                    });
                    $('.modal-footer button:eq(1)').bind('click', function () {
                        window.location.href = "index.html";
                    });
                    $('.modal-body p:eq(0)').text(res.biz.address);
                    $('.modal-body p:eq(1)').text(res.biz.mnemonic);
                    $('#myModal').modal('show');
                    $("#sub1").attr('disabled', false);
                } else {
                    $("#sub1").text("NEXT");
                    $("#sub1").attr('disabled', false);
                    alert(res.base.desc);
                }

            })
        }
    },
};

var Detail = {


    address: '',

    txList: [],

    init: function () {
        var that = this;
        that.getAccountDetail();
        that.jumpPage(1);

        $('.toast').toast({animation: true, autohide: true, delay: 1000})
        var clipboard1 = new ClipboardJS('.fa-copy');
        clipboard1.on('success', function (e) {
            $('#toast1 div:eq(0)').text('Copy successfully!');
            $('#toast1').toast('show')
        });


        var that = this;
        $('.pagination li:eq(0)').bind('click',function () {
            that.txPageNo = that.txPageNo - 1;
            if (that.txPageNo <= 0) {
                that.txPageNo = 1;
            }
            that.jumpPage(that.txPageNo);
        });
        $('.pagination li:eq(2)').bind('click',function () {
            if (that.txCount>0){
                that.txPageNo = that.txPageNo + 1;
                that.jumpPage(that.txPageNo);
            }
        });


        $('.backup').bind('click', function () {

            Common.post('path/keystore', {}, {}, function (res) {

                if (res.base.code === 'SUCCESS') {
                    $('.modal-body').empty();
                    $('.modal-title').text('Help');
                    $('.modal-body').append(`<p class="text-left">Please backup the keystore file in the folder :</p>`)
                    $('.modal-body').append(`<span style="color:#000">${res.biz}</span>`)

                    $('.modal').modal('show');

                }
            });

        });

        setInterval(function () {
            that.getAccountDetail();
            that.getTxList();
        }, 10000);
    },

    getAccountDetail: function () {

        var pk = GetQueryString("pk");
        var biz = {
            PK: pk,
        }
        $('.currency').empty();
        Common.post("account/detail", biz, {}, function (res) {

            if (res.base.code === "SUCCESS") {

                var pkr = res.biz.PkrBase58[res.biz.PkrBase58.length - 1];

                $('.fa-copy').attr('data-clipboard-text', pkr);
                $('.address').text(pkr);

                $('.fa-qrcode').bind('click', function () {
                    $('.modal-body div:eq(1)').text(pkr);
                    $('#qrcode').empty();
                    $('#qrcode').qrcode({
                        render: "canvas",
                        width: 200,
                        height: 200,
                        text: pkr
                    });
                    $('#myModal').modal('show')
                });

                $('.pk').text(res.biz.PK.substring(0, 8) + " ... " + res.biz.PK.substring(res.biz.PK.length - 8));

                var balanceObj = res.biz.Balance;

                var strMap = new Map();
                for (var k of Object.keys(balanceObj)) {
                    strMap.set(k, balanceObj[k]);
                    $('.currency').append(`
                        <div class="col-md-3 col-xl-3 mb-4">
                            <div class="card shadow border-left-success py-2">
                                <div class="card-body">
                                    <div class="row align-items-center no-gutters">
                                        <div class="col mr-2">
                                            <div class="text-uppercase text-success font-weight-bold text-xs mb-1"><span>${k}</span></div>
                                            <div class="text-dark font-weight-bold h5 mb-0"><span>${new BigNumber(balanceObj[k]).dividedBy(Common.baseDecimal).toFixed(6)}</span></div>
                                        </div>
                                        <div class="col-auto"><i class="fas fa-dollar-sign fa-2x text-gray-300"></i></div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    `);
                }

                if (strMap.size === 0) {

                    $('.currency').append(`
                        <div class="col-md-3 col-xl-3 mb-4">
                            <div class="card shadow border-left-success py-2">
                                <div class="card-body">
                                    <div class="row align-items-center no-gutters">
                                        <div class="col mr-2">
                                            <div class="text-uppercase text-success font-weight-bold text-xs mb-1"><span>SERO</span></div>
                                            <div class="text-dark font-weight-bold h5 mb-0"><span>0.000000</span></div>
                                        </div>
                                        <div class="col-auto"><i class="fas fa-dollar-sign fa-2x text-gray-300"></i></div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    `);
                }

            }
        });
    },

    txPageNo: 1,

    txPageSize: 10,

    txCount: 0,

    jumpPage: function (pageNo) {
        var that = this;
        if (pageNo <= 0) {
            pageNo = 1
        }
        that.txPageNo = pageNo;
        that.getTxList(function (page) {
            var count = page.count?page.count:0;
            that.txCount = count;
            if (count === 0) {
                if (page.page_no === 1) {
                    $('.pagination li:eq(0)').addClass('disabled');
                    $('.pagination li:eq(2)').addClass('disabled');
                } else {
                    $('.pagination li:eq(0)').removeClass('disabled');
                    if(count<page.page_size){
                        $('.pagination li:eq(2)').addClass('disabled');
                    }else{
                        $('.pagination li:eq(2)').removeClass('disabled');
                    }
                }
            } else {
                if (count > 0) {
                    if (page.page_no === 1) {
                        $('.pagination li:eq(0)').addClass('disabled');
                        $('.pagination li:eq(2)').removeClass('disabled');
                    } else {
                        $('.pagination li:eq(0)').removeClass('disabled');
                        $('.pagination li:eq(2)').removeClass('disabled');
                    }
                }
            }

            $('.pagination li:eq(1) a').text(page.page_no);

        });
    },


    getTxList: function (callback) {
        var that = this;

        var pk = GetQueryString("pk");
        var biz = {
            PK: pk,
        };
        var page = {
            page_no: that.txPageNo,
            page_size: that.txPageSize,
        }
        $('tbody').empty();

        Common.post("tx/list", biz, page, function (res) {
            if (res.base.code === "SUCCESS") {
                if (res.biz) {
                    var data = res.biz;
                    for (var i = 0; i < data.length; i++) {

                        var tx = data[i];

                        var tknObj = tx.Asset.Tkn;
                        var value = 0;
                        var strMap = new Map()
                        for (var k of Object.keys(tknObj)) {
                            strMap[k] = tknObj[k];
                        }
                        value = new BigNumber(strMap['Value']);
                        value = value.dividedBy(Common.baseDecimal).toFixed(6);

                        $('tbody').append(
                            `
                        <tr>
                            <td>${i + 1}</td>
                            <td><a href="https://explorer.web.sero.cash/txsInfo.html?hash=${tx.TxHash}" title="${tx.TxHash}">${tx.TxHash.substring(0, 5) + " ... " + tx.TxHash.substring(tx.TxHash.length - 5)}</a></td>
                            <td>${tx.Num}</td>
                            <td title="${tx.Pkr}">${tx.Pkr.substring(0, 5) + " ... " + tx.Pkr.substring(tx.Pkr.length - 5)}</td>
                            <td>${strMap['Currency']}</td>
                            <td><span class="text-success">${tx.Num > 0 ? 'Completed' : 'Pedding'}</span></td>
                            <td>${value}</td>
                        </tr>
                        `
                        );
                    }
                    ;


                    if (data.length > 0) {

                        $('.pagination')
                    }
                }
                if (callback) {
                    callback(res.page)
                }
            }
        });
    }
}

function GetQueryString(name) {
    var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)");
    var r = window.location.search.substr(1).match(reg);
    if (r != null) return unescape(r[2]);
    return null;
}


var Keystore = {

    file: '',

    init: function () {
        var that = this;

        $('.close').bind('click', function () {
            $('#myModal').modal('hide');
        });

        $('.modal-footer button:eq(1)').bind('click', function () {
            window.location.href = 'index.html';
        });


        $("#i-file").bind("change", function () {
            that.file = this.files[0];
        });
    },

    import: function () {
        var that = this;
        var password = $('#password').val();
        var formData = new FormData();
        formData.append("passphrase", password);
        formData.append("uploadFile", that.file);

        $.ajax({
            url: Common.host + "/account/import/keystore",
            dataType: 'json',
            type: 'POST',
            async: false,
            data: formData,
            processData: false,
            contentType: false,
            success: function (data) {
                if (data.responseText === 'INVALID_FILE_TYPE') {
                    $('.modal-title').text("Warning");
                    $('.modal-body').text("Password given is incorrect!");
                } else if (data.responseText === 'SUCCESS') {
                    $('.modal-title').text("Successful");
                    $('.modal-body').text("Successfully imported!");
                } else {
                    $('.modal-title').text("Error");
                    $('.modal-body').text("Import failed,Incorrect file type");
                }
                $('#myModal').modal('show');
            },
            error: function (data) {
                if (data.responseText === 'INVALID_FILE_TYPE') {
                    $('.modal-title').text("Warning");
                    $('.modal-body').text("Password given is incorrect!");
                } else if (data.responseText === 'SUCCESS') {
                    $('.modal-title').text("Successful");
                    $('.modal-body').text("Successfully imported!");
                } else {
                    $('.modal-title').text("Error");
                    $('.modal-body').text("Import failed,Incorrect file type");
                }
                $('#myModal').modal('show');
            }
        });

        $("#sub1").attr('disabled', false);
    },


};

var Mnemnic = {

    init: function () {
        $('.close').bind('click', function () {
            $('.modal').modal('hide');
        });

        $('.modal-footer button:eq(1)').bind('click', function () {
            window.location.href = 'index.html';
        });

    },

    import: function () {
        var mnemonic = $('#mnemonic').val();
        var password = $('#password').val();

        var biz = {
            mnemonic: mnemonic,
            passphrase: password,
        }

        Common.post('account/import/mnemonic', biz, {}, function (res) {

            if (res.base.code === 'SUCCESS') {
                var address = res.biz.address
                $('.modal-title').text("Import Successful");
                $('.modal-body p:eq(0)').text(address);
            } else {
                $('.modal-title').text("ERROR");
                $('.modal-body p:eq(0)').text(res.base.desc);
            }
            $('.modal').modal('show');
            $("#sub1").attr('disabled', false);
        });

    }

}

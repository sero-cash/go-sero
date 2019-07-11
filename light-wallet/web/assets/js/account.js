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

    txList:[],

    init: function () {
        var that = this;
        that.getAccountDetail();
        that.getTxList();

        setInterval(function () {
            that.getAccountDetail();
            that.getTxList();
        },1000);
    },

    getAccountDetail: function () {

        var pk = GetQueryString("pk");
        var biz = {
            PK: pk,
        }
        $('.currency').empty();
        Common.post("account/detail", biz, {}, function (res) {

            if (res.base.code === "SUCCESS") {


                $('.address').text(res.biz.PkrBase58[res.biz.PkrBase58.length-1]);
                $('.pk').text(res.biz.PK.substring(0,8) + " ... " + res.biz.PK.substring(res.biz.PK.length-8));

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

    getTxList: function () {

        var pk = GetQueryString("pk");
        var biz = {
            PK: pk,
        };
        var page = {
            page_no: 1,
            page_size: 10,
        }
        $('tbody').empty();

        Common.post("tx/list", biz, page, function (res) {
            if (res.base.code === "SUCCESS") {
                if(res.biz){
                    var data = res.biz;
                    for (var i=0;i<data.length;i++){

                        var tx = data[i];

                        var tknObj = tx.Asset.Tkn;
                        var value = 0;
                        var strMap = new Map()
                        for (var k of Object.keys(tknObj)) {
                            strMap[k]=tknObj[k];
                        }
                        value = new BigNumber(strMap['Value']);
                        value = value.dividedBy(Common.baseDecimal).toFixed(6);

                        $('tbody').append(
                            `
                        <tr>
                            <td>${i+1}</td>
                            <td><a href="tx-detail.html">${tx.TxHash.substring(0,5) + " ... " + tx.TxHash.substring(tx.TxHash.length-5)}</a></td>
                            <td><a href="https://explorer.web.sero.cash/txsInfo.html?hash=${tx.TxHash}" target="_blank">${tx.Num}</a></td>
                            <td>${tx.Pkr.substring(0,5) + " ... " + tx.Pkr.substring(tx.Pkr.length-5)}</td>
                            <td>${strMap['Currency']}</td>
                            <td><span class="text-success">${tx.Num>0?'Completed':'Pedding'}</span></td>
                            <td>${value}</td>
                        </tr>
                        `
                        );
                    }
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

    file :'',

    init: function () {
        var that = this;

        $('.close').bind('click', function () {
            $('.modal').hide();
        });

        $("#i-file").bind("change",function () {
            that.file = this.files[0];
        });
    },

    import: function () {
        var that = this;
        var password = $('#password').val();
        var formData = new FormData();
        formData.append("passphrase",password);
        formData.append("uploadFile",that.file);

        $.ajax({
            url:Common.host + "/account/import/keystore",
            dataType:'json',
            type:'POST',
            async: false,
            data: formData,
            processData : false,
            contentType : false,
            success: function(data){
                console.log("data:: ",data);
                if (data.responseText === 'INVALID_FILE_TYPE'){
                    alert('密码不对！');
                }else if (data.responseText === 'SUCCESS') {
                    alert('导入成功！');
                }else{
                    alert('导入失败！');
                }
            },
            error:function(data){
                if (data.responseText === 'INVALID_FILE_TYPE'){
                    alert('密码不对！');
                }else if (data.responseText === 'SUCCESS') {
                    alert('导入成功！');
                }else{
                    alert('导入失败！');
                }
            }
        });

        $("#sub1").attr('disabled',false);
    },


};

var Mnemnic = {

    init: function () {
        $('.close').bind('click', function () {
            $('.modal').hide();
        });

    },

    import: function () {

    }

}

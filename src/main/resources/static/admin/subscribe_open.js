
/**
 * 新用户报装
 * @param session
 * @constructor
 */
toughradius.admin.subscribe.OpenSubscribeForm = function(session){
    var winid = "toughradius.admin.subscribe.OpenSubscribeForm";
    if($$(winid))
        return;
    var formid = winid+"_form";
    var idcard_base64_pic = webix.uid();
    var updateFeeValue = function(){
        var _params = {
            fee_ids:$$(formid).elements['fee_ids'].getValue(),
            order_num:$$(formid).elements['order_num'].getValue(),
            product_id:$$(formid).elements['product_id'].getValue(),
            device_sn:$$(formid).elements['device_sn'].getValue()
        };
        webix.ajax().get('/admin/subscribe/calc', _params).then(function (result) {
            var resp = result.json();
            var data = resp.data;
            console.log(data);
            $$(formid).elements['fees_value'].setValue(data.fees_value);
            $$(formid).elements['product_fee'].setValue(data.product_fee);
            $$(formid).elements['total_fee'].setValue(data.total_fee);
            $$(formid).elements['expire_time'].setValue(data.expire_time);
            $$(formid).elements['device_fee'].setValue(data.device_fee);
            $$(formid).refresh();
            if(resp.code>0&&resp.msg){
                webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
            }
        })

    };
    var getDeviceSn = function(sn){
        var _params = {
            sn: sn,
            node_id : $$(formid).elements['node_id'].getValue(),
            area_id:$$(formid).elements['area_id'].getValue()
        };
        if(!_params.node_id||!_params.area_id){
            webix.message({ type: "error", text: "请选择组织和区域", expire: 3000 });
            return;
        }
        webix.ajax().get('/admin/device/get',_params).then(function (result) {
            var resp = result.json();
            var data = resp.data;
            console.log(data);
            if(resp.code>0&&resp.msg){
                webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
            }else{
                webix.message({ type: "info", text: "设备可用", expire: 3000 });
            }
            updateFeeValue();
        })
    };
    webix.ui({
        id:winid,
        view: "window",
        css:"win-body",
        move:true,
        width:800,
        height:645,
        position: "center",
        head: {
            view: "toolbar",
            css:"win-toolbar",

            cols: [
                {view: "icon", icon: "laptop", css: "alter"},
                {view: "label", label: "用户报装"},
                {view: "icon", icon: "times-circle", css: "alter", click: function(){
                    $$(winid).close();
                }}
            ]
        },
        body: {
            rows:[
                {
                    id: formid,
                    view: "form",
                    scroll: 'y',
                    elementsConfig: { labelWidth: 85 },
                    elements: [
                        { view: "fieldset", label: "基本信息", body: {
                            cols:[
                                {view:"template", id:idcard_base64_pic, borderless:false, width:110,height:120, hidden:session.system_config['USE_IDCARD_READER']!=='enabled'},
                                {width:5},
                                {
                                    rows:[
                                        { view: "text", hidden:true, name:"idcard_pic" },
                                        {
                                            cols:[
                                                {
                                                    cols:[
                                                        {view: "text",hidden:true, name: "node_id", value:session.node_id||""},
                                                        {view: "text", name: "node_name", readonly:true, label: "组织节点",  validate:webix.rules.isNotEmpty, value:session.node_name||""},
                                                        {view: "button", label: "选择", type: "icon", icon: "angle-down", borderless: true, width: 66,click:function(){
                                                            toughradius.admin.methods.openNodeTree(session.node_id,this.$view,function(item){
                                                                $$(formid).elements['node_id'].setValue(item.id);
                                                                $$(formid).elements['node_name'].setValue(item.value);
                                                                var list = $$(formid).elements['area_id'].getPopup().getList();
                                                                list.clearAll();
                                                                list.load("/admin/area/options?node_id=" + item.id);
                                                                var list2 = $$(formid).elements['product_id'].getPopup().getList();
                                                                list2.clearAll();
                                                                list2.load("/admin/product/options?node_id="+item.id);
                                                                var list3 = $$(formid).elements['fee_ids'].getPopup().getList();
                                                                list3.clearAll();
                                                                list3.load("/admin/fees/options?node_id"+item.id);
                                                                var list4 = $$(formid).elements['issues_opr'].getPopup().getList();
                                                                list4.clearAll();
                                                                list4.load("/admin/opr/options?node_id"+item.id);
                                                            });
                                                        }},
                                                    ]
                                                },
                                                { view: "combo", name: "area_id", label: "区域(*)", icon: "caret-down", validate:webix.rules.isNotEmpty,on:{
                                                    onChange:function(newv, oldv){
                                                        var list = $$(formid).elements['zone_id'].getPopup().getList();
                                                        list.clearAll();
                                                        list.load("/admin/zone/options?area_id=" + newv);
                                                    }
                                                },options: {
                                                    view:"suggest", url:"/admin/area/options?node_id=" + session.node_id
                                                }}

                                            ]
                                        },
                                        {
                                            cols:[
                                                { view: "combo", name: "zone_id", label: "小区", icon: "caret-down", options: {
                                                    view:"suggest",data:[]
                                                }},
                                                {
                                                    cols:[
                                                        {view: "text", name: "realname", label: "客户名称(*)", width:250, placeholder: "客户名称",validate:webix.rules.isNotEmpty},
                                                        {view:"combo", name:"gender", labelWidth:0, label: "", value:"male",options:[{id:'male',value:"男"}, {id:'female',value:"女"}]},
                                                    ]
                                                },
                                            ]
                                        },
                                        {
                                            cols:[
                                                {view: "text", name: "idcard", label: "证件号码", placeholder: "证件号码"},
                                                {view: "text", name: "mobile", label: "手机号码", placeholder: "手机号码"},
                                            ]
                                        },
                                        {
                                            cols:[
                                                {view: "textarea", name: "address", label: "地址", placeholder: "地址", height:40},
                                                { view: "textarea", name: "remark", label: "备注", placeholder: "备注", height: 40}
                                            ]
                                        }
                                    ]
                                }
                            ]

                        }},
                        { view: "fieldset", label: "授权信息", body: {
                            rows:[
                                {
                                    cols:[
                                        { view: "richselect", name: "product_id", label: "商品(*)", icon: "caret-down", validate:webix.rules.isNotEmpty,
                                             options: {
                                                view:"suggest",url:"/admin/product/options?node_id=" + session.node_id
                                             },
                                            on:{
                                                onChange:function(newv, oldv){
                                                    updateFeeValue()
                                                 }
                                             }
                                         },
                                        {view: "counter", name: "order_num", label: "购买数量(*)",  value:1, min:1, max:1000,on:{
                                            onChange:function(newv, oldv){
                                                updateFeeValue()
                                             }
                                         }},
                                    ]
                                },
                                {
                                    cols:[
                                        {view: "text", name: "subscriber", label: "授权帐号(*)", placeholder: "授权帐号",validate:webix.rules.isNotEmpty},
                                        {view: "text", name: "password", type:"password", label: "授权密码(*)", placeholder: "授权密码",validate:webix.rules.isNotEmpty}
                                    ]
                                },
                                {
                                    cols:[
                                        {view: "datepicker",name:"expire_time",timepicker :true, readonly:!hasPerms(session,['subscribe_expire_modify']),width:210,
                                                label:"过期时间",stringResult:true,format: session.system_config.SYSTEM_USER_EXPORT_FORMAT, validate:webix.rules.isNotEmpty},
                                        {width:10},
                                        {view:"radio", name:"device_type", label: "", value:"deposit",options:[{id:'deposit',value:"租用终端"}, {id:'purchase',value:"购买终端"}]},
                                        {view: "text", name: "device_sn", label: "", labelWidth:0},
                                        {
                                            view: "button", type: "icon", width: 90, icon: "calculator", label: "检查库存", click: function () {
                                                var sn = $$(formid).elements['device_sn'].getValue();
                                                if(!sn){
                                                    webix.message({ type: "error", text: "请输入设备序列号", expire: 3000 });
                                                }else{
                                                    getDeviceSn(sn);
                                                }
                                            }
                                        },
                                        {
                                            view: "button", type: "icon", width: 55, icon: "remove", label: "清除", click: function () {
                                                $$(formid).elements['device_sn'].setValue("");
                                                updateFeeValue();
                                            }
                                        },
                                    ]
                                }
                            ]
                        }},
                        {
                            cols:[

                                {
                                    view: "fieldset", label: "缴费信息",  body: {
                                        rows: [
                                            {
                                                view: "multiselect", name: "fee_ids", label: "收费项", icon: "caret-down",
                                                options: "/admin/fees/options?node_id"+session.node_id, on: {
                                                    onChange: function (newv, oldv) {
                                                        console.log(this.getValue())
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            { view: "text", name: "fees_value", label: "收费项合计", readonly: true, value: "0.00" },
                                            { view: "text", name: "device_fee", label: "终端费用(*)", placeholder: "终端费用", readonly: true, validate: webix.rules.isNotEmpty },
                                            { view: "text", name: "product_fee", label: "资费金额(*)", placeholder: "资费金额", readonly: true, validate: webix.rules.isNotEmpty },
                                            {
                                                cols:[
                                                    { view: "text", name: "total_fee", label: "合计", readonly: !hasPerms(session, ['subscribe_fee_modify']), validate: webix.rules.isNotEmpty },
                                                    {
                                                        view: "button", type: "icon", icon: "calculator", label: "计算", width:70,click: function () {
                                                            updateFeeValue();
                                                        }
                                                    },
                                                ]
                                            },
                                            { view: "label", css: "form-desc", label: "注意： 总费用 = 终端设备费用 + 收费项目合计费用 + 资费费用" },
                                            {
                                                view: "richselect", name: "pay_type", label: "缴费方式", value: 'cash',
                                                options: [{ id: 'cash', value: "现金" }, { id: 'bank', value: "银行卡" }, { id: 'alipay', value: "支付宝" }, { id: 'wxpay', value: "微信支付" }]
                                            },
                                            { view: "radio", name: "pay_status", label: "支付状态",hidden:false, value: 'done', options: [{ id: 'done', value: "已缴费" }, { id: 'padding', value: "未缴费" }] }
                                        ]
                                    }
                                },
                                {width:20},
                                {
                                    view: "fieldset", label: "工单信息",  body: {
                                        rows: [
                                            {
                                                view: "richselect", name: "issues_opr", label: "委派操作员", icon: "caret-down",
                                                options: [], on: {
                                                    onChange: function (newv, oldv) {
                                                        console.log(this.getValue());
                                                        updateFeeValue();
                                                    }
                                                }
                                            },
                                            { view: "radio", name: "issues_type", label: "工单类型", value: 'install', options: [{ id: 'install', value: "新装" }] },
                                            { view: "radio", name: "issues_status", label: "状态", value: 'padding', options: [{ id: 'padding', value: "未完成" }, { id: 'done', value: "已完成" }] },
                                            { view: "textarea", name: "issues_remark", label: "备注", placeholder: "备注", height: 145 }
                                        ]
                                    }
                                }
                            ]
                        }

                    ]
                },
                {
                    view: "toolbar",
                    height:42,
                    css: "page-toolbar",
                    cols: [
                        {view:"label",css: "form-desc",label:"资料录入 (*)必填"}, {},
                        {
                            view: "button", type: "form", width: 100, icon: "check-circle", label: "读取身份证",
                            hidden:session.system_config['USE_IDCARD_READER']!=='enabled',click: function () {
                                var btn = this;
                                btn.disable();
                                var CVR_IDCard = document.getElementById("CVR_IDCard");
                                var strReadResult = CVR_IDCard.ReadCard();
                                if(strReadResult == "0")
                                {
                                    $$(formid).elements['realname'].setValue(CVR_IDCard.Name);
                                    $$(formid).elements['gender'].setValue(CVR_IDCard.Sex=="男"?"male":"female");
                                    $$(formid).elements['idcard'].setValue(CVR_IDCard.CardNo);
                                    $$(formid).elements['idcard_pic'].setValue(CVR_IDCard.Picture);
                                    $$(idcard_base64_pic).define('template', "<img src='data:image/jpeg;base64,"+CVR_IDCard.Picture+"'/>")
                                    $$(idcard_base64_pic).refresh()
                                    btn.enable();
                                  }
                                  else
                                  {
                                      btn.enable();
                                      webix.message({ type: "error", text: strReadResult, expire: 2000 });
                                  }
                            }
                        },
                        {
                            view: "button", type: "form", width: 100, icon: "check-circle", label: "提交", click: function () {
                                if (!$$(formid).validate()) {
                                    webix.message({ type: "error", text: "请正确填写资料", expire: 1000 });
                                    return false;
                                }
                                var btn = this;
                                btn.disable();
                                var params = $$(formid).getValues();
                                webix.ajax().post('/admin/subscribe/create', params).then(function (result) {
                                    btn.enable();
                                    var resp = result.json();
                                    webix.message({ type: resp.msgtype, text: resp.msg, expire: 3000 });
                                    if (resp.code === 0) {
                                        toughradius.admin.subscribe.loadPage(session, params.subscriber);
                                        $$(winid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", type: "base", width: 100, icon: "times-circle", label: "取消", click: function () {
                                 $$(winid).close();
                            }
                        }
                    ]
                },
            ]
        }

    }).show();
};

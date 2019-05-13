webix.ready(function() {
    var logviewId = webix.uid();
    var doLogin = function (formValues){
        webix.ajax().post('/admin/login',formValues).then(function (result) {
            var resp = result.json();
            if (resp.code===0){
                window.location.href = "/admin";
            }else{
                webix.message({type: resp.msgtype, text:resp.msg,expire:2000});
            }
        }).fail(function (xhr) {
            webix.message({type: 'error', text: "登录失败:"+xhr.statusText,expire:2000});
        });
    };


    webix.ui({
        css:"login-page",
        rows:[
            {gravity: 1},
            {
                align: "center,middle",
                body:{
                    css:"login-win",
                    width: 360,
                    rows:[
                        {
                            id:logviewId,
                            view: "form",
                            scroll: false,
                            autoHeight: true,
                            paddingX: 30,
                            paddingY: 30,
                            elements: [
                                {
                                    height:90,
                                    cols:[
                                        {css:"login-logo"}
                                    ]
                                },
                                {view: "text",name:"username", value: '', placeholder: "帐 号", height:35},
                                {view: "text", name:"password",type: 'password', value: '', placeholder: "密 码",height:35},
                                {
                                    margin: 0, cols: [
                                        {
                                            id:"login_btn",
                                            view: "button",
                                            label: "登 录",
                                            value: "Login",
                                            type: "form",
                                            height:39,
                                            click:function(){
                                                doLogin($$(logviewId).getValues())
                                            },
                                            hotkey: "enter",
                                            on:{
                                                onKeyPress:function(){
                                                    doLogin($$(logviewId).getValues())
                                                }
                                            }
                                        }
                                    ]
                                },
                            ]
                        }
                    ]
                }
            },
            {gravity: 2}
        ]
    });
});

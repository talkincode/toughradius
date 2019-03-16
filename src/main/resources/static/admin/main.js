if (!window.toughradius)
    toughradius={};

// if (!webix.env.touch && webix.ui.scrollSize){
//     webix.CustomScroll.init();
// }


currentLang = navigator.language;
if(!currentLang){
    currentLang = navigator.browserLanguage;
}
webix.i18n.setLocale(currentLang);

toughradius.admin = {};
toughradius.admin.pageId = "toughradius.admin-main-page";
toughradius.admin.panelId = "toughradius.admin-main-panel";
toughradius.admin.toolbarId = "toughradius.admin-main-toolbar";
toughradius.admin.actions = {};
toughradius.admin.methods = {};

toughradius.admin.methods.setToolbar = function(icon, title, help){
    $$(toughradius.admin.toolbarId+"_icon").define("icon",icon);
    $$(toughradius.admin.toolbarId+"_icon").refresh();
    $$(toughradius.admin.toolbarId+"_title").define("label",title);
    $$(toughradius.admin.toolbarId+"_title").refresh();
};

toughradius.admin.methods.doLogin = function (formValues){
    webix.ajax().post('/login',formValues).then(function (result) {
        var resp = result.json();
        if (resp.code===0){
            window.location.href = "/admin";
        }else{
            webix.message({type: resp.msgtype, text:resp.msg,expire:500});
        }
    }).fail(function (xhr) {
        webix.message({type: 'error', text: "登录失败:"+xhr.statusText,expire:500});
    });
};

toughradius.admin.methods.showBusyBar = function (viewid,delay, callback){
    $$(viewid).disable();
    $$(viewid).showProgress({
        type:"top",
        delay:delay,
        hide:true
    });
    setTimeout(function(){
        callback();
        $$(viewid).enable();
    }, delay);
};

toughradius.admin.initUploadApi = function(uid, uploadurl, callback){
     webix.ui({
        id:uid,
        view:"uploader",
        upload:uploadurl,
        on:{
            onBeforeFileAdd:function(item){
                 item.formData = {};
                 webix.message({type: "info", text: "正在上传..", expire: 3000})
            },
            onFileUpload:function(item){
                if(callback){
                    callback(item);
                }
            },
            onFileUploadError:function(item){
                webix.message({type:"error",text:"Error during file upload",expire:3000});
            },
            onUploadComplete:function(resp){
                webix.message({type: resp.msgtype, text: resp.msg, expire: 5000});
            }
        },
        apiOnly:true
    });
};

toughradius.admin.methods.requirejs = function(jsname, session,callback){
    console.log("load admin/" + jsname + ".js");
    if(session.dev_mode === 'enabled'){
         webix.require("admin/" + jsname + ".js?rand="+new Date().getTime(), function () {
            callback();
         });
    }else{
        webix.require("admin/" + jsname + ".js", function () {
             callback();
        });
    }
};


webix.ready(function() {
    /**
     * 加载主界面
     */
    webix.ajax().get('/admin/session',{}).then(function (result) {
        var resp = result.json();
        if(resp.code===1){
            webix.message({type:"error",text:resp.msg});
            setTimeout(function(){window.location.href = "/admin";},2000);
            return false;
        }
        var session = resp.data;
        webix.require("sidebar.js", function () {
            webix.require("css/sidebar.css");
            webix.ui({
                rows: [
                    {
                        view: "toolbar",
                        padding: 3,
                        height: 44,
                        css: "page-nav",
                        elements: [
                            {
                                cols: [
                                    { view: "template", css: "nav-logo", maxWidth:188, template: "<a href='/admin'><img src='/static/imgs/logo.png' width='156' height='25'/></a>", height:40},
                                    {
                                        view: "button", type: "icon", icon: "bars", width: 37, align: "left", css: "nav-item-color", click: function () {
                                            $$("$sidebar1").toggle()
                                        }
                                    },
                                    {},
                                    {
                                        view: "button", css: "nav-item-color", type: "icon", width: 90, maxWidth: 200, icon: "key",align:"right",
                                        label: "修改密码", click: function () {
                                        }
                                    },
                                    {
                                        view: "button", css: "nav-item-color", type: "icon", width: 70, icon: "sign-out",align:"right",
                                        label: "退出", click: function () {
                                            window.location.href = "/admin/logout";
                                        }
                                    }
                                ]
                            }

                        ]
                    },
                    {
                        borderless:true,
                        cols: [
                            {
                                rows:[
                                    {
                                        rows: [
                                            { view: "label", height:40, css: "sideber-label", label: "<i class=\"fa fa-bars\" aria-hidden=\"true\"></i> 功能导航" },
                                            {
                                                view: "sidebar",
                                                scroll:"auto",
                                                width: 180,
                                                data: session.menudata,
                                                on: {
                                                    onAfterSelect: function (id) {
                                                        try {
                                                            console.log("action = " + id);
                                                            webix.require("admin/" + id + ".js?rand="+new Date().getTime(), function () {
                                                                toughradius.admin[id].loadPage(session);
                                                            });
                                                        } catch (err) {
                                                            console.log(err);
                                                        }
                                                    }
                                                },
                                                ready: function () {
                                                    webix.require("admin/dashboard.js?rand="+new Date().getTime(), function () {
                                                        toughradius.admin.dashboard.loadPage(session);
                                                    });
                                                }
                                            }
                                        ]
                                    }
                                ]
                            },
                            {
                                rows:[
                                    {
                                        height:40,
                                        css:"main-toolbar",
                                        cols:[
                                            { id:toughradius.admin.toolbarId+"_icon",view:"icon", icon:"home", width:45},
                                            { id:toughradius.admin.toolbarId+"_title", view: "label", label: ""},
                                            { },
                                            {
                                                view: "button", type: "icon", width: 100, icon: "book", label: "在线文档",  click: function () {
                                                    var bookurl = "https://docs.toughradius.net/zh/";
                                                    windowObjectReference = window.open(bookurl,"在线文档","resizable,scrollbars,status");
                                                }
                                            }
                                        ]
                                    },
                                    {height:5},
                                    {
                                        id: toughradius.admin.pageId,
                                        css:"main-page",
                                        // paddingY:3,
                                        rows:[
                                            {
                                                id: toughradius.admin.panelId,
                                                template: ""
                                            }
                                        ]
                                    }
                                ]
                            }

                        ]
                    }
                ]
            });

        });
    });

});


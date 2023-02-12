if (!window.OprPanelUI)
    window.OprPanelUI = {};

OprPanelUI.accountSidebar = function (view) {
    let sideid = "opr.accountSidebar";
    if ($$(sideid) && $$(sideid).isVisible()) {
        $$(sideid).close();
        return;
    }
    webix.ui({
        view: "sidemenu",
        id: sideid,
        width: 340,
        position: "right",
        state: function (state) {
            let toolbarHeight = $$("main-toolbar").$height;
            state.top = toolbarHeight;
            state.height -= toolbarHeight;
        },
        body: {
            rows: [
                {
                    css: "panel-toolbar",
                    paddingX: 10,
                    cols: [
                        {view: "icon", icon: "mdi mdi-account", css: "alter webix_transparent"},
                        {
                            view: "label",
                            label: " <i class='mdi mdi-info'></i> " + tr("glabal", "My Profile"),
                            css: "dash-title-b",
                            inputWidth: 150,
                            align: "left"
                        },
                        {},
                        {
                            view: "icon", icon: "mdi mdi-close", css: "webix_transparent", click: function () {
                                $$(sideid).close();
                            }
                        },
                    ]
                },
                {
                    view: "form",
                    paddingX: 20,
                    elementsConfig: {
                        marginY: 0,
                    },
                    url: "/admin/opr/current",
                    elements: [
                        {
                            cols: [
                                {
                                    view: "template", css: "opr-head", borderless: true,
                                    template: "<img src='/static/images/head.png' width='88' height='88'/>", height: 100
                                },
                                {
                                    rows: [{},
                                        {
                                            view: "button", css: "webix_transparent", type: "icon", icon: "mdi mdi-onepassword",
                                            label: tr("global", "Change Password"), click: function () {
                                                OprPanelUI.updatePassword(this.$view);
                                            },
                                        },
                                        {
                                            view: "button", css: "webix_transparent",  type: "icon", icon: "mdi mdi-logout-variant",
                                            label: gtr("Logout Session"), click: function () {
                                                window.location.href = "/logout";
                                            },
                                        },
                                        {},
                                    ],
                                },
                            ],
                        },
                        {view: "text", name: "realname", label: gtr("Name"), css: "nborder-input", readonly: true},
                        {view: "text", name: "email", label: gtr("Email"), css: "nborder-input", readonly: true},
                        {view: "text", name: "mobile", label: gtr("Mobile"), css: "nborder-input", readonly: true},
                        {view: "text", name: "last_login", label: gtr("Last login"), css: "nborder-input", readonly: true},
                    ],
                },

            ]
        }
    }).show();
};


OprPanelUI.updatePassword = function (hnode) {
    let pwinid = "opr.updatePassword ";
    if ($$(pwinid)) {
        return;
    }
    let formid = webix.uid().toString();
    webix.ui({
        id: pwinid,
        view: "window",
        width: 420,
        height: 420,
        css: "win-body",
        move: true,
        position: "center",
        head: {
            view: "toolbar",
            css: "win-toolbar",
            cols: [
                {view: "icon", icon: "mdi mdi-file", css: "alter"},
                {view: "label", label: tr("global", "Change Password")},
                {
                    view: "icon", icon: "mdi mdi-close", css: "alter", click: function () {
                        $$(pwinid).close();
                    }
                }
            ]
        },
        body: {
            rows: [
                {
                    id: formid,
                    view: "form",
                    scroll: false,
                    elementsConfig: {
                        labelWidth: 180,
                    },
                    elements: [
                        {view: "text", name: "oldpassword", type: "password", label: gtr("Old Password"), validate: webix.rules.isNotEmpty},
                        {view: "text", name: "password", type: "password", label: gtr("New Password"), validate: webix.rules.isNotEmpty},
                        {view: "text", name: "cpassword", type: "password", label: gtr("New password confirmation"), validate: webix.rules.isNotEmpty}
                    ]
                },
                {
                    padding: 5,
                    cols: [{},
                        {
                            view: "button", name: "submit", css: "webix_transparent", value: gtr("Save"), width: 90, height: 36,
                            click: function () {
                                if (!$$(formid).validate()) {
                                    webix.message({type: "error", text: tr("global", "Please fill out the form correctly."), expire: 1000});
                                    return false;
                                }
                                let btn = this;
                                webix.ajax().post('/admin/opr/uppassword', $$(formid).getValues()).then(function (result) {
                                    btn.enable();
                                    let resp = result.json();
                                    webix.message({type: resp.msgtype, text: resp.msg, expire: 3000});
                                    if (resp.code === 0) {
                                        $$(pwinid).close();
                                    }
                                });
                            }
                        },
                        {
                            view: "button", css: "webix_transparent", icon: "times-circle", width: 70, label: gtr("Cancel"), click: function () {
                                $$(pwinid).close();
                            }
                        }
                    ]
                }
            ]
        }
    }).show();
};
